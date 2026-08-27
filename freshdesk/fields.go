package freshdesk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// TicketField is a field on the ticket form. Custom fields have a "custom_"
// type prefix and a generated name; default fields have Default set.
type TicketField struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	// Type is e.g. "default_subject", "custom_dropdown", "nested_field".
	Type     string `json:"type"`
	Position int    `json:"position"`
	Default  bool   `json:"default"`

	LabelForCustomers    string `json:"label_for_customers"`
	CustomersCanEdit     bool   `json:"customers_can_edit"`
	CustomersCanFilter   bool   `json:"customers_can_filter"`
	DisplayedToCustomers bool   `json:"displayed_to_customers"`
	RequiredForCustomers bool   `json:"required_for_customers"`
	RequiredForAgents    bool   `json:"required_for_agents"`
	RequiredForClosure   bool   `json:"required_for_closure"`

	// PortalCC and PortalCCTo configure the CC control on the portal form.
	PortalCC   bool   `json:"portal_cc"`
	PortalCCTo string `json:"portal_cc_to"`
	// Archived marks a field withdrawn from the form but kept on
	// existing tickets.
	Archived bool `json:"archived"`

	// Choices is raw because its shape depends on the field; use its Choices
	// method for a custom dropdown.
	Choices RawJSON `json:"choices"`
	// DependentFields carries the nested levels of a nested_field.
	DependentFields []TicketField `json:"dependent_fields"`
	// HasSection and Sections describe dynamic sections driven by this field.
	HasSection      bool             `json:"has_section"`
	Sections        []TicketSection  `json:"sections"`
	SectionMappings []SectionMapping `json:"section_mappings"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// SectionMapping places a field inside a section at a position.
type SectionMapping struct {
	SectionID int64 `json:"section_id"`
	Position  int   `json:"position"`
}

// TicketSection is a dynamic section shown for particular choices of a field.
type TicketSection struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	// ParentTicketFieldID is the field whose choices reveal this section.
	ParentTicketFieldID int64 `json:"parent_ticket_field_id"`
	// ChoiceIDs are the parent-field choices that display the section.
	ChoiceIDs []int64 `json:"choice_ids"`
	// TicketFieldIDs are the fields contained in the section.
	TicketFieldIDs []int64 `json:"ticket_field_ids"`
	IsFSM          bool    `json:"is_fsm"`
}

// TicketFieldRequest is the create/update payload for a ticket field.
type TicketFieldRequest struct {
	Label                *string `json:"label,omitempty"`
	LabelForCustomers    *string `json:"label_for_customers,omitempty"`
	Type                 *string `json:"type,omitempty"`
	Position             *int    `json:"position,omitempty"`
	Description          *string `json:"description,omitempty"`
	CustomersCanEdit     *bool   `json:"customers_can_edit,omitempty"`
	DisplayedToCustomers *bool   `json:"displayed_to_customers,omitempty"`
	RequiredForCustomers *bool   `json:"required_for_customers,omitempty"`
	RequiredForAgents    *bool   `json:"required_for_agents,omitempty"`
	RequiredForClosure   *bool   `json:"required_for_closure,omitempty"`

	Choices         []Choice         `json:"choices,omitempty"`
	DependentFields []map[string]any `json:"dependent_fields,omitempty"`
	SectionMappings []SectionMapping `json:"section_mappings,omitempty"`
}

const ticketFieldsPath = "admin/ticket_fields"

// GetTicketField fetches a ticket field. Set includeSection to sideload the
// dynamic sections the field drives.
func (c *Client) GetTicketField(ctx context.Context, id int64, includeSection bool) (*TicketField, error) {
	q := url.Values{}
	if includeSection {
		q.Set("include", "section")
	}
	return getResource[TicketField](ctx, c, ticketFieldsPath, id, q)
}

// ListTicketFields returns every ticket field through the admin API.
func (c *Client) ListTicketFields(ctx context.Context, opts ListOptions) ([]TicketField, error) {
	return listAll[TicketField](ctx, c, ticketFieldsPath, opts)
}

// ListTicketFieldsLegacy returns ticket fields through the older
// /ticket_fields endpoint, which some accounts still rely on.
func (c *Client) ListTicketFieldsLegacy(ctx context.Context, opts ListOptions) ([]TicketField, error) {
	return listAll[TicketField](ctx, c, "ticket_fields", opts)
}

// CreateTicketField creates a ticket field.
func (c *Client) CreateTicketField(ctx context.Context, req TicketFieldRequest) (*TicketField, error) {
	return createResource[TicketField](ctx, c, ticketFieldsPath, req)
}

// UpdateTicketField applies a partial update to a ticket field.
func (c *Client) UpdateTicketField(ctx context.Context, id int64, req TicketFieldRequest) (*TicketField, error) {
	return updateResource[TicketField](ctx, c, ticketFieldsPath, id, req)
}

// DeleteTicketField deletes a ticket field.
func (c *Client) DeleteTicketField(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, ticketFieldsPath, id)
}

// SectionRequest is the create/update payload for a dynamic section.
type SectionRequest struct {
	Label     *string `json:"label,omitempty"`
	ChoiceIDs []int64 `json:"choice_ids,omitempty"`
}

// sectionsPath builds the sections collection path for a parent field.
func sectionsPath(fieldID int64) string {
	return pathFor(ticketFieldsPath, fieldID) + "/sections"
}

// ListSections lists the dynamic sections driven by a ticket field.
//
// The published documentation wraps the collection in a "sections" envelope,
// but the API returns a bare array. Both are accepted.
func (c *Client) ListSections(ctx context.Context, fieldID int64) ([]TicketSection, error) {
	var raw json.RawMessage
	if err := c.Get(ctx, sectionsPath(fieldID), nil, &raw); err != nil {
		return nil, err
	}

	var bare []TicketSection
	if err := json.Unmarshal(raw, &bare); err == nil {
		return bare, nil
	}

	var envelope struct {
		Sections []TicketSection `json:"sections"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decoding the ticket field's sections: %w", err)
	}

	return envelope.Sections, nil
}

