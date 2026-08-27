package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*companyResource)(nil)
	_ resource.ResourceWithConfigure   = (*companyResource)(nil)
	_ resource.ResourceWithImportState = (*companyResource)(nil)
)

type companyResource = crud[companyModel, freshdesk.Company, *companyModel]

type companyModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Note         types.String `tfsdk:"note"`
	Domains      types.Set    `tfsdk:"domains"`
	HealthScore  types.String `tfsdk:"health_score"`
	AccountTier  types.String `tfsdk:"account_tier"`
	Industry     types.String `tfsdk:"industry"`
	RenewalDate  types.String `tfsdk:"renewal_date"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *companyModel) GetID() types.String { return m.ID }

// Apply copies an API company into the model.
func (m *companyModel) Apply(c *freshdesk.Company) {
	m.ID = idString(c.ID)
	m.Name = types.StringValue(c.Name)
	m.Description = optString(c.Description)
	m.Note = optString(c.Note)
	m.HealthScore = optString(c.HealthScore)
	m.AccountTier = optString(c.AccountTier)
	m.Industry = optString(c.Industry)
	m.CustomFields = applyMap(m.CustomFields, c.CustomFields)
	m.CreatedAt = timeString(c.CreatedAt)
	m.UpdatedAt = timeString(c.UpdatedAt)

	m.Domains = applyStringSet(m.Domains, c.Domains)

	// The API echoes the renewal date as a full timestamp; keep the date form
	// the practitioner configured so the plan does not churn.
	if !c.RenewalDate.IsZero() {
		m.RenewalDate = types.StringValue(c.RenewalDate.UTC().Format("2006-01-02"))
	} else {
		m.RenewalDate = types.StringNull()
	}
}

func companyRequest(
	ctx context.Context,
	plan, prior *companyModel,
	diags *diagnostics,
) freshdesk.CompanyRequest {
	req := freshdesk.CompanyRequest{
		Name:         strPtr(plan.Name),
		Description:  strPtr(plan.Description),
		Note:         strPtr(plan.Note),
		Domains:      toStringSlice(ctx, plan.Domains, diags),
		HealthScore:  strPtr(plan.HealthScore),
		AccountTier:  strPtr(plan.AccountTier),
		Industry:     strPtr(plan.Industry),
		RenewalDate:  strPtr(plan.RenewalDate),
		CustomFields: mapToCustomFields(ctx, plan.CustomFields, diags),
	}
	if prior != nil {
		req.ClearDomains = plan.Domains.IsNull() && !prior.Domains.IsNull()
		req.ClearRenewalDate = plan.RenewalDate.IsNull() && !prior.RenewalDate.IsNull()
		req.Description = clearableString(plan.Description, prior.Description)
		req.Note = clearableString(plan.Note, prior.Note)
	}

	return req
}

// NewCompanyResource returns the freshdesk_company resource.
func NewCompanyResource() resource.Resource {
	return &companyResource{
		name:  "company",
		label: "company",
		schema: schema.Schema{
			MarkdownDescription: "A company that contacts can belong to.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("company"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the company. Must be unique across the helpdesk.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the company.",
				},
				"note": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Free-text note about the company.",
				},
				"domains": schema.SetAttribute{
					Optional:    true,
					ElementType: types.StringType,
					MarkdownDescription: "Email domains belonging to the company. Contacts with a " +
						"matching address are associated automatically.",
				},
				"health_score": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Relationship strength. The allowed values are the choices " +
						"configured on the account's Health Score field.",
				},
				"account_tier": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Account tier. The allowed values are the choices " +
						"configured on the account's Account Tier field.",
				},
				"industry": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Industry. The allowed values are the choices configured " +
						"on the account's Industry field.",
				},
				"renewal_date": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Contract renewal date, as `YYYY-MM-DD`.",
				},
				"custom_fields": schema.MapAttribute{
					Optional:    true,
					ElementType: types.StringType,
					MarkdownDescription: "Custom field values keyed by field name. Non-string values " +
						"are given as JSON, for example `\"42\"` or `\"true\"`.",
				},
			}, timestampAttributes("company")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *companyModel, d *diagnostics,
		) (*freshdesk.Company, error) {
			return c.CreateCompany(ctx, companyRequest(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *companyModel,
		) (*freshdesk.Company, error) {
			return c.GetCompany(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, prior *companyModel, d *diagnostics,
		) (*freshdesk.Company, error) {
			return c.UpdateCompany(ctx, id, companyRequest(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *companyModel) error {
			return c.DeleteCompany(ctx, id)
		},
	}
}
