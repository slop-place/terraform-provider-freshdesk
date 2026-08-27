package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*slaPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*slaPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*slaPolicyResource)(nil)
	_ resource.Resource                = (*automationRuleResource)(nil)
)

// --- SLA policies --------------------------------------------------------

// slaTargetType is the object type of one priority's SLA targets.
//
//nolint:gochecknoglobals // an immutable schema type
var slaTargetType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"respond_within":      types.Int64Type,
	"resolve_within":      types.Int64Type,
	"next_respond_within": types.Int64Type,
	"business_hours":      types.BoolType,
	"escalation_enabled":  types.BoolType,
}}

type slaPolicyResource = crud[slaPolicyModel, freshdesk.SLAPolicy, *slaPolicyModel]

type slaPolicyModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Active      types.Bool   `tfsdk:"active"`
	Position    types.Int64  `tfsdk:"position"`
	// SLATarget maps "priority_1".."priority_4" to that priority's targets.
	SLATarget types.Map `tfsdk:"sla_target"`

	ApplicableToCompanyIDs  types.Set `tfsdk:"applicable_to_company_ids"`
	ApplicableToGroupIDs    types.Set `tfsdk:"applicable_to_group_ids"`
	ApplicableToProductIDs  types.Set `tfsdk:"applicable_to_product_ids"`
	ApplicableToSources     types.Set `tfsdk:"applicable_to_sources"`
	ApplicableToTicketTypes types.Set `tfsdk:"applicable_to_ticket_types"`

	// Escalation is JSON because its shape varies with the account's plan.
	Escalation jsonValueType `tfsdk:"escalation"`
	IsDefault  types.Bool    `tfsdk:"is_default"`
	CreatedAt  types.String  `tfsdk:"created_at"`
	UpdatedAt  types.String  `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *slaPolicyModel) GetID() types.String { return m.ID }

// Apply copies an API SLA policy into the model.
func (m *slaPolicyModel) Apply(p *freshdesk.SLAPolicy) {
	m.ID = idString(p.ID.Int64())
	m.Name = types.StringValue(p.Name)
	m.Description = optString(p.Description)
	m.Active = types.BoolValue(p.Active)
	m.Position = types.Int64Value(int64(p.Position))
	m.IsDefault = types.BoolValue(p.IsDefault)
	m.CreatedAt = timeString(p.CreatedAt)
	m.UpdatedAt = timeString(p.UpdatedAt)

	if p.SLATarget != nil {
		elems := make(map[string]attr.Value, len(p.SLATarget))
		for priority, target := range p.SLATarget {
			elems[priority] = types.ObjectValueMust(slaTargetType.AttrTypes, map[string]attr.Value{
				"respond_within":      types.Int64Value(int64(target.RespondWithin)),
				"resolve_within":      types.Int64Value(int64(target.ResolveWithin)),
				"next_respond_within": types.Int64Value(int64(target.NextRespondWithin)),
				"business_hours":      types.BoolValue(target.BusinessHours),
				"escalation_enabled":  types.BoolValue(target.EscalationEnabled),
			})
		}
		m.SLATarget = types.MapValueMust(slaTargetType, elems)
	}

	m.ApplicableToCompanyIDs = applyInt64Set(m.ApplicableToCompanyIDs, p.ApplicableTo.CompanyIDs)
	m.ApplicableToGroupIDs = applyInt64Set(m.ApplicableToGroupIDs, p.ApplicableTo.GroupIDs)
	m.ApplicableToProductIDs = applyInt64Set(m.ApplicableToProductIDs, p.ApplicableTo.ProductIDs)

	if p.ApplicableTo.Sources != nil {
		ids := make([]int64, 0, len(p.ApplicableTo.Sources))
		for _, s := range p.ApplicableTo.Sources {
			ids = append(ids, int64(s))
		}
		m.ApplicableToSources = int64Set(sortedInt64(ids))
	}

	m.ApplicableToTicketTypes = applyStringSet(
		m.ApplicableToTicketTypes, p.ApplicableTo.TicketTypes)

	if len(p.Escalation) > 0 {
		m.Escalation = jsonEncoded(p.Escalation)
	}
}

// slaTargetModel mirrors one entry of the sla_target map.
type slaTargetModel struct {
	RespondWithin     types.Int64 `tfsdk:"respond_within"`
	ResolveWithin     types.Int64 `tfsdk:"resolve_within"`
	NextRespondWithin types.Int64 `tfsdk:"next_respond_within"`
	BusinessHours     types.Bool  `tfsdk:"business_hours"`
	EscalationEnabled types.Bool  `tfsdk:"escalation_enabled"`
}

// NewSLAPolicyResource returns the freshdesk_sla_policy resource.
func NewSLAPolicyResource() resource.Resource {
	build := func(ctx context.Context, plan *slaPolicyModel, d *diagnostics) freshdesk.SLAPolicyRequest {
		req := freshdesk.SLAPolicyRequest{
			Name:        strPtr(plan.Name),
			Description: strPtr(plan.Description),
			Active:      boolPtr(plan.Active),
			Position:    intPtr(plan.Position),
		}

		if !plan.SLATarget.IsNull() && !plan.SLATarget.IsUnknown() {
			targets := map[string]slaTargetModel{}
			d.Append(plan.SLATarget.ElementsAs(ctx, &targets, false)...)

			if d.HasError() {
				return req
			}
			req.SLATarget = map[string]freshdesk.SLATarget{}

			for priority, t := range targets {
				req.SLATarget[priority] = freshdesk.SLATarget{
					RespondWithin:     int(t.RespondWithin.ValueInt64()),
					ResolveWithin:     int(t.ResolveWithin.ValueInt64()),
					NextRespondWithin: int(t.NextRespondWithin.ValueInt64()),
					BusinessHours:     t.BusinessHours.ValueBool(),
					EscalationEnabled: t.EscalationEnabled.ValueBool(),
				}
			}
		}

		applicable := freshdesk.SLAApplicableTo{
			CompanyIDs:  toInt64Slice(ctx, plan.ApplicableToCompanyIDs, d),
			GroupIDs:    toInt64Slice(ctx, plan.ApplicableToGroupIDs, d),
			ProductIDs:  toInt64Slice(ctx, plan.ApplicableToProductIDs, d),
			TicketTypes: toStringSlice(ctx, plan.ApplicableToTicketTypes, d),
		}

		for _, s := range toInt64Slice(ctx, plan.ApplicableToSources, d) {
			applicable.Sources = append(applicable.Sources, int(s))
		}

		if applicable.CompanyIDs != nil || applicable.GroupIDs != nil ||
			applicable.ProductIDs != nil || applicable.Sources != nil ||
			applicable.TicketTypes != nil {
			req.ApplicableTo = &applicable
		}

		if raw := jsonAttrPtr(plan.Escalation, d, "escalation"); raw != nil {
			obj, ok := raw.(map[string]any)
			if !ok {
				d.AddError("Invalid escalation", "The value must be a JSON object.")

				return req
			}
			req.Escalation = obj
		}

		return req
	}

	return &slaPolicyResource{
		name:  "sla_policy",
		label: "SLA policy",
		schema: schema.Schema{
			MarkdownDescription: "An SLA policy setting response and resolution targets for the " +
				"tickets it matches.\n\n" +
				"~> The Freshdesk API offers no way to delete an SLA policy. Destroying this " +
				"resource removes it from state and warns; the policy itself stays until it " +
				"is deleted in the portal. Set `active = false` to stop it applying.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("policy"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the policy.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the policy.",
				},
				"active": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the policy is in force.",
				},
				"position": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Evaluation order. The first matching policy wins, so " +
						"lower positions take precedence.",
				},
				"sla_target": schema.MapNestedAttribute{
					Optional: true,
					MarkdownDescription: "Targets keyed by priority: `priority_1` (low) through " +
						"`priority_4` (urgent). All durations are in seconds.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"respond_within": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Seconds allowed for the first response.",
							},
							"resolve_within": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Seconds allowed to resolve the ticket.",
							},
							"next_respond_within": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								MarkdownDescription: "Seconds allowed for each follow-up response. " +
									"Must be at least 30 when set.",
							},
							"business_hours": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								MarkdownDescription: "Whether the clock counts business hours only, " +
									"rather than calendar hours.",
							},
							"escalation_enabled": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether missing a target raises an escalation.",
							},
						},
					},
				},
				"applicable_to_company_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "Restrict the policy to tickets from these companies.",
				},
				"applicable_to_group_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "Restrict the policy to tickets in these groups.",
				},
				"applicable_to_product_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "Restrict the policy to tickets on these products.",
				},
				"applicable_to_sources": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "Restrict the policy to tickets from these source channels.",
				},
				"applicable_to_ticket_types": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Restrict the policy to these ticket types.",
				},
				"escalation": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					Computed:   true,
					MarkdownDescription: "Escalation recipients as a JSON object, keyed by the " +
						"target being missed.",
				},
				"is_default": schema.BoolAttribute{
					Computed: true,
					MarkdownDescription: "Whether this is the fallback policy applied when no " +
						"other matches.",
				},
			}, timestampAttributes("policy")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *slaPolicyModel, d *diagnostics,
		) (*freshdesk.SLAPolicy, error) {
			return c.CreateSLAPolicy(ctx, build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *slaPolicyModel,
		) (*freshdesk.SLAPolicy, error) {
			return c.GetSLAPolicy(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *slaPolicyModel, d *diagnostics,
		) (*freshdesk.SLAPolicy, error) {
			return c.UpdateSLAPolicy(ctx, id, build(ctx, plan, d))
		},
		deleteFn: deleteUnsupported[slaPolicyModel, freshdesk.SLAPolicy](
			"SLA policy", "from Admin > SLA Policies in the Freshdesk portal"),
	}
}

// --- automation rules ----------------------------------------------------

type automationRuleResource = crud[automationRuleModel, freshdesk.AutomationRule, *automationRuleModel]

type automationRuleModel struct {
	ID       types.String `tfsdk:"id"`
	RuleType types.Int64  `tfsdk:"rule_type"`
	Name     types.String `tfsdk:"name"`

	Description types.String `tfsdk:"description"`
	Position    types.Int64  `tfsdk:"position"`
	Active      types.Bool   `tfsdk:"active"`
	Operator    types.String `tfsdk:"operator"`

	// The rule body is expressed as JSON, because Freshdesk's condition,
	// event and action shapes vary by field type and by plan.
	Performer  jsonValueType `tfsdk:"performer"`
	Events     jsonValueType `tfsdk:"events"`
	Conditions jsonValueType `tfsdk:"conditions"`
	Actions    jsonValueType `tfsdk:"actions"`

	Outdated             types.Bool   `tfsdk:"outdated"`
	AffectedTicketsCount types.Int64  `tfsdk:"affected_tickets_count"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *automationRuleModel) GetID() types.String { return m.ID }

