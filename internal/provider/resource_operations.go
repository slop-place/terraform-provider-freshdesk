package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*ticketFormResource)(nil)
	_ resource.ResourceWithConfigure   = (*ticketFormResource)(nil)
	_ resource.ResourceWithImportState = (*ticketFormResource)(nil)
	_ resource.Resource                = (*cannedResponseFolderResource)(nil)
	_ resource.Resource                = (*cannedResponseResource)(nil)
	_ resource.Resource                = (*emailMailboxResource)(nil)
	_ resource.Resource                = (*timeEntryResource)(nil)
)

// --- ticket forms --------------------------------------------------------

type ticketFormResource = crud[ticketFormModel, freshdesk.TicketForm, *ticketFormModel]

type ticketFormModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Default     types.Bool   `tfsdk:"default"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *ticketFormModel) GetID() types.String { return m.ID }

// Apply copies an API ticket form into the model.
func (m *ticketFormModel) Apply(f *freshdesk.TicketForm) {
	m.ID = idString(f.ID)
	m.Name = types.StringValue(f.Name)
	m.Title = optString(f.Title)
	m.Description = optString(f.Description)
	m.Default = types.BoolValue(f.Default)
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)
}

// NewTicketFormResource returns the freshdesk_ticket_form resource.
func NewTicketFormResource() resource.Resource {
	build := func(plan *ticketFormModel) freshdesk.TicketFormRequest {
		return freshdesk.TicketFormRequest{
			Title:       strPtr(plan.Title),
			Description: strPtr(plan.Description),
		}
	}

	return &ticketFormResource{
		name:  "ticket_form",
		label: "ticket form",
		schema: schema.Schema{
			MarkdownDescription: "A ticket form shown in the customer portal.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("form"),
				"name": schema.StringAttribute{
					Computed: true,
					MarkdownDescription: "Internal name of the form. Freshdesk derives it from " +
						"`title` and does not accept it on a write.",
				},
				"title": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Heading shown above the form in the portal. Must be " +
						"unique across the helpdesk.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Text shown beneath the form's title.",
				},
				"default": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether this is the fallback form for new tickets.",
				},
			}, timestampAttributes("form")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *ticketFormModel, _ *diagnostics,
		) (*freshdesk.TicketForm, error) {
			return c.CreateTicketForm(ctx, build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *ticketFormModel,
		) (*freshdesk.TicketForm, error) {
			return c.GetTicketForm(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *ticketFormModel, _ *diagnostics,
		) (*freshdesk.TicketForm, error) {
			return c.UpdateTicketForm(ctx, id, build(plan))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *ticketFormModel) error {
			return c.DeleteTicketForm(ctx, id)
		},
	}
}

// --- canned response folders ---------------------------------------------

type cannedResponseFolderResource = crud[
	cannedResponseFolderModel, freshdesk.CannedResponseFolder, *cannedResponseFolderModel]

type cannedResponseFolderModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Personal       types.Bool   `tfsdk:"personal"`
	ResponsesCount types.Int64  `tfsdk:"responses_count"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *cannedResponseFolderModel) GetID() types.String { return m.ID }

// Apply copies an API canned response folder into the model.
func (m *cannedResponseFolderModel) Apply(f *freshdesk.CannedResponseFolder) {
	m.ID = idString(f.ID)
	m.Name = types.StringValue(f.Name)
	m.Personal = types.BoolValue(f.Personal)
	m.ResponsesCount = types.Int64Value(int64(f.ResponsesCount))
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)
}

// NewCannedResponseFolderResource returns the
// freshdesk_canned_response_folder resource.
func NewCannedResponseFolderResource() resource.Resource {
	build := func(plan *cannedResponseFolderModel) freshdesk.CannedResponseFolderRequest {
		return freshdesk.CannedResponseFolderRequest{Name: strPtr(plan.Name)}
	}

	return &cannedResponseFolderResource{
		name:  "canned_response_folder",
		label: "canned response folder",
		schema: schema.Schema{
			MarkdownDescription: "A folder grouping canned responses.\n\n" +
				"~> The Freshdesk API offers no way to delete a folder. Destroying this " +
				"resource removes it from state and warns; the folder itself stays until it " +
				"is deleted in the portal.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("folder"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the folder.",
				},
				"personal": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether this is an agent's private folder.",
				},
				"responses_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Number of canned responses in the folder.",
				},
			}, timestampAttributes("folder")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *cannedResponseFolderModel, _ *diagnostics,
		) (*freshdesk.CannedResponseFolder, error) {
			return c.CreateCannedResponseFolder(ctx, build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *cannedResponseFolderModel,
		) (*freshdesk.CannedResponseFolder, error) {
			return c.GetCannedResponseFolder(ctx, id)
		},
		updateFn: func(
			ctx context.Context,
			c *freshdesk.Client,
			id int64,
			plan, _ *cannedResponseFolderModel,
			_ *diagnostics,
		) (*freshdesk.CannedResponseFolder, error) {
			return c.UpdateCannedResponseFolder(ctx, id, build(plan))
		},
		deleteFn: deleteUnsupported[cannedResponseFolderModel, freshdesk.CannedResponseFolder](
			"canned response folder", "from Admin > Canned Responses in the Freshdesk portal"),
	}
}

