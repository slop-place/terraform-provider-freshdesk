package provider

import (
	"context"

	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*customObjectRecordResource)(nil)
	_ resource.ResourceWithConfigure   = (*customObjectRecordResource)(nil)
	_ resource.ResourceWithImportState = (*customObjectRecordResource)(nil)
)

// NewCustomObjectRecordResource returns the freshdesk_custom_object_record
// resource.
func NewCustomObjectRecordResource() resource.Resource { return &customObjectRecordResource{} }

// customObjectRecordResource manages one row of a custom object. It does not
// use the shared CRUD scaffold, because records are keyed by a string display
// ID rather than a number.
type customObjectRecordResource struct{ base }

type customObjectRecordModel struct {
	ID        types.String `tfsdk:"id"`
	SchemaID  types.String `tfsdk:"schema_id"`
	Data      types.Map    `tfsdk:"data"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// apply copies an API record into the model.
func (m *customObjectRecordModel) apply(r *freshdesk.CustomObjectRecord) {
	m.ID = types.StringValue(r.DisplayID)
	m.Data = customFieldsToMap(r.Data)
	m.CreatedAt = timeString(r.CreatedAt)
	m.UpdatedAt = timeString(r.UpdatedAt)
}

func (r *customObjectRecordResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_object_record"
}

func (r *customObjectRecordResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "One record of a custom object.\n\n" +
			"-> Custom objects require a plan that enables the feature. Accounts without " +
			"it answer `403` on these endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Display identifier Freshdesk assigned to the record, for " +
					"example `BKG-1`.",
				PlanModifiers: []planModifierString{stringplanmodifier.UseStateForUnknown()},
			},
			"schema_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "ID of the custom object schema. Changing it forces a new " +
					"record.",
				PlanModifiers: []planModifierString{stringplanmodifier.RequiresReplace()},
			},
			"data": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Field values keyed by field name. Non-string values are " +
					"given as JSON, for example `\"42\"` or `\"true\"`.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the record was created (RFC 3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the record was last changed (RFC 3339).",
			},
		},
	}
}

func (r *customObjectRecordResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan customObjectRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data := mapToCustomFields(ctx, plan.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.CreateCustomObjectRecord(ctx, plan.SchemaID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the Freshdesk custom object record", err.Error())

		return
	}

	plan.apply(record)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customObjectRecordResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state customObjectRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.GetCustomObjectRecord(ctx,
		state.SchemaID.ValueString(), state.ID.ValueString())
	if err != nil {
		if freshdesk.NotFound(err) {
			resp.State.RemoveResource(ctx)

			return
		}
		resp.Diagnostics.AddError("Unable to read the Freshdesk custom object record", err.Error())

		return
	}

	state.apply(record)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customObjectRecordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state customObjectRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data := mapToCustomFields(ctx, plan.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.UpdateCustomObjectRecord(ctx,
		plan.SchemaID.ValueString(), state.ID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update the Freshdesk custom object record", err.Error())

		return
	}

	plan.apply(record)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customObjectRecordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state customObjectRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCustomObjectRecord(ctx, state.SchemaID.ValueString(), state.ID.ValueString())
	if err != nil && !freshdesk.NotFound(err) {
		resp.Diagnostics.AddError("Unable to delete the Freshdesk custom object record", err.Error())
	}
}

func (r *customObjectRecordResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	schemaID, recordID, ok := strings.Cut(req.ID, ":")
	if !ok || schemaID == "" || recordID == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Import a custom object record as <schema_id>:<record_id>, for example "+
				"`terraform import freshdesk_custom_object_record.example "+
				"CO_123:BKG-1`. Got "+req.ID+".")

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_id"), schemaID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), recordID)...)
}