// Apply copies an API automation rule into the model.
func (m *automationRuleModel) Apply(r *freshdesk.AutomationRule) {
	m.ID = idString(r.ID)
	m.Name = types.StringValue(r.Name)
	m.Description = optString(r.Description)
	m.Position = types.Int64Value(int64(r.Position))
	m.Active = types.BoolValue(r.Active)
	m.Operator = optString(r.Operator)
	m.Outdated = types.BoolValue(r.Outdated)
	m.AffectedTicketsCount = types.Int64Value(int64(r.AffectedTicketsCount))
	m.CreatedAt = timeString(r.CreatedAt)
	m.UpdatedAt = timeString(r.UpdatedAt)

	if r.Performer != nil {
		m.Performer = jsonEncoded(r.Performer)
	}

	if r.Events != nil {
		m.Events = jsonEncoded(r.Events)
	}

	// A rule created without conditions still reads back with one empty set,
	// which would turn an unset attribute into a value and fail Terraform's
	// consistency check. Keep it unset unless the set carries real properties.
	if r.Conditions != nil && (!m.Conditions.IsNull() || !vacuousConditions(r.Conditions)) {
		m.Conditions = jsonEncoded(r.Conditions)
	}

	if r.Actions != nil {
		m.Actions = jsonEncoded(r.Actions)
	}
}

