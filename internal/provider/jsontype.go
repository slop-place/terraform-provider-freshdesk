package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Several Freshdesk objects carry a body whose shape depends on the fields it
// references — an automation rule's conditions, a skill's match rules, an SLA
// policy's escalation table. Those are surfaced as JSON strings.
//
// The API re-serialises what it is given, so key order and whitespace come back
// changed even when nothing else did. jsonType compares such attributes by
// their parsed value instead of their text, so a reordering is not a diff while
// a genuine change still is.

var (
	_ basetypes.StringTypable                    = jsonType{}
	_ basetypes.StringValuable                   = jsonValueType{}
	_ basetypes.StringValuableWithSemanticEquals = jsonValueType{}
)

// jsonType is the Terraform type of a JSON-carrying string attribute.
type jsonType struct{ basetypes.StringType }

// jsonAttr returns the attribute type to use for a JSON body.
func jsonAttr() jsonType { return jsonType{} }

func (t jsonType) Equal(o attr.Type) bool {
	other, ok := o.(jsonType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t jsonType) String() string { return "provider.jsonType" }

func (t jsonType) ValueFromString(
	_ context.Context,
	in basetypes.StringValue,
) (basetypes.StringValuable, diag.Diagnostics) {
	return jsonValueType{StringValue: in}, nil
}

func (t jsonType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("converting the Terraform value: %w", err)
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", ErrNotAString, attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("%w: %v", ErrConvertingValue, diags)
	}

	return stringValuable, nil
}

func (t jsonType) ValueType(_ context.Context) attr.Value { return jsonValueType{} }

// jsonValueType is one JSON body.
type jsonValueType struct{ basetypes.StringValue }

// jsonStringValue builds a jsonValueType, treating "" as null.
func jsonStringValue(s string) jsonValueType {
	if s == "" {
		return jsonValueType{StringValue: basetypes.NewStringNull()}
	}

	return jsonValueType{StringValue: basetypes.NewStringValue(s)}
}

// jsonEncoded renders an API value as a JSON attribute, or null when absent.
func jsonEncoded(v any) jsonValueType {
	s := jsonString(v)
	if s.IsNull() {
		return jsonValueType{StringValue: basetypes.NewStringNull()}
	}

	return jsonStringValue(s.ValueString())
}

func (v jsonValueType) Equal(o attr.Value) bool {
	other, ok := o.(jsonValueType)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v jsonValueType) Type(_ context.Context) attr.Type { return jsonType{} }

// StringSemanticEquals compares the two bodies as parsed JSON, so a difference
// in key order or whitespace is not treated as a change.
func (v jsonValueType) StringSemanticEquals(
	_ context.Context,
	newValuable basetypes.StringValuable,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(jsonValueType)
	if !ok {
		diags.AddError("Unexpected value type",
			fmt.Sprintf("Expected a jsonValueType, got %T. This is a bug in the provider.",
				newValuable))

		return false, diags
	}

	oldNormalised, oldErr := normaliseJSON(v.ValueString())
	newNormalised, newErr := normaliseJSON(newValue.ValueString())

	// If either side is not valid JSON, fall back to comparing the text; the
	// schema validator reports the malformed value separately.
	if oldErr != nil || newErr != nil {
		return v.ValueString() == newValue.ValueString(), diags
	}

	return bytes.Equal(oldNormalised, newNormalised), diags
}

// normaliseJSON re-encodes a JSON document with sorted keys and no whitespace,
// which is what makes two orderings comparable.
func normaliseJSON(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(s), &decoded); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// encoding/json sorts map keys, so a plain re-encode is enough.
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return nil, fmt.Errorf("encoding JSON: %w", err)
	}

	return encoded, nil
}

// jsonAttrPtr reads a JSON attribute into a decoded value.
func jsonAttrPtr(v jsonValueType, diags *diagnostics, attrName string) any {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}

	var out any
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		diags.AddError("Invalid JSON in "+attrName,
			fmt.Sprintf("The value of %s must be valid JSON: %s", attrName, err))

		return nil
	}

	return out
}
