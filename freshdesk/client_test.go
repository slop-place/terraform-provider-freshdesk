package freshdesk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient wires a Client to an httptest server with backoff disabled so
// retry paths run instantly.
func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c, err := New(Config{Domain: "acme", APIKey: "key", HTTPClient: srv.Client()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Point the client at the test server while keeping the /api/v2/ prefix.
	base := c.baseURL
	u := *base
	srvURL := strings.TrimPrefix(srv.URL, "http://")
	u.Scheme = "http"
	u.Host = srvURL
	c.baseURL = &u
	c.sleep = func(context.Context, time.Duration) error { return nil }

	return c
}

func TestNormalizeDomain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "acme", want: "acme.freshdesk.com"},
		{in: "acme.freshdesk.com", want: "acme.freshdesk.com"},
		{in: "https://acme.freshdesk.com", want: "acme.freshdesk.com"},
		{in: "http://acme.freshdesk.com/", want: "acme.freshdesk.com"},
		{in: "ACME", want: "acme.freshdesk.com"},
		{in: "support.acme.com", want: "support.acme.com"},
		{in: "  acme  ", want: "acme.freshdesk.com"},
		{in: "", wantErr: true},
		{in: "  ", wantErr: true},
		{in: "acme/tickets", wantErr: true},
		{in: "https://", wantErr: true},
	}
	for _, tc := range cases {
		got, err := NormalizeDomain(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeDomain(%q) = %q, want error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeDomain(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeDomain(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNewRejectsEmptyAPIKey(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{Domain: "acme", APIKey: "  "}); err == nil {
		t.Fatal("New with blank api key: want error, got nil")
	}
}

func TestAuthHeaderIsBasicKeyColonX(t *testing.T) {
	t.Parallel()

	var got string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{}`))
	})
	if _, err := c.GetAccount(context.Background()); err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("key:X"))
	if got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestGetRequestShape(t *testing.T) {
	t.Parallel()

	var path, query, accept, ua string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		accept, ua = r.Header.Get("Accept"), r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"id":1,"name":"Entertainers"}`))
	})
	g, err := c.GetGroup(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if path != "/api/v2/groups/1" {
		t.Errorf("path = %q, want /api/v2/groups/1", path)
	}
	if query != "" {
		t.Errorf("query = %q, want empty", query)
	}
	if accept != "application/json" {
		t.Errorf("Accept = %q", accept)
	}
	if ua != DefaultUserAgent {
		t.Errorf("User-Agent = %q, want %q", ua, DefaultUserAgent)
	}
	if g.Name != "Entertainers" {
		t.Errorf("Name = %q", g.Name)
	}
}

func TestErrorDecoding(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "req-42")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"description":"Validation failed","errors":[
			{"field":"name","message":"It should be a String","code":"datatype_mismatch"}]}`))
	})
	_, err := c.GetGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	apiErr, ok := AsError(err)
	if !ok {
		t.Fatalf("AsError: not an *Error: %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Validation failed" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if apiErr.RequestID != "req-42" {
		t.Errorf("RequestID = %q", apiErr.RequestID)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Field != "name" {
		t.Errorf("Errors = %+v", apiErr.Errors)
	}
	if !strings.Contains(err.Error(), "name: It should be a String") {
		t.Errorf("Error() missing field detail: %s", err)
	}
}

// errPlain stands in for a non-API error in classification tests.
var errPlain = errors.New("plain")

func TestNotFoundAndUnauthorized(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		code             int
		notFound, unauth bool
	}{
		{http.StatusNotFound, true, false},
		{http.StatusUnauthorized, false, true},
		{http.StatusForbidden, false, true},
		{http.StatusBadRequest, false, false},
	} {
		c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.code)
			_, _ = w.Write([]byte(`{"message":"nope"}`))
		})
		_, err := c.GetGroup(context.Background(), 1)
		if NotFound(err) != tc.notFound {
			t.Errorf("status %d: NotFound = %v, want %v", tc.code, NotFound(err), tc.notFound)
		}
		if Unauthorized(err) != tc.unauth {
			t.Errorf("status %d: Unauthorized = %v, want %v", tc.code, Unauthorized(err), tc.unauth)
		}
	}
	if NotFound(nil) || Unauthorized(nil) {
		t.Error("nil error must not classify as NotFound or Unauthorized")
	}
	if NotFound(errPlain) {
		t.Error("plain error must not classify as NotFound")
	}
}

func TestRetriesOn429ThenSucceeds(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":"throttled","message":"slow down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":7,"name":"ok"}`))
	})
	g, err := c.GetGroup(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
	if g.ID != 7 {
		t.Errorf("ID = %d", g.ID)
	}
}

