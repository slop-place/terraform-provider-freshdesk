package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// This file wires each entity descriptor to its singular and plural data
// sources. The scaffolding lives in datasource.go; the attribute definitions
// live in the entities_*.go files.

// listAll adapts a client list method that takes only ListOptions, so the
// common "list everything" data source needs no closure of its own. The
// parameter is a method expression, hence the leading receiver.
func listAll[A any](
	fn func(*freshdesk.Client, context.Context, freshdesk.ListOptions) ([]A, error),
) func(context.Context, *freshdesk.Client) ([]A, error) {
	return func(ctx context.Context, c *freshdesk.Client) ([]A, error) {
		return fn(c, ctx, freshdesk.ListOptions{})
	}
}

// --- groups --------------------------------------------------------------

// NewGroupDataSource returns the freshdesk_group data source.
func NewGroupDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Group]{
		name:        "group",
		description: "Look up a single agent group by ID.",
		entity:      groupEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Group, error) {
			return c.GetGroup(ctx, id)
		},
	}
}

// NewGroupsDataSource returns the freshdesk_groups data source.
func NewGroupsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Group]{
		name:        "groups",
		description: "Every agent group on the helpdesk.",
		itemsAttr:   "groups",
		entity:      groupEntity(),
		listFn:      listAll((*freshdesk.Client).ListGroups),
	}
}

// --- agents --------------------------------------------------------------

// NewAgentDataSource returns the freshdesk_agent data source.
func NewAgentDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Agent]{
		name:        "agent",
		description: "Look up a single agent by ID.",
		entity:      agentEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Agent, error) {
			return c.GetAgent(ctx, id)
		},
	}
}

// NewAgentsDataSource returns the freshdesk_agents data source.
func NewAgentsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Agent]{
		name:        "agents",
		description: "Every agent on the helpdesk.",
		itemsAttr:   "agents",
		entity:      agentEntity(),
		listFn: func(ctx context.Context, c *freshdesk.Client) ([]freshdesk.Agent, error) {
			return c.ListAgents(ctx, freshdesk.AgentListOptions{})
		},
	}
}

// --- roles ---------------------------------------------------------------

// NewRoleDataSource returns the freshdesk_role data source.
func NewRoleDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Role]{
		name:        "role",
		description: "Look up a single agent role by ID.",
		entity:      roleEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Role, error) {
			return c.GetRole(ctx, id)
		},
	}
}

// NewRolesDataSource returns the freshdesk_roles data source.
func NewRolesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Role]{
		name: "roles",
		description: "Every agent role on the helpdesk. Roles are defined by Freshdesk and " +
			"cannot be created through the API, so this is the way to find the IDs an " +
			"agent needs.",
		itemsAttr: "roles",
		entity:    roleEntity(),
		listFn:    listAll((*freshdesk.Client).ListRoles),
	}
}

// --- skills --------------------------------------------------------------

// NewSkillDataSource returns the freshdesk_skill data source.
func NewSkillDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Skill]{
		name:        "skill",
		description: "Look up a single routing skill by ID.",
		entity:      skillEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Skill, error) {
			return c.GetSkill(ctx, id)
		},
	}
}

// NewSkillsDataSource returns the freshdesk_skills data source.
func NewSkillsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Skill]{
		name:        "skills",
		description: "Every routing skill on the helpdesk.",
		itemsAttr:   "skills",
		entity:      skillEntity(),
		listFn:      listAll((*freshdesk.Client).ListSkills),
	}
}

// --- products ------------------------------------------------------------

// NewProductDataSource returns the freshdesk_product data source.
func NewProductDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Product]{
		name:        "product",
		description: "Look up a single product by ID.",
		entity:      productEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Product, error) {
			return c.GetProduct(ctx, id)
		},
	}
}

