package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*skillResource)(nil)
	_ resource.ResourceWithConfigure   = (*skillResource)(nil)
	_ resource.ResourceWithImportState = (*skillResource)(nil)
)

type skillResource = crud[skillModel, freshdesk.Skill, *skillModel]

type skillModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	MatchType types.String `tfsdk:"match_type"`
	AgentIDs  types.Set    `tfsdk:"agent_ids"`
	// Conditions is JSON because the shape varies by resource and field type.
	Conditions jsonValueType `tfsdk:"conditions"`
	Rank       types.Int64   `tfsdk:"rank"`
	CreatedAt  types.String  `tfsdk:"created_at"`
	UpdatedAt  types.String  `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *skillModel) GetID() types.String { return m.ID }

// Apply copies an API skill into the model.
func (m *skillModel) Apply(s *freshdesk.Skill) {
	m.ID = idString(s.ID)
	m.Name = types.StringValue(s.Name)
	m.MatchType = optString(s.MatchType)
	m.Rank = types.Int64Value(int64(s.Rank))
	m.CreatedAt = timeString(s.CreatedAt)
	m.UpdatedAt = timeString(s.UpdatedAt)

	ids := make([]int64, 0, len(s.Agents))
	for _, a := range s.Agents {
		ids = append(ids, a.ID)
	}
	m.AgentIDs = applyInt64Set(m.AgentIDs, ids)

	if s.Conditions != nil {
		m.Conditions = jsonEncoded(s.Conditions)
	}
}

func skillRequest(ctx context.Context, plan, prior *skillModel, diags *diagnostics) freshdesk.SkillRequest {
	req := freshdesk.SkillRequest{
		Name:      strPtr(plan.Name),
		MatchType: strPtr(plan.MatchType),
		Rank:      intPtr(plan.Rank),
	}

	for _, id := range toInt64Slice(ctx, plan.AgentIDs, diags) {
		req.Agents = append(req.Agents, freshdesk.SkillAgent{ID: id})
	}

	// The condition body is passed through as written, so both the classic and
	// the omniroute shapes reach the API unchanged.
	if raw := jsonAttrPtr(plan.Conditions, diags, "conditions"); raw != nil {
		if _, ok := raw.([]any); !ok {
			diags.AddError("Invalid conditions", "The value must be a JSON array.")

			return req
		}
		req.RawConditions = raw
	}

	if prior != nil {
		req.ClearAgents = plan.AgentIDs.IsNull() && !prior.AgentIDs.IsNull()
	}

	return req
}

// NewSkillResource returns the freshdesk_skill resource.
func NewSkillResource() resource.Resource {
	return &skillResource{
		name:  "skill",
		label: "skill",
		schema: schema.Schema{
			MarkdownDescription: "A routing skill that matches tickets to the agents who hold " +
				"it.\n\n" +
				"~> Freshdesk serves two different shapes here depending on whether the " +
				"account is on omniroute. Classic accounts take `match_type`, `agent_ids` and " +
				"a flat `conditions` array. Omniroute accounts reject all three and take " +
				"channel-scoped conditions instead:\n\n" +
				"```hcl\n" +
				"conditions = jsonencode([{\n" +
				"  channel = \"ticket\"\n" +
				"  channel_conditions = [{\n" +
				"    name       = \"condition_set_1\"\n" +
				"    match_type = \"all\"\n" +
				"    properties = [{\n" +
				"      resource_type = \"ticket\"\n" +
				"      field_name    = \"priority\"\n" +
				"      operator      = \"in\"\n" +
				"      value         = [4]\n" +
				"    }]\n" +
				"  }]\n" +
				"}])\n" +
				"```\n\n" +
				"Set only the attributes your account accepts; the provider omits the rest.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("skill"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the skill.",
				},
				"match_type": schema.StringAttribute{
					Optional:   true,
					Validators: []validatorString{matchTypeValidator()},
					MarkdownDescription: "Whether `all` or `any` of the conditions must match. " +
						"Classic accounts only; omniroute accounts reject this field and carry " +
						"the match type inside each condition set.",
				},
				"agent_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of the agents who hold the skill. Classic accounts " +
						"only; omniroute accounts reject this field and assign skills to agents " +
						"through `freshdesk_agent.skill_ids` instead.",
				},
				"conditions": schema.StringAttribute{
					Optional:   true,
					CustomType: jsonAttr(),
					MarkdownDescription: "Matching rules as a JSON array. On a classic account " +
						"each element takes `resource_type`, `field_name`, `operator` and " +
						"`value`, plus an optional `nested_fields` object for dependent fields. " +
						"On an omniroute account each element is channel-scoped; see the " +
						"example above.",
				},
				"rank": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Evaluation order of the skill. Lower ranks match first.",
				},
			}, timestampAttributes("skill")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *skillModel, d *diagnostics,
		) (*freshdesk.Skill, error) {
			return c.CreateSkill(ctx, skillRequest(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *skillModel,
		) (*freshdesk.Skill, error) {
			return c.GetSkill(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *skillModel, d *diagnostics,
		) (*freshdesk.Skill, error) {
			return c.UpdateSkill(ctx, id, skillRequest(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *skillModel) error {
			return c.DeleteSkill(ctx, id)
		},
	}
}
