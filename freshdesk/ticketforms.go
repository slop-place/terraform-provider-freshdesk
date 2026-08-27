package freshdesk

import "context"

// TicketForm is a portal ticket form.
type TicketForm struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Default marks the form used when no other applies.
	Default bool `json:"default"`
	// Fields is populated when the form is read individually.
	Fields        []TicketFormField `json:"fields,omitempty"`
	LastUpdatedBy int64             `json:"last_updated_by"`
	CreatedAt     Time              `json:"created_at"`
	UpdatedAt     Time              `json:"updated_at"`
}

// TicketFormField is a field's placement and behaviour on a form.
type TicketFormField struct {
	ID    int64  `json:"id"`
	Name  string `json:"name,omitempty"`
	Label string `json:"label,omitempty"`
	Type  string `json:"type,omitempty"`
	// Position orders the field on the form.
	Position int `json:"position,omitempty"`

	LabelForCustomers    string `json:"label_for_customers,omitempty"`
	RequiredForCustomers bool   `json:"required_for_customers"`
	DisplayedToCustomers bool   `json:"displayed_to_customers"`
	CustomersCanEdit     bool   `json:"customers_can_edit"`
	RequiredForAgents    bool   `json:"required_for_agents"`
	RequiredForClosure   bool   `json:"required_for_closure"`

	Choices RawJSON `json:"choices,omitempty"`
}

// TicketFormRequest is the create/update payload for a ticket form.
//
// Freshdesk derives the form's internal name from its title and rejects a
// name in the request body, so only Title and Description are writable.
type TicketFormRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

// TicketFormFieldRequest updates one field's placement on a form.
type TicketFormFieldRequest struct {
	Position             *int    `json:"position,omitempty"`
	LabelForCustomers    *string `json:"label_for_customers,omitempty"`
	RequiredForCustomers *bool   `json:"required_for_customers,omitempty"`
	DisplayedToCustomers *bool   `json:"displayed_to_customers,omitempty"`
	CustomersCanEdit     *bool   `json:"customers_can_edit,omitempty"`
	RequiredForAgents    *bool   `json:"required_for_agents,omitempty"`
	RequiredForClosure   *bool   `json:"required_for_closure,omitempty"`
}

const ticketFormsPath = "ticket-forms"

// GetTicketForm fetches a ticket form and its fields.
func (c *Client) GetTicketForm(ctx context.Context, id int64) (*TicketForm, error) {
	return getResource[TicketForm](ctx, c, ticketFormsPath, id, nil)
}

// ListTicketForms returns every ticket form.
func (c *Client) ListTicketForms(ctx context.Context, opts ListOptions) ([]TicketForm, error) {
	return listAll[TicketForm](ctx, c, ticketFormsPath, opts)
}

// CreateTicketForm creates a ticket form.
func (c *Client) CreateTicketForm(ctx context.Context, req TicketFormRequest) (*TicketForm, error) {
	return createResource[TicketForm](ctx, c, ticketFormsPath, req)
}

// UpdateTicketForm applies a partial update to a ticket form.
func (c *Client) UpdateTicketForm(ctx context.Context, id int64, req TicketFormRequest) (*TicketForm, error) {
	return updateResource[TicketForm](ctx, c, ticketFormsPath, id, req)
}

// DeleteTicketForm deletes a ticket form.
func (c *Client) DeleteTicketForm(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, ticketFormsPath, id)
}

// CloneTicketForm copies a form, giving the copy the supplied name.
func (c *Client) CloneTicketForm(ctx context.Context, id int64, name string) (*TicketForm, error) {
	var out TicketForm
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if err := c.Post(ctx, pathFor(ticketFormsPath, id)+"/clone", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// formFieldPath builds the path to one field's placement on a form.
func formFieldPath(formID, fieldID int64) string {
	return pathFor(ticketFormsPath, formID) + "/fields/" + idString(fieldID)
}

// GetTicketFormField reads one field's placement on a form.
func (c *Client) GetTicketFormField(ctx context.Context, formID, fieldID int64) (*TicketFormField, error) {
	var out TicketFormField
	if err := c.Get(ctx, formFieldPath(formID, fieldID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTicketFormField updates one field's placement on a form.
func (c *Client) UpdateTicketFormField(
	ctx context.Context,
	formID,
	fieldID int64,
	req TicketFormFieldRequest,
) (*TicketFormField, error) {
	var out TicketFormField
	if err := c.Put(ctx, formFieldPath(formID, fieldID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTicketFormField removes a field from a form.
func (c *Client) DeleteTicketFormField(ctx context.Context, formID, fieldID int64) error {
	return c.Delete(ctx, formFieldPath(formID, fieldID))
}
