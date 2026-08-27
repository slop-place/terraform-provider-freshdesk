package freshdesk

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeUnmarshalAcceptsFreshdeskLayouts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string // RFC3339 in UTC, or "" for zero
	}{
		{`"2015-08-28T11:47:58Z"`, "2015-08-28T11:47:58Z"},
		{`"2014-01-08T07:53:41+05:30"`, "2014-01-08T02:23:41Z"},
		{`"2019-12-18T11:45:25.017Z"`, "2019-12-18T11:45:25Z"},
		{`"2020-12-31"`, "2020-12-31T00:00:00Z"},
		{`"2015-08-24 13:49:37"`, "2015-08-24T13:49:37Z"},
		{`null`, ""},
		{`""`, ""},
	}
	for _, tc := range cases {
		var got Time
		if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", tc.in, err)
			continue
		}
		if tc.want == "" {
			if !got.IsZero() {
				t.Errorf("Unmarshal(%s) = %v, want zero", tc.in, got)
			}
			continue
		}
		if got.UTC().Format(time.RFC3339) != tc.want {
			t.Errorf("Unmarshal(%s) = %s, want %s", tc.in, got.UTC().Format(time.RFC3339), tc.want)
		}
	}
}

func TestTimeUnmarshalRejectsGarbage(t *testing.T) {
	t.Parallel()

	var got Time
	if err := json.Unmarshal([]byte(`"not-a-time"`), &got); err == nil {
		t.Fatal("want error for unparseable timestamp")
	}
}

func TestTimeMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	var zero Time
	b, err := json.Marshal(zero)
	if err != nil {
		t.Fatalf("marshal zero: %v", err)
	}
	if string(b) != "null" {
		t.Errorf("zero Time marshals to %s, want null", b)
	}
	if zero.String() != "" {
		t.Errorf("zero Time String() = %q, want empty", zero.String())
	}

	var ts Time
	if err := json.Unmarshal([]byte(`"2015-08-28T11:47:58Z"`), &ts); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	b, err = json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `"2015-08-28T11:47:58Z"` {
		t.Errorf("marshal = %s", b)
	}
	if ts.String() != "2015-08-28T11:47:58Z" {
		t.Errorf("String() = %q", ts.String())
	}
}

// TestIDAcceptsNumberOrString covers the inconsistency between endpoints:
// business_hours and sla_policies quote their IDs, everything else does not.
func TestIDAcceptsNumberOrString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want int64
	}{
		{`158000272633`, 158000272633},
		{`"158000272633"`, 158000272633},
		{`0`, 0},
		{`null`, 0},
		{`""`, 0},
	}
	for _, tc := range cases {
		var got ID
		if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", tc.in, err)
			continue
		}
		if got.Int64() != tc.want {
			t.Errorf("Unmarshal(%s) = %d, want %d", tc.in, got.Int64(), tc.want)
		}
	}
}

func TestIDRejectsNonNumeric(t *testing.T) {
	t.Parallel()

	var got ID
	if err := json.Unmarshal([]byte(`"abc"`), &got); err == nil {
		t.Fatal("want error for non-numeric id")
	}
}

func TestIDMarshalsAsNumber(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(ID(158000272633))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "158000272633" {
		t.Errorf("marshal = %s, want unquoted number", b)
	}
	if ID(42).String() != "42" {
		t.Errorf("String() = %q", ID(42).String())
	}
}

// TestBusinessHoursDecodesStringID pins the real shape returned by the API.
func TestBusinessHoursDecodesStringID(t *testing.T) {
	t.Parallel()

	const body = `{"id":"158000272633","name":"General working hours",
	"is_default":true,"time_zone":"Pacific Time (US & Canada)",
	"business_hours":{"monday":{"start_time":"08:00:00 am","end_time":"05:00:00 pm"}},
	"source_business_hours":[{"business_hours_type":"custom","is_default":true,
	  "business_hours":[{"day":"monday","time_slots":[{"start_time":"08:00","end_time":"17:00"}]}],
	  "sources":[]}]}`
	var bh BusinessHours
	if err := json.Unmarshal([]byte(body), &bh); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bh.ID.Int64() != 158000272633 {
		t.Errorf("ID = %d", bh.ID.Int64())
	}
	if bh.BusinessHours["monday"].StartTime != "08:00:00 am" {
		t.Errorf("monday = %+v", bh.BusinessHours["monday"])
	}
	if len(bh.SourceBusinessHours) != 1 ||
		bh.SourceBusinessHours[0].BusinessHours[0].TimeSlots[0].EndTime != "17:00" {
		t.Errorf("source_business_hours = %+v", bh.SourceBusinessHours)
	}
}