// GetSection fetches one dynamic section.
func (c *Client) GetSection(ctx context.Context, fieldID, sectionID int64) (*TicketSection, error) {
	return getResource[TicketSection](ctx, c, sectionsPath(fieldID), sectionID, nil)
}

// CreateSection creates a dynamic section under a ticket field.
//
// Freshdesk is inconsistent about what it answers with here: some accounts
// return the parent field's whole section list in a "sections" envelope, others
// return the created section on its own. Both are accepted, and if neither
// carries an identifier the section is looked up by label.
func (c *Client) CreateSection(
	ctx context.Context,
	fieldID int64,
	req SectionRequest,
) (*TicketSection, error) {
	var raw json.RawMessage
	if err := c.Post(ctx, sectionsPath(fieldID), req, &raw); err != nil {
		return nil, err
	}

	if section := sectionFromResponse(raw, req.Label); section != nil {
		return section, nil
	}

	// The response told us nothing useful; find the section by its label.
	sections, err := c.ListSections(ctx, fieldID)
	if err != nil {
		return nil, err
	}

	if req.Label != nil {
		for i := len(sections) - 1; i >= 0; i-- {
			if sections[i].Label == *req.Label {
				return &sections[i], nil
			}
		}
	}

	if len(sections) > 0 {
		return &sections[len(sections)-1], nil
	}

	return nil, &Error{
		StatusCode: http.StatusOK,
		Message:    "creating the section returned no section and none could be found",
	}
}

// sectionFromResponse pulls the created section out of either shape Freshdesk
// answers a section create with.
func sectionFromResponse(raw json.RawMessage, label *string) *TicketSection {
	var envelope struct {
		Sections []TicketSection `json:"sections"`
	}

	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Sections) > 0 {
		if label != nil {
			for i := len(envelope.Sections) - 1; i >= 0; i-- {
				if envelope.Sections[i].Label == *label {
					return &envelope.Sections[i]
				}
			}
		}

		return &envelope.Sections[len(envelope.Sections)-1]
	}

	var single TicketSection
	if err := json.Unmarshal(raw, &single); err == nil && single.ID != 0 {
		return &single
	}

	return nil
}

// UpdateSection updates a dynamic section.
func (c *Client) UpdateSection(
	ctx context.Context,
	fieldID,
	sectionID int64,
	req SectionRequest,
) (*TicketSection, error) {
	return updateResource[TicketSection](ctx, c, sectionsPath(fieldID), sectionID, req)
}

// DeleteSection deletes a dynamic section.
func (c *Client) DeleteSection(ctx context.Context, fieldID, sectionID int64) error {
	return deleteResource(ctx, c, sectionsPath(fieldID), sectionID)
}

// ContactField is a field on the contact form.
type ContactField struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Position int    `json:"position"`
	Default  bool   `json:"default"`

	LabelForCustomers     string `json:"label_for_customers"`
	CustomersCanEdit      bool   `json:"customers_can_edit"`
	DisplayedForCustomers bool   `json:"displayed_for_customers"`
	RequiredForCustomers  bool   `json:"required_for_customers"`
	RequiredForAgents     bool   `json:"required_for_agents"`
	AgentsCanEdit         bool   `json:"agents_can_edit"`
	DisplayedForAgents    bool   `json:"displayed_for_agents"`
	QuickAddForAgent      bool   `json:"quick_add_for_agent"`
	EditableInSignup      bool   `json:"editable_in_signup"`
	Unique                bool   `json:"unique"`

	// Choices is raw because its shape depends on the field; use its Choices
	// method for a custom dropdown.
	Choices   RawJSON `json:"choices"`
	CreatedAt Time    `json:"created_at"`
	UpdatedAt Time    `json:"updated_at"`
}

