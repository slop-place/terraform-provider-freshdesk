package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*adminGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*adminGroupResource)(nil)
	_ resource.ResourceWithImportState = (*adminGroupResource)(nil)
)

type adminGroupResource = crud[adminGroupModel, freshdesk.AdminGroup, *adminGroupModel]

type adminGroupModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Type           types.String `tfsdk:"type"`
	EscalateTo     types.Int64  `tfsdk:"escalate_to"`
	UnassignedFor  types.String `tfsdk:"unassigned_for"`
	AgentIDs       types.Set    `tfsdk:"agent_ids"`
	BusinessHourID types.Int64  `tfsdk:"business_hour_id"`
	// AutomaticAgentAssignment is JSON so the omniroute assignment options can
	// evolve without a provider release.
	AutomaticAgentAssignment jsonValueType `tfsdk:"automatic_agent_assignment"`
	CreatedAt                types.String  `tfsdk:"created_at"`
	UpdatedAt                types.String  `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *adminGroupModel) GetID() types.String { return m.ID }

// Apply copies an API admin group into the model.
func (m *adminGroupModel) Apply(g *freshdesk.AdminGroup) {
	m.ID = idString(g.ID)
	m.Name = types.StringValue(g.Name)
	m.Description = optString(g.Description)
	m.Type = optString(g.Type)
	m.EscalateTo = optInt64(g.EscalateTo)
	m.UnassignedFor = optString(g.UnassignedFor)
	m.BusinessHourID = optInt64(g.BusinessCalendarID)
	m.CreatedAt = timeString(g.CreatedAt)
	m.UpdatedAt = timeString(g.UpdatedAt)

	m.AgentIDs = applyInt64Set(m.AgentIDs, g.AgentIDs)

	if g.AutomaticAgentAssignment != nil {
		m.AutomaticAgentAssignment = jsonEncoded(g.AutomaticAgentAssignment)
	}
}

func adminGroupRequest(
	ctx context.Context,
	plan, prior *adminGroupModel,
	diags *diagnostics,
	creating bool,
) freshdesk.AdminGroupRequest {
	req := freshdesk.AdminGroupRequest{
		Name:               strPtr(plan.Name),
		Description:        strPtr(plan.Description),
		EscalateTo:         int64Ptr(plan.EscalateTo),
		UnassignedFor:      strPtr(plan.UnassignedFor),
		AgentIDs:           toInt64Slice(ctx, plan.AgentIDs, diags),
		BusinessCalendarID: int64Ptr(plan.BusinessHourID),
	}

	req.AutomaticAgentAssignment = decodeAgentAssignment(plan.AutomaticAgentAssignment, diags)

	// Freshdesk rejects a group type on update, even an unchanged one, so it
	// is only sent when the group is being created.
	if creating {
		req.Type = strPtr(plan.Type)
	}

	if prior != nil {
		req.ClearEscalateTo = plan.EscalateTo.IsNull() && !prior.EscalateTo.IsNull()
		req.ClearAgents = plan.AgentIDs.IsNull() && !prior.AgentIDs.IsNull()
	}

	return req
}

// decodeAgentAssignment reads the omniroute assignment settings from their
// JSON attribute and passes every key through, because the accepted options
// depend on the plan and on the assignment type.
func decodeAgentAssignment(
	raw jsonValueType,
	diags *diagnostics,
) *freshdesk.AutomaticAgentAssignment {
	decoded := jsonAttrPtr(raw, diags, "automatic_agent_assignment")
	if decoded == nil {
		return nil
	}

	obj, ok := decoded.(map[string]any)
	if !ok {
		diags.AddError("Invalid automatic_agent_assignment", "The value must be a JSON object.")

		return nil
	}

	return &obj
}

// NewAdminGroupResource returns the freshdesk_admin_group resource.
func NewAdminGroupResource() resource.Resource {
	return &adminGroupResource{
		name:  "admin_group",
		label: "admin group",
		schema: schema.Schema{
			MarkdownDescription: "An agent group managed through the admin API, which adds group " +
				"types and the omniroute automatic-assignment settings that `freshdesk_group` " +
				"cannot reach.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("group"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the group. Must be unique across the helpdesk.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the group.",
				},
				"type": schema.StringAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Group type: `support_agent_group` or " +
						"`field_agent_group`.\n\n~> Freshdesk accepts a type only when the group " +
						"is created, so changing it forces a new group.",
					Validators:    []validatorString{groupTypeValidator()},
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"escalate_to": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the agent emailed when a ticket stays unassigned. " +
						"Freshdesk defaults this to the account administrator.",
				},
				"unassigned_for": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorString{unassignedForValidator()},
					MarkdownDescription: "How long a ticket may stay unassigned before escalation. " +
						"One of `30m`, `1h`, `2h`, `4h`, `8h`, `12h`, `1d`, `2d`, `3d`.",
				},
				"agent_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the agents who belong to the group.",
				},
				"business_hour_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the business-hours calendar the group follows. " +
						"Defaults to the account's default calendar.\n\n-> The admin API calls " +
						"this field `business_calendar_id`; the attribute keeps the name " +
						"`business_hour_id` so it matches `freshdesk_group`.",
				},
				"automatic_agent_assignment": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Omniroute assignment settings as a JSON object, for " +
						"example `jsonencode({ enabled = true, assignment_type = 1 })`.",
				},
			}, timestampAttributes("group")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *adminGroupModel, d *diagnostics,
		) (*freshdesk.AdminGroup, error) {
			return c.CreateAdminGroup(ctx, adminGroupRequest(ctx, plan, nil, d, true))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *adminGroupModel,
		) (*freshdesk.AdminGroup, error) {
			return c.GetAdminGroup(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *adminGroupModel, d *diagnostics,
		) (*freshdesk.AdminGroup, error) {
			return c.UpdateAdminGroup(ctx, id, adminGroupRequest(ctx, plan, prior, d, false))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *adminGroupModel) error {
			return c.DeleteAdminGroup(ctx, id)
		},
	}
}
