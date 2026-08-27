package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// The Apply methods and request builders are pure functions, so they can be
// checked directly rather than only through an acceptance run.

// firstInt64 reads the first element of a set of IDs.
func firstInt64(t *testing.T, set types.Set) int64 {
	t.Helper()

	elems := set.Elements()
	if len(elems) == 0 {
		t.Fatal("the set is empty")
	}

	value, ok := elems[0].(types.Int64)
	if !ok {
		t.Fatalf("the first element is %T, want an integer", elems[0])
	}

	return value.ValueInt64()
}

func TestGroupModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var m groupModel

	m.Apply(&freshdesk.Group{
		ID: 158000001, Name: "Billing", Description: "", EscalateTo: 42,
		UnassignedFor: "1h", AgentIDs: []int64{3, 1, 2}, AutoTicketAssign: 1,
	})

	if m.GetID().ValueString() != "158000001" {
		t.Errorf("id = %q", m.GetID().ValueString())
	}

	if !m.Description.IsNull() {
		t.Error("an empty description must land as null")
	}

	if got := firstInt64(t, m.AgentIDs); got != 1 {
		t.Errorf("agent_ids starts at %d, want ascending order from 1", got)
	}

	if n := len(m.AgentIDs.Elements()); n != 3 {
		t.Errorf("agent_ids has %d entries, want 3", n)
	}

	var diags diag.Diagnostics

	// Removing the agents and the escalation target must clear them, which
	// Freshdesk only does when told explicitly.
	plan := groupModel{Name: types.StringValue("Billing")}

	req := groupRequest(ctx, &plan, &m, &diags)
	if !req.ClearAgents {
		t.Error("removing agent_ids must set ClearAgents")
	}

	if !req.ClearEscalateTo {
		t.Error("removing escalate_to must set ClearEscalateTo")
	}

	// On create there is no prior state, so nothing is cleared.
	created := groupRequest(ctx, &plan, nil, &diags)
	if created.ClearAgents || created.ClearEscalateTo {
		t.Error("a create must not ask the API to clear anything")
	}

	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
}

func TestAgentModelReadsEmbeddedContact(t *testing.T) {
	t.Parallel()

	var m agentModel

	m.Apply(&freshdesk.Agent{
		ID: 7, TicketScope: 2, Occasional: true, Type: "support_agent",
		RoleIDs: []int64{9, 8}, GroupIDs: []int64{1},
		Contact: freshdesk.AgentContact{
			Email: "sam@example.com", Name: "Sam", Active: true, JobTitle: "Lead",
		},
	})

	if m.Email.ValueString() != "sam@example.com" {
		t.Errorf("email = %q", m.Email.ValueString())
	}

	if m.Name.ValueString() != "Sam" {
		t.Errorf("name = %q, must come from the embedded contact", m.Name.ValueString())
	}

	if !m.Active.ValueBool() || m.JobTitle.ValueString() != "Lead" {
		t.Error("the embedded contact's fields must be surfaced")
	}

	if got := firstInt64(t, m.RoleIDs); got != 8 {
		t.Errorf("role_ids starts at %d, want ascending order from 8", got)
	}
}

func TestAgentRequestOmitsIdentityOnUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var diags diag.Diagnostics

	plan := agentModel{
		Email:       emailString("sam@example.com"),
		Name:        types.StringValue("Sam"),
		TicketScope: types.Int64Value(1),
	}

	created := agentRequest(ctx, &plan, nil, &diags, true)
	if created.Email == nil || created.Name == nil {
		t.Error("a create must carry the email and name")
	}

	updated := agentRequest(ctx, &plan, &plan, &diags, false)
	if updated.Email != nil || updated.Name != nil {
		t.Error("an update must omit the email and name, which Freshdesk rejects")
	}
}

func TestContactModelAndRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var m contactModel

	m.Apply(&freshdesk.Contact{
		ID: 5, Name: "Avery", Email: "avery@example.com", CompanyID: 9,
		Tags: []string{"vip"}, Active: true, ContactType: "contact",
	})

	if m.Email.ValueString() != "avery@example.com" {
		t.Errorf("email = %q", m.Email.ValueString())
	}

	if m.CompanyID.ValueInt64() != 9 {
		t.Errorf("company_id = %d", m.CompanyID.ValueInt64())
	}

	var diags diag.Diagnostics

	plan := contactModel{Name: types.StringValue("Avery")}

	req := contactRequest(ctx, &plan, &m, &diags)
	if !req.ClearTags {
		t.Error("removing tags must set ClearTags")
	}

	if !req.ClearCompany {
		t.Error("removing company_id must set ClearCompany")
	}
}

