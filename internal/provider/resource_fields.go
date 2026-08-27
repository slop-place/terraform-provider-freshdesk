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
	_ resource.Resource                = (*ticketFieldResource)(nil)
	_ resource.ResourceWithConfigure   = (*ticketFieldResource)(nil)
	_ resource.ResourceWithImportState = (*ticketFieldResource)(nil)
	_ resource.Resource                = (*contactFieldResource)(nil)
	_ resource.Resource                = (*companyFieldResource)(nil)
	_ resource.Resource                = (*ticketFieldSectionResource)(nil)
)

// choiceType is the object type of a dropdown choice, shared by the three
// field resources.
//
//nolint:gochecknoglobals // an immutable schema type
var choiceType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id":       types.Int64Type,
	"value":    types.StringType,
	"label":    types.StringType,
	"position": types.Int64Type,
}}

// choiceAttribute is the nested block describing a dropdown field's options.
func choiceAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Optional: true,
		MarkdownDescription: "Options of a dropdown field, in the order they are shown. Only " +
			"meaningful for the `custom_dropdown` and `nested_field` types.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Numeric identifier of the choice.",
				},
				"value": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Stored value of the choice.",
				},
				"label": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Label shown for the choice. Defaults to `value`.",
				},
				"position": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Position of the choice in the dropdown.",
				},
			},
		},
	}
}

// applyChoices writes a field's options into state, keeping the attribute null
// when Freshdesk answered with one of the built-in fields' map shapes rather
// than the array a custom dropdown uses.
func applyChoices(current types.List, raw freshdesk.RawJSON) types.List {
	choices := raw.Choices()
	if len(choices) == 0 && current.IsNull() {
		return current
	}

	return choicesToList(choices)
}

// choicesToList renders API choices as a Terraform list.
func choicesToList(choices []freshdesk.Choice) types.List {
	if choices == nil {
		return types.ListNull(choiceType)
	}

	elems := make([]attr.Value, 0, len(choices))
	for _, c := range choices {
		label := c.Label
		if label == "" {
			label = c.Value
		}
		elems = append(elems, types.ObjectValueMust(choiceType.AttrTypes, map[string]attr.Value{
			"id":       types.Int64Value(c.ID),
			"value":    types.StringValue(c.Value),
			"label":    types.StringValue(label),
			"position": types.Int64Value(int64(c.Position)),
		}))
	}

	return types.ListValueMust(choiceType, elems)
}

// resolveChoiceIDs fills in the API's identifier for every choice whose value
// already exists on the field.
//
// Terraform marks a nested computed attribute unknown as soon as the list
// around it changes, so a plan that edits the choices carries no IDs at all.
// Freshdesk treats an ID-less choice as a brand-new option and rejects the
// request as a duplicate, so the identifiers are recovered from the live field
// and matched by value.
func resolveChoiceIDs(existing freshdesk.RawJSON, planned []freshdesk.Choice) []freshdesk.Choice {
	current := existing.Choices()

	byValue := make(map[string]int64, len(current))
	for _, c := range current {
		byValue[c.Value] = c.ID
	}

	out := make([]freshdesk.Choice, 0, len(planned))
	for _, c := range planned {
		if c.ID == 0 {
			c.ID = byValue[c.Value]
		}
		out = append(out, c)
	}

	return out
}

// choiceModel mirrors one element of the choices list.
type choiceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Value    types.String `tfsdk:"value"`
	Label    types.String `tfsdk:"label"`
	Position types.Int64  `tfsdk:"position"`
}

// listToChoices reads a Terraform choices list into the API type.
func listToChoices(ctx context.Context, l types.List, diags *diagnostics) []freshdesk.Choice {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}

	var models []choiceModel
	diags.Append(l.ElementsAs(ctx, &models, false)...)

	if diags.HasError() {
		return nil
	}

	out := make([]freshdesk.Choice, 0, len(models))
	for i, m := range models {
		position := i + 1
		if !m.Position.IsNull() && !m.Position.IsUnknown() {
			position = int(m.Position.ValueInt64())
		}

		choice := freshdesk.Choice{
			Value:    m.Value.ValueString(),
			Label:    m.Label.ValueString(),
			Position: position,
		}
		// An existing choice must carry its ID, or Freshdesk treats it as a new
		// option and rejects the request as a duplicate value.
		if !m.ID.IsNull() && !m.ID.IsUnknown() {
			choice.ID = m.ID.ValueInt64()
		}
		out = append(out, choice)
	}

	return out
}