// NewProductsDataSource returns the freshdesk_products data source.
func NewProductsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Product]{
		name: "products",
		description: "Every product on the helpdesk. Products are managed in the admin " +
			"console rather than through the API.",
		itemsAttr: "products",
		entity:    productEntity(),
		listFn:    listAll((*freshdesk.Client).ListProducts),
	}
}

// --- business hours ------------------------------------------------------

// NewBusinessHoursDataSource returns the freshdesk_business_hours data source.
func NewBusinessHoursDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.BusinessHours]{
		name:        "business_hours",
		description: "Look up a single business-hours calendar by ID.",
		entity:      businessHoursEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.BusinessHours, error) {
			return c.GetBusinessHours(ctx, id)
		},
	}
}

// NewBusinessHoursListDataSource returns the freshdesk_business_hours_list
// data source.
func NewBusinessHoursListDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.BusinessHours]{
		name: "business_hours_list",
		description: "Every business-hours calendar on the helpdesk. Calendars are managed " +
			"in the admin console rather than through the API.",
		itemsAttr: "business_hours",
		entity:    businessHoursEntity(),
		listFn:    listAll((*freshdesk.Client).ListBusinessHours),
	}
}

// --- email configs -------------------------------------------------------

// NewEmailConfigDataSource returns the freshdesk_email_config data source.
func NewEmailConfigDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.EmailConfig]{
		name:        "email_config",
		description: "Look up a single email config by ID.",
		entity:      emailConfigEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.EmailConfig, error) {
			return c.GetEmailConfig(ctx, id)
		},
	}
}

// NewEmailConfigsDataSource returns the freshdesk_email_configs data source.
func NewEmailConfigsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.EmailConfig]{
		name:        "email_configs",
		description: "Every email config on the helpdesk.",
		itemsAttr:   "email_configs",
		entity:      emailConfigEntity(),
		listFn:      listAll((*freshdesk.Client).ListEmailConfigs),
	}
}

// --- companies and contacts ----------------------------------------------

// NewCompanyDataSource returns the freshdesk_company data source.
func NewCompanyDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Company]{
		name:        "company",
		description: "Look up a single company by ID.",
		entity:      companyEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Company, error) {
			return c.GetCompany(ctx, id)
		},
	}
}

// NewCompaniesDataSource returns the freshdesk_companies data source.
func NewCompaniesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Company]{
		name:        "companies",
		description: "Every company on the helpdesk.",
		itemsAttr:   "companies",
		entity:      companyEntity(),
		listFn:      listAll((*freshdesk.Client).ListCompanies),
	}
}

// NewContactDataSource returns the freshdesk_contact data source.
func NewContactDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Contact]{
		name:        "contact",
		description: "Look up a single contact by ID.",
		entity:      contactEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Contact, error) {
			return c.GetContact(ctx, id)
		},
	}
}

// NewContactsDataSource returns the freshdesk_contacts data source.
func NewContactsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Contact]{
		name: "contacts",
		description: "Every contact on the helpdesk.\n\n" +
			"~> On a busy helpdesk this walks every page of the contacts collection, " +
			"which can be a great many API calls.",
		itemsAttr: "contacts",
		entity:    contactEntity(),
		listFn: func(ctx context.Context, c *freshdesk.Client) ([]freshdesk.Contact, error) {
			return c.ListContacts(ctx, freshdesk.ContactListOptions{})
		},
	}
}

// --- tickets -------------------------------------------------------------

// NewTicketDataSource returns the freshdesk_ticket data source.
func NewTicketDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.Ticket]{
		name:        "ticket",
		description: "Look up a single ticket by ID.",
		entity:      ticketEntity(),
		getFn: func(ctx context.Context, c *freshdesk.Client, id int64) (*freshdesk.Ticket, error) {
			return c.GetTicket(ctx, id, freshdesk.TicketIncludes{})
		},
	}
}

