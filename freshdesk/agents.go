package freshdesk

import (
	"context"
	"net/url"
)

// Agent ticket scopes.
const (
	AgentScopeGlobal     = 1
	AgentScopeGroup      = 2
	AgentScopeRestricted = 3
)

// Agent types.
const (
	AgentTypeSupport      = 1
	AgentTypeField        = 2
	AgentTypeCollaborator = 3
)

// Agent is a Freshdesk agent. Identity attributes (name, phone, job title) live
// on the embedded Contact and are read-only through this API.
type Agent struct {
	ID          int64  `json:"id"`
	Available   bool   `json:"available"`
	Occasional  bool   `json:"occasional"`
	Signature   string `json:"signature"`
	TicketScope int    `json:"ticket_scope"`
	// Type is the string form ("support_agent", "field_agent", "collaborator").
	Type                 string  `json:"type"`
	GroupIDs             []int64 `json:"group_ids"`
	RoleIDs              []int64 `json:"role_ids"`
	SkillIDs             []int64 `json:"skill_ids"`
	ContributionGroupIDs []int64 `json:"contribution_group_ids"`
	FocusMode            bool    `json:"focus_mode"`
	AvailableSince       Time    `json:"available_since"`
	// Deactivated marks an agent whose access has been withdrawn.
	Deactivated bool `json:"deactivated"`
	// APIKeyEnabled reports whether the agent may use the API.
	APIKeyEnabled bool `json:"api_key_enabled"`
	// AgentOperationalStatus is the omniroute availability state.
	AgentOperationalStatus any `json:"agent_operational_status"`
	// LastActiveAt is when the agent last did anything.
	LastActiveAt Time         `json:"last_active_at"`
	Contact      AgentContact `json:"contact"`
	CreatedAt    Time         `json:"created_at"`
	UpdatedAt    Time         `json:"updated_at"`
}

