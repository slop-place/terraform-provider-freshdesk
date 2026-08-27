package freshdesk

import "context"

// Group is a Freshdesk agent group.
type Group struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// BusinessHourID is the business-hours calendar the group follows.
	BusinessHourID int64 `json:"business_hour_id"`
	// EscalateTo is the agent notified when a ticket stays unassigned.
	EscalateTo int64 `json:"escalate_to"`
	// UnassignedFor is how long a ticket may stay unassigned before escalation
	// ("30m", "1h", "2h", "4h", "8h", "12h", "1d", "2d", "3d").
	UnassignedFor string `json:"unassigned_for"`
	// AgentIDs lists the group's members. Freshdesk omits it on list responses.
	AgentIDs []int64 `json:"agent_ids"`
	// AutoTicketAssign is 0 (off) or 1 (on) on this endpoint. Richer assignment
	// types are only settable through the admin groups API.
	AutoTicketAssign int  `json:"auto_ticket_assign"`
	CreatedAt        Time `json:"created_at"`
	UpdatedAt        Time `json:"updated_at"`
}

// GroupRequest is the create/update payload for a group.
type GroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	// EscalateTo is explicitly nullable: send a nil *int64 inside a set pointer
	// to clear it. Use ClearEscalateTo for that.
	EscalateTo    *int64  `json:"escalate_to,omitempty"`
	UnassignedFor *string `json:"unassigned_for,omitempty"`
	// AgentIDs replaces the membership wholesale. A non-nil empty slice removes
	// every agent, which is why it is not omitempty-guarded by a pointer.
	AgentIDs         []int64 `json:"agent_ids,omitempty"`
	AutoTicketAssign *int    `json:"auto_ticket_assign,omitempty"`
	BusinessHourID   *int64  `json:"business_hour_id,omitempty"`

	// ClearEscalateTo sends "escalate_to": null, the documented way to unset it.
	ClearEscalateTo bool `json:"-"`
	// ClearAgents sends an empty agent_ids array, removing all members.
	ClearAgents bool `json:"-"`
}

// MarshalJSON renders the explicit-null and empty-array clearing semantics that
// Freshdesk requires, which struct tags alone cannot express.
func (r GroupRequest) MarshalJSON() ([]byte, error) {
	type alias GroupRequest // avoid recursing into this method
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}
	if r.ClearEscalateTo {
		m["escalate_to"] = nil
	}
	if r.ClearAgents {
		m["agent_ids"] = []int64{}
	}
	return marshalMap(m)
}

const groupsPath = "groups"

// GetGroup fetches a group by ID.
func (c *Client) GetGroup(ctx context.Context, id int64) (*Group, error) {
	return getResource[Group](ctx, c, groupsPath, id, nil)
}

// ListGroups returns every group.
func (c *Client) ListGroups(ctx context.Context, opts ListOptions) ([]Group, error) {
	return listAll[Group](ctx, c, groupsPath, opts)
}

// CreateGroup creates a group.
func (c *Client) CreateGroup(ctx context.Context, req GroupRequest) (*Group, error) {
	return createResource[Group](ctx, c, groupsPath, req)
}

// UpdateGroup applies a partial update to a group.
func (c *Client) UpdateGroup(ctx context.Context, id int64, req GroupRequest) (*Group, error) {
	return updateResource[Group](ctx, c, groupsPath, id, req)
}

// DeleteGroup deletes a group.
func (c *Client) DeleteGroup(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, groupsPath, id)
}

// Group types accepted by the admin groups API.
const (
	GroupTypeSupportAgent = "support_agent_group"
	GroupTypeFieldAgent   = "field_agent_group"
)

// AdminGroup is a group as modelled by the newer /admin/groups API, which adds
// group types and the richer automatic-assignment settings.
type AdminGroup struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Type           string  `json:"type"`
	BusinessHourID int64   `json:"business_hour_id"`
	EscalateTo     int64   `json:"escalate_to"`
	UnassignedFor  string  `json:"unassigned_for"`
	AgentIDs       []int64 `json:"agent_ids"`
	// AutomaticAgentAssignment carries the omniroute assignment configuration.
	AutomaticAgentAssignment *AutomaticAgentAssignment `json:"automatic_agent_assignment,omitempty"`
	CreatedAt                Time                      `json:"created_at"`
	UpdatedAt                Time                      `json:"updated_at"`
}

