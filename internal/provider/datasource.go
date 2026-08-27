package provider

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// entity describes how one Freshdesk object is presented to Terraform. A single
// entity powers both the singular and the plural data source for that object,
// so their attributes cannot drift apart.
type entity[A any] struct {
	// label names the object in documentation and diagnostics.
	label string
	// attrs are the object's computed attributes, excluding "id".
	attrs map[string]dschema.Attribute
	// types mirrors attrs as attribute types, excluding "id".
	types map[string]attr.Type
	// mapFn renders one API object as attribute values, excluding "id".
	mapFn func(*A) map[string]attr.Value
	// idFn renders the object's identifier.
	idFn func(*A) types.String
}

// objectType is the full object type of the entity, including "id".
func (e entity[A]) objectType() types.ObjectType {
	attrTypes := make(map[string]attr.Type, len(e.types)+1)
	maps.Copy(attrTypes, e.types)
	attrTypes["id"] = types.StringType

	return types.ObjectType{AttrTypes: attrTypes}
}

// object renders one API object as a Terraform object value.
func (e entity[A]) object(obj *A) types.Object {
	values := make(map[string]attr.Value, len(e.types)+1)
	maps.Copy(values, e.mapFn(obj))
	values["id"] = e.idFn(obj)

	return types.ObjectValueMust(e.objectType().AttrTypes, values)
}

// --- singular data sources -----------------------------------------------

var _ datasource.DataSourceWithConfigure = (*lookupDataSource[any])(nil)

// lookupDataSource reads one object, selected by its numeric ID.
type lookupDataSource[A any] struct {
	dataSourceBase

	// name is the data source type suffix, e.g. "group".
	name string
	// description introduces the data source in the documentation.
	description string
	// entity describes the object's attributes.
	entity entity[A]
	// getFn fetches the object by ID.
	getFn func(context.Context, *freshdesk.Client, int64) (*A, error)
}

func (d *lookupDataSource[A]) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *lookupDataSource[A]) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	attrs := make(map[string]dschema.Attribute, len(d.entity.attrs)+1)
	maps.Copy(attrs, d.entity.attrs)
	attrs["id"] = dschema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Numeric identifier of the " + d.entity.label + " to look up.",
	}

	resp.Schema = dschema.Schema{
		MarkdownDescription: d.description,
		Attributes:          attrs,
	}
}

func (d *lookupDataSource[A]) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	numericID, ok := readID(id, d.entity.label, &resp.Diagnostics)
	if !ok {
		return
	}

	obj, err := d.getFn(ctx, d.client, numericID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Freshdesk "+d.entity.label, err.Error())

		return
	}

	setObjectState(ctx, resp, d.entity, obj)
}

// setObjectState writes an entity's attributes into a data source's state.
func setObjectState[A any](
	ctx context.Context,
	resp *datasource.ReadResponse,
	e entity[A],
	obj *A,
) {
	for name, value := range e.mapFn(obj) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), value)...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), e.idFn(obj))...)
}

// --- plural data sources -------------------------------------------------

var _ datasource.DataSourceWithConfigure = (*collectionDataSource[any])(nil)

// collectionDataSource reads every object of one kind.
type collectionDataSource[A any] struct {
	dataSourceBase

	// name is the data source type suffix, e.g. "groups".
	name string
	// description introduces the data source in the documentation.
	description string
	// itemsAttr is the attribute holding the collection, e.g. "groups".
	itemsAttr string
	// entity describes each element's attributes.
	entity entity[A]
	// listFn fetches every object.
	listFn func(context.Context, *freshdesk.Client) ([]A, error)
	// extraAttrs are optional input attributes, such as a parent ID filter.
	extraAttrs map[string]dschema.Attribute
}