// --- canned responses ----------------------------------------------------

type cannedResponseResource = crud[cannedResponseModel, freshdesk.CannedResponse, *cannedResponseModel]

type cannedResponseModel struct {
	ID          types.String `tfsdk:"id"`
	FolderID    types.Int64  `tfsdk:"folder_id"`
	Title       types.String `tfsdk:"title"`
	ContentHTML types.String `tfsdk:"content_html"`
	Visibility  types.Int64  `tfsdk:"visibility"`
	GroupIDs    types.Set    `tfsdk:"group_ids"`
	Content     types.String `tfsdk:"content"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *cannedResponseModel) GetID() types.String { return m.ID }

// Apply copies an API canned response into the model.
func (m *cannedResponseModel) Apply(r *freshdesk.CannedResponse) {
	m.ID = idString(r.ID)
	m.Title = types.StringValue(r.Title)
	m.Content = optString(r.Content)
	// Freshdesk rewrites the markup it is given (a <p> comes back as a <div>),
	// so the configured value is kept and the server's rendering is exposed
	// through `content` instead. A body edited in the portal therefore cannot
	// be detected as drift.
	if m.ContentHTML.IsNull() {
		m.ContentHTML = optString(r.ContentHTML)
	}
	m.Visibility = types.Int64Value(int64(r.Visibility))
	m.CreatedAt = timeString(r.CreatedAt)
	m.UpdatedAt = timeString(r.UpdatedAt)

	if r.FolderID != 0 {
		m.FolderID = types.Int64Value(r.FolderID)
	}

	m.GroupIDs = applyInt64Set(m.GroupIDs, r.GroupIDs)
}

// NewCannedResponseResource returns the freshdesk_canned_response resource.
func NewCannedResponseResource() resource.Resource {
	build := func(ctx context.Context, plan *cannedResponseModel, d *diagnostics) freshdesk.CannedResponseRequest {
		return freshdesk.CannedResponseRequest{
			Title:       strPtr(plan.Title),
			ContentHTML: strPtr(plan.ContentHTML),
			FolderID:    int64Ptr(plan.FolderID),
			Visibility:  intPtr(plan.Visibility),
			GroupIDs:    toInt64Slice(ctx, plan.GroupIDs, d),
		}
	}

	return &cannedResponseResource{
		name:  "canned_response",
		label: "canned response",
		schema: schema.Schema{
			MarkdownDescription: "A reusable reply template agents can insert into a ticket.\n\n" +
				"~> The Freshdesk API offers no way to delete a canned response. Destroying " +
				"this resource removes it from state and warns; the response itself stays " +
				"until it is deleted in the portal.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("canned response"),
				"folder_id": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "ID of the folder holding the response.",
				},
				"title": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Title of the response, shown in the agent's picker.",
				},
				"content_html": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Body of the response, in HTML.\n\n" +
						"~> Freshdesk rewrites the markup it is given, so this attribute keeps " +
						"what you configured and `content` exposes the stored rendering. A body " +
						"edited in the portal cannot be detected as drift.",
				},
				"visibility": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorInt64{cannedResponseVisibilityValidator()},
					MarkdownDescription: "Who may use the response: `0` all agents, `1` only the " +
						"owner, `2` selected groups.",
				},
				"group_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of the groups that may use the response. Only used " +
						"when `visibility` is `2`.",
				},
				"content": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Plain-text rendering of the response as Freshdesk stores it.",
				},
			}, timestampAttributes("canned response")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *cannedResponseModel, d *diagnostics,
		) (*freshdesk.CannedResponse, error) {
			return c.CreateCannedResponse(ctx, build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *cannedResponseModel,
		) (*freshdesk.CannedResponse, error) {
			return c.GetCannedResponse(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *cannedResponseModel, d *diagnostics,
		) (*freshdesk.CannedResponse, error) {
			return c.UpdateCannedResponse(ctx, id, build(ctx, plan, d))
		},
		deleteFn: deleteUnsupported[cannedResponseModel, freshdesk.CannedResponse](
			"canned response", "from Admin > Canned Responses in the Freshdesk portal"),
	}
}

// --- email mailboxes -----------------------------------------------------

type emailMailboxResource = crud[emailMailboxModel, freshdesk.EmailMailbox, *emailMailboxModel]

type emailMailboxModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SupportEmail      types.String `tfsdk:"support_email"`
	GroupID           types.Int64  `tfsdk:"group_id"`
	ProductID         types.Int64  `tfsdk:"product_id"`
	DefaultReplyEmail types.Bool   `tfsdk:"default_reply_email"`
	Active            types.Bool   `tfsdk:"active"`
	MailboxType       types.String `tfsdk:"mailbox_type"`
	DisableVerify     types.Bool   `tfsdk:"disable_verify"`
	// CustomMailbox holds the IMAP and SMTP settings as JSON, because they
	// carry write-only credentials the API never echoes back.
	CustomMailbox jsonValueType `tfsdk:"custom_mailbox"`
	ForwardEmail  types.String  `tfsdk:"forward_email"`
	CreatedAt     types.String  `tfsdk:"created_at"`
	UpdatedAt     types.String  `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *emailMailboxModel) GetID() types.String { return m.ID }

