package freshdesk

import "context"

// Thread types.
const (
	// ThreadTypeForward is a thread that forwards a ticket outside the account.
	ThreadTypeForward = "forward"
	// ThreadTypeDiscussion is an internal discussion thread.
	ThreadTypeDiscussion = "discussion"
	// ThreadTypePrivate is a private note thread.
	ThreadTypePrivate = "private"
)

// Thread is a collaboration thread attached to a ticket.
type Thread struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	// ParentID is the ticket the thread hangs off.
	ParentID int64 `json:"parent_id"`
	// ParentType is the kind of object ParentID refers to, normally "ticket".
	ParentType string `json:"parent_type"`
	// Participants lists the agents and contacts on the thread.
	Participants []ThreadParticipant `json:"participants"`
	// AnchorID and AnchorType point at the conversation the thread branched from.
	AnchorID   int64  `json:"anchor_id,omitempty"`
	AnchorType string `json:"anchor_type,omitempty"`
	CreatedAt  Time   `json:"created_at"`
	UpdatedAt  Time   `json:"updated_at"`
}

// ThreadParticipant is one member of a collaboration thread.
type ThreadParticipant struct {
	ID    int64  `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	// Type is "agent" or "contact".
	Type string `json:"type,omitempty"`
}

// ThreadRequest is the create/update payload for a thread.
type ThreadRequest struct {
	Type         *string             `json:"type,omitempty"`
	Title        *string             `json:"title,omitempty"`
	ParentID     *int64              `json:"parent_id,omitempty"`
	ParentType   *string             `json:"parent_type,omitempty"`
	AnchorID     *int64              `json:"anchor_id,omitempty"`
	AnchorType   *string             `json:"anchor_type,omitempty"`
	Participants []ThreadParticipant `json:"participants,omitempty"`
	// AdditionalInfo carries thread-type-specific settings such as the
	// email configuration used by a forward thread.
	AdditionalInfo map[string]any `json:"additional_info,omitempty"`
}

// ThreadMessage is a message posted into a collaboration thread.
type ThreadMessage struct {
	ID       int64  `json:"id"`
	ThreadID int64  `json:"thread_id"`
	Body     string `json:"body"`
	BodyText string `json:"body_text"`
	// FullMessage includes quoted history where the channel provides it.
	FullMessage string `json:"full_message,omitempty"`
	// ParticipantsInfo records who the message was addressed to.
	ParticipantsInfo map[string]any `json:"participants_info,omitempty"`
	Attachments      []Attachment   `json:"attachments"`
	CreatedAt        Time           `json:"created_at"`
	UpdatedAt        Time           `json:"updated_at"`
}

// ThreadMessageRequest is the create/update payload for a thread message.
type ThreadMessageRequest struct {
	ThreadID    *int64  `json:"thread_id,omitempty"`
	Body        *string `json:"body,omitempty"`
	FullMessage *string `json:"full_message,omitempty"`
	// ParticipantsInfo addresses the message within the thread.
	ParticipantsInfo map[string]any `json:"participants_info,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

const (
	threadsPath        = "collaboration/threads"
	threadMessagesPath = "collaboration/messages"
)

// GetThread fetches a collaboration thread.
func (c *Client) GetThread(ctx context.Context, id int64) (*Thread, error) {
	return getResource[Thread](ctx, c, threadsPath, id, nil)
}

// CreateThread creates a collaboration thread on a ticket.
func (c *Client) CreateThread(ctx context.Context, req ThreadRequest) (*Thread, error) {
	return createResource[Thread](ctx, c, threadsPath, req)
}

// UpdateThread updates a collaboration thread.
func (c *Client) UpdateThread(ctx context.Context, id int64, req ThreadRequest) (*Thread, error) {
	return updateResource[Thread](ctx, c, threadsPath, id, req)
}

// DeleteThread deletes a collaboration thread.
func (c *Client) DeleteThread(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, threadsPath, id)
}

// GetThreadMessage fetches a message from a collaboration thread.
func (c *Client) GetThreadMessage(ctx context.Context, id int64) (*ThreadMessage, error) {
	return getResource[ThreadMessage](ctx, c, threadMessagesPath, id, nil)
}

// CreateThreadMessage posts a message into a collaboration thread.
func (c *Client) CreateThreadMessage(ctx context.Context, req ThreadMessageRequest) (*ThreadMessage, error) {
	if len(req.Attachments) > 0 {
		fields, err := multipartFields(req)
		if err != nil {
			return nil, err
		}
		var out ThreadMessage
		if err := c.PostMultipart(ctx, threadMessagesPath, fields,
			map[string][]string{AttachmentsField: req.Attachments}, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	return createResource[ThreadMessage](ctx, c, threadMessagesPath, req)
}

// UpdateThreadMessage edits a message in a collaboration thread.
func (c *Client) UpdateThreadMessage(ctx context.Context, id int64, req ThreadMessageRequest) (*ThreadMessage, error) {
	return updateResource[ThreadMessage](ctx, c, threadMessagesPath, id, req)
}

// DeleteThreadMessage deletes a message from a collaboration thread.
func (c *Client) DeleteThreadMessage(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, threadMessagesPath, id)
}

// GenerateMessageQuote returns the quoted form of a message, for use when
// replying to it.
func (c *Client) GenerateMessageQuote(ctx context.Context, id int64) (string, error) {
	var out struct {
		QuotedText string `json:"quoted_text"`
	}
	if err := c.Get(ctx, pathFor(threadMessagesPath, id)+"/generate-quote", nil, &out); err != nil {
		return "", err
	}
	return out.QuotedText, nil
}

// OutboundMessage is a proactive message sent to a customer on a channel.
type OutboundMessage struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Channel names the delivery channel, e.g. "whatsapp".
	Channel string `json:"channel"`
	// To identifies the recipient on that channel.
	To string `json:"to"`
	// TicketID is the ticket the message created or continued.
	TicketID  int64          `json:"ticket_id,omitempty"`
	Message   map[string]any `json:"message,omitempty"`
	CreatedAt Time           `json:"created_at"`
	UpdatedAt Time           `json:"updated_at"`
}

const outboundMessagesPath = "channels/outbound-messages"

// SendOutboundMessage sends a proactive outbound message.
func (c *Client) SendOutboundMessage(ctx context.Context, body map[string]any) (*OutboundMessage, error) {
	var out OutboundMessage
	if err := c.Post(ctx, outboundMessagesPath, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOutboundMessage fetches a previously sent outbound message.
func (c *Client) GetOutboundMessage(ctx context.Context, id string) (*OutboundMessage, error) {
	return getResource[OutboundMessage](ctx, c, outboundMessagesPath, id, nil)
}

// PublishContactActivities publishes omnichannel activity records for a
// contact, feeding the unified customer timeline.
func (c *Client) PublishContactActivities(ctx context.Context, body map[string]any) error {
	return c.Post(ctx, "contact-activities", body, nil)
}
