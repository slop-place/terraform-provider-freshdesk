package provider

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// apiModel is a Terraform model that round-trips one Freshdesk API object.
// Implementations are pointer receivers on the model struct.
type apiModel[A any] interface {
	// GetID returns the resource identifier currently held in the model. It is
	// named GetID rather than ID because models carry an ID field.
	GetID() types.String
	// Apply copies an API object into the model, including its ID.
	Apply(obj *A)
}

// modelPtr constrains a pointer to a model that implements apiModel.
type modelPtr[T, A any] interface {
	*T
	apiModel[A]
}

// crud wires one Freshdesk API object into the framework's resource lifecycle.
//
// T is the Terraform model struct, A the API object. The hooks carry the parts
// that differ per resource; everything else — state plumbing, drift detection,
// import — is shared.
type crud[T, A any, PT modelPtr[T, A]] struct {
	base

	// name is the resource type suffix, e.g. "group" for freshdesk_group.
	name string
	// label names the object in diagnostics, e.g. "group".
	label string
	// schema is the resource schema.
	schema rschema.Schema

	// createFn creates the object from a plan.
	createFn func(context.Context, *freshdesk.Client, PT, *diagnostics) (*A, error)
	// readFn fetches the object by ID. It also receives the model held in
	// state, so resources whose lookup needs a parent ID — a comment's topic,
	// a section's ticket field — can reach it.
	readFn func(context.Context, *freshdesk.Client, int64, PT) (*A, error)
	// updateFn applies a plan to an existing object. prior is the state being
	// replaced, so the hook can detect attributes the practitioner removed.
	updateFn func(context.Context, *freshdesk.Client, int64, PT, PT, *diagnostics) (*A, error)
	// deleteFn removes the object. It receives the model held in state so a
	// resource whose delete needs a parent ID can reach it. A nil deleteFn
	// means the object cannot be deleted through the API and is simply
	// dropped from state.
	deleteFn func(context.Context, *freshdesk.Client, int64, PT) error
	// importFn overrides the default numeric-ID import, for resources whose
	// identity needs more than one value.
	importFn func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
}

func (r *crud[T, A, PT]) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_" + r.name
}

func (r *crud[T, A, PT]) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema
}

func (r *crud[T, A, PT]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.createFn(ctx, r.client, PT(&model), &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the Freshdesk "+r.label, err.Error())

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	PT(&model).Apply(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *crud[T, A, PT]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, ok := readID(PT(&model).GetID(), r.label, &resp.Diagnostics)
	if !ok {
		return
	}

	obj, err := r.readFn(ctx, r.client, id, PT(&model))
	if err != nil {
		if freshdesk.NotFound(err) {
			// Deleted outside Terraform: drop it so the next plan recreates it.
			resp.State.RemoveResource(ctx)

			return
		}
		resp.Diagnostics.AddError("Unable to read the Freshdesk "+r.label, err.Error())

		return
	}

	PT(&model).Apply(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *crud[T, A, PT]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, ok := readID(PT(&state).GetID(), r.label, &resp.Diagnostics)
	if !ok {
		return
	}

	obj, err := r.updateFn(ctx, r.client, id, PT(&plan), PT(&state), &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update the Freshdesk "+r.label, err.Error())

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	PT(&plan).Apply(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *crud[T, A, PT]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.deleteFn == nil {
		// Nothing to do at the API; the framework removes it from state.
		return
	}

	id, ok := readID(PT(&model).GetID(), r.label, &resp.Diagnostics)
	if !ok {
		return
	}

	err := r.deleteFn(ctx, r.client, id, PT(&model))
	switch {
	case err == nil:
	// An already-absent object is a successful delete.
	case freshdesk.NotFound(err):
	default:
		var unsupported *unsupportedDeleteError
		if errors.As(err, &unsupported) {
			resp.Diagnostics.AddWarning(
				"The Freshdesk API cannot delete this "+r.label, unsupported.Error())

			return
		}
		resp.Diagnostics.AddError("Unable to delete the Freshdesk "+r.label, err.Error())
	}
}

func (r *crud[T, A, PT]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if r.importFn != nil {
		r.importFn(ctx, req, resp)

		return
	}

	if _, err := parseID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(
			"Import a %s by its numeric ID, for example "+
				"`terraform import freshdesk_%s.example 42`: %s", r.label, r.name, err))

		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readID converts a model's id attribute into the numeric ID the API uses.
func readID(id types.String, label string, diags *diagnostics) (int64, bool) {
	parsed, err := parseID(id.ValueString())
	if err != nil {
		diags.AddError("Invalid "+label+" ID in state", err.Error())

		return 0, false
	}

	return parsed, true
}

// idAttribute is the computed identifier every resource carries.
func idAttribute(label string) rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Numeric identifier of the " + label + ".",
		PlanModifiers:       []planModifierString{useStateForUnknown()},
	}
}

// timestampAttributes are the created/updated pair most objects carry.
func timestampAttributes(label string) map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"created_at": rschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "When the " + label + " was created (RFC 3339).",
		},
		"updated_at": rschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "When the " + label + " was last changed (RFC 3339).",
		},
	}
}