// --- ticket fields -------------------------------------------------------

type ticketFieldResource = crud[ticketFieldModel, freshdesk.TicketField, *ticketFieldModel]

type ticketFieldModel struct {
	ID                   types.String `tfsdk:"id"`
	Label                types.String `tfsdk:"label"`
	LabelForCustomers    types.String `tfsdk:"label_for_customers"`
	Type                 types.String `tfsdk:"type"`
	Description          types.String `tfsdk:"description"`
	Position             types.Int64  `tfsdk:"position"`
	CustomersCanEdit     types.Bool   `tfsdk:"customers_can_edit"`
	DisplayedToCustomers types.Bool   `tfsdk:"displayed_to_customers"`
	RequiredForCustomers types.Bool   `tfsdk:"required_for_customers"`
	RequiredForAgents    types.Bool   `tfsdk:"required_for_agents"`
	RequiredForClosure   types.Bool   `tfsdk:"required_for_closure"`
	Choices              types.List   `tfsdk:"choices"`

	Name      types.String `tfsdk:"name"`
	Default   types.Bool   `tfsdk:"default"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *ticketFieldModel) GetID() types.String { return m.ID }

// Apply copies an API ticket field into the model.
func (m *ticketFieldModel) Apply(f *freshdesk.TicketField) {
	m.ID = idString(f.ID)
	m.Label = types.StringValue(f.Label)
	m.LabelForCustomers = optString(f.LabelForCustomers)
	m.Type = types.StringValue(f.Type)
	m.Description = optString(f.Description)
	m.Position = types.Int64Value(int64(f.Position))
	m.CustomersCanEdit = types.BoolValue(f.CustomersCanEdit)
	m.DisplayedToCustomers = types.BoolValue(f.DisplayedToCustomers)
	m.RequiredForCustomers = types.BoolValue(f.RequiredForCustomers)
	m.RequiredForAgents = types.BoolValue(f.RequiredForAgents)
	m.RequiredForClosure = types.BoolValue(f.RequiredForClosure)
	m.Name = types.StringValue(f.Name)
	m.Default = types.BoolValue(f.Default)
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)

	m.Choices = applyChoices(m.Choices, f.Choices)
}

// NewTicketFieldResource returns the freshdesk_ticket_field resource.
func NewTicketFieldResource() resource.Resource {
	build := func(
		ctx context.Context,
		plan *ticketFieldModel,
		d *diagnostics,
		creating bool,
	) freshdesk.TicketFieldRequest {
		req := freshdesk.TicketFieldRequest{
			Label:                strPtr(plan.Label),
			LabelForCustomers:    strPtr(plan.LabelForCustomers),
			Description:          strPtr(plan.Description),
			Position:             intPtr(plan.Position),
			CustomersCanEdit:     boolPtr(plan.CustomersCanEdit),
			DisplayedToCustomers: boolPtr(plan.DisplayedToCustomers),
			RequiredForCustomers: boolPtr(plan.RequiredForCustomers),
			RequiredForAgents:    boolPtr(plan.RequiredForAgents),
			RequiredForClosure:   boolPtr(plan.RequiredForClosure),
			Choices:              listToChoices(ctx, plan.Choices, d),
		}

		// Freshdesk rejects a type on update, even an unchanged one, so it is
		// only sent when the field is being created.
		if creating {
			req.Type = strPtr(plan.Type)
		}

		return req
	}

	return &ticketFieldResource{
		name:  "ticket_field",
		label: "ticket field",
		schema: schema.Schema{
			MarkdownDescription: "A field on the ticket form.\n\n" +
				"~> Deleting a custom field permanently removes the data captured in it on " +
				"every existing ticket.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("field"),
				"label": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Label shown to agents.",
				},
				"label_for_customers": schema.StringAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Label shown in the customer portal. Required when " +
						"`displayed_to_customers` is `true`.",
				},
				"type": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Field type, for example `custom_text`, `custom_dropdown`, " +
						"`custom_checkbox`, `custom_date` or `nested_field`. Changing it forces a " +
						"new field.",
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Help text shown beneath the field.",
				},
				"position": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Position of the field on the form.",
				},
				"customers_can_edit": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether customers may edit the field in the portal.",
				},
				"displayed_to_customers": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the customer portal.",
				},
				"required_for_customers": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether customers must fill the field in.",
				},
				"required_for_agents": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether agents must fill the field in.",
				},
				"required_for_closure": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field must be set before a ticket is closed.",
				},
				"choices": choiceAttribute(),
				"name": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Internal name Freshdesk generated for the field.",
				},
				"default": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether this is a built-in field rather than a custom one.",
				},
			}, timestampAttributes("field")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *ticketFieldModel, d *diagnostics,
		) (*freshdesk.TicketField, error) {
			return c.CreateTicketField(ctx, build(ctx, plan, d, true))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *ticketFieldModel,
		) (*freshdesk.TicketField, error) {
			return c.GetTicketField(ctx, id, false)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *ticketFieldModel, d *diagnostics,
		) (*freshdesk.TicketField, error) {
			req := build(ctx, plan, d, false)
			if d.HasError() {
				// The diagnostics already carry the failure.
				return nil, nil
			}

			if len(req.Choices) > 0 {
				current, err := c.GetTicketField(ctx, id, false)
				if err != nil {
					return nil, fmt.Errorf("reading the field's current choices: %w", err)
				}
				req.Choices = resolveChoiceIDs(current.Choices, req.Choices)
			}

			return c.UpdateTicketField(ctx, id, req)
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *ticketFieldModel) error {
			return c.DeleteTicketField(ctx, id)
		},
	}
}

// --- contact fields ------------------------------------------------------

type contactFieldResource = crud[contactFieldModel, freshdesk.ContactField, *contactFieldModel]

type contactFieldModel struct {
	ID                    types.String `tfsdk:"id"`
	Label                 types.String `tfsdk:"label"`
	LabelForCustomers     types.String `tfsdk:"label_for_customers"`
	Type                  types.String `tfsdk:"type"`
	Position              types.Int64  `tfsdk:"position"`
	EditableInSignup      types.Bool   `tfsdk:"editable_in_signup"`
	RequiredForAgents     types.Bool   `tfsdk:"required_for_agents"`
	AgentsCanEdit         types.Bool   `tfsdk:"agents_can_edit"`
	DisplayedForAgents    types.Bool   `tfsdk:"displayed_for_agents"`
	QuickAddForAgent      types.Bool   `tfsdk:"quick_add_for_agent"`
	Unique                types.Bool   `tfsdk:"unique"`
	CustomersCanEdit      types.Bool   `tfsdk:"customers_can_edit"`
	RequiredForCustomers  types.Bool   `tfsdk:"required_for_customers"`
	DisplayedForCustomers types.Bool   `tfsdk:"displayed_for_customers"`
	Choices               types.List   `tfsdk:"choices"`

	Name      types.String `tfsdk:"name"`
	Default   types.Bool   `tfsdk:"default"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *contactFieldModel) GetID() types.String { return m.ID }

