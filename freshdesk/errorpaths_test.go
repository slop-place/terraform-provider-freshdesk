package freshdesk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestErrorPropagation checks that a failing response surfaces as an error from
// every shape of client method: envelope reads, filtered reads, multipart
// uploads and the plain CRUD wrappers.
func TestErrorPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	calls := []struct {
		name string
		call func(*Client) error
	}{
		{"GetAccount", func(c *Client) error { _, e := c.GetAccount(ctx); return e }},
		{"GetHelpdeskSettings", func(c *Client) error { _, e := c.GetHelpdeskSettings(ctx); return e }},
		{"ExportAccount", func(c *Client) error { _, e := c.ExportAccount(ctx, nil); return e }},
		{"Me", func(c *Client) error { _, e := c.Me(ctx); return e }},
		{"ListAgents", func(c *Client) error { _, e := c.ListAgents(ctx, AgentListOptions{}); return e }},
		{"GetAgentAvailability", func(c *Client) error { _, e := c.GetAgentAvailability(ctx, 1); return e }},
		{"UpdateAgentAvailability", func(c *Client) error {
			_, e := c.UpdateAgentAvailability(ctx, 1, nil)

			return e
		}},
		{"ListContacts", func(c *Client) error { _, e := c.ListContacts(ctx, ContactListOptions{}); return e }},
		{"CreateContact", func(c *Client) error { _, e := c.CreateContact(ctx, ContactRequest{}); return e }},
		{"MakeAgent", func(c *Client) error { _, e := c.MakeAgent(ctx, 1, MakeAgentRequest{}); return e }},
		{"FilterContacts", func(c *Client) error { _, e := c.FilterContacts(ctx, "x", 0); return e }},
		{"SearchCompanies", func(c *Client) error { _, e := c.SearchCompanies(ctx, "x"); return e }},
		{"FilterCompanies", func(c *Client) error { _, e := c.FilterCompanies(ctx, "x", 0); return e }},
		{"FilterTickets", func(c *Client) error { _, e := c.FilterTickets(ctx, "x", 0); return e }},
		{"CreateReply", func(c *Client) error { _, e := c.CreateReply(ctx, 1, ReplyRequest{}); return e }},
		{"ForwardTicket", func(c *Client) error { _, e := c.ForwardTicket(ctx, 1, ForwardRequest{}); return e }},
		{"ReplyToForward", func(c *Client) error { _, e := c.ReplyToForward(ctx, 1, ForwardRequest{}); return e }},
		{"ListSections", func(c *Client) error { _, e := c.ListSections(ctx, 1); return e }},
		{"CreateSection", func(c *Client) error { _, e := c.CreateSection(ctx, 1, SectionRequest{}); return e }},
		{"ListCustomObjectSchemas", func(c *Client) error { _, e := c.ListCustomObjectSchemas(ctx); return e }},
		{"ListCustomObjectRecords", func(c *Client) error {
			_, e := c.ListCustomObjectRecords(ctx, "s", nil)

			return e
		}},
		{"CountCustomObjectRecords", func(c *Client) error { _, e := c.CountCustomObjectRecords(ctx, "s"); return e }},
		{"GenerateMessageQuote", func(c *Client) error { _, e := c.GenerateMessageQuote(ctx, 1); return e }},
		{"SendOutboundMessage", func(c *Client) error { _, e := c.SendOutboundMessage(ctx, nil); return e }},
		{"GetNotificationBCC", func(c *Client) error { _, e := c.GetNotificationBCC(ctx); return e }},
		{"UpdateNotificationBCC", func(c *Client) error { _, e := c.UpdateNotificationBCC(ctx, nil); return e }},
		{"GetEmailSettings", func(c *Client) error { _, e := c.GetEmailSettings(ctx); return e }},
		{"UpdateEmailSettings", func(c *Client) error { _, e := c.UpdateEmailSettings(ctx, nil); return e }},
		{"GetSLAPolicy", func(c *Client) error { _, e := c.GetSLAPolicy(ctx, 1); return e }},
		{"GetTimeEntry", func(c *Client) error { _, e := c.GetTimeEntry(ctx, 1); return e }},
		{"ToggleTimer", func(c *Client) error { _, e := c.ToggleTimer(ctx, 1); return e }},
		{"CreateCannedResponses", func(c *Client) error { _, e := c.CreateCannedResponses(ctx, 1, nil); return e }},
		{"CreateAgents", func(c *Client) error { _, e := c.CreateAgents(ctx, nil); return e }},
		{"BulkUpdateTickets", func(c *Client) error { _, e := c.BulkUpdateTickets(ctx, BulkTicketUpdate{}); return e }},
		{"BulkDeleteTickets", func(c *Client) error { _, e := c.BulkDeleteTickets(ctx, nil); return e }},
		{"ExportContacts", func(c *Client) error { _, e := c.ExportContacts(ctx, nil, nil); return e }},
		{"ExportCompanies", func(c *Client) error { _, e := c.ExportCompanies(ctx, nil, nil); return e }},
		{"GetSolutionCategory", func(c *Client) error { _, e := c.GetSolutionCategory(ctx, 1, ""); return e }},
		{"GetSolutionFolder", func(c *Client) error { _, e := c.GetSolutionFolder(ctx, 1, ""); return e }},
		{"GetSolutionArticle", func(c *Client) error { _, e := c.GetSolutionArticle(ctx, 1, ""); return e }},
		{"UpdateSolutionCategory", func(c *Client) error {
			_, e := c.UpdateSolutionCategory(ctx, 1, "", SolutionCategoryRequest{})

			return e
		}},
		{"CreateSolutionCategoryTranslation", func(c *Client) error {
			_, e := c.CreateSolutionCategoryTranslation(ctx, 1, "es", SolutionCategoryRequest{})

			return e
		}},
		{"GetTicketFormField", func(c *Client) error { _, e := c.GetTicketFormField(ctx, 1, 2); return e }},
		{"UpdateTicketFormField", func(c *Client) error {
			_, e := c.UpdateTicketFormField(ctx, 1, 2, TicketFormFieldRequest{})

			return e
		}},
		{"CloneTicketForm", func(c *Client) error { _, e := c.CloneTicketForm(ctx, 1, "x"); return e }},
		{"ListAgentsInGroup", func(c *Client) error { _, e := c.ListAgentsInGroup(ctx, 1); return e }},
		{"SearchAgents", func(c *Client) error { _, e := c.SearchAgents(ctx, "x"); return e }},
		{"SearchContacts", func(c *Client) error { _, e := c.SearchContacts(ctx, "x"); return e }},
		{"CreateSurveyResponse", func(c *Client) error { _, e := c.CreateSurveyResponse(ctx, "s", nil); return e }},
		{"CreateThreadMessage", func(c *Client) error {
			_, e := c.CreateThreadMessage(ctx, ThreadMessageRequest{})

			return e
		}},
		{"CreateCannedResponse", func(c *Client) error {
			_, e := c.CreateCannedResponse(ctx, CannedResponseRequest{})

			return e
		}},
		{"CreateTicket", func(c *Client) error { _, e := c.CreateTicket(ctx, TicketRequest{}); return e }},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"access_denied","message":"nope"}`))
			})
			err := tc.call(c)
			if err == nil {
				t.Fatalf("%s: want error, got nil", tc.name)
			}
			if !Unauthorized(err) {
				t.Errorf("%s: err = %v, want a 403", tc.name, err)
			}
		})
	}
}

// TestMultipartPathsOnError covers the attachment branches of the create
// methods, which take a different code path from the JSON ones.
func TestMultipartPathsOnError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	ctx := context.Background()
	calls := []struct {
		name string
		call func(*Client) error
	}{
		{"CreateTicket", func(c *Client) error {
			_, e := c.CreateTicket(ctx, TicketRequest{Attachments: []string{file}})

			return e
		}},
		{"CreateContact", func(c *Client) error {
			_, e := c.CreateContact(ctx, ContactRequest{Avatar: file})

			return e
		}},
		{"CreateCannedResponse", func(c *Client) error {
			_, e := c.CreateCannedResponse(ctx, CannedResponseRequest{Attachments: []string{file}})

			return e
		}},
		{"CreateThreadMessage", func(c *Client) error {
			_, e := c.CreateThreadMessage(ctx, ThreadMessageRequest{Attachments: []string{file}})

			return e
		}},
		{"CreateReply", func(c *Client) error {
			_, e := c.CreateReply(ctx, 1, ReplyRequest{Attachments: []string{file}})

			return e
		}},
		{"ForwardTicket", func(c *Client) error {
			_, e := c.ForwardTicket(ctx, 1, ForwardRequest{Attachments: []string{file}})

			return e
		}},
		{"ReplyToForward", func(c *Client) error {
			_, e := c.ReplyToForward(ctx, 1, ForwardRequest{Attachments: []string{file}})

			return e
		}},
	}

	for _, tc := range calls {
		t.Run(tc.name+"/success", func(t *testing.T) {
			t.Parallel()

			var ct string
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				ct = r.Header.Get("Content-Type")
				_, _ = w.Write([]byte(`{"id":1}`))
			})
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.HasPrefix(ct, "multipart/form-data") {
				t.Errorf("%s: Content-Type = %q, want multipart", tc.name, ct)
			}
		})

		t.Run(tc.name+"/error", func(t *testing.T) {
			t.Parallel()

			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})
			if err := tc.call(c); err == nil {
				t.Fatalf("%s: want error", tc.name)
			}
		})
	}
}

func TestDecodeFailureIsNotRetried(t *testing.T) {
	t.Parallel()

	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{not json`))
	})
	_, err := c.GetGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("want a decode error")
	}
	if !strings.Contains(err.Error(), "decoding freshdesk response") {
		t.Errorf("err = %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (a malformed body must not be retried)", calls)
	}
}