func TestCompanyModelFormatsRenewalDate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var renewal freshdesk.Time
	if err := renewal.UnmarshalJSON([]byte(`"2030-12-31T00:00:00Z"`)); err != nil {
		t.Fatalf("parsing the renewal date: %v", err)
	}

	var m companyModel

	m.Apply(&freshdesk.Company{
		ID: 3, Name: "Northwind", Description: "Wholesale", RenewalDate: renewal,
	})

	// The API echoes a full timestamp; state keeps the date form that was
	// configured so the plan does not churn.
	if got := m.RenewalDate.ValueString(); got != "2030-12-31" {
		t.Errorf("renewal_date = %q, want the date form", got)
	}

	var diags diag.Diagnostics

	plan := companyModel{Name: types.StringValue("Northwind")}

	req := companyRequest(ctx, &plan, &m, &diags)
	if !req.ClearRenewalDate {
		t.Error("removing renewal_date must set ClearRenewalDate")
	}

	if req.Description == nil || *req.Description != "" {
		t.Error("removing the description must send an empty string to clear it")
	}
}

func TestTicketModelKeepsConfiguredDescription(t *testing.T) {
	t.Parallel()

	m := ticketModel{Description: types.StringValue("<p>Hello</p>")}

	m.Apply(&freshdesk.Ticket{
		ID: 11, Subject: "Help",
		Description:     "<div>Hello</div>", // Freshdesk rewrote the markup.
		DescriptionText: "Hello",
		Status:          2, Priority: 3,
	})

	if got := m.Description.ValueString(); got != "<p>Hello</p>" {
		t.Errorf("description = %q, want the configured markup", got)
	}

	if got := m.DescriptionText.ValueString(); got != "Hello" {
		t.Errorf("description_text = %q, want the stored rendering", got)
	}

	// A ticket read fresh from the API, with nothing configured, takes the
	// server's value.
	var imported ticketModel

	imported.Apply(&freshdesk.Ticket{ID: 11, Description: "<div>Hello</div>"})

	if got := imported.Description.ValueString(); got != "<div>Hello</div>" {
		t.Errorf("imported description = %q", got)
	}
}

func TestTopicModelKeepsConfiguredMessage(t *testing.T) {
	t.Parallel()

	m := topicModel{Message: types.StringValue("<p>Hi</p>")}

	// The API never echoes the message back.
	m.Apply(&freshdesk.Topic{ID: 4, Title: "Question", ForumID: 2, StampType: 7})

	if got := m.Message.ValueString(); got != "<p>Hi</p>" {
		t.Errorf("message = %q, want the configured value kept", got)
	}

	if m.StampType.ValueInt64() != 7 {
		t.Errorf("stamp_type = %d", m.StampType.ValueInt64())
	}
}

func TestSLAPolicyModelReadsTargets(t *testing.T) {
	t.Parallel()

	var m slaPolicyModel

	m.Apply(&freshdesk.SLAPolicy{
		ID: freshdesk.ID(158000322059), Name: "Default", Active: true, Position: 1,
		SLATarget: map[string]freshdesk.SLATarget{
			"priority_4": {RespondWithin: 300, ResolveWithin: 7200, BusinessHours: true},
		},
		ApplicableTo: freshdesk.SLAApplicableTo{Sources: []int{1, 2}},
	})

	if m.GetID().ValueString() != "158000322059" {
		t.Errorf("id = %q, must decode the quoted API id", m.GetID().ValueString())
	}

	targets := m.SLATarget.Elements()
	if len(targets) != 1 {
		t.Fatalf("sla_target = %v", targets)
	}

	sources := m.ApplicableToSources.Elements()
	if len(sources) != 2 {
		t.Errorf("applicable_to_sources = %v", sources)
	}
}

