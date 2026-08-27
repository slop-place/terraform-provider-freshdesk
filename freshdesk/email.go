package freshdesk

import "context"

// EmailConfig is a support address mapped to a product and group. Configs are
// read-only through this API; use the mailboxes API to manage them.
type EmailConfig struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ToEmail    string `json:"to_email"`
	ReplyEmail string `json:"reply_email"`
	ProductID  int64  `json:"product_id"`
	GroupID    int64  `json:"group_id"`
	// PrimaryRole marks the account's default support address.
	PrimaryRole bool `json:"primary_role"`
	Active      bool `json:"active"`
	// WorkspaceID scopes the config on multi-workspace accounts.
	WorkspaceID int64 `json:"workspace_id"`
	CreatedAt   Time  `json:"created_at"`
	UpdatedAt   Time  `json:"updated_at"`
}

// GetEmailConfig fetches an email config.
func (c *Client) GetEmailConfig(ctx context.Context, id int64) (*EmailConfig, error) {
	return getResource[EmailConfig](ctx, c, "email_configs", id, nil)
}

// ListEmailConfigs returns every email config.
func (c *Client) ListEmailConfigs(ctx context.Context, opts ListOptions) ([]EmailConfig, error) {
	return listAll[EmailConfig](ctx, c, "email_configs", opts)
}

// Mailbox types.
const (
	// MailboxTypeFreshdesk is a Freshdesk-hosted forwarding address.
	MailboxTypeFreshdesk = "freshdesk_mailbox"
	// MailboxTypeCustom is a customer-owned mailbox reached over IMAP/SMTP.
	MailboxTypeCustom = "custom_mailbox"
)

