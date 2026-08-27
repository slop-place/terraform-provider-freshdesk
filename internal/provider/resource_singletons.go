package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*emailSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*emailSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*emailSettingsResource)(nil)
	_ resource.Resource                = (*notificationBCCResource)(nil)
	_ resource.ResourceWithConfigure   = (*notificationBCCResource)(nil)
	_ resource.ResourceWithImportState = (*notificationBCCResource)(nil)
)

// singletonID is the fixed resource ID used by account-wide settings, which
// have no identifier of their own.
const singletonID = "freshdesk"

// --- email settings ------------------------------------------------------

// NewEmailSettingsResource returns the freshdesk_email_settings resource.
func NewEmailSettingsResource() resource.Resource { return &emailSettingsResource{} }

type emailSettingsResource struct{ base }

type emailSettingsModel struct {
	ID types.String `tfsdk:"id"`

	PersonalizedEmailReplies             types.Bool `tfsdk:"personalized_email_replies"`
	CreateRequesterUsingReplyTo          types.Bool `tfsdk:"create_requester_using_reply_to"`
	AllowAgentToInitiateConversation     types.Bool `tfsdk:"allow_agent_to_initiate_conversation"`
	OriginalSenderAsRequesterForForward  types.Bool `tfsdk:"original_sender_as_requester_for_forward"`
	AllowWildcardTicketCreate            types.Bool `tfsdk:"allow_wildcard_ticket_create"`
	SkipTicketThreading                  types.Bool `tfsdk:"skip_ticket_threading"`
	ThreadingWithoutUserCheck            types.Bool `tfsdk:"threading_without_user_check"`
	ThreadingWithoutTicketIDCheck        types.Bool `tfsdk:"threading_without_ticket_id_check"`
	ExtendedQuotedText                   types.Bool `tfsdk:"extended_quoted_text"`
	AutoResponseDetectorToggle           types.Bool `tfsdk:"auto_response_detector_toggle"`
	EmailSubjectMatch                    types.Bool `tfsdk:"email_subject_match"`
	MultipleTo                           types.Bool `tfsdk:"multiple_to"`
	PrioritizeTicketsByOutlookImportance types.Bool `tfsdk:"prioritize_tickets_by_outlook_importance"`
}

// apply copies API email settings into the model.
func (m *emailSettingsModel) apply(s *freshdesk.EmailSettings) {
	m.ID = types.StringValue(singletonID)
	m.PersonalizedEmailReplies = types.BoolValue(s.PersonalizedEmailReplies)
	m.CreateRequesterUsingReplyTo = types.BoolValue(s.CreateRequesterUsingReplyTo)
	m.AllowAgentToInitiateConversation = types.BoolValue(s.AllowAgentToInitiateConversation)
	m.OriginalSenderAsRequesterForForward = types.BoolValue(s.OriginalSenderAsRequesterForForward)
	m.AllowWildcardTicketCreate = types.BoolValue(s.AllowWildcardTicketCreate)
	m.SkipTicketThreading = types.BoolValue(s.SkipTicketThreading)
	m.ThreadingWithoutUserCheck = types.BoolValue(s.ThreadingWithoutUserCheck)
	m.ThreadingWithoutTicketIDCheck = types.BoolValue(s.ThreadingWithoutTicketIDCheck)
	m.ExtendedQuotedText = types.BoolValue(s.ExtendedQuotedText)
	m.AutoResponseDetectorToggle = types.BoolValue(s.AutoResponseDetectorToggle)
	m.EmailSubjectMatch = types.BoolValue(s.EmailSubjectMatch)
	m.MultipleTo = types.BoolValue(s.MultipleTo)
	m.PrioritizeTicketsByOutlookImportance = types.BoolValue(s.PrioritizeTicketsByOutlookImportance)
}

// settings renders the model as the sparse map the API expects, carrying only
// the toggles the practitioner actually set.
func (m *emailSettingsModel) settings() map[string]bool {
	out := map[string]bool{}
	for key, value := range map[string]types.Bool{
		"personalized_email_replies":               m.PersonalizedEmailReplies,
		"create_requester_using_reply_to":          m.CreateRequesterUsingReplyTo,
		"allow_agent_to_initiate_conversation":     m.AllowAgentToInitiateConversation,
		"original_sender_as_requester_for_forward": m.OriginalSenderAsRequesterForForward,
		"allow_wildcard_ticket_create":             m.AllowWildcardTicketCreate,
		"skip_ticket_threading":                    m.SkipTicketThreading,
		"threading_without_user_check":             m.ThreadingWithoutUserCheck,
		"threading_without_ticket_id_check":        m.ThreadingWithoutTicketIDCheck,
		"extended_quoted_text":                     m.ExtendedQuotedText,
		"auto_response_detector_toggle":            m.AutoResponseDetectorToggle,
		"email_subject_match":                      m.EmailSubjectMatch,
		"multiple_to":                              m.MultipleTo,
		"prioritize_tickets_by_outlook_importance": m.PrioritizeTicketsByOutlookImportance,
	} {
		if !value.IsNull() && !value.IsUnknown() {
			out[key] = value.ValueBool()
		}
	}

	return out
}

