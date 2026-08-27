package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

func TestParseID(t *testing.T) {
	t.Parallel()

	if got, err := parseID("158018521243"); err != nil || got != 158018521243 {
		t.Errorf("parseID = (%d, %v)", got, err)
	}

	for _, in := range []string{"", "abc", "1.5", "12a"} {
		if _, err := parseID(in); err == nil {
			t.Errorf("parseID(%q): want error", in)
		}
	}
}

func TestOptionalScalars(t *testing.T) {
	t.Parallel()

	if !optString("").IsNull() {
		t.Error(`optString("") must be null`)
	}

	if optString("x").ValueString() != "x" {
		t.Error(`optString("x") must carry the value`)
	}

	if !optInt64(0).IsNull() {
		t.Error("optInt64(0) must be null")
	}

	if optInt64(7).ValueInt64() != 7 {
		t.Error("optInt64(7) must carry the value")
	}

	if !timeString(freshdesk.Time{}).IsNull() {
		t.Error("a zero timestamp must be null")
	}
}

func TestPointerHelpersSkipUnknownAndNull(t *testing.T) {
	t.Parallel()

	if strPtr(types.StringNull()) != nil || strPtr(types.StringUnknown()) != nil {
		t.Error("strPtr must skip null and unknown")
	}

	if got := strPtr(types.StringValue("x")); got == nil || *got != "x" {
		t.Errorf("strPtr = %v", got)
	}

	if intPtr(types.Int64Null()) != nil || int64Ptr(types.Int64Unknown()) != nil {
		t.Error("int pointers must skip null and unknown")
	}

	if boolPtr(types.BoolNull()) != nil {
		t.Error("boolPtr must skip null")
	}

	if got := boolPtr(types.BoolValue(true)); got == nil || !*got {
		t.Errorf("boolPtr = %v", got)
	}
}

func TestClearableString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		plan, prior types.String
		want        *string
	}{
		{"set", types.StringValue("new"), types.StringValue("old"), freshdesk.Ptr("new")},
		{"removed", types.StringNull(), types.StringValue("old"), freshdesk.Ptr("")},
		{"never set", types.StringNull(), types.StringNull(), nil},
		{"was empty", types.StringNull(), types.StringValue(""), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := clearableString(tc.plan, tc.prior)
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("got %q, want nil", *got)
			case tc.want != nil && got == nil:
				t.Errorf("got nil, want %q", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Errorf("got %q, want %q", *got, *tc.want)
			}
		})
	}
}

// TestApplyCollectionsPreserveNull covers the rule that keeps an unset
// attribute unset when the API answers with an empty collection.
func TestApplyCollectionsPreserveNull(t *testing.T) {
	t.Parallel()

	if got := applyStringSet(types.SetNull(types.StringType), nil); !got.IsNull() {
		t.Error("an empty API collection must leave a null attribute null")
	}

	if got := applyStringSet(types.SetNull(types.StringType), []string{"a"}); got.IsNull() {
		t.Error("a non-empty API collection must populate the attribute")
	}

	configured := stringSet([]string{"a"})
	if got := applyStringSet(configured, nil); got.IsNull() {
		t.Error("a configured collection cleared at the API must become empty, not null")
	}

	if got := applyInt64Set(types.SetNull(types.Int64Type), nil); !got.IsNull() {
		t.Error("an empty API ID collection must leave a null attribute null")
	}

	if got := applyMap(types.MapNull(types.StringType), nil); !got.IsNull() {
		t.Error("empty custom fields must leave a null attribute null")
	}
}

func TestCustomFieldsRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fields := freshdesk.CustomFields{
		"cf_text":   "hello",
		"cf_number": float64(42),
		"cf_bool":   true,
		"cf_unset":  nil,
		"cf_object": map[string]any{"a": 1},
	}

	m := customFieldsToMap(fields)
	if m.IsNull() {
		t.Fatal("custom fields must render")
	}

	elems := m.Elements()
	if _, ok := elems["cf_unset"]; ok {
		t.Error("a null custom field must be left out of state")
	}

	for key, want := range map[string]string{"cf_text": "hello", "cf_number": "42"} {
		value, ok := elems[key].(types.String)
		if !ok {
			t.Errorf("%s is %T, want a string", key, elems[key])

			continue
		}

		if value.ValueString() != want {
			t.Errorf("%s = %q, want %q", key, value.ValueString(), want)
		}
	}

	var diags diag.Diagnostics

	back := mapToCustomFields(ctx, m, &diags)
	if diags.HasError() {
		t.Fatalf("mapToCustomFields: %v", diags)
	}

	if back["cf_text"] != "hello" {
		t.Errorf("cf_text round-tripped as %#v", back["cf_text"])
	}

	if back["cf_bool"] != true {
		t.Errorf("cf_bool round-tripped as %#v, want a real boolean", back["cf_bool"])
	}

	if _, ok := back["cf_object"].(map[string]any); !ok {
		t.Errorf("cf_object round-tripped as %#v, want an object", back["cf_object"])
	}
}

func TestSortedInt64(t *testing.T) {
	t.Parallel()

	if got := sortedInt64(nil); got != nil {
		t.Errorf("sortedInt64(nil) = %v, want nil", got)
	}

	in := []int64{3, 1, 2}

	got := sortedInt64(in)
	if got[0] != 1 || got[2] != 3 {
		t.Errorf("sortedInt64 = %v", got)
	}

	if in[0] != 3 {
		t.Error("sortedInt64 must not reorder its argument")
	}
}

