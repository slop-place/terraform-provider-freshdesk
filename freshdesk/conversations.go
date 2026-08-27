package freshdesk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Conversation is a note, reply, or forward on a ticket.
type Conversation struct {
	ID       int64  `json:"id"`
	TicketID int64  `json:"ticket_id"`
	UserID   int64  `json:"user_id"`
	Body     string `json:"body"`
	BodyText string `json:"body_text"`
	// StructuredBody carries rich content for webchat and mobile SDK tickets.
	StructuredBody *StructuredBody `json:"structured_body,omitempty"`
	// Incoming marks a message that originated outside the web portal.
	Incoming bool `json:"incoming"`
	// Private is true for internal notes.
	Private bool `json:"private"`
	// Source distinguishes reply (0) from note (2) and the other channels.
	Source       int          `json:"source"`
	SupportEmail string       `json:"support_email"`
	FromEmail    string       `json:"from_email"`
	ToEmails     []string     `json:"to_emails"`
	CCEmails     []string     `json:"cc_emails"`
	BccEmails    []string     `json:"bcc_emails"`
	NotifiedTo   []string     `json:"notified_to"`
	Attachments  []Attachment `json:"attachments"`
	CreatedAt    Time         `json:"created_at"`
	UpdatedAt    Time         `json:"updated_at"`
}

// StructuredBody is the rich-content form of a conversation body.
type StructuredBody struct {
	BodyContents []StructuredBodyContent `json:"body_contents"`
}

// StructuredBodyContent is one block within a StructuredBody.
type StructuredBodyContent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// NoteRequest creates a note on a ticket. Either Body or StructuredBody must
// be set.
type NoteRequest struct {
	Body           *string         `json:"body,omitempty"`
	StructuredBody *StructuredBody `json:"structured_body,omitempty"`
	// Private defaults to true at the API when omitted.
	Private      *bool    `json:"private,omitempty"`
	Incoming     *bool    `json:"incoming,omitempty"`
	NotifyEmails []string `json:"notify_emails,omitempty"`
	UserID       *int64   `json:"user_id,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

// ReplyRequest creates a public reply on a ticket.
type ReplyRequest struct {
	Body           *string         `json:"body,omitempty"`
	StructuredBody *StructuredBody `json:"structured_body,omitempty"`
	FromEmail      *string         `json:"from_email,omitempty"`
	UserID         *int64          `json:"user_id,omitempty"`
	CCEmails       []string        `json:"cc_emails,omitempty"`
	BccEmails      []string        `json:"bcc_emails,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

// ForwardRequest forwards a ticket to other recipients.
type ForwardRequest struct {
	Body      *string  `json:"body,omitempty"`
	ToEmails  []string `json:"to_emails,omitempty"`
	CCEmails  []string `json:"cc_emails,omitempty"`
	BccEmails []string `json:"bcc_emails,omitempty"`
	// AgentID identifies the agent the forward is sent as.
	AgentID *int64 `json:"agent_id,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

// ConversationUpdate edits the body of an existing note.
type ConversationUpdate struct {
	Body           *string         `json:"body,omitempty"`
	StructuredBody *StructuredBody `json:"structured_body,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

// ListTicketConversations returns every conversation on a ticket.
func (c *Client) ListTicketConversations(
	ctx context.Context,
	ticketID int64,
	opts ListOptions,
) ([]Conversation, error) {
	return listAll[Conversation](ctx, c, pathFor(ticketsPath, ticketID)+"/conversations", opts)
}

// ListArchivedTicketConversations returns the conversations of an archived ticket.
func (c *Client) ListArchivedTicketConversations(
	ctx context.Context,
	ticketID int64,
	opts ListOptions,
) ([]Conversation, error) {
	return listAll[Conversation](ctx, c, "tickets/archived/"+idString(ticketID)+"/conversations", opts)
}

// CreateNote adds a note to a ticket.
func (c *Client) CreateNote(ctx context.Context, ticketID int64, req NoteRequest) (*Conversation, error) {
	path := pathFor(ticketsPath, ticketID) + "/notes"
	if len(req.Attachments) > 0 {
		return postConversationMultipart(ctx, c, path, req, req.Attachments)
	}
	return createResource[Conversation](ctx, c, path, req)
}

// CreateReply adds a public reply to a ticket.
func (c *Client) CreateReply(ctx context.Context, ticketID int64, req ReplyRequest) (*Conversation, error) {
	path := pathFor(ticketsPath, ticketID) + "/reply"
	if len(req.Attachments) > 0 {
		return postConversationMultipart(ctx, c, path, req, req.Attachments)
	}
	return createResource[Conversation](ctx, c, path, req)
}

// ForwardTicket forwards a ticket to additional recipients.
func (c *Client) ForwardTicket(ctx context.Context, ticketID int64, req ForwardRequest) (*Conversation, error) {
	path := pathFor(ticketsPath, ticketID) + "/forward"
	if len(req.Attachments) > 0 {
		return postConversationMultipart(ctx, c, path, req, req.Attachments)
	}
	return createResource[Conversation](ctx, c, path, req)
}

// ReplyToForward replies to a forwarded ticket.
func (c *Client) ReplyToForward(ctx context.Context, ticketID int64, req ForwardRequest) (*Conversation, error) {
	path := pathFor(ticketsPath, ticketID) + "/reply_to_forward"
	if len(req.Attachments) > 0 {
		return postConversationMultipart(ctx, c, path, req, req.Attachments)
	}
	return createResource[Conversation](ctx, c, path, req)
}

// UpdateConversation edits an existing note. Only notes can be edited.
func (c *Client) UpdateConversation(ctx context.Context, id int64, req ConversationUpdate) (*Conversation, error) {
	return updateResource[Conversation](ctx, c, "conversations", id, req)
}

// DeleteConversation deletes a conversation.
func (c *Client) DeleteConversation(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "conversations", id)
}

// DeleteAttachment deletes an attachment by ID.
func (c *Client) DeleteAttachment(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "attachments", id)
}

