package freshdesk

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// MaxPerPage is the largest page size Freshdesk accepts on list endpoints.
const MaxPerPage = 100

// pageCaps records the endpoints whose per_page limit differs from MaxPerPage.
// Freshdesk does not document these; they were found by probing a live
// account, and sending a larger page size is answered with a 400.
//
// A cap of zero means the endpoint rejects per_page altogether.
//
//nolint:gochecknoglobals // an immutable lookup table
var pageCaps = map[string]int{
	"ticket-forms":        ticketFormsPageCap,
	"admin/ticket_fields": noPaging,
}

const (
	// ticketFormsPageCap is the largest page the ticket-forms endpoint accepts.
	ticketFormsPageCap = 50
	// noPaging marks an endpoint that rejects the pagination parameters
	// altogether and answers with the entire collection.
	noPaging = 0
)

// pageCapFor returns the largest page size an endpoint accepts, and whether
// per_page may be sent at all.
func pageCapFor(path string) (int, bool) {
	limit, ok := pageCaps[strings.Trim(path, "/")]
	if !ok {
		return MaxPerPage, true
	}

	// A zero cap marks an endpoint that rejects per_page outright.
	if limit == 0 {
		return 0, false
	}

	return limit, true
}

// maxListPages bounds pagination. Freshdesk hard-caps deep pagination on
// several endpoints, so we stop before hammering it with requests that can
// only fail.
const maxListPages = 1000

// linkNextMatchLen is the length of a successful rel="next" submatch: the
// whole match plus the captured URL.
const linkNextMatchLen = 2

// ListOptions are the pagination and filter parameters common to list endpoints.
type ListOptions struct {
	// Page is the 1-based page number. Zero means "start at the first page".
	Page int
	// PerPage is the page size, capped at MaxPerPage. Zero means MaxPerPage.
	PerPage int
	// UpdatedSince restricts results to records changed at or after an
	// ISO-8601 timestamp.
	UpdatedSince string
	// OrderBy names the sort field, where the endpoint supports one.
	OrderBy string
	// OrderType is "asc" or "desc".
	OrderType string
	// Extra carries endpoint-specific query parameters.
	Extra url.Values
}

// Values renders the options as a query string.
func (o ListOptions) Values() url.Values {
	return o.valuesFor("")
}

// valuesFor renders the options for a particular endpoint, honouring that
// endpoint's per_page limit.
func (o ListOptions) valuesFor(path string) url.Values {
	v := url.Values{}
	for k, vals := range o.Extra {
		for _, s := range vals {
			v.Add(k, s)
		}
	}
	// An endpoint that rejects per_page rejects page with it, and answers with
	// the whole collection in one response.
	limit, allowed := pageCapFor(path)
	if allowed {
		if o.Page > 0 {
			v.Set("page", strconv.Itoa(o.Page))
		}

		perPage := o.PerPage
		if perPage <= 0 || perPage > limit {
			perPage = limit
		}

		v.Set("per_page", strconv.Itoa(perPage))
	}
	if o.UpdatedSince != "" {
		v.Set("updated_since", o.UpdatedSince)
	}
	if o.OrderBy != "" {
		v.Set("order_by", o.OrderBy)
	}
	if o.OrderType != "" {
		v.Set("order_type", o.OrderType)
	}
	return v
}

// linkNextRe extracts the rel="next" URL from an RFC 5988 Link header.
var linkNextRe = regexp.MustCompile(`<([^>]+)>\s*;\s*rel\s*=\s*"?next"?`)

// nextPageURL returns the rel="next" link from a Link header, or "".
func nextPageURL(link string) string {
	if m := linkNextRe.FindStringSubmatch(link); len(m) == linkNextMatchLen {
		return m[1]
	}
	return ""
}

// listPage fetches one page and returns the raw next-page link.
func listPage[T any](ctx context.Context, c *Client, path string, q url.Values) ([]T, string, error) {
	var out []T
	hdr, err := c.Do(ctx, Request{Method: "GET", Path: path, Query: q}, &out)
	if err != nil {
		return nil, "", err
	}
	var next string
	if hdr != nil {
		next = nextPageURL(hdr.Get("Link"))
	}
	return out, next, nil
}

// listAll walks every page of a collection endpoint and returns the union.
//
// Freshdesk signals "more pages" with a Link header on most endpoints, but not
// all of them; where it is missing we fall back to walking page numbers until a
// short (or empty) page comes back.
func listAll[T any](ctx context.Context, c *Client, path string, opts ListOptions) ([]T, error) {
	limit, pagingAllowed := pageCapFor(path)

	perPage := opts.PerPage
	if perPage <= 0 || perPage > limit {
		perPage = limit
	}
	opts.PerPage = perPage

	// An explicit page means the caller wants exactly that page.
	explicitPage := opts.Page > 0
	page := opts.Page
	if page <= 0 {
		page = 1
	}

	var all []T
	for {
		o := opts
		o.Page = page
		items, next, err := listPage[T](ctx, c, path, o.valuesFor(path))
		if err != nil {
			return nil, err
		}
		all = append(all, items...)

		if explicitPage || !pagingAllowed {
			// An endpoint that rejects per_page returns everything at once.
			return all, nil
		}
		// Prefer the server's own signal; fall back to page-size heuristics.
		if next == "" && len(items) < perPage {
			return all, nil
		}
		if len(items) == 0 {
			return all, nil
		}
		page++
		if page > maxListPages {
			return all, fmt.Errorf("%w: %d pages for %s", ErrPaginationLimit, maxListPages, path)
		}
	}
}
