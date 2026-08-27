package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = (*accountDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*helpdeskSettingsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*emailSettingsDataSource)(nil)
)

// --- account -------------------------------------------------------------

// NewAccountDataSource returns the freshdesk_account data source.
func NewAccountDataSource() datasource.DataSource { return &accountDataSource{} }

type accountDataSource struct{ dataSourceBase }

func (d *accountDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *accountDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Details of the helpdesk account the provider is configured for. " +
			"Useful for asserting that a configuration is pointed at the right tenant, and " +
			"for reading the plan tier that gates several features.",
		Attributes: map[string]dschema.Attribute{
			"id":                computedString("Numeric identifier of the account."),
			"account_name":      computedString("Name of the account."),
			"account_domain":    computedString("Primary Freshdesk domain of the account."),
			"organisation_id":   computedString("Identifier of the Freshworks organisation."),
			"organisation_name": computedString("Name of the Freshworks organisation."),
			"tier_type":         computedString("Plan the account is on, for example `Enterprise Omni`."),
			"type":              computedString("Whether the account is classic or unified-omni."),
			"data_center":       computedString("Region the account is hosted in."),
			"timezone":          computedString("Default time zone of the helpdesk."),
			"hipaa_compliant":   computedBool("Whether the account is provisioned for HIPAA."),
			"full_time_agents":  computedInt64("Number of full-time agent seats in use."),
			"occasional_agents": computedInt64("Number of occasional agent seats in use."),
			"field_agents":      computedInt64("Number of field service agent seats in use."),
			"collaborators":     computedInt64("Number of collaborator seats in use."),
			"contact_email":     computedString("Email address of the account's primary contact."),
			"contact_name":      computedString("Name of the account's primary contact."),
		},
	}
}

func (d *accountDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	account, err := d.client.GetAccount(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk account", err.Error())

		return
	}

	contactName := account.ContactPerson.FirstName
	if account.ContactPerson.LastName != "" {
		contactName += " " + account.ContactPerson.LastName
	}

	for name, value := range map[string]attrValue{
		"id":                idString(account.AccountID),
		"account_name":      optString(account.AccountName),
		"account_domain":    optString(account.AccountDomain),
		"organisation_id":   idString(account.OrganisationID),
		"organisation_name": optString(account.OrganisationName),
		"tier_type":         optString(account.TierType),
		"type":              optString(account.Type),
		"data_center":       optString(account.DataCenter),
		"timezone":          optString(account.Timezone),
		"hipaa_compliant":   types.BoolValue(account.HIPAACompliant),
		"full_time_agents":  types.Int64Value(int64(account.TotalAgents.FullTime)),
		"occasional_agents": types.Int64Value(int64(account.TotalAgents.Occasional)),
		"field_agents":      types.Int64Value(int64(account.TotalAgents.FieldService)),
		"collaborators":     types.Int64Value(int64(account.TotalAgents.Collaborators)),
		"contact_email":     optString(account.ContactPerson.Email),
		"contact_name":      optString(contactName),
	} {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot(name), value)...)
	}
}

// --- helpdesk settings ---------------------------------------------------

// NewHelpdeskSettingsDataSource returns the freshdesk_helpdesk_settings
// data source.
func NewHelpdeskSettingsDataSource() datasource.DataSource { return &helpdeskSettingsDataSource{} }

type helpdeskSettingsDataSource struct{ dataSourceBase }

func (d *helpdeskSettingsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_helpdesk_settings"
}

func (d *helpdeskSettingsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "The helpdesk's language configuration.",
		Attributes: map[string]dschema.Attribute{
			"id":                    computedString("Always `freshdesk`; these settings are account-wide."),
			"primary_language":      computedString("Default language of the helpdesk."),
			"supported_languages":   computedStringSet("Languages agents may work in."),
			"portal_languages":      computedStringSet("Languages the customer portal is offered in."),
			"help_widget_languages": computedStringSet("Languages the help widget is offered in."),
		},
	}
}