// automationRuleRequest maps a plan into the API payload. The rule body is
// carried as JSON, so each part is decoded and validated here.
func automationRuleRequest(
	plan *automationRuleModel,
	d *diagnostics,
) freshdesk.AutomationRuleRequest {
	req := freshdesk.AutomationRuleRequest{
		Name:        strPtr(plan.Name),
		Description: strPtr(plan.Description),
		Position:    intPtr(plan.Position),
		Active:      boolPtr(plan.Active),
		Operator:    strPtr(plan.Operator),
	}

	if raw := jsonAttrPtr(plan.Performer, d, "performer"); raw != nil {
		if obj, ok := raw.(map[string]any); ok {
			req.Performer = obj
		} else {
			d.AddError("Invalid performer", "The value must be a JSON object.")
		}
	}

	if raw := jsonAttrPtr(plan.Events, d, "events"); raw != nil {
		req.Events = decodeObjectList(raw, "events", d)
	}

	req.Actions = decodeAutomationActions(plan.Actions, d)
	req.Conditions = decodeAutomationConditions(plan.Conditions, d)

	return req
}

// vacuousConditions reports whether a condition set carries no actual rules.
// Freshdesk answers a rule that was created without conditions with a single
// named set whose properties are empty.
func vacuousConditions(sets []freshdesk.AutomationConditionSet) bool {
	for _, set := range sets {
		if len(set.Properties) > 0 {
			return false
		}
	}

	return true
}

// decodeAutomationActions reads the actions array from its JSON attribute.
// Every key is passed through verbatim — see freshdesk.AutomationAction.
func decodeAutomationActions(raw jsonValueType, d *diagnostics) []freshdesk.AutomationAction {
	decoded := jsonAttrPtr(raw, d, "actions")
	if decoded == nil {
		return nil
	}

	objects := decodeObjectList(decoded, "actions", d)

	actions := make([]freshdesk.AutomationAction, 0, len(objects))
	for _, obj := range objects {
		if _, ok := obj["field_name"].(string); !ok {
			d.AddError("Invalid action", "Each action must carry a string field_name.")
			continue
		}
		actions = append(actions, obj)
	}

	return actions
}