// AgentContact is the identity record embedded in an Agent.
type AgentContact struct {
	Active      bool   `json:"active"`
	Email       string `json:"email"`
	JobTitle    string `json:"job_title"`
	Language    string `json:"language"`
	LastLoginAt Time   `json:"last_login_at"`
	Mobile      string `json:"mobile"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	TimeZone    string `json:"time_zone"`
	CreatedAt   Time   `json:"created_at"`
	UpdatedAt   Time   `json:"updated_at"`
}

// AgentRequest is the create/update payload for an agent.
type AgentRequest struct {
	// Email is required on create.
	Email *string `json:"email,omitempty"`
	// Name is accepted on create only; Freshdesk rejects it on update.
	Name *string `json:"name,omitempty"`
	// TicketScope is required on create.
	TicketScope *int    `json:"ticket_scope,omitempty"`
	Occasional  *bool   `json:"occasional,omitempty"`
	Signature   *string `json:"signature,omitempty"`
	Language    *string `json:"language,omitempty"`
	TimeZone    *string `json:"time_zone,omitempty"`
	FocusMode   *bool   `json:"focus_mode,omitempty"`
	// AgentType is the numeric form used on writes.
	AgentType            *int    `json:"agent_type,omitempty"`
	RoleIDs              []int64 `json:"role_ids,omitempty"`
	GroupIDs             []int64 `json:"group_ids,omitempty"`
	SkillIDs             []int64 `json:"skill_ids,omitempty"`
	ContributionGroupIDs []int64 `json:"contribution_group_ids,omitempty"`

	// ClearGroups sends an empty group_ids array, removing every group.
	ClearGroups bool `json:"-"`
	// ClearSkills sends an empty skill_ids array.
	ClearSkills bool `json:"-"`
}

// MarshalJSON applies the empty-array clearing semantics.
func (r AgentRequest) MarshalJSON() ([]byte, error) {
	type alias AgentRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}
	if r.ClearGroups {
		m["group_ids"] = []int64{}
	}
	if r.ClearSkills {
		m["skill_ids"] = []int64{}
	}
	return marshalMap(m)
}

const agentsPath = "agents"

// GetAgent fetches an agent by ID.
func (c *Client) GetAgent(ctx context.Context, id int64) (*Agent, error) {
	return getResource[Agent](ctx, c, agentsPath, id, nil)
}

// Me returns the agent the API key authenticates as.
func (c *Client) Me(ctx context.Context) (*Agent, error) {
	var out Agent
	if err := c.Get(ctx, agentsPath+"/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentListOptions filters the agent collection endpoint.
type AgentListOptions struct {
	ListOptions

	Email  string
	Mobile string
	Phone  string
	// State is "fulltime" or "occasional".
	State string
}

// ListAgents returns every agent matching opts.
func (c *Client) ListAgents(ctx context.Context, opts AgentListOptions) ([]Agent, error) {
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
	if opts.State != "" {
		v.Set("state", opts.State)
	}
	base.Extra = v
	return listAll[Agent](ctx, c, agentsPath, base)
}

// CreateAgent creates an agent.
func (c *Client) CreateAgent(ctx context.Context, req AgentRequest) (*Agent, error) {
	return createResource[Agent](ctx, c, agentsPath, req)
}

// UpdateAgent applies a partial update to an agent.
func (c *Client) UpdateAgent(ctx context.Context, id int64, req AgentRequest) (*Agent, error) {
	return updateResource[Agent](ctx, c, agentsPath, id, req)
}

// DeleteAgent downgrades an agent to a contact.
func (c *Client) DeleteAgent(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, agentsPath, id)
}

// SearchAgents matches agents by name or email prefix.
func (c *Client) SearchAgents(ctx context.Context, term string) ([]Agent, error) {
	var out []Agent
	q := url.Values{"term": {term}}
	err := c.Get(ctx, agentsPath+"/autocomplete", q, &out)
	return out, err
}

// CreateAgents creates many agents in one asynchronous call.
func (c *Client) CreateAgents(ctx context.Context, reqs []AgentRequest) (*BulkJob, error) {
	var out BulkJob
	body := map[string]any{"agents": reqs}
	if err := c.Post(ctx, agentsPath+"/bulk", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentAvailability describes an agent's omniroute load settings.
type AgentAvailability struct {
	ID   int64 `json:"id"`
	User struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	Channels map[string]AgentChannelAvailability `json:"channels"`
}

// AgentChannelAvailability is the per-channel load configuration.
type AgentChannelAvailability struct {
	Status     string `json:"status,omitempty"`
	Load       int    `json:"load,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Assignment bool   `json:"assignment,omitempty"`
}

// GetAgentAvailability reads an agent's availability and load settings.
func (c *Client) GetAgentAvailability(ctx context.Context, id int64) (*AgentAvailability, error) {
	var out AgentAvailability
	if err := c.Get(ctx, pathFor(agentsPath, id)+"/availability", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAgentAvailability writes an agent's load settings.
func (c *Client) UpdateAgentAvailability(
	ctx context.Context,
	id int64,
	body map[string]any,
) (*AgentAvailability, error) {
	var out AgentAvailability
	if err := c.Put(ctx, pathFor(agentsPath, id)+"/availability", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAgentAvailability reads availability for every agent.
func (c *Client) ListAgentAvailability(ctx context.Context, opts ListOptions) ([]AgentAvailability, error) {
	v := opts.Values()
	v.Set("only", "availability")
	opts.Extra = v
	return listAll[AgentAvailability](ctx, c, agentsPath, opts)
}

// Role is a Freshdesk agent role. Roles are read-only through the API.
type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
	// AgentType is the agent kind the role applies to, matching AgentType*.
	AgentType int  `json:"agent_type"`
	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// GetRole fetches a role by ID.
func (c *Client) GetRole(ctx context.Context, id int64) (*Role, error) {
	return getResource[Role](ctx, c, "roles", id, nil)
}

// ListRoles returns every role.
func (c *Client) ListRoles(ctx context.Context, opts ListOptions) ([]Role, error) {
	return listAll[Role](ctx, c, "roles", opts)
}