// NewTicketsDataSource returns the freshdesk_tickets data source.
func NewTicketsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Ticket]{
		name: "tickets",
		description: "Every ticket the API returns.\n\n" +
			"~> Freshdesk only lists tickets updated in the last 30 days, and this walks " +
			"every page of that collection. Prefer a narrower lookup where you can.",
		itemsAttr: "tickets",
		entity:    ticketEntity(),
		listFn: func(ctx context.Context, c *freshdesk.Client) ([]freshdesk.Ticket, error) {
			return c.ListTickets(ctx, freshdesk.TicketListOptions{})
		},
	}
}

// --- fields and forms ----------------------------------------------------

// NewTicketFieldDataSource returns the freshdesk_ticket_field data source.
func NewTicketFieldDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.TicketField]{
		name:        "ticket_field",
		description: "Look up a single ticket field by ID.",
		entity:      ticketFieldEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.TicketField, error) {
			return c.GetTicketField(ctx, id, true)
		},
	}
}

// NewTicketFieldsDataSource returns the freshdesk_ticket_fields data source.
func NewTicketFieldsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.TicketField]{
		name: "ticket_fields",
		description: "Every field on the ticket form, built-in and custom. Useful for " +
			"finding the generated `name` of a custom field.",
		itemsAttr: "ticket_fields",
		entity:    ticketFieldEntity(),
		listFn:    listAll((*freshdesk.Client).ListTicketFields),
	}
}

// NewContactFieldsDataSource returns the freshdesk_contact_fields data source.
func NewContactFieldsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.ContactField]{
		name:        "contact_fields",
		description: "Every field on the contact form, built-in and custom.",
		itemsAttr:   "contact_fields",
		entity:      contactFieldEntity(),
		listFn:      listAll((*freshdesk.Client).ListContactFields),
	}
}

// NewCompanyFieldsDataSource returns the freshdesk_company_fields data source.
func NewCompanyFieldsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.CompanyField]{
		name:        "company_fields",
		description: "Every field on the company form, built-in and custom.",
		itemsAttr:   "company_fields",
		entity:      companyFieldEntity(),
		listFn:      listAll((*freshdesk.Client).ListCompanyFields),
	}
}

// NewTicketFormDataSource returns the freshdesk_ticket_form data source.
func NewTicketFormDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.TicketForm]{
		name:        "ticket_form",
		description: "Look up a single ticket form by ID, including its fields.",
		entity:      ticketFormEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.TicketForm, error) {
			return c.GetTicketForm(ctx, id)
		},
	}
}

// NewTicketFormsDataSource returns the freshdesk_ticket_forms data source.
func NewTicketFormsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.TicketForm]{
		name:        "ticket_forms",
		description: "Every ticket form on the helpdesk.",
		itemsAttr:   "ticket_forms",
		entity:      ticketFormEntity(),
		listFn:      listAll((*freshdesk.Client).ListTicketForms),
	}
}

// --- policies and automations --------------------------------------------

// NewSLAPolicyDataSource returns the freshdesk_sla_policy data source.
func NewSLAPolicyDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.SLAPolicy]{
		name:        "sla_policy",
		description: "Look up a single SLA policy by ID.",
		entity:      slaPolicyEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.SLAPolicy, error) {
			return c.GetSLAPolicy(ctx, id)
		},
	}
}

// NewSLAPoliciesDataSource returns the freshdesk_sla_policies data source.
func NewSLAPoliciesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.SLAPolicy]{
		name:        "sla_policies",
		description: "Every SLA policy on the helpdesk, in evaluation order.",
		itemsAttr:   "sla_policies",
		entity:      slaPolicyEntity(),
		listFn:      listAll((*freshdesk.Client).ListSLAPolicies),
	}
}

// NewScenarioAutomationsDataSource returns the
// freshdesk_scenario_automations data source.
func NewScenarioAutomationsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.ScenarioAutomation]{
		name: "scenario_automations",
		description: "Every scenario automation on the helpdesk. Scenarios are read-only " +
			"through the API.",
		itemsAttr: "scenario_automations",
		entity:    scenarioAutomationEntity(),
		listFn:    listAll((*freshdesk.Client).ListScenarioAutomations),
	}
}

