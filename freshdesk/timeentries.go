package freshdesk

import (
	"context"
	"net/http"
	"strconv"
)

// TimeEntry records effort logged against a ticket.
type TimeEntry struct {
	ID       int64  `json:"id"`
	TicketID int64  `json:"ticket_id"`
	AgentID  int64  `json:"agent_id"`
	Note     string `json:"note"`
	// TimeSpent is "HH:MM".
	TimeSpent string `json:"time_spent"`
	// TimeSpentInSeconds is the same duration as a number.
	TimeSpentInSeconds int `json:"time_spent_in_seconds"`
	// CompanyID is the company of the ticket's requester.
	CompanyID int64 `json:"company_id"`
	// Billable marks the entry as chargeable.
	Billable bool `json:"billable"`
	// TimerRunning is true while the entry's timer is counting.
	TimerRunning bool `json:"timer_running"`
	// StartTime is when the timer was started.
	StartTime  Time `json:"start_time"`
	ExecutedAt Time `json:"executed_at"`
	CreatedAt  Time `json:"created_at"`
	UpdatedAt  Time `json:"updated_at"`
}

// TimeEntryRequest is the create/update payload for a time entry.
type TimeEntryRequest struct {
	Note      *string `json:"note,omitempty"`
	TimeSpent *string `json:"time_spent,omitempty"`
	Billable  *bool   `json:"billable,omitempty"`
	AgentID   *int64  `json:"agent_id,omitempty"`
	// TimerRunning starts the entry with a running timer.
	TimerRunning *bool `json:"timer_running,omitempty"`
	// ExecutedAt backdates the entry.
	ExecutedAt *string `json:"executed_at,omitempty"`
	StartTime  *string `json:"start_time,omitempty"`
}

const timeEntriesPath = "time_entries"

// GetTimeEntry fetches a time entry. Freshdesk has no single-entry read, so
// this filters the collection.
func (c *Client) GetTimeEntry(ctx context.Context, id int64) (*TimeEntry, error) {
	entries, err := c.ListTimeEntries(ctx, TimeEntryListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range entries {
		if entries[i].ID == id {
			return &entries[i], nil
		}
	}
	return nil, &Error{StatusCode: http.StatusNotFound, Code: "not_found",
		Message: "time entry " + idString(id) + " not found"}
}

// TimeEntryListOptions filters the time-entry collection endpoint.
type TimeEntryListOptions struct {
	ListOptions

	CompanyID      int64
	AgentID        int64
	Billable       *bool
	ExecutedAfter  string
	ExecutedBefore string
}

// ListTimeEntries returns every time entry matching opts.
func (c *Client) ListTimeEntries(ctx context.Context, opts TimeEntryListOptions) ([]TimeEntry, error) {
	base := opts.ListOptions
	v := base.Values()
	if opts.CompanyID > 0 {
		v.Set("company_id", strconv.FormatInt(opts.CompanyID, 10))
	}
	if opts.AgentID > 0 {
		v.Set("agent_id", strconv.FormatInt(opts.AgentID, 10))
	}
	if opts.Billable != nil {
		v.Set("billable", strconv.FormatBool(*opts.Billable))
	}
	if opts.ExecutedAfter != "" {
		v.Set("executed_after", opts.ExecutedAfter)
	}
	if opts.ExecutedBefore != "" {
		v.Set("executed_before", opts.ExecutedBefore)
	}
	base.Extra = v
	return listAll[TimeEntry](ctx, c, timeEntriesPath, base)
}

// ListTicketTimeEntries returns the time entries on a ticket.
func (c *Client) ListTicketTimeEntries(ctx context.Context, ticketID int64, opts ListOptions) ([]TimeEntry, error) {
	return listAll[TimeEntry](ctx, c, pathFor(ticketsPath, ticketID)+"/time_entries", opts)
}

// CreateTimeEntry logs time against a ticket.
func (c *Client) CreateTimeEntry(ctx context.Context, ticketID int64, req TimeEntryRequest) (*TimeEntry, error) {
	return createResource[TimeEntry](ctx, c, pathFor(ticketsPath, ticketID)+"/time_entries", req)
}

// UpdateTimeEntry applies a partial update to a time entry.
func (c *Client) UpdateTimeEntry(ctx context.Context, id int64, req TimeEntryRequest) (*TimeEntry, error) {
	return updateResource[TimeEntry](ctx, c, timeEntriesPath, id, req)
}

// ToggleTimer starts a stopped timer or stops a running one.
func (c *Client) ToggleTimer(ctx context.Context, id int64) (*TimeEntry, error) {
	var out TimeEntry
	if err := c.Put(ctx, pathFor(timeEntriesPath, id)+"/toggle_timer", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTimeEntry deletes a time entry.
func (c *Client) DeleteTimeEntry(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, timeEntriesPath, id)
}
