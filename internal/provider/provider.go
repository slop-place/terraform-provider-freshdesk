// Package provider implements the Terraform provider for Freshdesk.
package provider

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// Ensure the provider satisfies the framework interfaces.
var (
	_ provider.Provider = (*freshdeskProvider)(nil)
)

// freshdeskProvider is the Terraform provider implementation.
type freshdeskProvider struct {
	// version is stamped at build time and reported to Terraform.
	version string
}

// New returns a provider factory for the given version string.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &freshdeskProvider{version: version}
	}
}

// providerModel maps the provider configuration block.
type providerModel struct {
	Domain     types.String `tfsdk:"domain"`
	APIKey     types.String `tfsdk:"api_key"`
	UserAgent  types.String `tfsdk:"user_agent"`
	MaxRetries types.Int64  `tfsdk:"max_retries"`
	Timeout    types.Int64  `tfsdk:"timeout_seconds"`
}

func (p *freshdeskProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "freshdesk"
	resp.Version = p.version
}

func (p *freshdeskProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a [Freshdesk](https://freshdesk.com) helpdesk: agents, groups, " +
			"ticket fields, SLA policies, automations, solutions, forums, and more.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Freshdesk domain. Accepts a bare subdomain (`acme`) or a full " +
					"host (`acme.freshdesk.com`). May also be set with the `FRESHDESK_DOMAIN` " +
					"environment variable (`FRESHDESK_URL` is accepted as an alias).",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: "Freshdesk API key, found under Profile Settings in the agent " +
					"portal. May also be set with the `FRESHDESK_API_KEY` environment variable.",
			},
			"user_agent": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Value sent in the `User-Agent` header. Defaults to " +
					"`terraform-provider-freshdesk`.",
			},
			"max_retries": schema.Int64Attribute{
				Optional: true,
				MarkdownDescription: "How many times to retry a throttled (HTTP 429) or transient " +
					"(HTTP 5xx) response. Defaults to `5`. Set to `-1` to disable retries.",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Per-request timeout in seconds. Defaults to `60`.",
			},
		},
	}
}

func (p *freshdeskProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unknown values mean another resource must be applied first; Terraform
	// will call Configure again once they are resolved.
	if cfg.Domain.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("domain"),
			"Freshdesk domain is not known at plan time",
			"The provider cannot be configured from a value that is only known after apply. "+
				"Set the domain to a literal, or use the FRESHDESK_DOMAIN environment variable.")
	}
	if cfg.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"),
			"Freshdesk API key is not known at plan time",
			"The provider cannot be configured from a value that is only known after apply. "+
				"Set the API key to a literal, or use the FRESHDESK_API_KEY environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration wins over the environment.
	domain := firstNonEmpty(cfg.Domain.ValueString(),
		os.Getenv("FRESHDESK_DOMAIN"), os.Getenv("FRESHDESK_URL"))
	apiKey := firstNonEmpty(cfg.APIKey.ValueString(), os.Getenv("FRESHDESK_API_KEY"))

	if domain == "" {
		resp.Diagnostics.AddAttributeError(path.Root("domain"),
			"Missing Freshdesk domain",
			"Set the `domain` argument on the provider, or the FRESHDESK_DOMAIN environment variable.")
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"),
			"Missing Freshdesk API key",
			"Set the `api_key` argument on the provider, or the FRESHDESK_API_KEY environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	clientCfg := freshdesk.Config{
		Domain:    domain,
		APIKey:    apiKey,
		UserAgent: cfg.UserAgent.ValueString(),
	}
	if !cfg.MaxRetries.IsNull() {
		clientCfg.MaxRetries = int(cfg.MaxRetries.ValueInt64())
	}
	if !cfg.Timeout.IsNull() && cfg.Timeout.ValueInt64() > 0 {
		clientCfg.Timeout = time.Duration(cfg.Timeout.ValueInt64()) * time.Second
	}
	if clientCfg.UserAgent == "" {
		clientCfg.UserAgent = freshdesk.DefaultUserAgent + "/" + p.version
	}

	client, err := freshdesk.New(clientCfg)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the Freshdesk client", err.Error())
		return
	}

	tflog.Debug(ctx, "configured freshdesk client", map[string]any{
		"base_url":    client.BaseURL(),
		"max_retries": strconv.Itoa(clientCfg.MaxRetries),
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

// firstNonEmpty returns the first argument that is not the empty string.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (p *freshdeskProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewAdminGroupResource,
		NewAgentResource,
		NewSkillResource,
		NewContactResource,
		NewCompanyResource,
		NewTicketResource,
		NewTicketFieldResource,
		NewTicketFieldSectionResource,
		NewContactFieldResource,
		NewCompanyFieldResource,
		NewTicketFormResource,
		NewSLAPolicyResource,
		NewAutomationRuleResource,
		NewCannedResponseFolderResource,
		NewCannedResponseResource,
		NewSolutionCategoryResource,
		NewSolutionFolderResource,
		NewSolutionArticleResource,
		NewForumCategoryResource,
		NewForumResource,
		NewTopicResource,
		NewCommentResource,
		NewEmailMailboxResource,
		NewTimeEntryResource,
		NewEmailSettingsResource,
		NewNotificationBCCResource,
		NewCustomObjectRecordResource,
	}
}

func (p *freshdeskProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
		NewHelpdeskSettingsDataSource,
		NewGroupDataSource,
		NewGroupsDataSource,
		NewAgentDataSource,
		NewAgentsDataSource,
		NewRoleDataSource,
		NewRolesDataSource,
		NewSkillDataSource,
		NewSkillsDataSource,
		NewProductDataSource,
		NewProductsDataSource,
		NewBusinessHoursDataSource,
		NewBusinessHoursListDataSource,
		NewEmailConfigDataSource,
		NewEmailConfigsDataSource,
		NewCompanyDataSource,
		NewCompaniesDataSource,
		NewContactDataSource,
		NewContactsDataSource,
		NewTicketDataSource,
		NewTicketsDataSource,
		NewTicketFieldDataSource,
		NewTicketFieldsDataSource,
		NewContactFieldsDataSource,
		NewCompanyFieldsDataSource,
		NewTicketFormDataSource,
		NewTicketFormsDataSource,
		NewSLAPolicyDataSource,
		NewSLAPoliciesDataSource,
		NewAutomationRulesDataSource,
		NewScenarioAutomationsDataSource,
		NewCannedResponseFoldersDataSource,
		NewSolutionCategoryDataSource,
		NewSolutionCategoriesDataSource,
		NewSolutionFolderDataSource,
		NewSolutionArticleDataSource,
		NewForumCategoriesDataSource,
		NewSurveysDataSource,
		NewSatisfactionRatingsDataSource,
		NewTimeEntriesDataSource,
		NewEmailMailboxDataSource,
		NewEmailMailboxesDataSource,
		NewEmailSettingsDataSource,
		NewCustomObjectSchemasDataSource,
	}
}