func TestRawJSONString(t *testing.T) {
	t.Parallel()

	if !rawJSONString(nil).IsNull() {
		t.Error("absent raw JSON must be null")
	}

	if !rawJSONString(freshdesk.RawJSON("null")).IsNull() {
		t.Error("a JSON null must be null")
	}

	if got := rawJSONString(freshdesk.RawJSON(`{"a":1}`)).ValueString(); got != `{"a":1}` {
		t.Errorf("rawJSONString = %q", got)
	}
}

func TestParseTimestamp(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics

	if got := parseTimestamp(types.StringNull(), "due_by", &diags); got != nil {
		t.Error("a null timestamp must map to nil")
	}

	got := parseTimestamp(types.StringValue("2030-01-31T15:04:05Z"), "due_by", &diags)
	if diags.HasError() || got == nil {
		t.Fatalf("parseTimestamp: %v", diags)
	}

	if got.UTC().Year() != 2030 {
		t.Errorf("parsed year = %d", got.UTC().Year())
	}

	parseTimestamp(types.StringValue("not-a-time"), "due_by", &diags)

	if !diags.HasError() {
		t.Error("an unparseable timestamp must report a diagnostic")
	}
}

// TestEmailSemanticEquals covers the case-insensitive comparison that keeps a
// configuration written with capitals from showing perpetual drift.
func TestEmailSemanticEquals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		a, b string
		want bool
	}{
		{"Sam@Example.com", "sam@example.com", true},
		{"sam@example.com", "sam@example.com", true},
		{"sam@example.com", "other@example.com", false},
		{"", "", true},
	}
	for _, tc := range cases {
		equal, diags := emailString(tc.a).StringSemanticEquals(ctx, emailString(tc.b))
		if diags.HasError() {
			t.Fatalf("semantic equals: %v", diags)
		}

		if equal != tc.want {
			t.Errorf("%q == %q: got %v, want %v", tc.a, tc.b, equal, tc.want)
		}
	}
}

func TestEmailValueBasics(t *testing.T) {
	t.Parallel()

	if !emailString("").IsNull() {
		t.Error(`emailString("") must be null`)
	}

	ctx := context.Background()
	if _, ok := email().ValueType(ctx).(emailValue); !ok {
		t.Error("ValueType must be an emailValue")
	}

	if !email().Equal(emailType{}) {
		t.Error("emailType must equal itself")
	}

	if email().Equal(basetypes.StringType{}) {
		t.Error("emailType must not equal a plain string type")
	}

	if emailString("a").Equal(types.StringValue("a")) {
		t.Error("an emailValue must not equal a plain string value")
	}
}

// TestJSONSemanticEquals covers the comparison that stops a re-ordered API
// response reading as a change.
func TestJSONSemanticEquals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"reordered keys", `{"a":1,"b":2}`, `{"b":2,"a":1}`, true},
		{"whitespace", `{"a": 1}`, `{"a":1}`, true},
		{"nested reorder",
			`[{"name":"s","match_type":"all"}]`, `[{"match_type":"all","name":"s"}]`, true},
		{"different value", `{"a":1}`, `{"a":2}`, false},
		{"array order matters", `[1,2]`, `[2,1]`, false},
		{"invalid json compares literally", `{oops`, `{oops`, true},
		{"invalid vs valid", `{oops`, `{"a":1}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			equal, diags := jsonStringValue(tc.a).StringSemanticEquals(ctx, jsonStringValue(tc.b))
			if diags.HasError() {
				t.Fatalf("semantic equals: %v", diags)
			}

			if equal != tc.want {
				t.Errorf("%s == %s: got %v, want %v", tc.a, tc.b, equal, tc.want)
			}
		})
	}
}

func TestJSONValueBasics(t *testing.T) {
	t.Parallel()

	if !jsonStringValue("").IsNull() {
		t.Error(`jsonStringValue("") must be null`)
	}

	if !jsonEncoded(nil).IsNull() {
		t.Error("jsonEncoded(nil) must be null")
	}

	if got := jsonEncoded(map[string]int{"a": 1}).ValueString(); got != `{"a":1}` {
		t.Errorf("jsonEncoded = %q", got)
	}

	var diags diag.Diagnostics

	if got := jsonAttrPtr(jsonStringValue(""), &diags, "x"); got != nil {
		t.Error("an empty JSON attribute must decode to nil")
	}

	jsonAttrPtr(jsonStringValue("{oops"), &diags, "conditions")

	if !diags.HasError() {
		t.Error("malformed JSON must report a diagnostic")
	}
}

func TestToSlicesReadCollections(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var diags diag.Diagnostics

	if got := toStringSlice(ctx, types.SetNull(types.StringType), &diags); got != nil {
		t.Errorf("a null set must read as nil, got %v", got)
	}

	strs := toStringSlice(ctx, stringSet([]string{"a", "b"}), &diags)
	if len(strs) != 2 {
		t.Errorf("toStringSlice = %v", strs)
	}

	ids := toInt64Slice(ctx, int64Set([]int64{1, 2, 3}), &diags)
	if len(ids) != 3 {
		t.Errorf("toInt64Slice = %v", ids)
	}

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// A value of the wrong shape must be reported rather than silently ignored.
	toStringSlice(ctx, types.BoolValue(true), &diags)

	if !diags.HasError() {
		t.Error("a non-collection must report a diagnostic")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()

	if got := firstNonEmpty("", "", "third"); got != "third" {
		t.Errorf("firstNonEmpty = %q", got)
	}

	if got := firstNonEmpty("first", "second"); got != "first" {
		t.Errorf("firstNonEmpty = %q", got)
	}

	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("firstNonEmpty = %q, want empty", got)
	}
}

// ensure attr stays imported for the collection helpers above.
var _ attr.Value = types.StringNull()
