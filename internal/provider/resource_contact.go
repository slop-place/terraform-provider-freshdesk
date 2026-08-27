package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*contactResource)(nil)
	_ resource.ResourceWithConfigure   = (*contactResource)(nil)
	_ resource.ResourceWithImportState = (*contactResource)(nil)
)

type contactResource = crud[contactModel, freshdesk.Contact, *contactModel]

type contactModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Email            emailValue   `tfsdk:"email"`
	Phone            types.String `tfsdk:"phone"`
	Mobile           types.String `tfsdk:"mobile"`
	TwitterID        types.String `tfsdk:"twitter_id"`
	UniqueExternalID types.String `tfsdk:"unique_external_id"`
	Address          types.String `tfsdk:"address"`
	Description      types.String `tfsdk:"description"`
	JobTitle         types.String `tfsdk:"job_title"`
	Language         types.String `tfsdk:"language"`
	TimeZone         types.String `tfsdk:"time_zone"`
	OtherEmails      types.Set    `tfsdk:"other_emails"`
	Tags             types.Set    `tfsdk:"tags"`
	CompanyID        types.Int64  `tfsdk:"company_id"`
	ViewAllTickets   types.Bool   `tfsdk:"view_all_tickets"`
	CustomFields     types.Map    `tfsdk:"custom_fields"`

	Active      types.Bool   `tfsdk:"active"`
	ContactType types.String `tfsdk:"contact_type"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *contactModel) GetID() types.String { return m.ID }

// Apply copies an API contact into the model.
func (m *contactModel) Apply(c *freshdesk.Contact) {
	m.ID = idString(c.ID)
	m.Name = types.StringValue(c.Name)
	m.Email = emailString(c.Email)
	m.Phone = optString(c.Phone)
	m.Mobile = optString(c.Mobile)
	m.TwitterID = optString(c.TwitterID)
	m.UniqueExternalID = optString(c.UniqueExternalID)
	m.Address = optString(c.Address)
	m.Description = optString(c.Description)
	m.JobTitle = optString(c.JobTitle)
	m.Language = optString(c.Language)
	m.TimeZone = optString(c.TimeZone)
	m.CompanyID = optInt64(c.CompanyID)
	m.ViewAllTickets = types.BoolValue(c.ViewAllTickets)
	m.CustomFields = applyMap(m.CustomFields, c.CustomFields)
	m.Active = types.BoolValue(c.Active)
	m.ContactType = optString(c.ContactType)
	m.CreatedAt = timeString(c.CreatedAt)
	m.UpdatedAt = timeString(c.UpdatedAt)

	m.OtherEmails = applyStringSet(m.OtherEmails, c.OtherEmails)
	m.Tags = applyStringSet(m.Tags, c.Tags)
}

func contactRequest(
	ctx context.Context,
	plan, prior *contactModel,
	diags *diagnostics,
) freshdesk.ContactRequest {
	req := freshdesk.ContactRequest{
		Name:             strPtr(plan.Name),
		Email:            emailPtr(plan.Email),
		Phone:            strPtr(plan.Phone),
		Mobile:           strPtr(plan.Mobile),
		TwitterID:        strPtr(plan.TwitterID),
		UniqueExternalID: strPtr(plan.UniqueExternalID),
		Address:          strPtr(plan.Address),
		Description:      strPtr(plan.Description),
		JobTitle:         strPtr(plan.JobTitle),
		Language:         strPtr(plan.Language),
		TimeZone:         strPtr(plan.TimeZone),
		OtherEmails:      toStringSlice(ctx, plan.OtherEmails, diags),
		Tags:             toStringSlice(ctx, plan.Tags, diags),
		CompanyID:        int64Ptr(plan.CompanyID),
		ViewAllTickets:   boolPtr(plan.ViewAllTickets),
		CustomFields:     mapToCustomFields(ctx, plan.CustomFields, diags),
	}
	if prior != nil {
		req.ClearTags = plan.Tags.IsNull() && !prior.Tags.IsNull()
		req.ClearCompany = plan.CompanyID.IsNull() && !prior.CompanyID.IsNull()
	}

	return req
}

// NewContactResource returns the freshdesk_contact resource.
func NewContactResource() resource.Resource {
	return &contactResource{
		name:  "contact",
		label: "contact",
		schema: schema.Schema{
			MarkdownDescription: "A contact who raises tickets.\n\n" +
				"~> Deleting this resource soft-deletes the contact. Freshdesk keeps it " +
				"recoverable for 30 days before removing it permanently.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("contact"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the contact.",
				},
				"email": schema.StringAttribute{
					Optional:   true,
					CustomType: email(),
					MarkdownDescription: "Primary email address. At least one of `email`, `phone`, " +
						"`mobile`, `twitter_id` or `unique_external_id` must be set.\n\n" +
						"-> Freshdesk stores addresses folded to lower case. This attribute " +
						"compares case-insensitively, so writing capitals does not produce a " +
						"perpetual diff.",
				},
				"phone": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Telephone number of the contact.",
				},
				"mobile": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Mobile number of the contact.",
				},
				"twitter_id": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Twitter handle of the contact.",
				},
				"unique_external_id": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Identifier of the contact in an external system.",
				},
				"address": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Postal address of the contact.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Free-text note about the contact.",
				},
				"job_title": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Job title of the contact.",
				},
				"language": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Language code of the contact, for example `en`.",
				},
				"time_zone": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Time zone of the contact. Defaults to the helpdesk's.",
				},
				"other_emails": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Additional email addresses for the contact.",
				},
				"tags": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Tags applied to the contact.",
				},
				"company_id": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "ID of the company the contact belongs to.",
				},
				"view_all_tickets": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Whether the contact can see every ticket raised by their " +
						"company, rather than only their own.",
				},
				"custom_fields": schema.MapAttribute{
					Optional:    true,
					ElementType: types.StringType,
					MarkdownDescription: "Custom field values keyed by field name. Non-string values " +
						"are given as JSON.",
				},
				"active": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the contact has activated their portal account.",
				},
				"contact_type": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Contact type reported by the API.",
				},
			}, timestampAttributes("contact")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *contactModel, d *diagnostics,
		) (*freshdesk.Contact, error) {
			return c.CreateContact(ctx, contactRequest(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *contactModel,
		) (*freshdesk.Contact, error) {
			return c.GetContact(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *contactModel, d *diagnostics,
		) (*freshdesk.Contact, error) {
			return c.UpdateContact(ctx, id, contactRequest(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *contactModel) error {
			return c.DeleteContact(ctx, id)
		},
	}
}
