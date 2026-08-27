package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*agentResource)(nil)
	_ resource.ResourceWithConfigure   = (*agentResource)(nil)
	_ resource.ResourceWithImportState = (*agentResource)(nil)
)

type agentResource = crud[agentModel, freshdesk.Agent, *agentModel]

type agentModel struct {
	ID                   types.String `tfsdk:"id"`
	Email                emailValue   `tfsdk:"email"`
	Name                 types.String `tfsdk:"name"`
	TicketScope          types.Int64  `tfsdk:"ticket_scope"`
	Occasional           types.Bool   `tfsdk:"occasional"`
	Signature            types.String `tfsdk:"signature"`
	Language             types.String `tfsdk:"language"`
	TimeZone             types.String `tfsdk:"time_zone"`
	FocusMode            types.Bool   `tfsdk:"focus_mode"`
	AgentType            types.Int64  `tfsdk:"agent_type"`
	RoleIDs              types.Set    `tfsdk:"role_ids"`
	GroupIDs             types.Set    `tfsdk:"group_ids"`
	SkillIDs             types.Set    `tfsdk:"skill_ids"`
	ContributionGroupIDs types.Set    `tfsdk:"contribution_group_ids"`

	Type        types.String `tfsdk:"type"`
	Available   types.Bool   `tfsdk:"available"`
	Active      types.Bool   `tfsdk:"active"`
	JobTitle    types.String `tfsdk:"job_title"`
	Mobile      types.String `tfsdk:"mobile"`
	Phone       types.String `tfsdk:"phone"`
	LastLoginAt types.String `tfsdk:"last_login_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *agentModel) GetID() types.String { return m.ID }

// Apply copies an API agent into the model.
func (m *agentModel) Apply(a *freshdesk.Agent) {
	m.ID = idString(a.ID)
	m.TicketScope = types.Int64Value(int64(a.TicketScope))
	m.Occasional = types.BoolValue(a.Occasional)
	m.Signature = optString(a.Signature)
	m.FocusMode = types.BoolValue(a.FocusMode)
	m.Type = optString(a.Type)
	m.Available = types.BoolValue(a.Available)

	m.RoleIDs = int64Set(sortedInt64(a.RoleIDs))
	m.GroupIDs = applyInt64Set(m.GroupIDs, a.GroupIDs)
	m.SkillIDs = applyInt64Set(m.SkillIDs, a.SkillIDs)
	m.ContributionGroupIDs = applyInt64Set(m.ContributionGroupIDs, a.ContributionGroupIDs)

	// Identity lives on the embedded contact and is read-only after creation.
	m.Email = emailString(a.Contact.Email)
	m.Name = optString(a.Contact.Name)
	m.Language = optString(a.Contact.Language)
	m.TimeZone = optString(a.Contact.TimeZone)
	m.Active = types.BoolValue(a.Contact.Active)
	m.JobTitle = optString(a.Contact.JobTitle)
	m.Mobile = optString(a.Contact.Mobile)
	m.Phone = optString(a.Contact.Phone)
	m.LastLoginAt = timeString(a.Contact.LastLoginAt)
	m.CreatedAt = timeString(a.CreatedAt)
	m.UpdatedAt = timeString(a.UpdatedAt)
}

func agentRequest(
	ctx context.Context,
	plan, prior *agentModel,
	diags *diagnostics,
	creating bool,
) freshdesk.AgentRequest {
	req := freshdesk.AgentRequest{
		TicketScope:          intPtr(plan.TicketScope),
		Occasional:           boolPtr(plan.Occasional),
		Signature:            strPtr(plan.Signature),
		Language:             strPtr(plan.Language),
		TimeZone:             strPtr(plan.TimeZone),
		FocusMode:            boolPtr(plan.FocusMode),
		AgentType:            intPtr(plan.AgentType),
		RoleIDs:              toInt64Slice(ctx, plan.RoleIDs, diags),
		GroupIDs:             toInt64Slice(ctx, plan.GroupIDs, diags),
		SkillIDs:             toInt64Slice(ctx, plan.SkillIDs, diags),
		ContributionGroupIDs: toInt64Slice(ctx, plan.ContributionGroupIDs, diags),
	}

	if creating {
		// Freshdesk accepts email and name only on create; sending them on an
		// update is rejected, so they are set once here.
		req.Email = emailPtr(plan.Email)
		req.Name = strPtr(plan.Name)
	}

	if prior != nil {
		req.ClearGroups = plan.GroupIDs.IsNull() && !prior.GroupIDs.IsNull()
		req.ClearSkills = plan.SkillIDs.IsNull() && !prior.SkillIDs.IsNull()
	}

	return req
}

// NewAgentResource returns the freshdesk_agent resource.
func NewAgentResource() resource.Resource {
	return &agentResource{
		name:  "agent",
		label: "agent",
		schema: schema.Schema{
			MarkdownDescription: "An agent who works on tickets.\n\n" +
				"~> Deleting this resource downgrades the agent to a contact rather than " +
				"removing the person from the helpdesk, which is what the Freshdesk API does.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("agent"),
				"email": schema.StringAttribute{
					Required:   true,
					CustomType: email(),
					MarkdownDescription: "Email address of the agent. Changing it forces a new " +
						"agent, because Freshdesk does not allow it to be updated through the " +
						"API.\n\n-> Freshdesk stores addresses folded to lower case. This " +
						"attribute compares case-insensitively.",
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"name": schema.StringAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Display name of the agent. Changing it forces a new agent: " +
						"Freshdesk only accepts a name when the agent is created, and it is " +
						"otherwise managed in the Neo Admin Center.",
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"ticket_scope": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "Ticket permission: `1` global access, `2` group access, " +
						"`3` restricted to assigned tickets.",
					Validators: []validatorInt64{agentScopeValidator()},
				},
				"occasional": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the agent is occasional rather than full-time.",
				},
				"signature": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Email signature of the agent, in HTML.",
				},
				"language": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Language code of the agent, for example `en`.",
				},
				"time_zone": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Time zone of the agent. Defaults to the helpdesk's.",
				},
				"focus_mode": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether focus mode is enabled for the agent.",
				},
				"agent_type": schema.Int64Attribute{
					Optional: true,
					MarkdownDescription: "Agent type: `1` support agent, `2` field agent, " +
						"`3` collaborator.",
					Validators:    []validatorInt64{agentTypeValidator()},
					PlanModifiers: []planmodifier.Int64{requiresReplaceInt64()},
				},
				"role_ids": schema.SetAttribute{
					Required:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the roles granted to the agent. At least one is required.",
				},
				"group_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the groups the agent belongs to.",
				},
				"skill_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the skills assigned to the agent.",
				},
				"contribution_group_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of groups the agent has view-only access to. Only " +
						"meaningful when `ticket_scope` is `2`.",
				},
				"type": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Agent type in string form, as reported by the API.",
				},
				"available": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the agent is accepting omniroute assignments.",
				},
				"active": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the agent has activated their account.",
				},
				"job_title": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Job title, managed in the Neo Admin Center.",
				},
				"mobile": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Mobile number, managed in the Neo Admin Center.",
				},
				"phone": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Phone number, managed in the Neo Admin Center.",
				},
				"last_login_at": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "When the agent last signed in (RFC 3339).",
				},
			}, timestampAttributes("agent")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *agentModel, d *diagnostics,
		) (*freshdesk.Agent, error) {
			return c.CreateAgent(ctx, agentRequest(ctx, plan, nil, d, true))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *agentModel,
		) (*freshdesk.Agent, error) {
			return c.GetAgent(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *agentModel, d *diagnostics,
		) (*freshdesk.Agent, error) {
			return c.UpdateAgent(ctx, id, agentRequest(ctx, plan, prior, d, false))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *agentModel) error {
			return c.DeleteAgent(ctx, id)
		},
	}
}
