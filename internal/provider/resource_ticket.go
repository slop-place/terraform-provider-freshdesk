package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*ticketResource)(nil)
	_ resource.ResourceWithConfigure   = (*ticketResource)(nil)
	_ resource.ResourceWithImportState = (*ticketResource)(nil)
)

type ticketResource = crud[ticketModel, freshdesk.Ticket, *ticketModel]

type ticketModel struct {
	ID          types.String `tfsdk:"id"`
	Subject     types.String `tfsdk:"subject"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Status      types.Int64  `tfsdk:"status"`
	Priority    types.Int64  `tfsdk:"priority"`
	Source      types.Int64  `tfsdk:"source"`

	RequesterID   types.Int64  `tfsdk:"requester_id"`
	ResponderID   types.Int64  `tfsdk:"responder_id"`
	GroupID       types.Int64  `tfsdk:"group_id"`
	CompanyID     types.Int64  `tfsdk:"company_id"`
	ProductID     types.Int64  `tfsdk:"product_id"`
	EmailConfigID types.Int64  `tfsdk:"email_config_id"`
	Email         types.String `tfsdk:"email"`
	Name          types.String `tfsdk:"name"`
	Phone         types.String `tfsdk:"phone"`

	CCEmails     types.Set    `tfsdk:"cc_emails"`
	Tags         types.Set    `tfsdk:"tags"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
	DueBy        types.String `tfsdk:"due_by"`
	FrDueBy      types.String `tfsdk:"fr_due_by"`
	ParentID     types.Int64  `tfsdk:"parent_id"`

	DescriptionText types.String `tfsdk:"description_text"`
	Spam            types.Bool   `tfsdk:"spam"`
	IsEscalated     types.Bool   `tfsdk:"is_escalated"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *ticketModel) GetID() types.String { return m.ID }

// Apply copies an API ticket into the model.
func (m *ticketModel) Apply(t *freshdesk.Ticket) {
	m.ID = idString(t.ID)
	m.Subject = optString(t.Subject)
	m.DescriptionText = optString(t.DescriptionText)
	// Freshdesk rewrites the markup it is given (a <p> comes back as a <div>),
	// so the configured body is kept and the server's plain-text rendering is
	// exposed through description_text instead.
	if m.Description.IsNull() {
		m.Description = optString(t.Description)
	}
	m.Type = optString(t.Type)
	m.Status = types.Int64Value(int64(t.Status))
	m.Priority = types.Int64Value(int64(t.Priority))
	m.Source = types.Int64Value(int64(t.Source))

	m.RequesterID = optInt64(t.RequesterID)
	m.ResponderID = optInt64(t.ResponderID)
	m.GroupID = optInt64(t.GroupID)
	m.CompanyID = optInt64(t.CompanyID)
	m.ProductID = optInt64(t.ProductID)
	m.EmailConfigID = optInt64(t.EmailConfigID)
	m.Email = optString(t.Email)
	m.Name = optString(t.Name)
	m.Phone = optString(t.Phone)

	m.CustomFields = applyMap(m.CustomFields, t.CustomFields)
	m.DueBy = timeString(t.DueBy)
	m.FrDueBy = timeString(t.FrDueBy)
	m.ParentID = optInt64(t.ParentID)
	m.Spam = types.BoolValue(t.Spam)
	m.IsEscalated = types.BoolValue(t.IsEscalated)
	m.CreatedAt = timeString(t.CreatedAt)
	m.UpdatedAt = timeString(t.UpdatedAt)

	m.Tags = applyStringSet(m.Tags, t.Tags)
	m.CCEmails = applyStringSet(m.CCEmails, t.CCEmails)
}

// NewTicketResource returns the freshdesk_ticket resource.
func NewTicketResource() resource.Resource {
	build := func(ctx context.Context, plan *ticketModel, d *diagnostics) freshdesk.TicketRequest {
		req := freshdesk.TicketRequest{
			Subject:       strPtr(plan.Subject),
			Description:   strPtr(plan.Description),
			Type:          strPtr(plan.Type),
			Status:        intPtr(plan.Status),
			Priority:      intPtr(plan.Priority),
			Source:        intPtr(plan.Source),
			RequesterID:   int64Ptr(plan.RequesterID),
			ResponderID:   int64Ptr(plan.ResponderID),
			GroupID:       int64Ptr(plan.GroupID),
			CompanyID:     int64Ptr(plan.CompanyID),
			ProductID:     int64Ptr(plan.ProductID),
			EmailConfigID: int64Ptr(plan.EmailConfigID),
			Email:         strPtr(plan.Email),
			Name:          strPtr(plan.Name),
			Phone:         strPtr(plan.Phone),
			CCEmails:      toStringSlice(ctx, plan.CCEmails, d),
			Tags:          toStringSlice(ctx, plan.Tags, d),
			CustomFields:  mapToCustomFields(ctx, plan.CustomFields, d),
			ParentID:      int64Ptr(plan.ParentID),
		}

		if due := parseTimestamp(plan.DueBy, "due_by", d); due != nil {
			req.DueBy = due
		}

		if fr := parseTimestamp(plan.FrDueBy, "fr_due_by", d); fr != nil {
			req.FrDueBy = fr
		}

		return req
	}

	return &ticketResource{
		name:  "ticket",
		label: "ticket",
		schema: schema.Schema{
			MarkdownDescription: "A support ticket.\n\n" +
				"-> Tickets are usually created by customers rather than declared in " +
				"configuration. This resource is most useful for seeding fixtures and for " +
				"automation that must own a ticket's lifecycle.\n\n" +
				"~> Deleting this resource moves the ticket to the trash, where Freshdesk " +
				"keeps it for 30 days before removing it permanently.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("ticket"),
				"subject": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Subject line of the ticket.",
				},
				"description": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Opening message of the ticket, in HTML.\n\n" +
						"~> Freshdesk rewrites the markup it is given, so this attribute keeps " +
						"what you configured and `description_text` exposes the stored plain-text " +
						"rendering. A body edited in the portal cannot be detected as drift.",
				},
				"type": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Ticket type. The allowed values are the choices " +
						"configured on the account's Type field.",
				},
				"status": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Default:  int64default.StaticInt64(freshdesk.TicketStatusOpen),
					MarkdownDescription: "Status: `2` open, `3` pending, `4` resolved, `5` closed. " +
						"Accounts may define further custom statuses. Freshdesk requires a status " +
						"on every ticket, so this defaults to `2`.",
				},
				"priority": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Default:    int64default.StaticInt64(freshdesk.TicketPriorityLow),
					Validators: []validatorInt64{ticketPriorityValidator()},
					MarkdownDescription: "Priority: `1` low, `2` medium, `3` high, `4` urgent. " +
						"Freshdesk requires a priority on every ticket, so this defaults to `1`.",
				},
				"source": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorInt64{ticketSourceValidator()},
					MarkdownDescription: "Channel the ticket arrived on: `1` email, `2` portal, " +
						"`3` phone, `7` chat, `9` feedback widget, `10` outbound email, " +
						"`11` e-commerce.",
				},
				"requester_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the contact who raised the ticket. One of " +
						"`requester_id`, `email`, `phone` or `name` identifies the requester.",
				},
				"responder_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the agent the ticket is assigned to. Freshdesk " +
						"fills this in when an automation assigns the ticket.",
				},
				"group_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the group the ticket is assigned to. Freshdesk " +
						"fills this in when an automation routes the ticket.",
				},
				"company_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the company the ticket belongs to.",
				},
				"product_id": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "ID of the product the ticket concerns. Freshdesk fills " +
						"this in from the email config when it is not given.",
				},
				"email_config_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the email config the ticket is associated with.",
				},
				"email": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Email address of the requester, when no `requester_id` is given.",
				},
				"name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Name of the requester, used when creating one.",
				},
				"phone": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Phone number of the requester, when no `requester_id` is given.",
				},
				"cc_emails": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Addresses copied on the ticket's correspondence.",
				},
				"tags": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Tags applied to the ticket.",
				},
				"custom_fields": schema.MapAttribute{
					Optional:    true,
					ElementType: types.StringType,
					MarkdownDescription: "Custom field values keyed by field name. Non-string " +
						"values are given as JSON.",
				},
				"due_by": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "When the ticket must be resolved (RFC 3339).",
				},
				"fr_due_by": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "When the first response is due (RFC 3339).",
				},
				"parent_id": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "ID of the parent ticket, for a child ticket.",
				},
				"description_text": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Plain-text rendering of `description`.",
				},
				"spam": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the ticket has been marked as spam.",
				},
				"is_escalated": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the ticket has breached its resolution SLA.",
				},
			}, timestampAttributes("ticket")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *ticketModel, d *diagnostics,
		) (*freshdesk.Ticket, error) {
			return c.CreateTicket(ctx, build(ctx, plan, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *ticketModel,
		) (*freshdesk.Ticket, error) {
			return c.GetTicket(ctx, id, freshdesk.TicketIncludes{})
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *ticketModel, d *diagnostics,
		) (*freshdesk.Ticket, error) {
			return c.UpdateTicket(ctx, id, build(ctx, plan, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *ticketModel) error {
			return c.DeleteTicket(ctx, id)
		},
	}
}
