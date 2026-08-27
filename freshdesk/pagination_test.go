package freshdesk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func TestListOptionsValues(t *testing.T) {
	t.Parallel()

	got := ListOptions{
		Page:         2,
		PerPage:      30,
		UpdatedSince: "2024-01-01T00:00:00Z",
		OrderBy:      "created_at",
		OrderType:    "desc",
		Extra:        url.Values{"filter": {"spam"}},
	}.Values()

	want := map[string]string{
		"page": "2", "per_page": "30", "updated_since": "2024-01-01T00:00:00Z",
		"order_by": "created_at", "order_type": "desc", "filter": "spam",
	}
	for k, v := range want {
		if got.Get(k) != v {
			t.Errorf("%s = %q, want %q", k, got.Get(k), v)
		}
	}
}

func TestListOptionsClampsPerPage(t *testing.T) {
	t.Parallel()

	for _, in := range []int{0, -5, 101, 1000} {
		got := ListOptions{PerPage: in}.Values().Get("per_page")
		if got != strconv.Itoa(MaxPerPage) {
			t.Errorf("PerPage %d -> per_page = %q, want %d", in, got, MaxPerPage)
		}
	}
	if got := (ListOptions{PerPage: 25}).Values().Get("per_page"); got != "25" {
		t.Errorf("PerPage 25 -> %q", got)
	}
}

func TestListOptionsOmitsUnsetPage(t *testing.T) {
	t.Parallel()

	if v := (ListOptions{}).Values(); v.Has("page") {
		t.Errorf("page must be omitted when unset, got %q", v.Get("page"))
	}
}

func TestNextPageURL(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{`<https://acme.freshdesk.com/api/v2/tickets?page=2>; rel="next"`,
			"https://acme.freshdesk.com/api/v2/tickets?page=2"},
		{`<https://x/a?page=3>;rel=next`, "https://x/a?page=3"},
		{`<https://x/a?page=1>; rel="prev", <https://x/a?page=3>; rel="next"`,
			"https://x/a?page=3"},
		{`<https://x/a?page=1>; rel="prev"`, ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := nextPageURL(tc.in); got != tc.want {
			t.Errorf("nextPageURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestListAllFollowsLinkHeader walks three pages advertised via Link headers.
func TestListAllFollowsLinkHeader(t *testing.T) {
	t.Parallel()

	var pages []string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		switch page {
		case "1":
			w.Header().Set("Link", `<https://x/api/v2/groups?page=2>; rel="next"`)
			_, _ = fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		case "2":
			w.Header().Set("Link", `<https://x/api/v2/groups?page=3>; rel="next"`)
			_, _ = fmt.Fprint(w, `[{"id":3}]`)
		default:
			_, _ = fmt.Fprint(w, `[{"id":4}]`)
		}
	})

	got, err := c.ListGroups(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d groups, want 4: %+v", len(got), got)
	}
	if len(pages) != 3 {
		t.Errorf("requested pages %v, want 3 requests", pages)
	}
}

// TestListAllStopsOnShortPage covers endpoints that omit the Link header.
func TestListAllStopsOnShortPage(t *testing.T) {
	t.Parallel()

	var requests int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests++
		// A page shorter than per_page means the collection is exhausted.
		_, _ = fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})
	got, err := c.ListGroups(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(got) != 2 || requests != 1 {
		t.Errorf("got %d groups over %d requests, want 2 over 1", len(got), requests)
	}
}

// TestListAllHonoursExplicitPage fetches exactly the requested page.
func TestListAllHonoursExplicitPage(t *testing.T) {
	t.Parallel()

	var requests int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("page = %q, want 2", got)
		}
		// Advertise a next page; an explicit request must not follow it.
		w.Header().Set("Link", `<https://x/api/v2/groups?page=3>; rel="next"`)
		_, _ = fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})
	got, err := c.ListGroups(context.Background(), ListOptions{Page: 2})
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(got) != 2 || requests != 1 {
		t.Errorf("got %d groups over %d requests, want 2 over 1", len(got), requests)
	}
}

// TestListAllStopsOnEmptyPage guards against looping on a full-then-empty page.
func TestListAllStopsOnEmptyPage(t *testing.T) {
	t.Parallel()

	var requests int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Query().Get("page") == "1" {
			// A full page with a next link, then nothing.
			w.Header().Set("Link", `<https://x/api/v2/groups?page=2>; rel="next"`)
			items := make([]string, MaxPerPage)
			for i := range items {
				items[i] = fmt.Sprintf(`{"id":%d}`, i+1)
			}
			_, _ = fmt.Fprint(w, "[")
			for i, it := range items {
				if i > 0 {
					_, _ = fmt.Fprint(w, ",")
				}
				_, _ = fmt.Fprint(w, it)
			}
			_, _ = fmt.Fprint(w, "]")
			return
		}
		_, _ = fmt.Fprint(w, `[]`)
	})
	got, err := c.ListGroups(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(got) != MaxPerPage {
		t.Errorf("got %d groups, want %d", len(got), MaxPerPage)
	}
	if requests != 2 {
		t.Errorf("requests = %d, want 2", requests)
	}
}

func TestListAllPropagatesErrors(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"message":"bad filter"}`)
	})
	if _, err := c.ListGroups(context.Background(), ListOptions{}); err == nil {
		t.Fatal("want error")
	}
}