// Apply copies an API mailbox into the model.
func (m *emailMailboxModel) Apply(b *freshdesk.EmailMailbox) {
	m.ID = idString(b.ID)
	m.Name = types.StringValue(b.Name)
	m.SupportEmail = types.StringValue(b.SupportEmail)
	m.GroupID = optInt64(b.GroupID)
	m.ProductID = optInt64(b.ProductID)
	m.DefaultReplyEmail = types.BoolValue(b.DefaultReplyEmail)
	m.Active = types.BoolValue(b.Active)
	m.MailboxType = optString(b.MailboxType)
	m.DisableVerify = types.BoolValue(b.DisableVerify)
	m.CreatedAt = timeString(b.CreatedAt)
	m.UpdatedAt = timeString(b.UpdatedAt)

	if b.FreshdeskMailbox != nil {
		m.ForwardEmail = optString(b.FreshdeskMailbox.ForwardEmail)
	}
	// custom_mailbox is deliberately not refreshed from the API: the response
	// omits the password, so echoing it back would wipe it from state.
}

// emailMailboxRequestFor maps a plan into the API payload. The connection
// settings pass through verbatim, because the keys Freshdesk accepts differ by
// mail provider and by authentication method.
func emailMailboxRequestFor(
	plan *emailMailboxModel,
	d *diagnostics,
) freshdesk.EmailMailboxRequest {
	req := freshdesk.EmailMailboxRequest{
		Name:              strPtr(plan.Name),
		SupportEmail:      strPtr(plan.SupportEmail),
		GroupID:           int64Ptr(plan.GroupID),
		ProductID:         int64Ptr(plan.ProductID),
		DefaultReplyEmail: boolPtr(plan.DefaultReplyEmail),
		Active:            boolPtr(plan.Active),
		MailboxType:       strPtr(plan.MailboxType),
		DisableVerify:     boolPtr(plan.DisableVerify),
	}

	if raw := jsonAttrPtr(plan.CustomMailbox, d, "custom_mailbox"); raw != nil {
		obj, ok := raw.(map[string]any)
		if !ok {
			d.AddError("Invalid custom_mailbox", "The value must be a JSON object.")

			return req
		}
		req.CustomMailbox = &obj
	}

	return req
}

