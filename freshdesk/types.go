package freshdesk

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time wraps time.Time so the several timestamp spellings Freshdesk emits all
// decode cleanly.
type Time struct{ time.Time }

// freshdeskTimeLayouts lists every timestamp format observed from the API.
//
//nolint:gochecknoglobals // an immutable lookup table
var freshdeskTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// UnmarshalJSON accepts any of the layouts Freshdesk uses, plus null.
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}
	for _, layout := range freshdeskTimeLayouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	// Fall back to epoch seconds, which a few endpoints use.
	if secs, err := strconv.ParseInt(s, 10, 64); err == nil {
		t.Time = time.Unix(secs, 0).UTC()
		return nil
	}
	return &time.ParseError{Layout: time.RFC3339, Value: s}
}

// MarshalJSON renders the timestamp as RFC 3339, or null when zero.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	b, err := json.Marshal(t.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("encoding freshdesk timestamp: %w", err)
	}

	return b, nil
}

// String renders the timestamp as RFC 3339, or "" when zero.
func (t Time) String() string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// CustomFields holds an account's custom field values, keyed by field name.
type CustomFields map[string]any

// Attachment is a file attached to a ticket, note, or article.
type Attachment struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ContentType   string `json:"content_type"`
	Size          int64  `json:"size"`
	AttachmentURL string `json:"attachment_url"`
	CreatedAt     Time   `json:"created_at"`
	UpdatedAt     Time   `json:"updated_at"`
}

// Avatar is a profile image on an agent or contact.
type Avatar struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ContentType   string `json:"content_type"`
	Size          int64  `json:"size"`
	AvatarURL     string `json:"avatar_url"`
	AttachmentURL string `json:"attachment_url"`
	CreatedAt     Time   `json:"created_at"`
	UpdatedAt     Time   `json:"updated_at"`
}

// RawJSON is a field whose shape varies too much to type.
//
// Freshdesk's `choices` is the clearest example: a custom dropdown returns an
// array of Choice objects, but the built-in fields return a string-to-string
// map (time_zone), a string-to-number map (priority), a string-to-array map
// (status), or a plain array of strings (health_score). RawJSON keeps whatever
// arrived so nothing is lost, and Choices decodes the array form when that is
// what the caller needs.
type RawJSON json.RawMessage

// MarshalJSON writes the value through unchanged.
func (r RawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}

	return r, nil
}

// UnmarshalJSON keeps the value verbatim.
func (r *RawJSON) UnmarshalJSON(b []byte) error {
	*r = append((*r)[:0], b...)

	return nil
}

// String renders the raw value, or "" when it is absent or null.
func (r RawJSON) String() string {
	s := strings.TrimSpace(string(r))
	if s == "null" {
		return ""
	}

	return s
}

// Choices decodes the value as the array of options a custom dropdown field
// uses. It returns nil for the built-in fields, whose choices take one of the
// map shapes instead.
func (r RawJSON) Choices() []Choice {
	if len(r) == 0 {
		return nil
	}

	var out []Choice
	if err := json.Unmarshal(r, &out); err != nil {
		return nil
	}

	return out
}

// Choice is one option of a dropdown-style field.
type Choice struct {
	ID       int64    `json:"id,omitempty"`
	Value    string   `json:"value,omitempty"`
	Label    string   `json:"label,omitempty"`
	Position int      `json:"position,omitempty"`
	Choices  []Choice `json:"choices,omitempty"`
}

// Ptr returns a pointer to v. It is exported for callers assembling the
// pointer-valued optional fields on the *Request types.
func Ptr[T any](v T) *T { return &v }

// ID is a Freshdesk numeric identifier.
//
// Freshdesk is not consistent about how it encodes these: most endpoints emit a
// JSON number, but business_hours and sla_policies emit the same value as a
// quoted string. ID accepts either spelling on the wire and always writes a
// number, so callers can treat every numeric ID uniformly.
type ID int64

// UnmarshalJSON accepts a JSON number, a quoted number, or null.
func (i *ID) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*i = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing freshdesk id %s: %w", b, err)
	}
	*i = ID(n)
	return nil
}

// MarshalJSON writes the identifier as a JSON number.
func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

// Int64 returns the identifier as a plain int64.
func (i ID) Int64() int64 { return int64(i) }

// String renders the identifier in base 10.
func (i ID) String() string { return strconv.FormatInt(int64(i), 10) }

// FlexString is a string field that Freshdesk sometimes sends as a number.
//
// org_company_id is the clearest case: the create response encodes it as a
// JSON number while the list response quotes it. FlexString accepts either and
// always presents a string.
type FlexString string

// UnmarshalJSON accepts a JSON string, a JSON number, or null.
func (f *FlexString) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*f = ""

		return nil
	}

	if strings.HasPrefix(s, `"`) {
		var unquoted string
		if err := json.Unmarshal(b, &unquoted); err != nil {
			return fmt.Errorf("parsing freshdesk string field: %w", err)
		}
		*f = FlexString(unquoted)

		return nil
	}

	// A bare number, object or array is kept in its literal form.
	*f = FlexString(s)

	return nil
}

// MarshalJSON writes the value as a JSON string.
func (f FlexString) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(string(f))
	if err != nil {
		return nil, fmt.Errorf("encoding freshdesk string field: %w", err)
	}

	return b, nil
}

// String returns the underlying value.
func (f FlexString) String() string { return string(f) }