// --- other collections ---------------------------------------------------

// NewCannedResponseFoldersDataSource returns the
// freshdesk_canned_response_folders data source.
func NewCannedResponseFoldersDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.CannedResponseFolder]{
		name:        "canned_response_folders",
		description: "Every canned response folder on the helpdesk.",
		itemsAttr:   "canned_response_folders",
		entity:      cannedResponseFolderEntity(),
		listFn:      listAll((*freshdesk.Client).ListCannedResponseFolders),
	}
}

// NewSolutionCategoryDataSource returns the freshdesk_solution_category
// data source.
func NewSolutionCategoryDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.SolutionCategory]{
		name:        "solution_category",
		description: "Look up a single knowledge-base category by ID.",
		entity:      solutionCategoryEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.SolutionCategory, error) {
			return c.GetSolutionCategory(ctx, id, "")
		},
	}
}

// NewSolutionCategoriesDataSource returns the freshdesk_solution_categories
// data source.
func NewSolutionCategoriesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.SolutionCategory]{
		name:        "solution_categories",
		description: "Every knowledge-base category on the helpdesk.",
		itemsAttr:   "solution_categories",
		entity:      solutionCategoryEntity(),
		listFn: func(
			ctx context.Context, c *freshdesk.Client,
		) ([]freshdesk.SolutionCategory, error) {
			return c.ListSolutionCategories(ctx, "", freshdesk.ListOptions{})
		},
	}
}

// NewSolutionFolderDataSource returns the freshdesk_solution_folder data source.
func NewSolutionFolderDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.SolutionFolder]{
		name:        "solution_folder",
		description: "Look up a single knowledge-base folder by ID.",
		entity:      solutionFolderEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.SolutionFolder, error) {
			return c.GetSolutionFolder(ctx, id, "")
		},
	}
}

// NewSolutionArticleDataSource returns the freshdesk_solution_article
// data source.
func NewSolutionArticleDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.SolutionArticle]{
		name:        "solution_article",
		description: "Look up a single knowledge-base article by ID.",
		entity:      solutionArticleEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.SolutionArticle, error) {
			return c.GetSolutionArticle(ctx, id, "")
		},
	}
}

// NewForumCategoriesDataSource returns the freshdesk_forum_categories
// data source.
func NewForumCategoriesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.ForumCategory]{
		name:        "forum_categories",
		description: "Every community forum category on the helpdesk.",
		itemsAttr:   "forum_categories",
		entity:      forumCategoryEntity(),
		listFn:      listAll((*freshdesk.Client).ListForumCategories),
	}
}

// NewSurveysDataSource returns the freshdesk_surveys data source.
func NewSurveysDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.Survey]{
		name:        "surveys",
		description: "Every customer-satisfaction survey on the helpdesk.",
		itemsAttr:   "surveys",
		entity:      surveyEntity(),
		listFn:      listAll((*freshdesk.Client).ListSurveys),
	}
}

// NewSatisfactionRatingsDataSource returns the freshdesk_satisfaction_ratings
// data source.
func NewSatisfactionRatingsDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.SatisfactionRating]{
		name:        "satisfaction_ratings",
		description: "Every customer-satisfaction rating the API returns.",
		itemsAttr:   "satisfaction_ratings",
		entity:      satisfactionRatingEntity(),
		listFn: func(
			ctx context.Context, c *freshdesk.Client,
		) ([]freshdesk.SatisfactionRating, error) {
			return c.ListSatisfactionRatings(ctx, freshdesk.SatisfactionRatingListOptions{})
		},
	}
}

// NewTimeEntriesDataSource returns the freshdesk_time_entries data source.
func NewTimeEntriesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.TimeEntry]{
		name:        "time_entries",
		description: "Every time entry the API returns.",
		itemsAttr:   "time_entries",
		entity:      timeEntryEntity(),
		listFn: func(ctx context.Context, c *freshdesk.Client) ([]freshdesk.TimeEntry, error) {
			return c.ListTimeEntries(ctx, freshdesk.TimeEntryListOptions{})
		},
	}
}

