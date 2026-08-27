package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// The provider's worst bug was a request type that enumerated a few fields of
// an open-ended object and silently dropped the rest, so a webhook action
// reached Freshdesk carrying only its field_name.
//
// These tests encode the rule that prevents a repeat: for every attribute the
// practitioner writes as JSON, whatever they put in must come out the other
// side byte-for-byte. They compare parsed JSON, so key order does not matter,
// but a missing key fails.

// assertNoKeysLost marshals what the client would send and checks that every
// key of the original survives, at every depth.
func assertNoKeysLost(t *testing.T, label string, original any, sent []byte) {
	t.Helper()

	var got any
	if err := json.Unmarshal(sent, &got); err != nil {
		t.Fatalf("%s: the request body is not valid JSON: %v", label, err)
	}

	compareKeys(t, label, original, got)
}

// compareKeys walks two decoded JSON values and reports anything in want that
// is absent or altered in got.
func compareKeys(t *testing.T, path string, want, got any) {
	t.Helper()

	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			t.Errorf("%s: expected an object, got %T", path, got)

			return
		}

		for k, v := range w {
			sub, present := g[k]
			if !present {
				t.Errorf("%s.%s was dropped", path, k)

				continue
			}
			compareKeys(t, path+"."+k, v, sub)
		}
	case []any:
		g, ok := got.([]any)
		if !ok {
			t.Errorf("%s: expected an array, got %T", path, got)

			return
		}

		if len(g) != len(w) {
			t.Errorf("%s: length %d, want %d", path, len(g), len(w))

			return
		}

		for i := range w {
			compareKeys(t, path+"["+string(rune('0'+i))+"]", w[i], g[i])
		}
	default:
		if !jsonEqual(want, got) {
			t.Errorf("%s = %#v, want %#v", path, got, want)
		}
	}
}

func jsonEqual(a, b any) bool {
	ab, aerr := json.Marshal(a)
	bb, berr := json.Marshal(b)

	return aerr == nil && berr == nil && string(ab) == string(bb)
}

// decodeJSON is a helper for building the "what the practitioner wrote" side.
func decodeJSON(t *testing.T, s string) any {
	t.Helper()

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("test fixture is not valid JSON: %v", err)
	}

	return v
}