// Apply copies an API contact field into the model.
func (m *contactFieldModel) Apply(f *freshdesk.ContactField) {
	m.ID = idString(f.ID)
	m.Label = types.StringValue(f.Label)
	m.LabelForCustomers = optString(f.LabelForCustomers)
	m.Type = types.StringValue(f.Type)
	m.Position = types.Int64Value(int64(f.Position))
	m.EditableInSignup = types.BoolValue(f.EditableInSignup)
	m.RequiredForAgents = types.BoolValue(f.RequiredForAgents)
	m.AgentsCanEdit = types.BoolValue(f.AgentsCanEdit)
	m.DisplayedForAgents = types.BoolValue(f.DisplayedForAgents)
	m.QuickAddForAgent = types.BoolValue(f.QuickAddForAgent)
	m.Unique = types.BoolValue(f.Unique)
	m.CustomersCanEdit = types.BoolValue(f.CustomersCanEdit)
	m.RequiredForCustomers = types.BoolValue(f.RequiredForCustomers)
	m.DisplayedForCustomers = types.BoolValue(f.DisplayedForCustomers)
	m.Name = types.StringValue(f.Name)
	m.Default = types.BoolValue(f.Default)
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)

	m.Choices = applyChoices(m.Choices, f.Choices)
}

// NewContactFieldResource returns the freshdesk_contact_field resource.
func NewContactFieldResource() resource.Resource {
	build := func(ctx context.Context, plan *contactFieldModel, d *diagnostics) freshdesk.ContactFieldRequest {
		return freshdesk.ContactFieldRequest{
			Label:                 strPtr(plan.Label),
			LabelForCustomers:     strPtr(plan.LabelForCustomers),
			Type:                  strPtr(plan.Type),
			Position:              intPtr(plan.Position),
			EditableInSignup:      boolPtr(plan.EditableInSignup),
			RequiredForAgents:     boolPtr(plan.RequiredForAgents),
			AgentsCanEdit:         boolPtr(plan.AgentsCanEdit),
			DisplayedForAgents:    boolPtr(plan.DisplayedForAgents),
			QuickAddForAgent:      boolPtr(plan.QuickAddForAgent),
			Unique:                boolPtr(plan.Unique),
			CustomersCanEdit:      boolPtr(plan.CustomersCanEdit),
			RequiredForCustomers:  boolPtr(plan.RequiredForCustomers),
			DisplayedForCustomers: boolPtr(plan.DisplayedForCustomers),
			Choices:               listToChoices(ctx, plan.Choices, d),
		}
	}

	return &contactFieldResource{
		name:  "contact_field",
		label: "contact field",
		schema: schema.Schema{
			MarkdownDescription: "A field on the contact form.\n\n" +
				"-> `agents_can_edit`, `displayed_for_agents`, `quick_add_for_agent` and " +
				"`unique` require the newer Contacts and Companies experience.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("field"),
				"label": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Label shown to agents.",
				},
				"label_for_customers": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Label shown in the customer portal.",
				},
				"type": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Field type, for example `custom_text`, `custom_dropdown` " +
						"or `custom_date`. Changing it forces a new field.",
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"position": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Position of the field on the form.",
				},
				"editable_in_signup": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether customers may set the field when signing up.",
				},
				"required_for_agents": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether agents must fill the field in.",
				},
				"agents_can_edit": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether agents may edit the field.",
				},
				"displayed_for_agents": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the agent interface.",
				},
				"quick_add_for_agent": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the quick-add contact form.",
				},
				"unique": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Whether values must be unique across contacts, which " +
						"prevents duplicates.",
				},
				"customers_can_edit": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether customers may edit the field in the portal.",
				},
				"required_for_customers": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether customers must fill the field in.",
				},
				"displayed_for_customers": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the customer portal.",
				},
				"choices": choiceAttribute(),
				"name": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Internal name Freshdesk generated for the field.",
				},
				"default": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether this is a built-in field rather than a custom one.",
				},
			}, timestampAttributes("field")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *contactFieldModel, d *diagnostics,
		) (*freshdesk.ContactField, error) {
			return c.CreateContactField(ctx, build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *contactFieldModel,
		) (*freshdesk.ContactField, error) {
			return c.GetContactField(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *contactFieldModel, d *diagnostics,
		) (*freshdesk.ContactField, error) {
			req := build(ctx, plan, d)
			if d.HasError() {
				// The diagnostics already carry the failure.
				return nil, nil
			}

			if len(req.Choices) > 0 {
				current, err := c.GetContactField(ctx, id)
				if err != nil {
					return nil, fmt.Errorf("reading the field's current choices: %w", err)
				}
				req.Choices = resolveChoiceIDs(current.Choices, req.Choices)
			}

			return c.UpdateContactField(ctx, id, req)
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *contactFieldModel) error {
			return c.DeleteContactField(ctx, id)
		},
	}
}