func TestTransportFailureIsRetried(t *testing.T) {
	t.Parallel()

	c, err := New(Config{
		Domain: "acme", APIKey: "k", MaxRetries: 2,
		// A client pointed at a closed port fails at the transport layer.
		HTTPClient: &http.Client{Timeout: 50 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var slept int
	c.sleep = func(context.Context, time.Duration) error { slept++; return nil }

	if _, gerr := c.GetGroup(context.Background(), 1); gerr == nil {
		t.Fatal("want a transport error")
	}
	if slept != 2 {
		t.Errorf("slept %d times, want 2 retries", slept)
	}
}

func TestMultipartScalarEncodings(t *testing.T) {
	t.Parallel()

	got, err := multipartFields(struct {
		S    string           `json:"s"`
		B    bool             `json:"b"`
		N    int              `json:"n"`
		F    float64          `json:"f"`
		Arr  []int            `json:"arr"`
		Objs []map[string]int `json:"objs"`
		Nil  *string          `json:"nil,omitempty"`
	}{S: "x", B: true, N: 7, F: 1.5, Arr: []int{1, 2}, Objs: []map[string]int{{"a": 1}}})
	if err != nil {
		t.Fatalf("multipartFields: %v", err)
	}

	want := map[string]string{"s": "x", "b": "true", "n": "7", "f": "1.5"}
	for k, v := range want {
		if got[k][0] != v {
			t.Errorf("%s = %v, want %s", k, got[k], v)
		}
	}
	if len(got["arr[]"]) != 2 || got["arr[]"][1] != "2" {
		t.Errorf("arr[] = %v", got["arr[]"])
	}
	if !strings.Contains(got["objs[]"][0], `"a":1`) {
		t.Errorf("objs[] = %v, want nested JSON", got["objs[]"])
	}
	if _, ok := got["nil"]; ok {
		t.Errorf("nil field must be omitted, got %v", got["nil"])
	}
}

func TestStructToMapRejectsUnencodable(t *testing.T) {
	t.Parallel()

	if _, err := structToMap(struct {
		C chan int `json:"c"`
	}{}); err == nil {
		t.Fatal("want an encoding error for a channel field")
	}
}

func TestSleepCtxHonoursCancellation(t *testing.T) {
	t.Parallel()

	if err := sleepCtx(context.Background(), 0); err != nil {
		t.Errorf("zero delay: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := sleepCtx(ctx, time.Minute); err == nil {
		t.Error("want an error from a cancelled context")
	}
}

func TestTransportErrorUnwraps(t *testing.T) {
	t.Parallel()

	inner := os.ErrDeadlineExceeded
	err := &transportError{err: inner}

	if !strings.Contains(err.Error(), "freshdesk request failed") {
		t.Errorf("Error() = %q", err.Error())
	}
	if !errors.Is(err.Unwrap(), inner) {
		t.Error("Unwrap did not return the inner error")
	}
	if !retryable(err) {
		t.Error("a transport error must be retryable")
	}
	if retryable(json.Unmarshal([]byte(`x`), &struct{}{})) {
		t.Error("a decode error must not be retryable")
	}
}