func postConversationMultipart[T any](
	ctx context.Context,
	c *Client,
	path string,
	req T,
	files []string,
) (*Conversation, error) {
	fields, err := multipartFields(req)
	if err != nil {
		return nil, err
	}
	var out Conversation
	if err := c.PostMultipart(ctx, path, fields, map[string][]string{AttachmentsField: files}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// multipartFields flattens a request struct into multipart form fields.
//
// Freshdesk expects scalars as plain values, string slices as repeated
// "name[]" fields, and nested objects as JSON-encoded strings.
func multipartFields(v any) (map[string][]string, error) {
	m, err := structToMap(v)
	if err != nil {
		return nil, err
	}
	fields := map[string][]string{}
	for key, val := range m {
		if val == nil {
			continue
		}
		switch tv := val.(type) {
		case string:
			fields[key] = []string{tv}
		case bool:
			fields[key] = []string{strconv.FormatBool(tv)}
		case json.Number:
			fields[key] = []string{tv.String()}
		case []any:
			name := key + "[]"
			for _, item := range tv {
				s, err := multipartScalar(item)
				if err != nil {
					return nil, fmt.Errorf("encoding multipart field %q: %w", key, err)
				}
				fields[name] = append(fields[name], s)
			}
		case map[string]any:
			// Nested objects (custom_fields, structured_body) travel as JSON.
			b, err := marshalMap(tv)
			if err != nil {
				return nil, fmt.Errorf("encoding multipart field %q: %w", key, err)
			}
			fields[key] = []string{string(b)}
		default:
			s, err := multipartScalar(val)
			if err != nil {
				return nil, fmt.Errorf("encoding multipart field %q: %w", key, err)
			}
			fields[key] = []string{s}
		}
	}
	return fields, nil
}

func multipartScalar(v any) (string, error) {
	switch tv := v.(type) {
	case string:
		return tv, nil
	case bool:
		return strconv.FormatBool(tv), nil
	case json.Number:
		return tv.String(), nil
	case nil:
		return "", nil
	default:
		// Structs and maps inside arrays are sent as compact JSON.
		if rv := reflect.ValueOf(v); rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice {
			b, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("encoding multipart value: %w", err)
			}

			return string(b), nil
		}
		return strings.TrimSpace(fmt.Sprint(v)), nil
	}
}
