package freshdesk

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// Ticket status values. Accounts may define additional custom statuses.
const (
	TicketStatusOpen     = 2
	TicketStatusPending  = 3
	TicketStatusResolved = 4
	TicketStatusClosed   = 5
)

// Ticket priority values.
const (
	TicketPriorityLow    = 1
	TicketPriorityMedium = 2
	TicketPriorityHigh   = 3
	TicketPriorityUrgent = 4
)

// Ticket source values.
const (
	TicketSourceEmail          = 1
	TicketSourcePortal         = 2
	TicketSourcePhone          = 3
	TicketSourceChat           = 7
	TicketSourceFeedbackWidget = 9
	TicketSourceOutboundEmail  = 10
	TicketSourceEcommerce      = 11
)

// Ticket association types.
const (
	TicketAssociationParent  = 1
	TicketAssociationChild   = 2
	TicketAssociationTracker = 3
	TicketAssociationRelated = 4
)

// Ticket is a Freshdesk support ticket.
type Ticket struct {
	ID              int64  `json:"id"`
	Subject         string `json:"subject"`
	Description     string `json:"description"`
	DescriptionText string `json:"description_text"`
	Type            string `json:"type"`
	Status          int    `json:"status"`
	Priority        int    `json:"priority"`
	Source          int    `json:"source"`

	RequesterID   int64 `json:"requester_id"`
	ResponderID   int64 `json:"responder_id"`
	GroupID       int64 `json:"group_id"`
	CompanyID     int64 `json:"company_id"`
	ProductID     int64 `json:"product_id"`
	EmailConfigID int64 `json:"email_config_id"`
	// FormID is the ticket form the ticket was raised through.
	FormID int64 `json:"form_id"`

	Name             string `json:"name"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	UniqueExternalID string `json:"unique_external_id"`
	FacebookID       string `json:"facebook_id"`
	TwitterID        string `json:"twitter_id"`

	ToEmails        []string `json:"to_emails"`
	CCEmails        []string `json:"cc_emails"`
	FwdEmails       []string `json:"fwd_emails"`
	ReplyCCEmails   []string `json:"reply_cc_emails"`
	TicketCCEmails  []string `json:"ticket_cc_emails"`
	TicketBccEmails []string `json:"ticket_bcc_emails"`
	// SupportEmail is the address the ticket arrived on.
	SupportEmail string `json:"support_email"`

	Tags         []string     `json:"tags"`
	CustomFields CustomFields `json:"custom_fields"`
	Attachments  []Attachment `json:"attachments"`

	DueBy       Time `json:"due_by"`
	FrDueBy     Time `json:"fr_due_by"`
	NrDueBy     Time `json:"nr_due_by"`
	IsEscalated bool `json:"is_escalated"`
	FrEscalated bool `json:"fr_escalated"`
	NrEscalated bool `json:"nr_escalated"`

	Spam    bool `json:"spam"`
	Deleted bool `json:"deleted"`

	AssociationType        int     `json:"association_type"`
	AssociatedTicketsCount int     `json:"associated_tickets_count"`
	AssociatedTicketIDs    []int64 `json:"associated_ticket_ids"`
	ParentID               int64   `json:"parent_id"`

	InternalAgentID int64 `json:"internal_agent_id"`
	InternalGroupID int64 `json:"internal_group_id"`

	SentimentScore        int `json:"sentiment_score"`
	InitialSentimentScore int `json:"initial_sentiment_score"`

	// SourceInfo carries channel-specific metadata about how the ticket arrived.
	SourceInfo map[string]any `json:"source_info"`
	// StructuredDescription holds rich content for webchat and mobile SDK tickets.
	StructuredDescription *StructuredBody `json:"structured_description"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`

	// Populated only when requested through TicketIncludes.
	Stats         *TicketStats     `json:"stats,omitempty"`
	Requester     *TicketRequester `json:"requester,omitempty"`
	Company       *TicketCompany   `json:"company,omitempty"`
	Conversations []Conversation   `json:"conversations,omitempty"`
}

// TicketStats holds lifecycle timestamps, returned with include=stats.
type TicketStats struct {
	AgentRespondedAt     Time `json:"agent_responded_at"`
	RequesterRespondedAt Time `json:"requester_responded_at"`
	FirstRespondedAt     Time `json:"first_responded_at"`
	StatusUpdatedAt      Time `json:"status_updated_at"`
	ReopenedAt           Time `json:"reopened_at"`
	ResolvedAt           Time `json:"resolved_at"`
	ClosedAt             Time `json:"closed_at"`
	PendingSince         Time `json:"pending_since"`
}

// TicketRequester is the abbreviated requester returned with include=requester.
type TicketRequester struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
	Phone  string `json:"phone"`
	Active bool   `json:"active"`
}

// TicketCompany is the abbreviated company returned with include=company.
type TicketCompany struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TicketRequest is the create/update payload for a ticket. Pointer fields are
// omitted when nil, so an update sends only what the caller set.
type TicketRequest struct {
	Subject     *string `json:"subject,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
	Source      *int    `json:"source,omitempty"`

	RequesterID   *int64 `json:"requester_id,omitempty"`
	ResponderID   *int64 `json:"responder_id,omitempty"`
	GroupID       *int64 `json:"group_id,omitempty"`
	CompanyID     *int64 `json:"company_id,omitempty"`
	ProductID     *int64 `json:"product_id,omitempty"`
	EmailConfigID *int64 `json:"email_config_id,omitempty"`

	Name             *string `json:"name,omitempty"`
	Email            *string `json:"email,omitempty"`
	Phone            *string `json:"phone,omitempty"`
	UniqueExternalID *string `json:"unique_external_id,omitempty"`
	FacebookID       *string `json:"facebook_id,omitempty"`
	TwitterID        *string `json:"twitter_id,omitempty"`

	CCEmails []string `json:"cc_emails,omitempty"`

	Tags         []string     `json:"tags,omitempty"`
	CustomFields CustomFields `json:"custom_fields,omitempty"`

	DueBy   *Time `json:"due_by,omitempty"`
	FrDueBy *Time `json:"fr_due_by,omitempty"`

	ParentID        *int64 `json:"parent_id,omitempty"`
	InternalAgentID *int64 `json:"internal_agent_id,omitempty"`
	InternalGroupID *int64 `json:"internal_group_id,omitempty"`

	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

// TicketIncludes selects the optional sideloaded bodies on ticket reads.
type TicketIncludes struct {
	Stats         bool
	Requester     bool
	Company       bool
	Description   bool
	Conversations bool
	Tags          bool
}

func (i TicketIncludes) values() []string {
	var out []string
	if i.Stats {
		out = append(out, "stats")
	}
	if i.Requester {
		out = append(out, "requester")
	}
	if i.Company {
		out = append(out, "company")
	}
	if i.Description {
		out = append(out, "description")
	}
	if i.Conversations {
		out = append(out, "conversations")
	}
	if i.Tags {
		out = append(out, "tags")
	}
	return out
}

const ticketsPath = "tickets"

// GetTicket fetches a ticket by ID.
func (c *Client) GetTicket(ctx context.Context, id int64, includes TicketIncludes) (*Ticket, error) {
	q := url.Values{}
	if v := includes.values(); len(v) > 0 {
		q.Set("include", strings.Join(v, ","))
	}
	return getResource[Ticket](ctx, c, ticketsPath, id, q)
}

// TicketListOptions filters the ticket collection endpoint.
type TicketListOptions struct {
	ListOptions

	// Filter is one of "new_and_my_open", "watching", "spam", "deleted".
	Filter string
	// RequesterID, Email, CompanyID and UniqueExternalID narrow by requester.
	RequesterID      int64
	Email            string
	CompanyID        int64
	UniqueExternalID string
	// Includes selects sideloaded bodies.
	Includes TicketIncludes
}

func (o TicketListOptions) values() url.Values {
	v := o.Values()
	if o.Filter != "" {
		v.Set("filter", o.Filter)
	}
	if o.RequesterID > 0 {
		v.Set("requester_id", strconv.FormatInt(o.RequesterID, 10))
	}
	if o.Email != "" {
		v.Set("email", o.Email)
	}
	if o.CompanyID > 0 {
		v.Set("company_id", strconv.FormatInt(o.CompanyID, 10))
	}
	if o.UniqueExternalID != "" {
		v.Set("unique_external_id", o.UniqueExternalID)
	}
	if iv := o.Includes.values(); len(iv) > 0 {
		v.Set("include", strings.Join(iv, ","))
	}
	return v
}

// ListTickets returns every ticket matching opts, following pagination.
func (c *Client) ListTickets(ctx context.Context, opts TicketListOptions) ([]Ticket, error) {
	extra := opts.values()
	// Fold the ticket-specific parameters into ListOptions.Extra so the shared
	// pagination walker carries them on every page.
	base := opts.ListOptions
	base.Extra = extra
	return listAll[Ticket](ctx, c, ticketsPath, base)
}

// CreateTicket creates a ticket.
func (c *Client) CreateTicket(ctx context.Context, req TicketRequest) (*Ticket, error) {
	if len(req.Attachments) > 0 {
		var out Ticket
		fields, err := multipartFields(req)
		if err != nil {
			return nil, err
		}
		if err := c.PostMultipart(ctx, ticketsPath, fields,
			map[string][]string{AttachmentsField: req.Attachments}, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	return createResource[Ticket](ctx, c, ticketsPath, req)
}

// UpdateTicket applies a partial update to a ticket.
func (c *Client) UpdateTicket(ctx context.Context, id int64, req TicketRequest) (*Ticket, error) {
	return updateResource[Ticket](ctx, c, ticketsPath, id, req)
}

// DeleteTicket moves a ticket to the trash; it can be restored for 30 days.
func (c *Client) DeleteTicket(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, ticketsPath, id)
}

// RestoreTicket undoes DeleteTicket.
func (c *Client) RestoreTicket(ctx context.Context, id int64) error {
	return c.Put(ctx, pathFor(ticketsPath, id)+"/restore", nil, nil)
}

// WatchTicket subscribes the calling agent to a ticket.
func (c *Client) WatchTicket(ctx context.Context, id int64) error {
	return c.Put(ctx, pathFor(ticketsPath, id)+"/watch", nil, nil)
}

// UnwatchTicket unsubscribes the calling agent from a ticket.
func (c *Client) UnwatchTicket(ctx context.Context, id int64) error {
	return c.Put(ctx, pathFor(ticketsPath, id)+"/unwatch", nil, nil)
}

// TicketMergeRequest describes a merge of secondary tickets into a primary one.
type TicketMergeRequest struct {
	PrimaryID int64   `json:"primary_id"`
	TicketIDs []int64 `json:"ticket_ids"`
	// ConvertRecipientsToCC moves the secondary tickets' requesters onto the
	// primary ticket as CCs. The wire field carries Freshdesk's own spelling.
	//nolint:misspell // the wire field carries Freshdesk's own spelling
	ConvertRecipientsToCC bool             `json:"convert_recepients_to_cc,omitempty"`
	NoteInPrimary         *TicketMergeNote `json:"note_in_primary,omitempty"`
	NoteInSecondary       *TicketMergeNote `json:"note_in_secondary,omitempty"`
}

// TicketMergeNote is the optional note added to tickets during a merge.
type TicketMergeNote struct {
	Body    string `json:"body"`
	Private bool   `json:"private"`
}

// MergeTickets merges secondary tickets into a primary ticket.
func (c *Client) MergeTickets(ctx context.Context, req TicketMergeRequest) error {
	return c.Put(ctx, ticketsPath+"/merge", req, nil)
}

// BulkTicketUpdate is the payload for BulkUpdateTickets.
type BulkTicketUpdate struct {
	IDs        []int64        `json:"ids"`
	Properties map[string]any `json:"properties"`
}

// BulkJob identifies an asynchronous bulk operation.
type BulkJob struct {
	JobID  string `json:"job_id"`
	HREF   string `json:"href"`
	Status string `json:"status"`
}

// BulkUpdateTickets applies the same property changes to many tickets.
func (c *Client) BulkUpdateTickets(ctx context.Context, req BulkTicketUpdate) (*BulkJob, error) {
	var out BulkJob
	body := map[string]any{"bulk_action": req}
	if err := c.Put(ctx, ticketsPath+"/bulk_update", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkDeleteTickets deletes many tickets in one call.
func (c *Client) BulkDeleteTickets(ctx context.Context, ids []int64) (*BulkJob, error) {
	var out BulkJob
	body := map[string]any{"bulk_action": map[string]any{"ids": ids}}
	if err := c.Post(ctx, ticketsPath+"/bulk_delete", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetArchivedTicket fetches a ticket that has been archived.
func (c *Client) GetArchivedTicket(ctx context.Context, id int64) (*Ticket, error) {
	return getResource[Ticket](ctx, c, "tickets/archived", id, nil)
}

// DeleteArchivedTicket permanently removes an archived ticket.
func (c *Client) DeleteArchivedTicket(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "tickets/archived", id)
}

// FilterTickets runs a Freshdesk query-language filter over tickets.
//
// The query uses the API's filter syntax, e.g.
//
//	"priority:3 AND status:2"
func (c *Client) FilterTickets(ctx context.Context, query string, page int) (*TicketSearchResult, error) {
	q := url.Values{}
	q.Set("query", quoteQuery(query))
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var out TicketSearchResult
	if err := c.Get(ctx, "search/tickets", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TicketSearchResult is the envelope returned by the ticket search endpoint.
type TicketSearchResult struct {
	Total   int      `json:"total"`
	Results []Ticket `json:"results"`
}

// quoteQuery wraps a filter expression in the double quotes the search
// endpoints require, unless the caller already did.
func quoteQuery(q string) string {
	q = strings.TrimSpace(q)
	if strings.HasPrefix(q, `"`) && strings.HasSuffix(q, `"`) && len(q) >= 2 {
		return q
	}
	return `"` + q + `"`
}