// EmailMailbox is a support mailbox.
type EmailMailbox struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SupportEmail string `json:"support_email"`
	GroupID      int64  `json:"group_id"`
	ProductID    int64  `json:"product_id"`
	// DefaultReplyEmail marks the address outgoing replies come from.
	DefaultReplyEmail bool `json:"default_reply_email"`
	Active            bool `json:"active"`
	// MailboxType is MailboxTypeFreshdesk or MailboxTypeCustom.
	MailboxType string `json:"mailbox_type"`
	// FailureCode reports why a custom mailbox stopped working, when it has.
	FailureCode any `json:"failure_code"`
	// DisableVerify skips ownership verification on create.
	DisableVerify bool `json:"disable_verify"`

	// FreshdeskMailbox is populated for MailboxTypeFreshdesk.
	FreshdeskMailbox *Mailbox `json:"freshdesk_mailbox,omitempty"`
	// CustomMailbox is populated for MailboxTypeCustom.
	CustomMailbox *CustomMailbox `json:"custom_mailbox,omitempty"`
	// CSATSettings configures the satisfaction survey sent from this mailbox.
	CSATSettings *MailboxCSATSettings `json:"csat_settings,omitempty"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// Mailbox holds the forwarding address of a Freshdesk-hosted mailbox.
type Mailbox struct {
	ForwardEmail string `json:"forward_email"`
}

// CustomMailbox holds the IMAP/SMTP settings of a customer-owned mailbox.
type CustomMailbox struct {
	IncomingSettings *MailServerSettings `json:"incoming,omitempty"`
	OutgoingSettings *MailServerSettings `json:"outgoing,omitempty"`
	// AccessTokenID references stored OAuth credentials, where used.
	AccessTokenID int64 `json:"access_token_id,omitempty"`
}

// MailServerSettings describes one leg of a custom mailbox connection.
type MailServerSettings struct {
	MailServer string `json:"mail_server,omitempty"`
	Port       int    `json:"port,omitempty"`
	UserName   string `json:"user_name,omitempty"`
	// Password is write-only; Freshdesk never returns it.
	Password           string `json:"password,omitempty"`
	UseSSL             *bool  `json:"use_ssl,omitempty"`
	DeleteFromServer   *bool  `json:"delete_from_server,omitempty"`
	AuthenticationType string `json:"authentication_type,omitempty"`
}

// MailboxCSATSettings configures a mailbox's satisfaction survey.
type MailboxCSATSettings struct {
	SurveyEnabled             bool   `json:"survey_enabled"`
	SurveyID                  any    `json:"survey_id"`
	MinimumAgentResponseCount int    `json:"minimum_agent_response_count"`
	UserResponseInterval      int    `json:"user_response_interval"`
	UserResponseIntervalUnit  string `json:"user_response_interval_unit"`
	SurveyExpiryTime          int    `json:"survey_expiry_time"`
	SurveyExpiryTimeUnit      string `json:"survey_expiry_time_unit"`
	RetakeSurveyEnabled       bool   `json:"retake_survey_enabled"`
}

// EmailMailboxRequest is the create/update payload for a mailbox.
type EmailMailboxRequest struct {
	Name              *string `json:"name,omitempty"`
	SupportEmail      *string `json:"support_email,omitempty"`
	GroupID           *int64  `json:"group_id,omitempty"`
	ProductID         *int64  `json:"product_id,omitempty"`
	DefaultReplyEmail *bool   `json:"default_reply_email,omitempty"`
	Active            *bool   `json:"active,omitempty"`
	MailboxType       *string `json:"mailbox_type,omitempty"`
	DisableVerify     *bool   `json:"disable_verify,omitempty"`

	CustomMailbox *CustomMailbox       `json:"custom_mailbox,omitempty"`
	CSATSettings  *MailboxCSATSettings `json:"csat_settings,omitempty"`
}

const mailboxesPath = "email/mailboxes"

// GetEmailMailbox fetches a mailbox.
func (c *Client) GetEmailMailbox(ctx context.Context, id int64) (*EmailMailbox, error) {
	return getResource[EmailMailbox](ctx, c, mailboxesPath, id, nil)
}

// ListEmailMailboxes returns every mailbox.
func (c *Client) ListEmailMailboxes(ctx context.Context, opts ListOptions) ([]EmailMailbox, error) {
	return listAll[EmailMailbox](ctx, c, mailboxesPath, opts)
}

// CreateEmailMailbox creates a mailbox.
func (c *Client) CreateEmailMailbox(ctx context.Context, req EmailMailboxRequest) (*EmailMailbox, error) {
	return createResource[EmailMailbox](ctx, c, mailboxesPath, req)
}

// UpdateEmailMailbox applies a partial update to a mailbox.
func (c *Client) UpdateEmailMailbox(ctx context.Context, id int64, req EmailMailboxRequest) (*EmailMailbox, error) {
	return updateResource[EmailMailbox](ctx, c, mailboxesPath, id, req)
}

// DeleteEmailMailbox deletes a mailbox.
func (c *Client) DeleteEmailMailbox(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, mailboxesPath, id)
}

// EmailSettings holds account-wide email handling toggles.
type EmailSettings struct {
	PersonalizedEmailReplies             bool `json:"personalized_email_replies"`
	CreateRequesterUsingReplyTo          bool `json:"create_requester_using_reply_to"`
	AllowAgentToInitiateConversation     bool `json:"allow_agent_to_initiate_conversation"`
	OriginalSenderAsRequesterForForward  bool `json:"original_sender_as_requester_for_forward"`
	AllowWildcardTicketCreate            bool `json:"allow_wildcard_ticket_create"`
	SkipTicketThreading                  bool `json:"skip_ticket_threading"`
	ThreadingWithoutUserCheck            bool `json:"threading_without_user_check"`
	ThreadingWithoutTicketIDCheck        bool `json:"threading_without_ticket_id_check"`
	ExtendedQuotedText                   bool `json:"extended_quoted_text"`
	AutoResponseDetectorToggle           bool `json:"auto_response_detector_toggle"`
	EmailSubjectMatch                    bool `json:"email_subject_match"`
	MultipleTo                           bool `json:"multiple_to"`
	PrioritizeTicketsByOutlookImportance bool `json:"prioritize_tickets_by_outlook_importance"`
}

// GetEmailSettings reads the account-wide email settings.
func (c *Client) GetEmailSettings(ctx context.Context) (*EmailSettings, error) {
	var out EmailSettings
	if err := c.Get(ctx, "email/settings", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEmailSettings writes the account-wide email settings. Only the keys
// present in settings are changed.
func (c *Client) UpdateEmailSettings(ctx context.Context, settings map[string]bool) (*EmailSettings, error) {
	var out EmailSettings
	if err := c.Put(ctx, "email/settings", settings, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNotificationBCC reads the addresses BCC'd on outgoing notifications.
func (c *Client) GetNotificationBCC(ctx context.Context) ([]string, error) {
	var out struct {
		Emails []string `json:"emails"`
	}
	if err := c.Get(ctx, "notifications/email/bcc", nil, &out); err != nil {
		return nil, err
	}
	return out.Emails, nil
}

// UpdateNotificationBCC replaces the automatic BCC address list.
func (c *Client) UpdateNotificationBCC(ctx context.Context, emails []string) ([]string, error) {
	if emails == nil {
		emails = []string{}
	}
	var out struct {
		Emails []string `json:"emails"`
	}
	if err := c.Put(ctx, "notifications/email/bcc", map[string]any{"emails": emails}, &out); err != nil {
		return nil, err
	}
	return out.Emails, nil
}