func (d *collectionDataSource[A]) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *collectionDataSource[A]) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	elemAttrs := make(map[string]dschema.Attribute, len(d.entity.attrs)+1)
	maps.Copy(elemAttrs, d.entity.attrs)
	elemAttrs["id"] = dschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Numeric identifier of the " + d.entity.label + ".",
	}

	attrs := map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Placeholder identifier for the data source. Always " +
				"`" + d.itemsAttr + "`.",
		},
		d.itemsAttr: dschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "Every " + d.entity.label + " the API returned.",
			NestedObject:        dschema.NestedAttributeObject{Attributes: elemAttrs},
		},
	}

	maps.Copy(attrs, d.extraAttrs)

	resp.Schema = dschema.Schema{
		MarkdownDescription: d.description,
		Attributes:          attrs,
	}
}

func (d *collectionDataSource[A]) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	objects, err := d.listFn(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to list the Freshdesk %ss", d.entity.label), err.Error())

		return
	}

	elems := make([]attr.Value, 0, len(objects))
	for i := range objects {
		elems = append(elems, d.entity.object(&objects[i]))
	}

	list, diags := types.ListValue(d.entity.objectType(), elems)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"),
		types.StringValue(d.itemsAttr))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(d.itemsAttr), list)...)
}

// --- shared attribute helpers -------------------------------------------

// computedString is a read-only string attribute.
func computedString(description string) dschema.StringAttribute {
	return dschema.StringAttribute{Computed: true, MarkdownDescription: description}
}

// computedInt64 is a read-only integer attribute.
func computedInt64(description string) dschema.Int64Attribute {
	return dschema.Int64Attribute{Computed: true, MarkdownDescription: description}
}

// computedBool is a read-only boolean attribute.
func computedBool(description string) dschema.BoolAttribute {
	return dschema.BoolAttribute{Computed: true, MarkdownDescription: description}
}

// computedStringSet is a read-only set of strings.
func computedStringSet(description string) dschema.SetAttribute {
	return dschema.SetAttribute{
		Computed:            true,
		ElementType:         types.StringType,
		MarkdownDescription: description,
	}
}

// computedInt64Set is a read-only set of integers.
func computedInt64Set(description string) dschema.SetAttribute {
	return dschema.SetAttribute{
		Computed:            true,
		ElementType:         types.Int64Type,
		MarkdownDescription: description,
	}
}

// computedStringMap is a read-only map of strings.
func computedStringMap(description string) dschema.MapAttribute {
	return dschema.MapAttribute{
		Computed:            true,
		ElementType:         types.StringType,
		MarkdownDescription: description,
	}
}

// timestampDataAttributes are the created/updated pair on a read-only object.
func timestampDataAttributes(label string) map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"created_at": computedString("When the " + label + " was created (RFC 3339)."),
		"updated_at": computedString("When the " + label + " was last changed (RFC 3339)."),
	}
}

// timestampDataTypes mirrors timestampDataAttributes as attribute types.
func timestampDataTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"created_at": types.StringType,
		"updated_at": types.StringType,
	}
}

// mergeAttrs combines attribute maps into a new map.
func mergeAttrs(sources ...map[string]dschema.Attribute) map[string]dschema.Attribute {
	out := map[string]dschema.Attribute{}
	for _, m := range sources {
		maps.Copy(out, m)
	}

	return out
}

// mergeTypes combines attribute type maps into a new map.
func mergeTypes(sources ...map[string]attr.Type) map[string]attr.Type {
	out := map[string]attr.Type{}
	for _, m := range sources {
		maps.Copy(out, m)
	}

	return out
}

// mergeValues combines attribute value maps into a new map.
func mergeValues(sources ...map[string]attr.Value) map[string]attr.Value {
	out := map[string]attr.Value{}
	for _, m := range sources {
		maps.Copy(out, m)
	}

	return out
}

// timestampValues renders the created/updated pair.
func timestampValues(created, updated freshdesk.Time) map[string]attr.Value {
	return map[string]attr.Value{
		"created_at": timeString(created),
		"updated_at": timeString(updated),
	}
}

// attrValue aliases the framework's attribute value interface.
type attrValue = attr.Value

// pathRoot names a top-level attribute.
func pathRoot(name string) path.Path { return path.Root(name) }
