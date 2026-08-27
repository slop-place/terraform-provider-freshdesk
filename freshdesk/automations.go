package freshdesk

import (
	"context"
	"strconv"
)

// Automation rule type IDs, used as the first path segment of the rules API.
const (
	// AutomationTypeTicketCreation runs when a ticket is created.
	AutomationTypeTicketCreation = 1
	// AutomationTypeTimeTriggered runs on a schedule.
	AutomationTypeTimeTriggered = 3
	// AutomationTypeTicketUpdate runs when a ticket is updated.
	AutomationTypeTicketUpdate = 4
)

// AutomationRule is a supervisor, dispatcher, or observer rule.
type AutomationRule struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	Active      bool   `json:"active"`
	// Outdated marks a rule referencing fields that no longer exist.
	Outdated bool `json:"outdated"`
	// Operator joins the condition sets, "AND" or "OR".
	Operator string `json:"operator,omitempty"`
	// Performer identifies who must act for an update rule to fire.
	Performer map[string]any `json:"performer,omitempty"`
	// Events are the field changes an update rule watches.
	Events []map[string]any `json:"events,omitempty"`
	// Conditions are the named condition sets the rule matches on.
	Conditions []AutomationConditionSet `json:"conditions,omitempty"`
	// Actions are applied when the rule matches.
	Actions []AutomationAction `json:"actions,omitempty"`
	// Summary is a human-readable rendering, returned on reads.
	Summary map[string]any `json:"summary,omitempty"`

	AffectedTicketsCount int   `json:"affected_tickets_count,omitempty"`
	LastUpdatedBy        int64 `json:"last_updated_by,omitempty"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// AutomationConditionSet is one named group of conditions.
type AutomationConditionSet struct {
	// Name is "condition_set_1" or "condition_set_2".
	Name string `json:"name"`
	// MatchType is "all" or "any".
	MatchType  string           `json:"match_type"`
	Properties []map[string]any `json:"properties"`
}

// AutomationAction is one action applied by a rule. Action shapes vary
// wildly per field_name (webhook actions alone carry request_type, url,
// content_type, content_layout, content and custom_headers), so the map
// passes every key through verbatim instead of enumerating them.
type AutomationAction = map[string]any

// AutomationRuleRequest is the create/update payload for an automation rule.
type AutomationRuleRequest struct {
	Name        *string                  `json:"name,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Position    *int                     `json:"position,omitempty"`
	Active      *bool                    `json:"active,omitempty"`
	Operator    *string                  `json:"operator,omitempty"`
	Performer   map[string]any           `json:"performer,omitempty"`
	Events      []map[string]any         `json:"events,omitempty"`
	Conditions  []AutomationConditionSet `json:"conditions,omitempty"`
	Actions     []AutomationAction       `json:"actions,omitempty"`
}

// automationRulesPath builds the rules collection path for a rule type.
func automationRulesPath(automationTypeID int) string {
	return "automations/" + strconv.Itoa(automationTypeID) + "/rules"
}

// GetAutomationRule fetches one automation rule.
func (c *Client) GetAutomationRule(ctx context.Context, automationTypeID int, id int64) (*AutomationRule, error) {
	return getResource[AutomationRule](ctx, c, automationRulesPath(automationTypeID), id, nil)
}

// ListAutomationRules returns every rule of one automation type.
func (c *Client) ListAutomationRules(
	ctx context.Context,
	automationTypeID int,
	opts ListOptions,
) ([]AutomationRule, error) {
	return listAll[AutomationRule](ctx, c, automationRulesPath(automationTypeID), opts)
}

// CreateAutomationRule creates an automation rule.
func (c *Client) CreateAutomationRule(
	ctx context.Context,
	automationTypeID int,
	req AutomationRuleRequest,
) (*AutomationRule, error) {
	return createResource[AutomationRule](ctx, c, automationRulesPath(automationTypeID), req)
}

// UpdateAutomationRule applies a partial update to an automation rule.
func (c *Client) UpdateAutomationRule(
	ctx context.Context,
	automationTypeID int,
	id int64,
	req AutomationRuleRequest,
) (*AutomationRule, error) {
	return updateResource[AutomationRule](ctx, c, automationRulesPath(automationTypeID), id, req)
}

// DeleteAutomationRule deletes an automation rule.
func (c *Client) DeleteAutomationRule(ctx context.Context, automationTypeID int, id int64) error {
	return deleteResource(ctx, c, automationRulesPath(automationTypeID), id)
}

// ScenarioAutomation is a saved set of actions an agent can apply to a ticket.
// Scenarios are read-only through the API.
type ScenarioAutomation struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Actions     []ScenarioAction `json:"actions"`
	Private     bool             `json:"private"`
	CreatedAt   Time             `json:"created_at"`
	UpdatedAt   Time             `json:"updated_at"`
}

// ScenarioAction is one action within a scenario automation. As with
// AutomationAction, the shape varies by action type, so every key is kept.
type ScenarioAction = map[string]any

// ListScenarioAutomations returns every scenario automation.
func (c *Client) ListScenarioAutomations(ctx context.Context, opts ListOptions) ([]ScenarioAutomation, error) {
	return listAll[ScenarioAutomation](ctx, c, "scenario_automations", opts)
}

// GetScenarioAutomation fetches one scenario automation.
func (c *Client) GetScenarioAutomation(ctx context.Context, id int64) (*ScenarioAutomation, error) {
	return getResource[ScenarioAutomation](ctx, c, "scenario_automations", id, nil)
}
