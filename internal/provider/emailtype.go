package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Freshdesk stores email addresses folded to lower case. A configuration
// written with capitals would therefore always disagree with the value the API
// returns, and Terraform would report perpetual drift.
//
// emailType solves this with semantic equality: state keeps whatever Freshdesk
// returned, configuration keeps whatever the practitioner wrote, and the two
// compare equal when they differ only by case. Genuine changes still show up as
// a diff.

// Sentinel errors for the custom-type conversions the framework drives.
var (
	// ErrNotAString is returned when the framework hands over a non-string.
	ErrNotAString = errors.New("expected a string value")
	// ErrConvertingValue is returned when a value conversion reports
	// diagnostics rather than an error.
	ErrConvertingValue = errors.New("converting the string value")
)

var (
	_ basetypes.StringTypable                    = emailType{}
	_ basetypes.StringValuable                   = emailValue{}
	_ basetypes.StringValuableWithSemanticEquals = emailValue{}
)

// emailType is the Terraform type of a case-insensitive email attribute.
type emailType struct{ basetypes.StringType }

// email returns the attribute type to use for an email address.
func email() emailType { return emailType{} }

func (t emailType) Equal(o attr.Type) bool {
	other, ok := o.(emailType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t emailType) String() string { return "provider.emailType" }

func (t emailType) ValueFromString(
	_ context.Context,
	in basetypes.StringValue,
) (basetypes.StringValuable, diag.Diagnostics) {
	return emailValue{StringValue: in}, nil
}

func (t emailType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t emailType) ValueType(_ context.Context) attr.Value { return emailValue{} }

// emailValue is one email address.
type emailValue struct{ basetypes.StringValue }

// emailString builds an emailValue, treating "" as null.
func emailString(s string) emailValue {
	if s == "" {
		return emailValue{StringValue: basetypes.NewStringNull()}
	}

	return emailValue{StringValue: basetypes.NewStringValue(s)}
}

func (v emailValue) Equal(o attr.Value) bool {
	other, ok := o.(emailValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v emailValue) Type(_ context.Context) attr.Type { return emailType{} }

// StringSemanticEquals treats two addresses that differ only by case as equal,
// which is exactly how Freshdesk treats them.
func (v emailValue) StringSemanticEquals(
	_ context.Context,
	newValuable basetypes.StringValuable,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(emailValue)
	if !ok {
		diags.AddError("Unexpected value type",
			fmt.Sprintf("Expected an emailValue, got %T. This is a bug in the provider.",
				newValuable))

		return false, diags
	}

	return strings.EqualFold(v.ValueString(), newValue.ValueString()), diags
}