// --- company fields ------------------------------------------------------

type companyFieldResource = crud[companyFieldModel, freshdesk.CompanyField, *companyFieldModel]

type companyFieldModel struct {
	ID                 types.String `tfsdk:"id"`
	Label              types.String `tfsdk:"label"`
	Type               types.String `tfsdk:"type"`
	Position           types.Int64  `tfsdk:"position"`
	RequiredForAgents  types.Bool   `tfsdk:"required_for_agents"`
	AgentsCanEdit      types.Bool   `tfsdk:"agents_can_edit"`
	DisplayedForAgents types.Bool   `tfsdk:"displayed_for_agents"`
	QuickAddForAgent   types.Bool   `tfsdk:"quick_add_for_agent"`
	Unique             types.Bool   `tfsdk:"unique"`
	Choices            types.List   `tfsdk:"choices"`

	Name      types.String `tfsdk:"name"`
	Default   types.Bool   `tfsdk:"default"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *companyFieldModel) GetID() types.String { return m.ID }

// Apply copies an API company field into the model.
func (m *companyFieldModel) Apply(f *freshdesk.CompanyField) {
	m.ID = idString(f.ID)
	m.Label = types.StringValue(f.Label)
	m.Type = types.StringValue(f.Type)
	m.Position = types.Int64Value(int64(f.Position))
	m.RequiredForAgents = types.BoolValue(f.RequiredForAgents)
	m.AgentsCanEdit = types.BoolValue(f.AgentsCanEdit)
	m.DisplayedForAgents = types.BoolValue(f.DisplayedForAgents)
	m.QuickAddForAgent = types.BoolValue(f.QuickAddForAgent)
	m.Unique = types.BoolValue(f.Unique)
	m.Name = types.StringValue(f.Name)
	m.Default = types.BoolValue(f.Default)
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)

	m.Choices = applyChoices(m.Choices, f.Choices)
}

// NewCompanyFieldResource returns the freshdesk_company_field resource.
func NewCompanyFieldResource() resource.Resource {
	build := func(ctx context.Context, plan *companyFieldModel, d *diagnostics) freshdesk.CompanyFieldRequest {
		return freshdesk.CompanyFieldRequest{
			Label:              strPtr(plan.Label),
			Type:               strPtr(plan.Type),
			Position:           intPtr(plan.Position),
			RequiredForAgents:  boolPtr(plan.RequiredForAgents),
			AgentsCanEdit:      boolPtr(plan.AgentsCanEdit),
			DisplayedForAgents: boolPtr(plan.DisplayedForAgents),
			QuickAddForAgent:   boolPtr(plan.QuickAddForAgent),
			Unique:             boolPtr(plan.Unique),
			Choices:            listToChoices(ctx, plan.Choices, d),
		}
	}

	return &companyFieldResource{
		name:  "company_field",
		label: "company field",
		schema: schema.Schema{
			MarkdownDescription: "A field on the company form.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("field"),
				"label": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Label shown to agents.",
				},
				"type": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Field type, for example `custom_text` or " +
						"`custom_dropdown`. Changing it forces a new field.",
					PlanModifiers: []planModifierString{requiresReplace()},
				},
				"position": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Position of the field on the form.",
				},
				"required_for_agents": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether agents must fill the field in.",
				},
				"agents_can_edit": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether agents may edit the field.",
				},
				"displayed_for_agents": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the agent interface.",
				},
				"quick_add_for_agent": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the field appears in the quick-add company form.",
				},
				"unique": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether values must be unique across companies.",
				},
				"choices": choiceAttribute(),
				"name": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Internal name Freshdesk generated for the field.",
				},
				"default": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether this is a built-in field rather than a custom one.",
				},
			}, timestampAttributes("field")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *companyFieldModel, d *diagnostics,
		) (*freshdesk.CompanyField, error) {
			return c.CreateCompanyField(ctx, build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *companyFieldModel,
		) (*freshdesk.CompanyField, error) {
			return c.GetCompanyField(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *companyFieldModel, d *diagnostics,
		) (*freshdesk.CompanyField, error) {
			req := build(ctx, plan, d)
			if d.HasError() {
				// The diagnostics already carry the failure.
				return nil, nil
			}

			if len(req.Choices) > 0 {
				current, err := c.GetCompanyField(ctx, id)
				if err != nil {
					return nil, fmt.Errorf("reading the field's current choices: %w", err)
				}
				req.Choices = resolveChoiceIDs(current.Choices, req.Choices)
			}

			return c.UpdateCompanyField(ctx, id, req)
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *companyFieldModel) error {
			return c.DeleteCompanyField(ctx, id)
		},
	}
}

// --- ticket field sections -----------------------------------------------

type ticketFieldSectionResource = crud[ticketFieldSectionModel, freshdesk.TicketSection, *ticketFieldSectionModel]

type ticketFieldSectionModel struct {
	ID             types.String `tfsdk:"id"`
	TicketFieldID  types.Int64  `tfsdk:"ticket_field_id"`
	Label          types.String `tfsdk:"label"`
	ChoiceIDs      types.Set    `tfsdk:"choice_ids"`
	TicketFieldIDs types.Set    `tfsdk:"ticket_field_ids"`
}

// GetID reports the identifier held in state.
func (m *ticketFieldSectionModel) GetID() types.String { return m.ID }

// Apply copies an API section into the model.
func (m *ticketFieldSectionModel) Apply(s *freshdesk.TicketSection) {
	m.ID = idString(s.ID)
	m.Label = types.StringValue(s.Label)

	if s.ParentTicketFieldID != 0 {
		m.TicketFieldID = types.Int64Value(s.ParentTicketFieldID)
	}

	m.ChoiceIDs = applyInt64Set(m.ChoiceIDs, s.ChoiceIDs)

	m.TicketFieldIDs = int64Set(sortedInt64(s.TicketFieldIDs))
}

// NewTicketFieldSectionResource returns the freshdesk_ticket_field_section
// resource.
func NewTicketFieldSectionResource() resource.Resource {
	build := func(ctx context.Context, plan *ticketFieldSectionModel, d *diagnostics) freshdesk.SectionRequest {
		return freshdesk.SectionRequest{
			Label:     strPtr(plan.Label),
			ChoiceIDs: toInt64Slice(ctx, plan.ChoiceIDs, d),
		}
	}

	return &ticketFieldSectionResource{
		name:  "ticket_field_section",
		label: "ticket field section",
		schema: schema.Schema{
			MarkdownDescription: "A dynamic section revealed when a ticket field takes particular " +
				"values, letting a form show extra fields only when they are relevant.",
			Attributes: map[string]schema.Attribute{
				"id": idAttribute("section"),
				"ticket_field_id": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "ID of the dropdown field whose choices reveal the " +
						"section. Changing it forces a new section.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"label": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the section.",
				},
				"choice_ids": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the parent field's choices that display the section.",
				},
				"ticket_field_ids": schema.SetAttribute{
					Computed:            true,
					ElementType:         types.Int64Type,
					MarkdownDescription: "IDs of the fields shown inside the section.",
				},
			},
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *ticketFieldSectionModel, d *diagnostics,
		) (*freshdesk.TicketSection, error) {
			return c.CreateSection(ctx, plan.TicketFieldID.ValueInt64(), build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, state *ticketFieldSectionModel,
		) (*freshdesk.TicketSection, error) {
			return c.GetSection(ctx, state.TicketFieldID.ValueInt64(), id)
		},
		updateFn: func(
			ctx context.Context,
			c *freshdesk.Client,
			id int64,
			plan, state *ticketFieldSectionModel,
			d *diagnostics,
		) (*freshdesk.TicketSection, error) {
			return c.UpdateSection(ctx, state.TicketFieldID.ValueInt64(), id, build(ctx, plan, d))
		},
		deleteFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, state *ticketFieldSectionModel,
		) error {
			return c.DeleteSection(ctx, state.TicketFieldID.ValueInt64(), id)
		},
		// A section is identified by its parent field as well as its own ID, so
		// it is imported as "<ticket_field_id>:<section_id>".
		importFn: importCompositeID(
			"freshdesk_ticket_field_section", []string{"ticket_field_id", "id"}),
	}
}