func (d *helpdeskSettingsDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	settings, err := d.client.GetHelpdeskSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk helpdesk settings", err.Error())

		return
	}

	for name, value := range map[string]attrValue{
		"id":                    types.StringValue(singletonID),
		"primary_language":      optString(settings.PrimaryLanguage),
		"supported_languages":   stringSet(settings.SupportedLanguages),
		"portal_languages":      stringSet(settings.PortalLanguages),
		"help_widget_languages": stringSet(settings.HelpWidgetLanguages),
	} {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot(name), value)...)
	}
}

// --- email settings ------------------------------------------------------

// NewEmailSettingsDataSource returns the freshdesk_email_settings data source.
func NewEmailSettingsDataSource() datasource.DataSource { return &emailSettingsDataSource{} }

type emailSettingsDataSource struct{ dataSourceBase }

func (d *emailSettingsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_email_settings"
}

func (d *emailSettingsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "The account's email handling settings.",
		Attributes: map[string]dschema.Attribute{
			"id": computedString("Always `freshdesk`; these settings are account-wide."),
			"personalized_email_replies": computedBool(
				"Whether replies are sent from the agent's own name."),
			"create_requester_using_reply_to": computedBool(
				"Whether the Reply-To address becomes the requester."),
			"allow_agent_to_initiate_conversation": computedBool(
				"Whether agents may email a customer without an existing ticket."),
			"original_sender_as_requester_for_forward": computedBool(
				"Whether a forwarded email's original sender becomes the requester."),
			"allow_wildcard_ticket_create": computedBool(
				"Whether mail to any address on the domain creates a ticket."),
			"skip_ticket_threading": computedBool(
				"Whether replies always open a new ticket instead of threading."),
			"threading_without_user_check": computedBool(
				"Whether replies thread onto a ticket even from a different sender."),
			"threading_without_ticket_id_check": computedBool(
				"Whether replies thread onto a ticket without a matching ticket ID."),
			"extended_quoted_text": computedBool(
				"Whether more of the quoted history is kept on replies."),
			"auto_response_detector_toggle": computedBool(
				"Whether automatic replies are detected and suppressed."),
			"email_subject_match": computedBool(
				"Whether a matching subject threads a reply onto an existing ticket."),
			"multiple_to": computedBool(
				"Whether a message with several To addresses creates one ticket per address."),
			"prioritize_tickets_by_outlook_importance": computedBool(
				"Whether Outlook's importance flag sets the ticket priority."),
		},
	}
}

func (d *emailSettingsDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	s, err := d.client.GetEmailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk email settings", err.Error())

		return
	}

	for name, value := range map[string]attrValue{
		"id":                                   types.StringValue(singletonID),
		"personalized_email_replies":           types.BoolValue(s.PersonalizedEmailReplies),
		"create_requester_using_reply_to":      types.BoolValue(s.CreateRequesterUsingReplyTo),
		"allow_agent_to_initiate_conversation": types.BoolValue(s.AllowAgentToInitiateConversation),
		"original_sender_as_requester_for_forward": types.BoolValue(
			s.OriginalSenderAsRequesterForForward),
		"allow_wildcard_ticket_create":      types.BoolValue(s.AllowWildcardTicketCreate),
		"skip_ticket_threading":             types.BoolValue(s.SkipTicketThreading),
		"threading_without_user_check":      types.BoolValue(s.ThreadingWithoutUserCheck),
		"threading_without_ticket_id_check": types.BoolValue(s.ThreadingWithoutTicketIDCheck),
		"extended_quoted_text":              types.BoolValue(s.ExtendedQuotedText),
		"auto_response_detector_toggle":     types.BoolValue(s.AutoResponseDetectorToggle),
		"email_subject_match":               types.BoolValue(s.EmailSubjectMatch),
		"multiple_to":                       types.BoolValue(s.MultipleTo),
		"prioritize_tickets_by_outlook_importance": types.BoolValue(
			s.PrioritizeTicketsByOutlookImportance),
	} {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot(name), value)...)
	}
}