// decodeAutomationConditions reads the condition sets from their JSON attribute.
func decodeAutomationConditions(
	raw jsonValueType,
	d *diagnostics,
) []freshdesk.AutomationConditionSet {
	decoded := jsonAttrPtr(raw, d, "conditions")
	if decoded == nil {
		return nil
	}

	objects := decodeObjectList(decoded, "conditions", d)

	sets := make([]freshdesk.AutomationConditionSet, 0, len(objects))
	for _, obj := range objects {
		set := freshdesk.AutomationConditionSet{
			Properties: decodeObjectList(obj["properties"], "conditions.properties", d),
		}
		if name, ok := obj["name"].(string); ok {
			set.Name = name
		}

		if match, ok := obj["match_type"].(string); ok {
			set.MatchType = match
		}
		sets = append(sets, set)
	}

	return sets
}

// NewAutomationRuleResource returns the freshdesk_automation_rule resource.
func NewAutomationRuleResource() resource.Resource {
	build := automationRuleRequest

	return &automationRuleResource{
		name:  "automation_rule",
		label: "automation rule",
		schema: schema.Schema{
			MarkdownDescription: "An automation rule: a dispatcher (`rule_type = 1`), a " +
				"time-triggered rule (`3`), or an observer (`4`).\n\n" +
				"-> The rule body is given as JSON because Freshdesk's condition and action " +
				"shapes depend on the fields they reference. Build them with `jsonencode`.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("rule"),
				"rule_type": schema.Int64Attribute{
					Required:   true,
					Validators: []validatorInt64{automationTypeValidator()},
					MarkdownDescription: "Which automation the rule belongs to: `1` ticket " +
						"creation (dispatcher), `3` time triggers, `4` ticket updates (observer). " +
						"Changing it forces a new rule.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the rule.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the rule.",
				},
				"position": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Evaluation order within the automation.",
				},
				"active": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the rule is running.",
				},
				"operator": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "How the condition sets combine: `AND` or `OR`.",
				},
				"performer": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Who must act for an observer rule to fire, as a JSON " +
						"object, for example `jsonencode({ type = 1 })`.",
				},
				"events": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Field changes an observer rule watches, as a JSON array " +
						"of objects.",
				},
				"conditions": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Condition sets as a JSON array. Each element takes " +
						"`name` (`condition_set_1` or `condition_set_2`), `match_type` (`all` or " +
						"`any`) and a `properties` array.",
				},
				"actions": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Actions as a JSON array. Each element takes `field_name` " +
						"and `value`, plus an optional `email` object for email actions.",
				},
				"outdated": schema.BoolAttribute{
					Computed: true,
					MarkdownDescription: "Whether the rule references a field that no longer " +
						"exists, which stops it firing.",
				},
				"affected_tickets_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "How many tickets the rule has acted on.",
				},
			}, timestampAttributes("rule")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *automationRuleModel, d *diagnostics,
		) (*freshdesk.AutomationRule, error) {
			return c.CreateAutomationRule(ctx, int(plan.RuleType.ValueInt64()), build(plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, state *automationRuleModel,
		) (*freshdesk.AutomationRule, error) {
			return c.GetAutomationRule(ctx, int(state.RuleType.ValueInt64()), id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *automationRuleModel, d *diagnostics,
		) (*freshdesk.AutomationRule, error) {
			return c.UpdateAutomationRule(ctx, int(plan.RuleType.ValueInt64()), id, build(plan, d))
		},
		deleteFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, state *automationRuleModel,
		) error {
			return c.DeleteAutomationRule(ctx, int(state.RuleType.ValueInt64()), id)
		},
		// A rule is addressed by its automation type as well as its own ID.
		importFn: importCompositeID("freshdesk_automation_rule", []string{"rule_type", "id"}),
	}
}

// decodeObjectList converts a decoded JSON array into a list of objects,
// reporting a diagnostic when an element is not an object.
func decodeObjectList(raw any, attrName string, d *diagnostics) []map[string]any {
	if raw == nil {
		return nil
	}

	list, ok := raw.([]any)
	if !ok {
		d.AddError("Invalid "+attrName, "The value must be a JSON array.")

		return nil
	}

	out := make([]map[string]any, 0, len(list))
	for i, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			d.AddError("Invalid "+attrName,
				fmt.Sprintf("Element %d must be a JSON object.", i))

			return nil
		}
		out = append(out, obj)
	}

	return out
}
