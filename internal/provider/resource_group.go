package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*groupResource)(nil)
	_ resource.ResourceWithConfigure   = (*groupResource)(nil)
	_ resource.ResourceWithImportState = (*groupResource)(nil)
)

type groupResource = crud[groupModel, freshdesk.Group, *groupModel]

type groupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	EscalateTo       types.Int64  `tfsdk:"escalate_to"`
	UnassignedFor    types.String `tfsdk:"unassigned_for"`
	AgentIDs         types.Set    `tfsdk:"agent_ids"`
	AutoTicketAssign types.Int64  `tfsdk:"auto_ticket_assign"`
	BusinessHourID   types.Int64  `tfsdk:"business_hour_id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *groupModel) GetID() types.String { return m.ID }

// Apply copies an API group into the model.
func (m *groupModel) Apply(g *freshdesk.Group) {
	m.ID = idString(g.ID)
	m.Name = types.StringValue(g.Name)
	m.Description = optString(g.Description)
	m.EscalateTo = optInt64(g.EscalateTo)
	m.UnassignedFor = optString(g.UnassignedFor)
	m.AutoTicketAssign = types.Int64Value(int64(g.AutoTicketAssign))
	m.BusinessHourID = optInt64(g.BusinessHourID)
	m.CreatedAt = timeString(g.CreatedAt)
	m.UpdatedAt = timeString(g.UpdatedAt)
	// Freshdesk omits agent_ids from list responses; keep the configured value
	// rather than blanking it when the API does not report one.
	m.AgentIDs = applyInt64Set(m.AgentIDs, g.AgentIDs)
}

// groupRequest maps a plan into the API payload. prior is the state being
// replaced on update, or nil on create; it detects attributes the practitioner
// removed, which Freshdesk only clears when told to explicitly.
func groupRequest(ctx context.Context, plan, prior *groupModel, diags *diagnostics) freshdesk.GroupRequest {
	req := freshdesk.GroupRequest{
		Name:             strPtr(plan.Name),
		Description:      strPtr(plan.Description),
		EscalateTo:       int64Ptr(plan.EscalateTo),
		UnassignedFor:    strPtr(plan.UnassignedFor),
		AgentIDs:         toInt64Slice(ctx, plan.AgentIDs, diags),
		AutoTicketAssign: intPtr(plan.AutoTicketAssign),
		BusinessHourID:   int64Ptr(plan.BusinessHourID),
	}
	if prior != nil {
		req.ClearEscalateTo = plan.EscalateTo.IsNull() && !prior.EscalateTo.IsNull()
		req.ClearAgents = plan.AgentIDs.IsNull() && !prior.AgentIDs.IsNull()
	}

	return req
}

// NewGroupResource returns the freshdesk_group resource.
func NewGroupResource() resource.Resource {
	return &groupResource{
		name:  "group",
		label: "group",
		schema: schema.Schema{
			MarkdownDescription: "An agent group that tickets can be assigned to.",
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
				"escalate_to": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the agent emailed when a ticket in this group stays " +
						"unassigned past `unassigned_for`. Freshdesk defaults this to the account " +
						"administrator.",
				},
				"unassigned_for": schema.StringAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "How long a ticket may stay unassigned before the escalation " +
						"email is sent. One of `30m`, `1h`, `2h`, `4h`, `8h`, `12h`, `1d`, `2d`, `3d`.",
					Validators: []validatorString{unassignedForValidator()},
				},
				"agent_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the agents who belong to the group.",
				},
				"auto_ticket_assign": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Automatic ticket assignment: `0` disables it, `1` enables " +
						"round-robin. Richer assignment types require `freshdesk_admin_group`.",
				},
				"business_hour_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the business-hours calendar the group follows. " +
						"Defaults to the account's default calendar.",
				},
			}, timestampAttributes("group")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *groupModel, d *diagnostics,
		) (*freshdesk.Group, error) {
			return c.CreateGroup(ctx, groupRequest(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *groupModel,
		) (*freshdesk.Group, error) {
			return c.GetGroup(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *groupModel, d *diagnostics,
		) (*freshdesk.Group, error) {
			return c.UpdateGroup(ctx, id, groupRequest(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *groupModel) error {
			return c.DeleteGroup(ctx, id)
		},
	}
}