func (r *emailSettingsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_email_settings"
}

func (r *emailSettingsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	boolAttr := func(description string) schema.BoolAttribute {
		return schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: description,
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Account-wide email handling settings.\n\n" +
			"~> This is a singleton: only one instance should exist, and destroying it " +
			"leaves the settings as they are rather than resetting them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `freshdesk`; these settings are account-wide.",
				PlanModifiers:       []planModifierString{stringplanmodifier.UseStateForUnknown()},
			},
			"personalized_email_replies": boolAttr(
				"Whether replies are sent from the agent's own name."),
			"create_requester_using_reply_to": boolAttr(
				"Whether the Reply-To address becomes the requester."),
			"allow_agent_to_initiate_conversation": boolAttr(
				"Whether agents may email a customer without an existing ticket."),
			"original_sender_as_requester_for_forward": boolAttr(
				"Whether a forwarded email's original sender becomes the requester."),
			"allow_wildcard_ticket_create": boolAttr(
				"Whether mail to any address on the domain creates a ticket."),
			"skip_ticket_threading": boolAttr(
				"Whether replies always open a new ticket instead of threading."),
			"threading_without_user_check": boolAttr(
				"Whether replies thread onto a ticket even from a different sender."),
			"threading_without_ticket_id_check": boolAttr(
				"Whether replies thread onto a ticket without a matching ticket ID."),
			"extended_quoted_text": boolAttr(
				"Whether more of the quoted history is kept on replies."),
			"auto_response_detector_toggle": boolAttr(
				"Whether automatic replies such as out-of-office are detected and suppressed."),
			"email_subject_match": boolAttr(
				"Whether a matching subject line threads a reply onto an existing ticket."),
			"multiple_to": boolAttr(
				"Whether a message with several To addresses creates one ticket per address."),
			"prioritize_tickets_by_outlook_importance": boolAttr(
				"Whether Outlook's importance flag sets the ticket priority."),
		},
	}
}

func (r *emailSettingsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan emailSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emailSettingsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state emailSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetEmailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk email settings", err.Error())

		return
	}

	state.apply(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emailSettingsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan emailSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete leaves the account's settings untouched: there is nothing to remove.
func (r *emailSettingsResource) Delete(
	_ context.Context,
	_ resource.DeleteRequest,
	_ *resource.DeleteResponse,
) {
}

func (r *emailSettingsResource) ImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), singletonID)...)
}

// write pushes the configured toggles and refreshes the model from the result.
func (r *emailSettingsResource) write(ctx context.Context, plan *emailSettingsModel, diags *diagnostics) {
	settings, err := r.client.UpdateEmailSettings(ctx, plan.settings())
	if err != nil {
		diags.AddError("Unable to update the Freshdesk email settings", err.Error())

		return
	}
	plan.apply(settings)
}

// --- automatic BCC -------------------------------------------------------

// NewNotificationBCCResource returns the freshdesk_notification_bcc resource.
func NewNotificationBCCResource() resource.Resource { return &notificationBCCResource{} }

type notificationBCCResource struct{ base }

type notificationBCCModel struct {
	ID     types.String `tfsdk:"id"`
	Emails types.Set    `tfsdk:"emails"`
}

func (r *notificationBCCResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_bcc"
}

func (r *notificationBCCResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Addresses blind-copied on every outgoing notification, typically " +
			"for archiving.\n\n" +
			"~> This is a singleton: only one instance should exist. Destroying it clears " +
			"the address list.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `freshdesk`; this list is account-wide.",
				PlanModifiers:       []planModifierString{stringplanmodifier.UseStateForUnknown()},
			},
			"emails": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Addresses to blind-copy. An empty set disables the feature.",
			},
		},
	}
}

func (r *notificationBCCResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan notificationBCCModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationBCCResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state notificationBCCModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	emails, err := r.client.GetNotificationBCC(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk BCC addresses", err.Error())

		return
	}

	state.ID = types.StringValue(singletonID)
	state.Emails = stringSet(emails)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationBCCResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan notificationBCCModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete clears the list, which is the closest the API has to removal.
func (r *notificationBCCResource) Delete(
	ctx context.Context,
	_ resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	if _, err := r.client.UpdateNotificationBCC(ctx, nil); err != nil {
		resp.Diagnostics.AddError("Unable to clear the Freshdesk BCC addresses", err.Error())
	}
}

func (r *notificationBCCResource) ImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), singletonID)...)
}

// write pushes the configured addresses and refreshes the model.
func (r *notificationBCCResource) write(
	ctx context.Context,
	plan *notificationBCCModel,
	diags *diagnostics,
) {
	emails, err := r.client.UpdateNotificationBCC(ctx, toStringSlice(ctx, plan.Emails, diags))
	if err != nil {
		diags.AddError("Unable to update the Freshdesk BCC addresses", err.Error())

		return
	}

	plan.ID = types.StringValue(singletonID)
	plan.Emails = stringSet(emails)
}