// AutomaticAgentAssignment configures omniroute ticket assignment for a group.
type AutomaticAgentAssignment struct {
	Enabled          bool           `json:"enabled"`
	AssignmentType   int            `json:"assignment_type,omitempty"`
	AssignmentConfig map[string]any `json:"assignment_config,omitempty"`
}

// AdminGroupRequest is the create/update payload for an admin group.
type AdminGroupRequest struct {
	Name                     *string                   `json:"name,omitempty"`
	Description              *string                   `json:"description,omitempty"`
	Type                     *string                   `json:"type,omitempty"`
	EscalateTo               *int64                    `json:"escalate_to,omitempty"`
	UnassignedFor            *string                   `json:"unassigned_for,omitempty"`
	AgentIDs                 []int64                   `json:"agent_ids,omitempty"`
	BusinessHourID           *int64                    `json:"business_hour_id,omitempty"`
	AutomaticAgentAssignment *AutomaticAgentAssignment `json:"automatic_agent_assignment,omitempty"`

	// ClearEscalateTo sends "escalate_to": null.
	ClearEscalateTo bool `json:"-"`
	// ClearAgents sends an empty agent_ids array.
	ClearAgents bool `json:"-"`
}

// MarshalJSON applies the null/empty clearing semantics.
func (r AdminGroupRequest) MarshalJSON() ([]byte, error) {
	type alias AdminGroupRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}
	if r.ClearEscalateTo {
		m["escalate_to"] = nil
	}
	if r.ClearAgents {
		m["agent_ids"] = []int64{}
	}
	return marshalMap(m)
}

const adminGroupsPath = "admin/groups"

// GetAdminGroup fetches a group through the admin API.
func (c *Client) GetAdminGroup(ctx context.Context, id int64) (*AdminGroup, error) {
	return getResource[AdminGroup](ctx, c, adminGroupsPath, id, nil)
}

// ListAdminGroups returns every group through the admin API.
func (c *Client) ListAdminGroups(ctx context.Context, opts ListOptions) ([]AdminGroup, error) {
	return listAll[AdminGroup](ctx, c, adminGroupsPath, opts)
}

// CreateAdminGroup creates a group through the admin API.
func (c *Client) CreateAdminGroup(ctx context.Context, req AdminGroupRequest) (*AdminGroup, error) {
	return createResource[AdminGroup](ctx, c, adminGroupsPath, req)
}

// UpdateAdminGroup updates a group through the admin API.
func (c *Client) UpdateAdminGroup(ctx context.Context, id int64, req AdminGroupRequest) (*AdminGroup, error) {
	return updateResource[AdminGroup](ctx, c, adminGroupsPath, id, req)
}

// DeleteAdminGroup deletes a group through the admin API.
func (c *Client) DeleteAdminGroup(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, adminGroupsPath, id)
}

// GroupAgent is a group membership entry.
type GroupAgent struct {
	ID int64 `json:"id"`
	// Role distinguishes members from group leaders where the plan supports it.
	Role string `json:"role,omitempty"`
}

// ListAgentsInGroup lists the agents belonging to a group.
func (c *Client) ListAgentsInGroup(ctx context.Context, groupID int64) ([]GroupAgent, error) {
	var out []GroupAgent
	err := c.Get(ctx, pathFor(adminGroupsPath, groupID)+"/agents", nil, &out)
	return out, err
}

// UpdateGroupAgents adds or removes agents in a group. Agents named in add are
// added; agents named in remove are removed.
func (c *Client) UpdateGroupAgents(ctx context.Context, groupID int64, add, remove []int64) error {
	body := map[string]any{}
	if len(add) > 0 {
		body["add"] = add
	}
	if len(remove) > 0 {
		body["remove"] = remove
	}
	return c.Put(ctx, pathFor(adminGroupsPath, groupID)+"/agents", body, nil)
}