func TestRetriesExhaustedReturnsError(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	_, err := c.GetGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("want error after exhausting retries")
	}
	// One initial attempt plus DefaultMaxRetries retries.
	if got, want := calls.Load(), int32(DefaultMaxRetries+1); got != want {
		t.Errorf("calls = %d, want %d", got, want)
	}
}

func TestDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	})
	if _, err := c.GetGroup(context.Background(), 1); err == nil {
		t.Fatal("want error")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("calls = %d, want 1 (400 must not be retried)", got)
	}
}

func TestContextCancellationStopsRetries(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	ctx, cancel := context.WithCancel(context.Background())
	c.sleep = func(ctx context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	}
	if _, err := c.GetGroup(ctx, 1); err == nil {
		t.Fatal("want error from cancelled context")
	}
}

func TestBackoffHonoursRetryAfter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		attempt    int
		retryAfter string
		want       time.Duration
	}{
		{0, "", time.Second},
		{1, "", 2 * time.Second},
		{2, "", 4 * time.Second},
		{10, "", 30 * time.Second}, // capped
		{0, "7", 7 * time.Second},
		{0, "  3 ", 3 * time.Second},
		{0, "not-a-number", time.Second},
		{0, "100000", 5 * time.Minute}, // capped
	}
	for _, tc := range cases {
		if got := backoff(tc.attempt, tc.retryAfter); got != tc.want {
			t.Errorf("backoff(%d, %q) = %v, want %v", tc.attempt, tc.retryAfter, got, tc.want)
		}
	}
}

func TestPostSendsJSONBody(t *testing.T) {
	t.Parallel()

	var body map[string]any
	var ct string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":3,"name":"Support"}`))
	})
	g, err := c.CreateGroup(context.Background(), GroupRequest{
		Name:     Ptr("Support"),
		AgentIDs: []int64{2, 15},
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if body["name"] != "Support" {
		t.Errorf("name = %v", body["name"])
	}
	if _, ok := body["description"]; ok {
		t.Error("unset optional field must be omitted from the body")
	}
	if g.ID != 3 {
		t.Errorf("ID = %d", g.ID)
	}
}

func TestClearSemantics(t *testing.T) {
	t.Parallel()

	t.Run("group clears escalate_to and agents", func(t *testing.T) {
		t.Parallel()

		b, err := json.Marshal(GroupRequest{
			Name:            Ptr("g"),
			ClearEscalateTo: true,
			ClearAgents:     true,
		})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		v, ok := m["escalate_to"]
		if !ok || v != nil {
			t.Errorf("escalate_to = %v (present=%v), want explicit null", v, ok)
		}
		arr, ok := m["agent_ids"].([]any)
		if !ok || len(arr) != 0 {
			t.Errorf("agent_ids = %v, want empty array", m["agent_ids"])
		}
	})

	t.Run("contact clears tags and company", func(t *testing.T) {
		t.Parallel()

		b, _ := json.Marshal(ContactRequest{ClearTags: true, ClearCompany: true})
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		if v, ok := m["company_id"]; !ok || v != nil {
			t.Errorf("company_id = %v (present=%v), want null", v, ok)
		}
		if arr, ok := m["tags"].([]any); !ok || len(arr) != 0 {
			t.Errorf("tags = %v, want empty array", m["tags"])
		}
	})

	t.Run("unset fields stay omitted", func(t *testing.T) {
		t.Parallel()

		b, _ := json.Marshal(GroupRequest{Name: Ptr("g")})
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		for _, k := range []string{"escalate_to", "agent_ids", "description", "unassigned_for"} {
			if _, ok := m[k]; ok {
				t.Errorf("%s must be omitted when unset, got %v", k, m[k])
			}
		}
	})
}

func TestMarshalDoesNotEscapeHTML(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(AgentRequest{Signature: Ptr(`<div dir="ltr">Hi</div>`)})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), `<`) {
		t.Errorf("signature was HTML-escaped: %s", b)
	}
}
