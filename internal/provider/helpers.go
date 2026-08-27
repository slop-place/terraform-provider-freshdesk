package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// ErrInvalidID is returned when a state or import ID is not a number.
var ErrInvalidID = errors.New("not a valid Freshdesk numeric ID")

// parseID converts a Terraform resource ID string into the int64 the API uses.
func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is %w", s, ErrInvalidID)
	}

	return id, nil
}

// idString renders an API identifier as a Terraform resource ID.
func idString(id int64) types.String { return types.StringValue(strconv.FormatInt(id, 10)) }

// --- scalars: API value -> Terraform value -------------------------------

// optString maps an API string to a Terraform string, treating "" as null so
// that an unset optional attribute stays unset in state.
func optString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// optInt64 maps an API integer to a Terraform integer, treating 0 as null.
// Freshdesk omits unset numeric references rather than sending null.
func optInt64(v int64) types.Int64 {
	if v == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}

// timeString renders an API timestamp as RFC 3339, or null when unset.
func timeString(t freshdesk.Time) types.String {
	if t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.String())
}

// --- scalars: Terraform value -> API value -------------------------------

// strPtr returns a pointer to the value of a known, non-null string.
func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return freshdesk.Ptr(v.ValueString())
}

// intPtr returns a pointer to the value of a known, non-null integer.
func intPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return freshdesk.Ptr(int(v.ValueInt64()))
}

// int64Ptr returns a pointer to the value of a known, non-null integer.
func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return freshdesk.Ptr(v.ValueInt64())
}

// boolPtr returns a pointer to the value of a known, non-null bool.
func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return freshdesk.Ptr(v.ValueBool())
}

// --- collections ---------------------------------------------------------

// stringSet converts a slice of API strings into a Terraform set.
func stringSet(vals []string) types.Set {
	if vals == nil {
		return types.SetNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(vals))
	for _, v := range vals {
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elems)
}

// int64Set converts a slice of API IDs into a Terraform set.
func int64Set(vals []int64) types.Set {
	if vals == nil {
		return types.SetNull(types.Int64Type)
	}
	elems := make([]attr.Value, 0, len(vals))
	for _, v := range vals {
		elems = append(elems, types.Int64Value(v))
	}
	return types.SetValueMust(types.Int64Type, elems)
}

// toStringSlice reads a Terraform set or list of strings into a Go slice.
func toStringSlice(ctx context.Context, v attr.Value, diags *diag.Diagnostics) []string {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out []string
	switch tv := v.(type) {
	case types.Set:
		diags.Append(tv.ElementsAs(ctx, &out, false)...)
	case types.List:
		diags.Append(tv.ElementsAs(ctx, &out, false)...)
	default:
		diags.AddError("Unexpected collection type",
			fmt.Sprintf("cannot read %T as a collection of strings", v))
	}
	return out
}

// toInt64Slice reads a Terraform set or list of integers into a Go slice.
func toInt64Slice(ctx context.Context, v attr.Value, diags *diag.Diagnostics) []int64 {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out []int64
	switch tv := v.(type) {
	case types.Set:
		diags.Append(tv.ElementsAs(ctx, &out, false)...)
	case types.List:
		diags.Append(tv.ElementsAs(ctx, &out, false)...)
	default:
		diags.AddError("Unexpected collection type",
			fmt.Sprintf("cannot read %T as a collection of integers", v))
	}
	return out
}

// --- custom fields -------------------------------------------------------
//
// Freshdesk custom fields are an open-ended JSON object whose value types vary
// per field. Terraform needs a concrete type, so they are surfaced as a
// map of JSON-encoded strings: a value that is a plain string stays plain,
// and anything richer round-trips as JSON.

// customFieldsToMap renders API custom fields as a Terraform string map.
func customFieldsToMap(cf freshdesk.CustomFields) types.Map {
	if cf == nil {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(cf))
	for k, v := range cf {
		if v == nil {
			// A null custom field is an unset one; keep it out of state so it
			// does not read as a value the practitioner set.
			continue
		}
		if s, ok := v.(string); ok {
			elems[k] = types.StringValue(s)
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			elems[k] = types.StringValue(fmt.Sprint(v))
			continue
		}
		elems[k] = types.StringValue(string(b))
	}
	if len(elems) == 0 {
		return types.MapNull(types.StringType)
	}
	return types.MapValueMust(types.StringType, elems)
}