// TestAutomationActionsSurviveVerbatim is the regression test for the 500:
// a trigger_webhook action carries six keys beyond field_name, and every one
// must reach the API.
func TestAutomationActionsSurviveVerbatim(t *testing.T) {
	t.Parallel()

	const actions = `[
	  {
	    "field_name": "trigger_webhook",
	    "request_type": "POST",
	    "content_type": "JSON",
	    "content_layout": "2",
	    "url": "https://hooks.example.test/x",
	    "content": {"ticket_id": "{{ticket.id}}", "nested": {"deep": [1, 2, 3]}},
	    "custom_headers": {"x-secret": "s", "x-other": "t"}
	  },
	  {"field_name": "priority", "value": 4},
	  {
	    "field_name": "send_email_to_agent",
	    "email_to": [1, 2],
	    "email_subject": "Ticket {{ticket.id}}",
	    "email_body": "<p>Body</p>"
	  }
	]`

	var diags diag.Diagnostics

	plan := automationRuleModel{
		Name:    types.StringValue("webhook"),
		Actions: jsonStringValue(actions),
	}

	req := automationRuleRequest(&plan, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	sent, err := json.Marshal(req.Actions)
	if err != nil {
		t.Fatalf("encoding the actions: %v", err)
	}

	assertNoKeysLost(t, "actions", decodeJSON(t, actions), sent)
}

// TestAutomationConditionsSurviveVerbatim covers the condition sets, whose
// properties are equally open-ended.
func TestAutomationConditionsSurviveVerbatim(t *testing.T) {
	t.Parallel()

	const conditions = `[
	  {
	    "name": "condition_set_1",
	    "match_type": "all",
	    "properties": [
	      {
	        "resource_type": "ticket",
	        "field_name": "cf_custom_thing",
	        "operator": "in",
	        "value": [1, 2],
	        "nested_fields": {"level2": {"field_name": "x", "operator": "is", "value": "y"}},
	        "case_sensitive": false
	      }
	    ]
	  }
	]`

	var diags diag.Diagnostics

	plan := automationRuleModel{Conditions: jsonStringValue(conditions)}

	req := automationRuleRequest(&plan, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	sent, err := json.Marshal(req.Conditions)
	if err != nil {
		t.Fatalf("encoding the conditions: %v", err)
	}

	assertNoKeysLost(t, "conditions", decodeJSON(t, conditions), sent)
}

// TestAutomationEventsAndPerformerSurviveVerbatim covers the remaining two
// JSON bodies on an automation rule.
func TestAutomationEventsAndPerformerSurviveVerbatim(t *testing.T) {
	t.Parallel()

	const events = `[
	  {"field_name": "status", "from": "--", "to": "--"},
	  {"field_name": "note_type", "value": "public"},
	  {"field_name": "reply_sent"}
	]`
	const performer = `{"type": 2, "members": [11, 22]}`

	var diags diag.Diagnostics

	plan := automationRuleModel{
		Events:    jsonStringValue(events),
		Performer: jsonStringValue(performer),
	}

	req := automationRuleRequest(&plan, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	sentEvents, err := json.Marshal(req.Events)
	if err != nil {
		t.Fatalf("encoding the events: %v", err)
	}

	assertNoKeysLost(t, "events", decodeJSON(t, events), sentEvents)

	sentPerformer, err := json.Marshal(req.Performer)
	if err != nil {
		t.Fatalf("encoding the performer: %v", err)
	}

	assertNoKeysLost(t, "performer", decodeJSON(t, performer), sentPerformer)
}

// TestSkillConditionsSurviveVerbatim covers both shapes Freshdesk serves.
func TestSkillConditionsSurviveVerbatim(t *testing.T) {
	t.Parallel()

	for name, conditions := range map[string]string{
		"omniroute": `[{"channel":"ticket","channel_conditions":[` +
			`{"name":"condition_set_1","match_type":"all","properties":` +
			`[{"resource_type":"ticket","field_name":"priority","operator":"in","value":[4]}]}]}]`,
		"classic": `[{"resource_type":"ticket","field_name":"cf_bank","operator":"is",` +
			`"value":"HDFC","nested_fields":{"level2":{"field_name":"cf_branch",` +
			`"operator":"is","value":"Chennai"}}}]`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			plan := skillModel{
				Name:       types.StringValue("s"),
				Conditions: jsonStringValue(conditions),
			}

			req := skillRequest(t.Context(), &plan, nil, &diags)
			if diags.HasError() {
				t.Fatalf("diagnostics: %v", diags)
			}

			sent, err := json.Marshal(req.RawConditions)
			if err != nil {
				t.Fatalf("encoding the conditions: %v", err)
			}

			assertNoKeysLost(t, "conditions", decodeJSON(t, conditions), sent)
		})
	}
}

// TestCustomMailboxSurvivesVerbatim covers the mailbox connection settings,
// whose accepted keys differ by provider and authentication method.
func TestCustomMailboxSurvivesVerbatim(t *testing.T) {
	t.Parallel()

	const mailbox = `{
	  "incoming": {
	    "mail_server": "imap.example.test",
	    "port": 993,
	    "user_name": "support@example.test",
	    "password": "secret",
	    "use_ssl": true,
	    "delete_from_server": false,
	    "authentication_type": "basic"
	  },
	  "outgoing": {
	    "mail_server": "smtp.example.test",
	    "port": 587,
	    "user_name": "support@example.test",
	    "password": "secret",
	    "use_ssl": true,
	    "authentication_type": "basic"
	  },
	  "access_token_id": 4242
	}`

	var diags diag.Diagnostics

	plan := emailMailboxModel{
		Name:          types.StringValue("support"),
		SupportEmail:  types.StringValue("support@example.test"),
		MailboxType:   types.StringValue(freshdesk.MailboxTypeCustom),
		CustomMailbox: jsonStringValue(mailbox),
	}

	req := emailMailboxRequestFor(&plan, &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	if req.CustomMailbox == nil {
		t.Fatal("the custom mailbox settings were dropped entirely")
	}

	sent, err := json.Marshal(req.CustomMailbox)
	if err != nil {
		t.Fatalf("encoding the mailbox: %v", err)
	}

	assertNoKeysLost(t, "custom_mailbox", decodeJSON(t, mailbox), sent)
}

// TestAgentAssignmentSurvivesVerbatim covers the omniroute assignment body.
func TestAgentAssignmentSurvivesVerbatim(t *testing.T) {
	t.Parallel()

	const assignment = `{
	  "enabled": true,
	  "assignment_type": 3,
	  "assignment_config": {"max_load": 12, "skill_based": true, "channels": ["email", "chat"]},
	  "some_future_key": "kept"
	}`

	var diags diag.Diagnostics

	got := decodeAgentAssignment(jsonStringValue(assignment), &diags)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}

	if got == nil {
		t.Fatal("the assignment settings were dropped entirely")
	}

	sent, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("encoding the assignment: %v", err)
	}

	assertNoKeysLost(t, "automatic_agent_assignment", decodeJSON(t, assignment), sent)
}