// NewEmailMailboxResource returns the freshdesk_email_mailbox resource.
func NewEmailMailboxResource() resource.Resource {
	return &emailMailboxResource{
		name:  "email_mailbox",
		label: "email mailbox",
		schema: schema.Schema{
			MarkdownDescription: "A support mailbox that turns inbound email into tickets.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("mailbox"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the mailbox.",
				},
				"support_email": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Address customers write to.",
				},
				"group_id": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "ID of the group tickets from this mailbox are assigned to.",
				},
				"product_id": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "ID of the product this mailbox belongs to.",
				},
				"default_reply_email": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether outgoing replies are sent from this address.",
				},
				"active": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the mailbox is accepting mail.",
				},
				"mailbox_type": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorString{mailboxTypeValidator()},
					MarkdownDescription: "`freshdesk_mailbox` for a Freshdesk-hosted forwarding " +
						"address, or `custom_mailbox` for one you own and reach over IMAP.",
				},
				"disable_verify": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether to skip ownership verification when creating.",
				},
				"custom_mailbox": schema.StringAttribute{
					Optional:   true,
					Sensitive:  true,
					CustomType: jsonAttr(),
					MarkdownDescription: "IMAP and SMTP settings as a JSON object with `incoming` " +
						"and `outgoing` keys, each taking `mail_server`, `port`, `user_name`, " +
						"`password`, `use_ssl` and `delete_from_server`. Only used when " +
						"`mailbox_type` is `custom_mailbox`.\n\n" +
						"~> Freshdesk never returns the password, so this attribute is not " +
						"refreshed from the API and drift in it cannot be detected.",
				},
				"forward_email": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Address to forward mail to, for a Freshdesk-hosted mailbox.",
				},
			}, timestampAttributes("mailbox")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *emailMailboxModel, d *diagnostics,
		) (*freshdesk.EmailMailbox, error) {
			return c.CreateEmailMailbox(ctx, emailMailboxRequestFor(plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *emailMailboxModel,
		) (*freshdesk.EmailMailbox, error) {
			return c.GetEmailMailbox(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *emailMailboxModel, d *diagnostics,
		) (*freshdesk.EmailMailbox, error) {
			return c.UpdateEmailMailbox(ctx, id, emailMailboxRequestFor(plan, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *emailMailboxModel) error {
			return c.DeleteEmailMailbox(ctx, id)
		},
	}
}

// --- time entries --------------------------------------------------------

type timeEntryResource = crud[timeEntryModel, freshdesk.TimeEntry, *timeEntryModel]

type timeEntryModel struct {
	ID           types.String `tfsdk:"id"`
	TicketID     types.Int64  `tfsdk:"ticket_id"`
	Note         types.String `tfsdk:"note"`
	TimeSpent    types.String `tfsdk:"time_spent"`
	Billable     types.Bool   `tfsdk:"billable"`
	AgentID      types.Int64  `tfsdk:"agent_id"`
	ExecutedAt   types.String `tfsdk:"executed_at"`
	TimerRunning types.Bool   `tfsdk:"timer_running"`
	StartTime    types.String `tfsdk:"start_time"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *timeEntryModel) GetID() types.String { return m.ID }

// Apply copies an API time entry into the model.
func (m *timeEntryModel) Apply(e *freshdesk.TimeEntry) {
	m.ID = idString(e.ID)
	m.Note = optString(e.Note)
	m.TimeSpent = optString(e.TimeSpent)
	m.Billable = types.BoolValue(e.Billable)
	m.AgentID = optInt64(e.AgentID)
	m.TimerRunning = types.BoolValue(e.TimerRunning)
	m.StartTime = timeString(e.StartTime)
	m.ExecutedAt = timeString(e.ExecutedAt)
	m.CreatedAt = timeString(e.CreatedAt)
	m.UpdatedAt = timeString(e.UpdatedAt)

	if e.TicketID != 0 {
		m.TicketID = types.Int64Value(e.TicketID)
	}
}

// NewTimeEntryResource returns the freshdesk_time_entry resource.
func NewTimeEntryResource() resource.Resource {
	build := func(plan *timeEntryModel) freshdesk.TimeEntryRequest {
		return freshdesk.TimeEntryRequest{
			Note:         strPtr(plan.Note),
			TimeSpent:    strPtr(plan.TimeSpent),
			Billable:     boolPtr(plan.Billable),
			AgentID:      int64Ptr(plan.AgentID),
			TimerRunning: boolPtr(plan.TimerRunning),
			ExecutedAt:   strPtr(plan.ExecutedAt),
			StartTime:    strPtr(plan.StartTime),
		}
	}

	return &timeEntryResource{
		name:  "time_entry",
		label: "time entry",
		schema: schema.Schema{
			MarkdownDescription: "Time logged by an agent against a ticket.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("time entry"),
				"ticket_id": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "ID of the ticket the time is logged against. Changing it " +
						"forces a new entry.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"note": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the work done.",
				},
				"time_spent": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Duration as `HH:MM`, for example `01:30`.",
				},
				"billable": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the time is chargeable to the customer.",
				},
				"agent_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the agent the time is logged for.",
				},
				"executed_at": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "When the work was done (RFC 3339), for backdating.",
				},
				"timer_running": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the entry's timer is currently counting.",
				},
				"start_time": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "When the timer was started (RFC 3339).",
				},
			}, timestampAttributes("time entry")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *timeEntryModel, _ *diagnostics,
		) (*freshdesk.TimeEntry, error) {
			return c.CreateTimeEntry(ctx, plan.TicketID.ValueInt64(), build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *timeEntryModel,
		) (*freshdesk.TimeEntry, error) {
			return c.GetTimeEntry(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *timeEntryModel, _ *diagnostics,
		) (*freshdesk.TimeEntry, error) {
			return c.UpdateTimeEntry(ctx, id, build(plan))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *timeEntryModel) error {
			return c.DeleteTimeEntry(ctx, id)
		},
	}
}
