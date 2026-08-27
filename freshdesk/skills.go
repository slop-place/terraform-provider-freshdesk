package freshdesk

import "context"

// Skill is a routing skill used to match tickets to agents.
type Skill struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// MatchType is "all" or "any". Omniroute accounts do not use it.
	MatchType string       `json:"match_type"`
	Agents    []SkillAgent `json:"agents"`
	// Conditions is left untyped because Freshdesk serves two different
	// shapes: a flat list of SkillCondition on classic accounts, and
	// channel-scoped groups on omniroute ones. Both decode as objects.
	Conditions []map[string]any `json:"conditions"`
	// Rank orders skills; lower ranks are evaluated first.
	Rank      int  `json:"rank"`
	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// SkillAgent references an agent assigned to a skill.
type SkillAgent struct {
	ID int64 `json:"id"`
}

// SkillCondition is one matching rule on a classic (non-omniroute) skill. It
// is provided for callers assembling a request; reads come back untyped in
// Skill.Conditions, because omniroute accounts use a different shape.
type SkillCondition struct {
	// ResourceType is "ticket", "contact", or "company".
	ResourceType string `json:"resource_type"`
	FieldName    string `json:"field_name"`
	// Operator is "is"/"is_not" for dependent fields, "in"/"not_in" otherwise.
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
	// NestedFields carries dependent-field levels, keyed "level2", "level3".
	NestedFields map[string]SkillNestedField `json:"nested_fields,omitempty"`
}

// SkillNestedField is one level of a dependent-field condition.
type SkillNestedField struct {
	FieldName string `json:"field_name"`
	Operator  string `json:"operator,omitempty"`
	Value     any    `json:"value,omitempty"`
}

// SkillRequest is the create/update payload for a skill.
//
// Freshdesk serves two shapes here. Classic accounts take MatchType, Agents and
// a flat Conditions array; omniroute accounts reject the first two and expect
// channel-scoped conditions. RawConditions carries whichever shape the caller
// needs through untouched.
type SkillRequest struct {
	Name      *string      `json:"name,omitempty"`
	MatchType *string      `json:"match_type,omitempty"`
	Agents    []SkillAgent `json:"agents,omitempty"`
	// Conditions is the classic, typed condition list.
	Conditions []SkillCondition `json:"-"`
	// RawConditions overrides Conditions when set, and is passed through as
	// given. Use it for the omniroute shape.
	RawConditions any  `json:"-"`
	Rank          *int `json:"rank,omitempty"`

	// ClearAgents sends an empty agents array, unassigning every agent.
	ClearAgents bool `json:"-"`
}

// MarshalJSON applies the empty-array clearing semantics and picks whichever
// condition representation the caller supplied.
func (r SkillRequest) MarshalJSON() ([]byte, error) {
	type alias SkillRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}

	if r.ClearAgents {
		m["agents"] = []SkillAgent{}
	}

	switch {
	case r.RawConditions != nil:
		m["conditions"] = r.RawConditions
	case r.Conditions != nil:
		m["conditions"] = r.Conditions
	}

	return marshalMap(m)
}

const skillsPath = "admin/skills"

// GetSkill fetches a skill by ID.
func (c *Client) GetSkill(ctx context.Context, id int64) (*Skill, error) {
	return getResource[Skill](ctx, c, skillsPath, id, nil)
}

// ListSkills returns every skill.
func (c *Client) ListSkills(ctx context.Context, opts ListOptions) ([]Skill, error) {
	return listAll[Skill](ctx, c, skillsPath, opts)
}

// CreateSkill creates a skill.
func (c *Client) CreateSkill(ctx context.Context, req SkillRequest) (*Skill, error) {
	return createResource[Skill](ctx, c, skillsPath, req)
}

// UpdateSkill applies a partial update to a skill.
//
// This endpoint takes PATCH rather than the PUT the rest of the API uses; a
// PUT is answered with 405.
func (c *Client) UpdateSkill(ctx context.Context, id int64, req SkillRequest) (*Skill, error) {
	return patchResource[Skill](ctx, c, skillsPath, id, req)
}

// DeleteSkill deletes a skill.
func (c *Client) DeleteSkill(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, skillsPath, id)
}

// legacySkillsPath is the pre-omniroute skills endpoint. Accounts migrated to
// the newer route answer 404 here; accounts still on the old one need it.
const legacySkillsPath = "skills"

// GetSkillLegacy fetches a skill through the legacy omni route.
func (c *Client) GetSkillLegacy(ctx context.Context, id int64) (*Skill, error) {
	return getResource[Skill](ctx, c, legacySkillsPath, id, nil)
}

// ListSkillsLegacy returns every skill through the legacy omni route.
func (c *Client) ListSkillsLegacy(ctx context.Context, opts ListOptions) ([]Skill, error) {
	return listAll[Skill](ctx, c, legacySkillsPath, opts)
}

// CreateSkillLegacy creates a skill through the legacy omni route.
func (c *Client) CreateSkillLegacy(ctx context.Context, req SkillRequest) (*Skill, error) {
	return createResource[Skill](ctx, c, legacySkillsPath, req)
}

// UpdateSkillLegacy updates a skill through the legacy omni route.
func (c *Client) UpdateSkillLegacy(ctx context.Context, id int64, req SkillRequest) (*Skill, error) {
	return patchResource[Skill](ctx, c, legacySkillsPath, id, req)
}

// DeleteSkillLegacy deletes a skill through the legacy omni route.
func (c *Client) DeleteSkillLegacy(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, legacySkillsPath, id)
}
