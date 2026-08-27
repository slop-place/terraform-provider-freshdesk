package freshdesk

import (
	"context"
	"net/http"
)

// SLAPolicy defines response and resolution targets for matching tickets.
//
// This endpoint encodes its ID as a quoted string, hence the ID type.
type SLAPolicy struct {
	ID          ID     `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	IsDefault   bool   `json:"is_default"`
	Position    int    `json:"position"`
	// SLATarget maps "priority_1".."priority_4" to that priority's targets.
	SLATarget map[string]SLATarget `json:"sla_target"`
	// ApplicableTo scopes the policy by company, group, product, or source.
	ApplicableTo SLAApplicableTo `json:"applicable_to"`
	// Escalation configures who is notified when a target is missed.
	Escalation map[string]any `json:"escalation"`
	// AdvancedConditions expresses the newer nested matching rules.
	AdvancedConditions map[string]any `json:"advanced_conditions,omitempty"`
	CreatedAt          Time           `json:"created_at"`
	UpdatedAt          Time           `json:"updated_at"`
}

// SLATarget is the target set for one priority.
type SLATarget struct {
	// RespondWithin, ResolveWithin and NextRespondWithin are in seconds.
	// NextRespondWithin is omitted when zero, because Freshdesk requires it to
	// be at least 30 seconds and defaults it when absent.
	RespondWithin     int  `json:"respond_within"`
	ResolveWithin     int  `json:"resolve_within"`
	NextRespondWithin int  `json:"next_respond_within,omitempty"`
	BusinessHours     bool `json:"business_hours"`
	EscalationEnabled bool `json:"escalation_enabled"`
}

// SLAApplicableTo scopes an SLA policy.
type SLAApplicableTo struct {
	CompanyIDs  []int64  `json:"company_ids,omitempty"`
	GroupIDs    []int64  `json:"group_ids,omitempty"`
	ProductIDs  []int64  `json:"product_ids,omitempty"`
	Sources     []int    `json:"sources,omitempty"`
	TicketTypes []string `json:"ticket_types,omitempty"`
}

// SLAPolicyRequest is the create/update payload for an SLA policy.
type SLAPolicyRequest struct {
	Name         *string              `json:"name,omitempty"`
	Description  *string              `json:"description,omitempty"`
	Active       *bool                `json:"active,omitempty"`
	Position     *int                 `json:"position,omitempty"`
	SLATarget    map[string]SLATarget `json:"sla_target,omitempty"`
	ApplicableTo *SLAApplicableTo     `json:"applicable_to,omitempty"`
	Escalation   map[string]any       `json:"escalation,omitempty"`
}

// GetSLAPolicy fetches an SLA policy. Freshdesk exposes no single-policy read,
// so this filters the collection.
func (c *Client) GetSLAPolicy(ctx context.Context, id int64) (*SLAPolicy, error) {
	policies, err := c.ListSLAPolicies(ctx, ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].ID.Int64() == id {
			return &policies[i], nil
		}
	}
	return nil, &Error{StatusCode: http.StatusNotFound, Code: "not_found",
		Message: "sla policy " + idString(id) + " not found"}
}

// ListSLAPolicies returns every SLA policy.
func (c *Client) ListSLAPolicies(ctx context.Context, opts ListOptions) ([]SLAPolicy, error) {
	return listAll[SLAPolicy](ctx, c, "sla_policies", opts)
}

// CreateSLAPolicy creates an SLA policy.
func (c *Client) CreateSLAPolicy(ctx context.Context, req SLAPolicyRequest) (*SLAPolicy, error) {
	return createResource[SLAPolicy](ctx, c, "sla_policies", req)
}

// UpdateSLAPolicy applies a partial update to an SLA policy.
func (c *Client) UpdateSLAPolicy(ctx context.Context, id int64, req SLAPolicyRequest) (*SLAPolicy, error) {
	return updateResource[SLAPolicy](ctx, c, "sla_policies", id, req)
}

// DeleteSLAPolicy deletes an SLA policy.
//
// Freshdesk answers 405 here: the endpoint accepts only PATCH and PUT. The
// method is kept for accounts where the behaviour differs, but callers should
// expect it to fail and deactivate the policy instead.
func (c *Client) DeleteSLAPolicy(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, "sla_policies", id)
}