// ContactFieldRequest is the create/update payload for a contact field.
type ContactFieldRequest struct {
	Label                 *string  `json:"label,omitempty"`
	LabelForCustomers     *string  `json:"label_for_customers,omitempty"`
	Type                  *string  `json:"type,omitempty"`
	Position              *int     `json:"position,omitempty"`
	EditableInSignup      *bool    `json:"editable_in_signup,omitempty"`
	RequiredForAgents     *bool    `json:"required_for_agents,omitempty"`
	AgentsCanEdit         *bool    `json:"agents_can_edit,omitempty"`
	DisplayedForAgents    *bool    `json:"displayed_for_agents,omitempty"`
	QuickAddForAgent      *bool    `json:"quick_add_for_agent,omitempty"`
	Unique                *bool    `json:"unique,omitempty"`
	CustomersCanEdit      *bool    `json:"customers_can_edit,omitempty"`
	RequiredForCustomers  *bool    `json:"required_for_customers,omitempty"`
	DisplayedForCustomers *bool    `json:"displayed_for_customers,omitempty"`
	Choices               []Choice `json:"choices,omitempty"`
}

// GetContactField fetches a contact field.
func (c *Client) GetContactField(ctx context.Context, id int64) (*ContactField, error) {
	return getResource[ContactField](ctx, c, "contact_fields", id, nil)
}

// ListContactFields returns every contact field.
func (c *Client) ListContactFields(ctx context.Context, opts ListOptions) ([]ContactField, error) {
	return listAll[ContactField](ctx, c, "contact_fields", opts)
}

// CreateContactField creates a contact field.
func (c *Client) CreateContactField(ctx context.Context, req ContactFieldRequest) (*ContactField, error) {
	return createResource[ContactField](ctx, c, "contact_fields", req)
}

// UpdateContactField applies a partial update to a contact field.
func (c *Client) UpdateContactField(ctx context.Context, id int64, req ContactFieldRequest) (*ContactField, error) {
	return updateResource[ContactField](ctx, c, "contact_fields", id, req)
}

// DeleteContactField deletes a contact field.
//
// The published documentation puts this verb on a singular "contact_field"
// path, but that answers 404 in practice; the plural path is the one the API
// serves.
func (c *Client) DeleteContactField(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "contact_fields", id)
}

// CompanyField is a field on the company form.
type CompanyField struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Position int    `json:"position"`
	Default  bool   `json:"default"`

	RequiredForAgents  bool `json:"required_for_agents"`
	AgentsCanEdit      bool `json:"agents_can_edit"`
	DisplayedForAgents bool `json:"displayed_for_agents"`
	QuickAddForAgent   bool `json:"quick_add_for_agent"`
	Unique             bool `json:"unique"`

	// Choices is raw because its shape depends on the field; use its Choices
	// method for a custom dropdown.
	Choices   RawJSON `json:"choices"`
	CreatedAt Time    `json:"created_at"`
	UpdatedAt Time    `json:"updated_at"`
}

// CompanyFieldRequest is the create/update payload for a company field.
type CompanyFieldRequest struct {
	Label              *string  `json:"label,omitempty"`
	Type               *string  `json:"type,omitempty"`
	Position           *int     `json:"position,omitempty"`
	RequiredForAgents  *bool    `json:"required_for_agents,omitempty"`
	AgentsCanEdit      *bool    `json:"agents_can_edit,omitempty"`
	DisplayedForAgents *bool    `json:"displayed_for_agents,omitempty"`
	QuickAddForAgent   *bool    `json:"quick_add_for_agent,omitempty"`
	Unique             *bool    `json:"unique,omitempty"`
	Choices            []Choice `json:"choices,omitempty"`
}

// GetCompanyField fetches a company field.
func (c *Client) GetCompanyField(ctx context.Context, id int64) (*CompanyField, error) {
	return getResource[CompanyField](ctx, c, "company_fields", id, nil)
}

// ListCompanyFields returns every company field.
func (c *Client) ListCompanyFields(ctx context.Context, opts ListOptions) ([]CompanyField, error) {
	return listAll[CompanyField](ctx, c, "company_fields", opts)
}

// CreateCompanyField creates a company field.
func (c *Client) CreateCompanyField(ctx context.Context, req CompanyFieldRequest) (*CompanyField, error) {
	return createResource[CompanyField](ctx, c, "company_fields", req)
}

// UpdateCompanyField applies a partial update to a company field.
//
// As with contact fields, the documented singular path answers 404; the plural
// path is the one the API serves.
func (c *Client) UpdateCompanyField(ctx context.Context, id int64, req CompanyFieldRequest) (*CompanyField, error) {
	return updateResource[CompanyField](ctx, c, "company_fields", id, req)
}

// DeleteCompanyField deletes a company field.
func (c *Client) DeleteCompanyField(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "company_fields", id)
}
