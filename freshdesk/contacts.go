package freshdesk

import (
	"context"
	"net/url"
	"strconv"
)

// Contact is a Freshdesk requester.
type Contact struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// FirstName and LastName are the split form of Name, maintained by
	// Freshdesk on accounts with the newer contacts model.
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Mobile      string `json:"mobile"`
	TwitterID   string `json:"twitter_id"`
	FacebookID  string `json:"facebook_id"`
	Address     string `json:"address"`
	Description string `json:"description"`
	JobTitle    string `json:"job_title"`
	Language    string `json:"language"`
	TimeZone    string `json:"time_zone"`
	// ContactType distinguishes contacts from agents downgraded to contacts.
	ContactType      string   `json:"contact_type"`
	UniqueExternalID string   `json:"unique_external_id"`
	OtherEmails      []string `json:"other_emails"`
	Tags             []string `json:"tags"`

	CompanyID      int64 `json:"company_id"`
	ViewAllTickets bool  `json:"view_all_tickets"`
	// OtherCompanies requires the Multiple Companies feature.
	OtherCompanies []ContactCompany `json:"other_companies"`
	// SocialHandler holds up to ten social profiles.
	SocialHandler []SocialHandle `json:"social_handler"`

	CustomFields CustomFields `json:"custom_fields"`
	Avatar       *Avatar      `json:"avatar"`

	Active  bool `json:"active"`
	Deleted bool `json:"deleted"`

	// OtherPhoneNumbers holds additional numbers on the newer contacts model.
	OtherPhoneNumbers []string `json:"other_phone_numbers"`
	// CSATRating is the contact's most recent satisfaction rating.
	CSATRating any `json:"csat_rating"`
	// PreferredSource is the channel the contact most often uses.
	PreferredSource string `json:"preferred_source"`
	FirstSeen       Time   `json:"first_seen"`
	LastSeen        Time   `json:"last_seen"`
	IPAddress       string `json:"ip_address"`
	VisitorID       string `json:"visitor_id"`
	// OrgContactID identifies the contact across the Freshworks organisation.
	OrgContactID    ID         `json:"org_contact_id"`
	OrgContactIDStr FlexString `json:"org_contact_id_str"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// ContactCompany associates a contact with an additional company.
type ContactCompany struct {
	CompanyID      int64 `json:"company_id"`
	ViewAllTickets bool  `json:"view_all_tickets"`
}

// SocialHandle is one social profile on a contact.
type SocialHandle struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ContactRequest is the create/update payload for a contact. At least one of
// Email, Phone, Mobile, TwitterID, or UniqueExternalID must be set on create.
type ContactRequest struct {
	Name             *string  `json:"name,omitempty"`
	Email            *string  `json:"email,omitempty"`
	Phone            *string  `json:"phone,omitempty"`
	Mobile           *string  `json:"mobile,omitempty"`
	TwitterID        *string  `json:"twitter_id,omitempty"`
	UniqueExternalID *string  `json:"unique_external_id,omitempty"`
	Address          *string  `json:"address,omitempty"`
	Description      *string  `json:"description,omitempty"`
	JobTitle         *string  `json:"job_title,omitempty"`
	Language         *string  `json:"language,omitempty"`
	TimeZone         *string  `json:"time_zone,omitempty"`
	OtherEmails      []string `json:"other_emails,omitempty"`
	Tags             []string `json:"tags,omitempty"`

	CompanyID      *int64           `json:"company_id,omitempty"`
	ViewAllTickets *bool            `json:"view_all_tickets,omitempty"`
	OtherCompanies []ContactCompany `json:"other_companies,omitempty"`
	SocialHandler  []SocialHandle   `json:"social_handler,omitempty"`

	CustomFields CustomFields `json:"custom_fields,omitempty"`
	// LookupParameter is "display_id" or "primary_field_value" and selects how
	// custom-object lookups in CustomFields are interpreted.
	LookupParameter *string `json:"lookup_parameter,omitempty"`

	// Avatar is a local image path; setting it forces a multipart request.
	Avatar string `json:"-"`

	// ClearTags sends an empty tags array, removing every tag.
	ClearTags bool `json:"-"`
	// ClearCompany sends "company_id": null, detaching the primary company.
	ClearCompany bool `json:"-"`
}

// MarshalJSON applies the null/empty clearing semantics.
func (r ContactRequest) MarshalJSON() ([]byte, error) {
	type alias ContactRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}
	if r.ClearTags {
		m["tags"] = []string{}
	}
	if r.ClearCompany {
		m["company_id"] = nil
	}
	return marshalMap(m)
}

const contactsPath = "contacts"

// GetContact fetches a contact by ID.
func (c *Client) GetContact(ctx context.Context, id int64) (*Contact, error) {
	return getResource[Contact](ctx, c, contactsPath, id, nil)
}

// ContactListOptions filters the contact collection endpoint.
type ContactListOptions struct {
	ListOptions

	Email            string
	Mobile           string
	Phone            string
	CompanyID        int64
	UniqueExternalID string
	// State is "verified", "unverified", "blocked", or "deleted".
	State string
}

// ListContacts returns every contact matching opts.
func (c *Client) ListContacts(ctx context.Context, opts ContactListOptions) ([]Contact, error) {
	base := opts.ListOptions
	v := base.Values()
	if opts.Email != "" {
		v.Set("email", opts.Email)
	}
	if opts.Mobile != "" {
		v.Set("mobile", opts.Mobile)
	}
	if opts.Phone != "" {
		v.Set("phone", opts.Phone)
	}
	if opts.CompanyID > 0 {
		v.Set("company_id", strconv.FormatInt(opts.CompanyID, 10))
	}
	if opts.UniqueExternalID != "" {
		v.Set("unique_external_id", opts.UniqueExternalID)
	}
	if opts.State != "" {
		v.Set("state", opts.State)
	}
	base.Extra = v
	return listAll[Contact](ctx, c, contactsPath, base)
}

// CreateContact creates a contact.
func (c *Client) CreateContact(ctx context.Context, req ContactRequest) (*Contact, error) {
	if req.Avatar != "" {
		var out Contact
		fields, err := multipartFields(req)
		if err != nil {
			return nil, err
		}
		if err := c.PostMultipart(ctx, contactsPath, fields,
			map[string][]string{"avatar": {req.Avatar}}, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	return createResource[Contact](ctx, c, contactsPath, req)
}

// UpdateContact applies a partial update to a contact.
func (c *Client) UpdateContact(ctx context.Context, id int64, req ContactRequest) (*Contact, error) {
	return updateResource[Contact](ctx, c, contactsPath, id, req)
}

// DeleteContact soft-deletes a contact; RestoreContact undoes it.
func (c *Client) DeleteContact(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, contactsPath, id)
}

// HardDeleteContact permanently deletes a contact. Setting force also deletes
// a contact that has not been soft-deleted first.
func (c *Client) HardDeleteContact(ctx context.Context, id int64, force bool) error {
	path := pathFor(contactsPath, id) + "/hard_delete"
	if force {
		path += "?force=true"
	}
	return c.Delete(ctx, path)
}

// RestoreContact undoes DeleteContact.
func (c *Client) RestoreContact(ctx context.Context, id int64) error {
	return c.Put(ctx, pathFor(contactsPath, id)+"/restore", nil, nil)
}

// SendContactInvite emails a contact an activation invitation.
func (c *Client) SendContactInvite(ctx context.Context, id int64) error {
	return c.Put(ctx, pathFor(contactsPath, id)+"/send_invite", nil, nil)
}

// MakeAgentRequest promotes a contact to an agent.
type MakeAgentRequest struct {
	Occasional  *bool   `json:"occasional,omitempty"`
	TicketScope *int    `json:"ticket_scope,omitempty"`
	RoleIDs     []int64 `json:"role_ids,omitempty"`
	GroupIDs    []int64 `json:"group_ids,omitempty"`
	SkillIDs    []int64 `json:"skill_ids,omitempty"`
	AgentType   *int    `json:"agent_type,omitempty"`
	Signature   *string `json:"signature,omitempty"`
}

// MakeAgent promotes a contact to an agent and returns the new agent.
func (c *Client) MakeAgent(ctx context.Context, id int64, req MakeAgentRequest) (*Agent, error) {
	var out Agent
	if err := c.Put(ctx, pathFor(contactsPath, id)+"/make_agent", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ContactMergeRequest merges secondary contacts into a primary contact.
type ContactMergeRequest struct {
	PrimaryContactID    int64   `json:"primary_contact_id"`
	SecondaryContactIDs []int64 `json:"secondary_contact_ids"`
	// Contact optionally overrides fields on the surviving contact.
	Contact map[string]any `json:"contact,omitempty"`
}

// MergeContacts merges contacts.
func (c *Client) MergeContacts(ctx context.Context, req ContactMergeRequest) error {
	return c.Post(ctx, contactsPath+"/merge", req, nil)
}

// SearchContacts matches contacts by name or email prefix.
func (c *Client) SearchContacts(ctx context.Context, term string) ([]Contact, error) {
	var out []Contact
	err := c.Get(ctx, contactsPath+"/autocomplete", url.Values{"term": {term}}, &out)
	return out, err
}

// ContactSearchResult is the envelope returned by the contact filter endpoint.
type ContactSearchResult struct {
	Total   int       `json:"total"`
	Results []Contact `json:"results"`
}

// FilterContacts runs a Freshdesk query-language filter over contacts.
func (c *Client) FilterContacts(ctx context.Context, query string, page int) (*ContactSearchResult, error) {
	q := url.Values{"query": {quoteQuery(query)}}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var out ContactSearchResult
	if err := c.Get(ctx, "search/contacts", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportJob identifies an asynchronous contact or company export.
type ExportJob struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// DownloadURL is populated once the export completes.
	DownloadURL string `json:"download_url"`
}

// ExportContacts starts a contact export covering the named default and custom
// fields, and returns the job to poll with GetContactExport.
func (c *Client) ExportContacts(ctx context.Context, defaultFields, customFields []string) (*ExportJob, error) {
	body := map[string]any{
		"fields": map[string]any{
			"default_fields": defaultFields,
			"custom_fields":  customFields,
		},
	}
	var out ExportJob
	if err := c.Post(ctx, contactsPath+"/export", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetContactExport polls a contact export job.
func (c *Client) GetContactExport(ctx context.Context, jobID string) (*ExportJob, error) {
	return getResource[ExportJob](ctx, c, contactsPath+"/export", jobID, nil)
}

// ImportJob identifies an asynchronous contact or company import.
type ImportJob struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	// Records reports progress once the import is running.
	Records map[string]any `json:"records,omitempty"`
}

// ImportContacts starts a contact import from an uploaded CSV.
func (c *Client) ImportContacts(ctx context.Context, csvPath string, fields map[string]string) (*ImportJob, error) {
	form := map[string][]string{}
	for k, v := range fields {
		form["fields["+k+"]"] = []string{v}
	}
	var out ImportJob
	if err := c.PostMultipart(ctx, contactsPath+"/imports", form,
		map[string][]string{"file": {csvPath}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetContactImport polls a contact import job.
func (c *Client) GetContactImport(ctx context.Context, id int64) (*ImportJob, error) {
	return getResource[ImportJob](ctx, c, contactsPath+"/imports", id, nil)
}

// CancelContactImport cancels a running contact import.
func (c *Client) CancelContactImport(ctx context.Context, id int64) error {
	return c.Post(ctx, pathFor(contactsPath+"/imports", id)+"/cancel", nil, nil)
}