// TestSLAPolicyDecodesStringID pins the SLA policy shape.
func TestSLAPolicyDecodesStringID(t *testing.T) {
	t.Parallel()

	const body = `{"id":"158000322059","name":"Default policy","active":true,
	"sla_target":{"priority_4":{"business_hours":true,"escalation_enabled":true,
	  "respond_within":30,"resolve_within":900,"next_respond_within":30}},
	"applicable_to":{"sources":[15,17]},"is_default":true,"position":1}`
	var p SLAPolicy
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ID.Int64() != 158000322059 {
		t.Errorf("ID = %d", p.ID.Int64())
	}
	if p.SLATarget["priority_4"].ResolveWithin != 900 {
		t.Errorf("priority_4 = %+v", p.SLATarget["priority_4"])
	}
	if len(p.ApplicableTo.Sources) != 2 {
		t.Errorf("sources = %v", p.ApplicableTo.Sources)
	}
}

// TestSurveyDecodesUUID pins the CSAT survey shape: a UUID identifier and a
// state string rather than a boolean.
func TestSurveyDecodesUUID(t *testing.T) {
	t.Parallel()

	const body = `{"id":"1e8f3a9c-5299-4732-949d-7fa9eea1e8c8","title":"Basic CSAT Survey",
	"description":"","header_message":"Thank you for your time.","state":"ACTIVE",
	"language":"en","questions":[{"type":"range","id":"Q_1","question_identifier":"CSAT",
	  "text":"How satisfied?","required":true}]}`

	var s Survey
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if s.ID != "1e8f3a9c-5299-4732-949d-7fa9eea1e8c8" {
		t.Errorf("ID = %q", s.ID)
	}

	if !s.Active() {
		t.Errorf("state %q should read as active", s.State)
	}

	if len(s.Questions) != 1 || s.Questions[0].QuestionIdentifier != "CSAT" {
		t.Errorf("questions = %+v", s.Questions)
	}

	var inactive Survey
	if err := json.Unmarshal([]byte(`{"state":"INACTIVE"}`), &inactive); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if inactive.Active() {
		t.Error("an INACTIVE survey must not read as active")
	}
}

// TestPageCaps pins the endpoints whose per_page limit differs from the
// documented maximum.
func TestPageCaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path    string
		size    int
		allowed bool
	}{
		{"groups", MaxPerPage, true},
		{"ticket-forms", ticketFormsPageCap, true},
		{"/ticket-forms", ticketFormsPageCap, true},
		{"admin/ticket_fields", 0, false},
	}
	for _, tc := range cases {
		size, allowed := pageCapFor(tc.path)
		if size != tc.size || allowed != tc.allowed {
			t.Errorf("pageCapFor(%q) = (%d, %v), want (%d, %v)",
				tc.path, size, allowed, tc.size, tc.allowed)
		}
	}

	if v := (ListOptions{}).valuesFor("admin/ticket_fields"); v.Has("per_page") {
		t.Errorf("admin/ticket_fields must not carry per_page, got %q", v.Get("per_page"))
	}

	if got := (ListOptions{PerPage: 100}).valuesFor("ticket-forms").Get("per_page"); got != "50" {
		t.Errorf("ticket-forms per_page = %q, want 50", got)
	}
}

func TestStructToMapPreservesLargeInts(t *testing.T) {
	t.Parallel()

	m, err := structToMap(struct {
		ID int64 `json:"id"`
	}{158018521243})
	if err != nil {
		t.Fatalf("structToMap: %v", err)
	}
	b, err := marshalMap(m)
	if err != nil {
		t.Fatalf("marshalMap: %v", err)
	}
	if string(b) != `{"id":158018521243}` {
		t.Errorf("marshalMap = %s, want exact integer", b)
	}
}

func TestPathForEscapesStringIDs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		collection string
		id         any
		want       string
	}{
		{"groups", int64(1), "groups/1"},
		{"groups", 2, "groups/2"},
		{"jobs", "abc-123", "jobs/abc-123"},
		{"records", "BKG 1", "records/BKG%201"},
		{"records", "a/b", "records/a%2Fb"},
	}
	for _, tc := range cases {
		if got := pathFor(tc.collection, tc.id); got != tc.want {
			t.Errorf("pathFor(%q, %v) = %q, want %q", tc.collection, tc.id, got, tc.want)
		}
	}
}

func TestQuoteQuery(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{`priority:3`, `"priority:3"`},
		{`"priority:3"`, `"priority:3"`},
		{`  status:2  `, `"status:2"`},
	}
	for _, tc := range cases {
		if got := quoteQuery(tc.in); got != tc.want {
			t.Errorf("quoteQuery(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