// withAttributes merges extra attributes into a base map, returning a new map.
func withAttributes(
	base map[string]rschema.Attribute,
	extra ...map[string]rschema.Attribute,
) map[string]rschema.Attribute {
	out := make(map[string]rschema.Attribute, len(base))
	maps.Copy(out, base)

	for _, m := range extra {
		maps.Copy(out, m)
	}

	return out
}

// importCompositeID builds an import handler for a resource whose identity
// needs more than one value, given as colon-separated parts.
//
// The parts are written to the named attributes in order; every part except the
// final ID is set as an int64, matching the parent-reference attributes.
func importCompositeID(
	resourceType string,
	attrs []string,
) func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse) {
	return func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
		parts := strings.Split(req.ID, ":")
		if len(parts) != len(attrs) {
			resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(
				"Import %s as %s, for example `terraform import %s.example %s`. Got %q.",
				resourceType, strings.Join(placeholders(attrs), ":"), resourceType,
				strings.Join(exampleParts(attrs), ":"), req.ID))

			return
		}

		for i, name := range attrs {
			value, err := parseID(parts[i])
			if err != nil {
				resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(
					"The %s part of the import ID must be numeric: %s", name, err))

				return
			}

			if name == "id" {
				resp.Diagnostics.Append(
					resp.State.SetAttribute(ctx, path.Root(name), strconv.FormatInt(value, 10))...)

				continue
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), value)...)
		}
	}
}

// placeholders renders attribute names as <name> for an error message.
func placeholders(attrs []string) []string {
	out := make([]string, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, "<"+a+">")
	}

	return out
}

// exampleIDStep spaces the stand-in IDs in an import example apart, so the
// parts of a composite ID read as clearly distinct values.
const exampleIDStep = 11

// exampleParts renders stand-in values for an import example.
func exampleParts(attrs []string) []string {
	out := make([]string, 0, len(attrs))
	for i := range attrs {
		out = append(out, strconv.Itoa((i+1)*exampleIDStep))
	}

	return out
}

// deleteUnsupported builds a deleteFn for an object Freshdesk offers no way to
// remove. Terraform drops the resource from state either way; the warning makes
// sure the practitioner knows the record itself is still there.
func deleteUnsupported[T, A any, PT modelPtr[T, A]](
	label, where string,
) func(context.Context, *freshdesk.Client, int64, PT) error {
	return func(_ context.Context, _ *freshdesk.Client, _ int64, _ PT) error {
		return &unsupportedDeleteError{label: label, where: where}
	}
}

// unsupportedDeleteError reports that the API has no delete for an object.
type unsupportedDeleteError struct{ label, where string }

func (e *unsupportedDeleteError) Error() string {
	return fmt.Sprintf(
		"the Freshdesk API provides no way to delete a %s: it has been removed from "+
			"Terraform state, but the record itself remains and must be deleted %s",
		e.label, e.where)
}