// mapToCustomFields reads a Terraform string map into API custom fields.
// Values that parse as JSON are sent as their decoded form so numbers, booleans
// and nested objects reach the API with the right type.
func mapToCustomFields(ctx context.Context, m types.Map, diags *diag.Diagnostics) freshdesk.CustomFields {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	raw := map[string]string{}
	diags.Append(m.ElementsAs(ctx, &raw, false)...)
	if diags.HasError() {
		return nil
	}
	out := freshdesk.CustomFields{}
	for k, v := range raw {
		var decoded any
		if err := json.Unmarshal([]byte(v), &decoded); err == nil {
			out[k] = decoded
			continue
		}
		out[k] = v
	}
	return out
}

// --- misc ----------------------------------------------------------------

// sortedInt64 returns a copy of ids in ascending order, so that state written
// from an API response does not churn on ordering alone.
func sortedInt64(ids []int64) []int64 {
	if ids == nil {
		return nil
	}
	out := append([]int64(nil), ids...)
	slices.Sort(out)
	return out
}

// jsonString renders an arbitrary API value as a JSON string attribute.
func jsonString(v any) types.String {
	if v == nil {
		return types.StringNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	if string(b) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

// parseTimestamp reads an RFC 3339 attribute into the API's time wrapper.
func parseTimestamp(v types.String, attrName string, diags *diagnostics) *freshdesk.Time {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, v.ValueString())
	if err != nil {
		diags.AddError("Invalid timestamp in "+attrName,
			fmt.Sprintf("The value of %s must be an RFC 3339 timestamp, "+
				"for example 2026-01-31T15:04:05Z: %s", attrName, err))

		return nil
	}

	return &freshdesk.Time{Time: parsed}
}

// clearableString maps an optional text attribute into a request pointer,
// sending an explicit empty string when the practitioner removed a value that
// was previously set.
//
// Freshdesk ignores fields a request omits, so without this a value deleted
// from the configuration would silently survive on the record.
func clearableString(plan, prior types.String) *string {
	if !plan.IsNull() && !plan.IsUnknown() {
		return freshdesk.Ptr(plan.ValueString())
	}

	if prior.IsNull() || prior.IsUnknown() || prior.ValueString() == "" {
		return nil
	}

	return freshdesk.Ptr("")
}

// applyStringSet writes an API collection into state without inventing a value
// the practitioner never wrote.
//
// Freshdesk answers with an empty array for a collection that was never set, so
// copying it straight into state would turn a null attribute into an empty set
// and fail Terraform's consistency check. Keeping null when both sides are
// empty leaves the two representations in agreement.
func applyStringSet(current types.Set, vals []string) types.Set {
	if len(vals) == 0 {
		// Never configured, and still empty: leave it unset.
		if current.IsNull() {
			return current
		}

		// Configured but empty at the API: an empty set, not a null one, so
		// the difference from the configured value is visible.
		return types.SetValueMust(types.StringType, []attr.Value{})
	}

	return stringSet(vals)
}

// applyInt64Set is applyStringSet for a collection of IDs.
func applyInt64Set(current types.Set, vals []int64) types.Set {
	if len(vals) == 0 {
		if current.IsNull() {
			return current
		}

		return types.SetValueMust(types.Int64Type, []attr.Value{})
	}

	return int64Set(sortedInt64(vals))
}

// applyMap is applyStringSet for a map of custom fields.
func applyMap(current types.Map, fields freshdesk.CustomFields) types.Map {
	next := customFieldsToMap(fields)
	if next.IsNull() && current.IsNull() {
		return current
	}

	return next
}

// rawJSONString renders a raw API field as a JSON string attribute.
func rawJSONString(raw freshdesk.RawJSON) types.String {
	s := raw.String()
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

// emailPtr returns a pointer to a known, non-null email address.
func emailPtr(v emailValue) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return freshdesk.Ptr(v.ValueString())
}