// NewEmailMailboxDataSource returns the freshdesk_email_mailbox data source.
func NewEmailMailboxDataSource() datasource.DataSource {
	return &lookupDataSource[freshdesk.EmailMailbox]{
		name:        "email_mailbox",
		description: "Look up a single support mailbox by ID.",
		entity:      emailMailboxEntity(),
		getFn: func(
			ctx context.Context, c *freshdesk.Client, id int64,
		) (*freshdesk.EmailMailbox, error) {
			return c.GetEmailMailbox(ctx, id)
		},
	}
}

// NewEmailMailboxesDataSource returns the freshdesk_email_mailboxes
// data source.
func NewEmailMailboxesDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.EmailMailbox]{
		name:        "email_mailboxes",
		description: "Every support mailbox on the helpdesk.",
		itemsAttr:   "email_mailboxes",
		entity:      emailMailboxEntity(),
		listFn:      listAll((*freshdesk.Client).ListEmailMailboxes),
	}
}

// NewCustomObjectSchemasDataSource returns the freshdesk_custom_object_schemas
// data source.
func NewCustomObjectSchemasDataSource() datasource.DataSource {
	return &collectionDataSource[freshdesk.CustomObjectSchema]{
		name: "custom_object_schemas",
		description: "Every custom object schema on the helpdesk.\n\n" +
			"-> Custom objects require a plan that enables the feature. Accounts without " +
			"it answer `403` on this endpoint.",
		itemsAttr: "custom_object_schemas",
		entity:    customObjectSchemaEntity(),
		listFn: func(
			ctx context.Context, c *freshdesk.Client,
		) ([]freshdesk.CustomObjectSchema, error) {
			return c.ListCustomObjectSchemas(ctx)
		},
	}
}

// NewAutomationRulesDataSource returns the freshdesk_automation_rules
// data source, which is scoped to one automation type.
func NewAutomationRulesDataSource() datasource.DataSource {
	return &automationRulesDataSource{}
}

var _ datasource.DataSourceWithConfigure = (*automationRulesDataSource)(nil)

// automationRulesDataSource lists the rules of one automation, so it takes a
// required rule_type input rather than listing everything.
type automationRulesDataSource struct{ dataSourceBase }

func (d *automationRulesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_automation_rules"
}

func (d *automationRulesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	e := automationRuleEntity()

	elemAttrs := mergeAttrs(e.attrs, map[string]dschema.Attribute{
		"id": computedString("Numeric identifier of the rule."),
	})

	resp.Schema = dschema.Schema{
		MarkdownDescription: "Every automation rule of one type.",
		Attributes: map[string]dschema.Attribute{
			"id": computedString("Placeholder identifier for the data source."),
			"rule_type": dschema.Int64Attribute{
				Required: true,
				MarkdownDescription: "Which automation to list: `1` ticket creation " +
					"(dispatcher), `3` time triggers, `4` ticket updates (observer).",
				Validators: []validatorInt64{automationTypeValidator()},
			},
			"automation_rules": dschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The rules belonging to the requested automation.",
				NestedObject:        dschema.NestedAttributeObject{Attributes: elemAttrs},
			},
		},
	}
}

func (d *automationRulesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var ruleType types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("rule_type"), &ruleType)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rules, err := d.client.ListAutomationRules(ctx,
		int(ruleType.ValueInt64()), freshdesk.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list the Freshdesk automation rules", err.Error())

		return
	}

	e := automationRuleEntity()

	elems := make([]attrValue, 0, len(rules))
	for i := range rules {
		elems = append(elems, e.object(&rules[i]))
	}

	list, diags := types.ListValue(e.objectType(), elems)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("id"),
		types.StringValue("automation_rules"))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("rule_type"), ruleType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx,
		pathRoot("automation_rules"), list)...)
}