func TestAutomationRuleRequestDecodesJSON(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics

	plan := automationRuleModel{
		Name: types.StringValue("Route urgent"),
		Conditions: jsonStringValue(
			`[{"name":"condition_set_1","match_type":"all","properties":[{"field_name":"priority"}]}]`),
		Actions:   jsonStringValue(`[{"field_name":"group_id","value":7}]`),
		Performer: jsonStringValue(`{"type":1}`),
	}

	req := automationRuleRequest(&plan, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	if len(req.Conditions) != 1 || req.Conditions[0].Name != "condition_set_1" {
		t.Errorf("conditions = %+v", req.Conditions)
	}

	if len(req.Actions) != 1 || req.Actions[0]["field_name"] != "group_id" {
		t.Errorf("actions = %+v", req.Actions)
	}

	webhook := automationRuleModel{Actions: jsonStringValue(
		`[{"field_name":"trigger_webhook","request_type":"POST","content_type":"JSON",` +
			`"content_layout":"2","url":"https://example.test/hooks/x",` +
			`"content":{"ticket_id":"{{ticket.id}}"},"custom_headers":{"x-secret":"s"}}]`)}

	var webhookDiags diag.Diagnostics

	webhookReq := automationRuleRequest(&webhook, &webhookDiags)
	if webhookDiags.HasError() {
		t.Fatalf("diagnostics: %v", webhookDiags)
	}

	if webhookReq.Actions[0]["content_type"] != "JSON" || webhookReq.Actions[0]["custom_headers"] == nil {
		t.Errorf("webhook action fields must pass through verbatim, got %+v", webhookReq.Actions[0])
	}

	if req.Performer["type"] != float64(1) {
		t.Errorf("performer = %+v", req.Performer)
	}

	// A malformed body must be reported, not silently dropped.
	bad := automationRuleModel{Actions: jsonStringValue(`{"not":"an array"}`)}

	var badDiags diag.Diagnostics

	automationRuleRequest(&bad, &badDiags)

	if !badDiags.HasError() {
		t.Error("a non-array actions body must report a diagnostic")
	}
}

func TestSkillRequestPassesConditionsThrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var diags diag.Diagnostics

	// The omniroute shape must survive untouched.
	const omniroute = `[{"channel":"ticket","channel_conditions":[` +
		`{"name":"condition_set_1","match_type":"all","properties":[]}]}]`

	plan := skillModel{
		Name:       types.StringValue("Escalations"),
		Conditions: jsonStringValue(omniroute),
	}

	req := skillRequest(ctx, &plan, nil, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	if req.RawConditions == nil {
		t.Fatal("conditions must be carried through as raw JSON")
	}

	list, ok := req.RawConditions.([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("RawConditions = %#v", req.RawConditions)
	}

	first, ok := list[0].(map[string]any)
	if !ok || first["channel"] != "ticket" {
		t.Errorf("the channel-scoped shape was not preserved: %#v", list[0])
	}
}

func TestFieldChoicesHandleEveryShape(t *testing.T) {
	t.Parallel()

	// A custom dropdown returns an array of objects.
	custom := freshdesk.RawJSON(`[{"id":1,"value":"Refund","position":1}]`)

	list := applyChoices(types.ListNull(choiceType), custom)
	if list.IsNull() || len(list.Elements()) != 1 {
		t.Errorf("a custom dropdown's choices must populate the attribute: %v", list)
	}

	// The built-in fields return maps, which have no place in the list.
	for _, raw := range []string{
		`{"Low":1,"Medium":2}`,
		`{"en":"English"}`,
		`{"2":["Open","Open"]}`,
		`["At risk","Happy"]`,
	} {
		got := applyChoices(types.ListNull(choiceType), freshdesk.RawJSON(raw))
		if !got.IsNull() {
			t.Errorf("choices %s must leave the attribute null, got %v", raw, got)
		}
	}
}

func TestResolveChoiceIDsMatchesByValue(t *testing.T) {
	t.Parallel()

	existing := freshdesk.RawJSON(`[{"id":11,"value":"Refund"},{"id":12,"value":"Faulty"}]`)

	planned := []freshdesk.Choice{
		{Value: "Refund", Position: 1},
		{Value: "Faulty", Position: 2},
		{Value: "New option", Position: 3},
	}

	got := resolveChoiceIDs(existing, planned)
	if got[0].ID != 11 || got[1].ID != 12 {
		t.Errorf("existing choices must keep their IDs: %+v", got)
	}

	if got[2].ID != 0 {
		t.Errorf("a new choice must carry no ID, got %d", got[2].ID)
	}
}

func TestEmailSettingsSendsOnlyConfiguredToggles(t *testing.T) {
	t.Parallel()

	m := emailSettingsModel{
		MultipleTo:          types.BoolValue(true),
		SkipTicketThreading: types.BoolValue(false),
		// Everything else is left null and must not be sent.
	}

	got := m.settings()
	if len(got) != 2 {
		t.Errorf("settings = %v, want only the two configured toggles", got)
	}

	if !got["multiple_to"] || got["skip_ticket_threading"] {
		t.Errorf("settings = %v", got)
	}
}
