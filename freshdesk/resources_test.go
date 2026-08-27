package freshdesk

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recorder captures the method, path and query of the last request.
type recorder struct {
	method, path, query string
	body                []byte
}

func newRecorder(t *testing.T, response string) (*Client, *recorder) {
	t.Helper()
	rec := &recorder{}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.method, rec.path, rec.query = r.Method, r.URL.Path, r.URL.RawQuery
		rec.body, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(response))
	})
	return c, rec
}

// TestEndpointRouting pins the verb and path of every CRUD entry point, which
// is where a typo would otherwise surface only against the live API.
func TestEndpointRouting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cases := []struct {
		name       string
		call       func(*Client) error
		wantMethod string
		wantPath   string
		response   string
	}{
		// Tickets
		{"GetTicket", func(c *Client) error { _, e := c.GetTicket(ctx, 1, TicketIncludes{}); return e },
			"GET", "/api/v2/tickets/1", `{}`},
		{"CreateTicket", func(c *Client) error { _, e := c.CreateTicket(ctx, TicketRequest{}); return e },
			"POST", "/api/v2/tickets", `{}`},
		{"UpdateTicket", func(c *Client) error { _, e := c.UpdateTicket(ctx, 1, TicketRequest{}); return e },
			"PUT", "/api/v2/tickets/1", `{}`},
		{"DeleteTicket", func(c *Client) error { return c.DeleteTicket(ctx, 1) },
			"DELETE", "/api/v2/tickets/1", ``},
		{"RestoreTicket", func(c *Client) error { return c.RestoreTicket(ctx, 1) },
			"PUT", "/api/v2/tickets/1/restore", ``},
		{"WatchTicket", func(c *Client) error { return c.WatchTicket(ctx, 1) },
			"PUT", "/api/v2/tickets/1/watch", ``},
		{"UnwatchTicket", func(c *Client) error { return c.UnwatchTicket(ctx, 1) },
			"PUT", "/api/v2/tickets/1/unwatch", ``},
		{"MergeTickets", func(c *Client) error { return c.MergeTickets(ctx, TicketMergeRequest{}) },
			"PUT", "/api/v2/tickets/merge", ``},
		{"BulkDeleteTickets", func(c *Client) error { _, e := c.BulkDeleteTickets(ctx, []int64{1}); return e },
			"POST", "/api/v2/tickets/bulk_delete", `{}`},
		{"BulkUpdateTickets", func(c *Client) error { _, e := c.BulkUpdateTickets(ctx, BulkTicketUpdate{}); return e },
			"PUT", "/api/v2/tickets/bulk_update", `{}`},
		{"GetArchivedTicket", func(c *Client) error { _, e := c.GetArchivedTicket(ctx, 1); return e },
			"GET", "/api/v2/tickets/archived/1", `{}`},
		{"DeleteAttachment", func(c *Client) error { return c.DeleteAttachment(ctx, 9) },
			"DELETE", "/api/v2/attachments/9", ``},

		// Conversations
		{"ListTicketConversations", func(c *Client) error {
			_, e := c.ListTicketConversations(ctx, 1, ListOptions{})
			return e
		},
			"GET", "/api/v2/tickets/1/conversations", `[]`},
		{"CreateNote", func(c *Client) error { _, e := c.CreateNote(ctx, 1, NoteRequest{}); return e },
			"POST", "/api/v2/tickets/1/notes", `{}`},
		{"CreateReply", func(c *Client) error { _, e := c.CreateReply(ctx, 1, ReplyRequest{}); return e },
			"POST", "/api/v2/tickets/1/reply", `{}`},
		{"ForwardTicket", func(c *Client) error { _, e := c.ForwardTicket(ctx, 1, ForwardRequest{}); return e },
			"POST", "/api/v2/tickets/1/forward", `{}`},
		{"ReplyToForward", func(c *Client) error { _, e := c.ReplyToForward(ctx, 1, ForwardRequest{}); return e },
			"POST", "/api/v2/tickets/1/reply_to_forward", `{}`},
		{"UpdateConversation", func(c *Client) error { _, e := c.UpdateConversation(ctx, 5, ConversationUpdate{}); return e },
			"PUT", "/api/v2/conversations/5", `{}`},
		{"DeleteConversation", func(c *Client) error { return c.DeleteConversation(ctx, 5) },
			"DELETE", "/api/v2/conversations/5", ``},

		// Contacts
		{"GetContact", func(c *Client) error { _, e := c.GetContact(ctx, 1); return e },
			"GET", "/api/v2/contacts/1", `{}`},
		{"CreateContact", func(c *Client) error { _, e := c.CreateContact(ctx, ContactRequest{}); return e },
			"POST", "/api/v2/contacts", `{}`},
		{"HardDeleteContact", func(c *Client) error { return c.HardDeleteContact(ctx, 1, false) },
			"DELETE", "/api/v2/contacts/1/hard_delete", ``},
		{"RestoreContact", func(c *Client) error { return c.RestoreContact(ctx, 1) },
			"PUT", "/api/v2/contacts/1/restore", ``},
		{"SendContactInvite", func(c *Client) error { return c.SendContactInvite(ctx, 1) },
			"PUT", "/api/v2/contacts/1/send_invite", ``},
		{"MakeAgent", func(c *Client) error { _, e := c.MakeAgent(ctx, 1, MakeAgentRequest{}); return e },
			"PUT", "/api/v2/contacts/1/make_agent", `{}`},
		{"MergeContacts", func(c *Client) error { return c.MergeContacts(ctx, ContactMergeRequest{}) },
			"POST", "/api/v2/contacts/merge", ``},
		{"SearchContacts", func(c *Client) error { _, e := c.SearchContacts(ctx, "bob"); return e },
			"GET", "/api/v2/contacts/autocomplete", `[]`},
		{"CancelContactImport", func(c *Client) error { return c.CancelContactImport(ctx, 7) },
			"POST", "/api/v2/contacts/imports/7/cancel", ``},

		// Companies
		{"GetCompany", func(c *Client) error { _, e := c.GetCompany(ctx, 1); return e },
			"GET", "/api/v2/companies/1", `{}`},
		{"CreateCompany", func(c *Client) error { _, e := c.CreateCompany(ctx, CompanyRequest{}); return e },
			"POST", "/api/v2/companies", `{}`},
		{"DeleteCompany", func(c *Client) error { return c.DeleteCompany(ctx, 1) },
			"DELETE", "/api/v2/companies/1", ``},
		{"SearchCompanies", func(c *Client) error { _, e := c.SearchCompanies(ctx, "acme"); return e },
			"GET", "/api/v2/companies/autocomplete", `{"companies":[]}`},

		// Agents, roles, skills, groups
		{"Me", func(c *Client) error { _, e := c.Me(ctx); return e },
			"GET", "/api/v2/agents/me", `{}`},
		{"CreateAgent", func(c *Client) error { _, e := c.CreateAgent(ctx, AgentRequest{}); return e },
			"POST", "/api/v2/agents", `{}`},
		{"DeleteAgent", func(c *Client) error { return c.DeleteAgent(ctx, 1) },
			"DELETE", "/api/v2/agents/1", ``},
		{"SearchAgents", func(c *Client) error { _, e := c.SearchAgents(ctx, "sue"); return e },
			"GET", "/api/v2/agents/autocomplete", `[]`},
		{"CreateAgents", func(c *Client) error { _, e := c.CreateAgents(ctx, nil); return e },
			"POST", "/api/v2/agents/bulk", `{}`},
		{"GetAgentAvailability", func(c *Client) error { _, e := c.GetAgentAvailability(ctx, 1); return e },
			"GET", "/api/v2/agents/1/availability", `{}`},
		{"GetRole", func(c *Client) error { _, e := c.GetRole(ctx, 1); return e },
			"GET", "/api/v2/roles/1", `{}`},
		{"GetSkill", func(c *Client) error { _, e := c.GetSkill(ctx, 1); return e },
			"GET", "/api/v2/admin/skills/1", `{}`},
		{"ListSkillsLegacy", func(c *Client) error { _, e := c.ListSkillsLegacy(ctx, ListOptions{}); return e },
			"GET", "/api/v2/skills", `[]`},
		{"CreateGroup", func(c *Client) error { _, e := c.CreateGroup(ctx, GroupRequest{}); return e },
			"POST", "/api/v2/groups", `{}`},
		{"CreateAdminGroup", func(c *Client) error { _, e := c.CreateAdminGroup(ctx, AdminGroupRequest{}); return e },
			"POST", "/api/v2/admin/groups", `{}`},
		{"ListAgentsInGroup", func(c *Client) error { _, e := c.ListAgentsInGroup(ctx, 3); return e },
			"GET", "/api/v2/admin/groups/3/agents", `[]`},
		{"UpdateGroupAgents", func(c *Client) error { return c.UpdateGroupAgents(ctx, 3, []int64{1}, nil) },
			"PUT", "/api/v2/admin/groups/3/agents", ``},

		// Fields, forms, sections
		{"GetTicketField", func(c *Client) error { _, e := c.GetTicketField(ctx, 1, false); return e },
			"GET", "/api/v2/admin/ticket_fields/1", `{}`},
		{"ListTicketFieldsLegacy", func(c *Client) error { _, e := c.ListTicketFieldsLegacy(ctx, ListOptions{}); return e },
			"GET", "/api/v2/ticket_fields", `[]`},
		{"ListSections", func(c *Client) error { _, e := c.ListSections(ctx, 22); return e },
			"GET", "/api/v2/admin/ticket_fields/22/sections", `[]`},
		{"UpdateSection", func(c *Client) error { _, e := c.UpdateSection(ctx, 22, 3, SectionRequest{}); return e },
			"PUT", "/api/v2/admin/ticket_fields/22/sections/3", `{}`},
		{"DeleteContactField", func(c *Client) error { return c.DeleteContactField(ctx, 4) },
			"DELETE", "/api/v2/contact_fields/4", ``},
		{"UpdateCompanyField", func(c *Client) error {
			_, e := c.UpdateCompanyField(ctx, 4, CompanyFieldRequest{})
			return e
		},
			"PUT", "/api/v2/company_fields/4", `{}`},
		{"CloneTicketForm", func(c *Client) error { _, e := c.CloneTicketForm(ctx, 2, "copy"); return e },
			"POST", "/api/v2/ticket-forms/2/clone", `{}`},
		{"UpdateTicketFormField", func(c *Client) error {
			_, e := c.UpdateTicketFormField(ctx, 1, 5, TicketFormFieldRequest{})
			return e
		}, "PUT", "/api/v2/ticket-forms/1/fields/5", `{}`},

		// Solutions
		{"GetSolutionCategory", func(c *Client) error { _, e := c.GetSolutionCategory(ctx, 3, ""); return e },
			"GET", "/api/v2/solutions/categories/3", `{}`},
		{"GetSolutionCategoryTranslated", func(c *Client) error { _, e := c.GetSolutionCategory(ctx, 3, "es"); return e },
			"GET", "/api/v2/solutions/categories/3/es", `{}`},
		{"ListSolutionFolders", func(c *Client) error { _, e := c.ListSolutionFolders(ctx, 3, "", ListOptions{}); return e },
			"GET", "/api/v2/solutions/categories/3/folders", `[]`},
		{"ListSolutionSubFolders", func(c *Client) error {
			_, e := c.ListSolutionSubFolders(ctx, 2, "", ListOptions{})
			return e
		},
			"GET", "/api/v2/solutions/folders/2/subfolders", `[]`},
		{"CreateSolutionArticle", func(c *Client) error {
			_, e := c.CreateSolutionArticle(ctx, 4, SolutionArticleRequest{})
			return e
		}, "POST", "/api/v2/solutions/folders/4/articles", `{}`},
		{"SearchSolutions", func(c *Client) error { _, e := c.SearchSolutions(ctx, "vpn", ListOptions{}); return e },
			"GET", "/api/v2/search/solutions", `[]`},

		// Discussions
		{"CreateForum", func(c *Client) error { _, e := c.CreateForum(ctx, 1, ForumRequest{}); return e },
			"POST", "/api/v2/discussions/categories/1/forums", `{}`},
		{"ListTopics", func(c *Client) error { _, e := c.ListTopics(ctx, 5, ListOptions{}); return e },
			"GET", "/api/v2/discussions/forums/5/topics", `[]`},
		{"FollowTopic", func(c *Client) error { return c.FollowTopic(ctx, 1, 2) },
			"POST", "/api/v2/discussions/topics/1/follow", ``},
		{"ListComments", func(c *Client) error { _, e := c.ListComments(ctx, 1, ListOptions{}); return e },
			"GET", "/api/v2/discussions/topics/1/comments", `[]`},
		{"ListTopicsFollowedBy", func(c *Client) error { _, e := c.ListTopicsFollowedBy(ctx, 9, ListOptions{}); return e },
			"GET", "/api/v2/discussions/topics/followed_by", `[]`},

		// Canned responses
		{"CreateCannedResponseFolder", func(c *Client) error {
			_, e := c.CreateCannedResponseFolder(ctx, CannedResponseFolderRequest{})
			return e
		}, "POST", "/api/v2/canned_response_folders", `{}`},
		{"ListCannedResponseFolders", func(c *Client) error {
			_, e := c.ListCannedResponseFolders(ctx, ListOptions{})
			return e
		}, "GET", "/api/v2/canned_response_folders", `[]`},
		{"ListCannedResponsesInFolder", func(c *Client) error {
			_, e := c.ListCannedResponsesInFolder(ctx, 1, ListOptions{})
			return e
		}, "GET", "/api/v2/canned_response_folders/1/responses", `[]`},
		{"CreateCannedResponses", func(c *Client) error { _, e := c.CreateCannedResponses(ctx, 1, nil); return e },
			"POST", "/api/v2/canned_responses/bulk", `[]`},

		// Time entries, SLA, automations
		{"CreateTimeEntry", func(c *Client) error { _, e := c.CreateTimeEntry(ctx, 1, TimeEntryRequest{}); return e },
			"POST", "/api/v2/tickets/1/time_entries", `{}`},
		{"ToggleTimer", func(c *Client) error { _, e := c.ToggleTimer(ctx, 3); return e },
			"PUT", "/api/v2/time_entries/3/toggle_timer", `{}`},
		{"CreateSLAPolicy", func(c *Client) error { _, e := c.CreateSLAPolicy(ctx, SLAPolicyRequest{}); return e },
			"POST", "/api/v2/sla_policies", `{}`},
		{"ListAutomationRules", func(c *Client) error {
			_, e := c.ListAutomationRules(ctx, AutomationTypeTicketUpdate, ListOptions{})
			return e
		}, "GET", "/api/v2/automations/4/rules", `[]`},
		{"CreateAutomationRule", func(c *Client) error {
			_, e := c.CreateAutomationRule(ctx, AutomationTypeTicketCreation, AutomationRuleRequest{})
			return e
		}, "POST", "/api/v2/automations/1/rules", `{}`},
		{"ListScenarioAutomations", func(c *Client) error {
			_, e := c.ListScenarioAutomations(ctx, ListOptions{})
			return e
		}, "GET", "/api/v2/scenario_automations", `[]`},

		// Email
		{"ListEmailConfigs", func(c *Client) error { _, e := c.ListEmailConfigs(ctx, ListOptions{}); return e },
			"GET", "/api/v2/email_configs", `[]`},
		{"CreateEmailMailbox", func(c *Client) error { _, e := c.CreateEmailMailbox(ctx, EmailMailboxRequest{}); return e },
			"POST", "/api/v2/email/mailboxes", `{}`},
		{"GetEmailSettings", func(c *Client) error { _, e := c.GetEmailSettings(ctx); return e },
			"GET", "/api/v2/email/settings", `{}`},
		{"GetNotificationBCC", func(c *Client) error { _, e := c.GetNotificationBCC(ctx); return e },
			"GET", "/api/v2/notifications/email/bcc", `{"emails":[]}`},

		// Surveys and satisfaction
		{"ListSurveys", func(c *Client) error { _, e := c.ListSurveys(ctx, ListOptions{}); return e },
			"GET", "/api/v2/customer-satisfaction/surveys", `{"data":[]}`},
		{"ListLegacySurveys", func(c *Client) error { _, e := c.ListLegacySurveys(ctx, ListOptions{}); return e },
			"GET", "/api/v2/surveys", `[]`},
		{"ListSatisfactionRatings", func(c *Client) error {
			_, e := c.ListSatisfactionRatings(ctx, SatisfactionRatingListOptions{})
			return e
		}, "GET", "/api/v2/surveys/satisfaction_ratings", `[]`},
		{"CreateSatisfactionRating", func(c *Client) error {
			_, e := c.CreateSatisfactionRating(ctx, 1, SatisfactionRatingRequest{})
			return e
		}, "POST", "/api/v2/tickets/1/satisfaction_ratings", `{}`},
		{"ListSurveyResponses", func(c *Client) error {
			_, e := c.ListSurveyResponses(ctx, "uuid-1", ListOptions{})
			return e
		},
			"GET", "/api/v2/customer-satisfaction/surveys/uuid-1/responses", `[]`},

		// Custom objects
		{"ListCustomObjectSchemas", func(c *Client) error { _, e := c.ListCustomObjectSchemas(ctx); return e },
			"GET", "/api/v2/custom_objects/schemas", `{"schemas":[]}`},
		{"ListCustomObjectRecords", func(c *Client) error {
			_, e := c.ListCustomObjectRecords(ctx, "sch-1", nil)
			return e
		}, "GET", "/api/v2/custom_objects/schemas/sch-1/records", `{"records":[]}`},
		{"CountCustomObjectRecords", func(c *Client) error { _, e := c.CountCustomObjectRecords(ctx, "sch-1"); return e },
			"GET", "/api/v2/custom_objects/schemas/sch-1/records/count", `{"count":0}`},
		{"DeleteCustomObjectRecord", func(c *Client) error { return c.DeleteCustomObjectRecord(ctx, "sch-1", "BKG-1") },
			"DELETE", "/api/v2/custom_objects/schemas/sch-1/records/BKG-1", ``},

		// Account, jobs, collaboration
		{"GetAccount", func(c *Client) error { _, e := c.GetAccount(ctx); return e },
			"GET", "/api/v2/account", `{}`},
		{"GetHelpdeskSettings", func(c *Client) error { _, e := c.GetHelpdeskSettings(ctx); return e },
			"GET", "/api/v2/settings/helpdesk", `{}`},
		{"GetJob", func(c *Client) error { _, e := c.GetJob(ctx, "abc"); return e },
			"GET", "/api/v2/jobs/abc", `{}`},
		{"CreateThread", func(c *Client) error { _, e := c.CreateThread(ctx, ThreadRequest{}); return e },
			"POST", "/api/v2/collaboration/threads", `{}`},
		{"CreateThreadMessage", func(c *Client) error {
			_, e := c.CreateThreadMessage(ctx, ThreadMessageRequest{})
			return e
		},
			"POST", "/api/v2/collaboration/messages", `{}`},
		{"GenerateMessageQuote", func(c *Client) error { _, e := c.GenerateMessageQuote(ctx, 2); return e },
			"GET", "/api/v2/collaboration/messages/2/generate-quote", `{"quoted_text":""}`},
		{"SendOutboundMessage", func(c *Client) error { _, e := c.SendOutboundMessage(ctx, nil); return e },
			"POST", "/api/v2/channels/outbound-messages", `{}`},
		{"PublishContactActivities", func(c *Client) error { return c.PublishContactActivities(ctx, nil) },
			"POST", "/api/v2/contact-activities", ``},

		// Search
		{"FilterTickets", func(c *Client) error { _, e := c.FilterTickets(ctx, "status:2", 0); return e },
			"GET", "/api/v2/search/tickets", `{"total":0,"results":[]}`},
		{"FilterContacts", func(c *Client) error { _, e := c.FilterContacts(ctx, "active:true", 0); return e },
			"GET", "/api/v2/search/contacts", `{"total":0,"results":[]}`},
		{"FilterCompanies", func(c *Client) error { _, e := c.FilterCompanies(ctx, "name:'x'", 0); return e },
			"GET", "/api/v2/search/companies", `{"total":0,"results":[]}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, rec := newRecorder(t, tc.response)
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if rec.method != tc.wantMethod {
				t.Errorf("method = %s, want %s", rec.method, tc.wantMethod)
			}
			if rec.path != tc.wantPath {
				t.Errorf("path = %s, want %s", rec.path, tc.wantPath)
			}
		})
	}
}

func TestSearchQueriesAreQuoted(t *testing.T) {
	t.Parallel()

	c, rec := newRecorder(t, `{"total":0,"results":[]}`)
	if _, err := c.FilterTickets(context.Background(), "priority:3", 2); err != nil {
		t.Fatalf("FilterTickets: %v", err)
	}
	if !strings.Contains(rec.query, `query=%22priority%3A3%22`) {
		t.Errorf("query = %q, want a quoted filter expression", rec.query)
	}
	if !strings.Contains(rec.query, "page=2") {
		t.Errorf("query = %q, want page=2", rec.query)
	}
}

func TestTicketIncludesAreSentAsCSV(t *testing.T) {
	t.Parallel()

	c, rec := newRecorder(t, `{}`)
	_, err := c.GetTicket(context.Background(), 1, TicketIncludes{Stats: true, Requester: true, Company: true})
	if err != nil {
		t.Fatalf("GetTicket: %v", err)
	}
	if !strings.Contains(rec.query, "include=stats%2Crequester%2Ccompany") {
		t.Errorf("query = %q, want include=stats,requester,company", rec.query)
	}
}

func TestHardDeleteContactForce(t *testing.T) {
	t.Parallel()

	c, rec := newRecorder(t, ``)
	if err := c.HardDeleteContact(context.Background(), 1, true); err != nil {
		t.Fatalf("HardDeleteContact: %v", err)
	}
	if rec.query != "force=true" {
		t.Errorf("query = %q, want force=true", rec.query)
	}
}

func TestGetSLAPolicyFiltersCollection(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"11","name":"a"},{"id":"22","name":"b"}]`))
	})
	got, err := c.GetSLAPolicy(context.Background(), 22)
	if err != nil {
		t.Fatalf("GetSLAPolicy: %v", err)
	}
	if got.Name != "b" {
		t.Errorf("Name = %q, want b", got.Name)
	}

	if _, err := c.GetSLAPolicy(context.Background(), 99); !NotFound(err) {
		t.Errorf("missing policy: err = %v, want NotFound", err)
	}
}

// TestCreateSectionAcceptsBothShapes covers the two things Freshdesk answers a
// section create with: the bare object the API returns, and the section-list
// envelope the documentation describes.
func TestCreateSectionAcceptsBothShapes(t *testing.T) {
	t.Parallel()

	t.Run("bare object", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(
				`{"id":7,"label":"Web","parent_ticket_field_id":22,"choice_ids":[31]}`))
		})

		got, err := c.CreateSection(context.Background(), 22, SectionRequest{Label: Ptr("Web")})
		if err != nil {
			t.Fatalf("CreateSection: %v", err)
		}

		if got.ID != 7 || got.ParentTicketFieldID != 22 {
			t.Errorf("section = %+v", got)
		}
	})

	t.Run("envelope matched by label", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"sections":[
				{"id":3,"label":"Mobile"},{"id":4,"label":"Web"}]}`))
		})

		got, err := c.CreateSection(context.Background(), 22, SectionRequest{Label: Ptr("Web")})
		if err != nil {
			t.Fatalf("CreateSection: %v", err)
		}

		if got.ID != 4 {
			t.Errorf("ID = %d, want 4 (matched by label)", got.ID)
		}
	})

	t.Run("falls back to listing", func(t *testing.T) {
		t.Parallel()

		var calls int

		c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			// The create says nothing useful; the list then finds the section.
			if r.Method == http.MethodPost {
				_, _ = w.Write([]byte(`{}`))

				return
			}
			_, _ = w.Write([]byte(`[{"id":9,"label":"Web"}]`))
		})

		got, err := c.CreateSection(context.Background(), 22, SectionRequest{Label: Ptr("Web")})
		if err != nil {
			t.Fatalf("CreateSection: %v", err)
		}

		if got.ID != 9 {
			t.Errorf("ID = %d, want 9 (found by listing)", got.ID)
		}

		if calls != 2 {
			t.Errorf("made %d calls, want a create then a list", calls)
		}
	})
}

func TestUpdateGroupAgentsBody(t *testing.T) {
	t.Parallel()

	c, rec := newRecorder(t, ``)
	if err := c.UpdateGroupAgents(context.Background(), 3, []int64{1, 2}, []int64{9}); err != nil {
		t.Fatalf("UpdateGroupAgents: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.body, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	add, addOK := body["add"].([]any)
	remove, removeOK := body["remove"].([]any)
	if !addOK || !removeOK || len(add) != 2 || len(remove) != 1 {
		t.Errorf("body = %v", body)
	}
}

func TestNotificationBCCSendsEmptyArray(t *testing.T) {
	t.Parallel()

	c, rec := newRecorder(t, `{"emails":[]}`)
	if _, err := c.UpdateNotificationBCC(context.Background(), nil); err != nil {
		t.Fatalf("UpdateNotificationBCC: %v", err)
	}
	if !strings.Contains(string(rec.body), `"emails":[]`) {
		t.Errorf("body = %s, want an explicit empty array", rec.body)
	}
}

// TestMultipartUpload checks that attachments switch the request to
// multipart/form-data with the file and the scalar fields both present.
func TestMultipartUpload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var gotFile, gotBody, gotNotify string
	var ct string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		mt, params, err := mime.ParseMediaType(ct)
		if err != nil || !strings.HasPrefix(mt, "multipart/") {
			t.Errorf("Content-Type = %q", ct)
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			data, _ := io.ReadAll(part)
			switch part.FormName() {
			case "attachments[]":
				gotFile = string(data)
				if part.FileName() != "note.txt" {
					t.Errorf("filename = %q", part.FileName())
				}
			case "body":
				gotBody = string(data)
			case "notify_emails[]":
				gotNotify = string(data)
			}
		}
		_, _ = w.Write([]byte(`{"id":5}`))
	})

	_, err := c.CreateNote(context.Background(), 20, NoteRequest{
		Body:         Ptr("Hi tom"),
		Private:      Ptr(false),
		NotifyEmails: []string{"tom@example.com"},
		Attachments:  []string{file},
	})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if gotFile != "hello" {
		t.Errorf("uploaded file content = %q", gotFile)
	}
	if gotBody != "Hi tom" {
		t.Errorf("body field = %q", gotBody)
	}
	if gotNotify != "tom@example.com" {
		t.Errorf("notify_emails[] = %q", gotNotify)
	}
}

func TestMultipartMissingFileReturnsError(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	_, err := c.CreateNote(context.Background(), 1, NoteRequest{
		Body:        Ptr("x"),
		Attachments: []string{filepath.Join(t.TempDir(), "missing.txt")},
	})
	if err == nil {
		t.Fatal("want error for missing attachment file")
	}
	if !strings.Contains(err.Error(), "opening attachment") {
		t.Errorf("err = %v, want an attachment-open failure", err)
	}
}

func TestMultipartFieldsEncoding(t *testing.T) {
	t.Parallel()

	got, err := multipartFields(NoteRequest{
		Body:         Ptr("hi"),
		Private:      Ptr(true),
		UserID:       Ptr(int64(158018521243)),
		NotifyEmails: []string{"a@x.com", "b@x.com"},
		StructuredBody: &StructuredBody{
			BodyContents: []StructuredBodyContent{{Type: "text", Data: map[string]any{"content": "hi"}}},
		},
	})
	if err != nil {
		t.Fatalf("multipartFields: %v", err)
	}
	if got["body"][0] != "hi" {
		t.Errorf("body = %v", got["body"])
	}
	if got["private"][0] != "true" {
		t.Errorf("private = %v", got["private"])
	}
	if got["user_id"][0] != "158018521243" {
		t.Errorf("user_id = %v, want an exact integer", got["user_id"])
	}
	if len(got["notify_emails[]"]) != 2 {
		t.Errorf("notify_emails[] = %v, want two repeated fields", got["notify_emails[]"])
	}
	if !strings.Contains(got["structured_body"][0], `"body_contents"`) {
		t.Errorf("structured_body = %v, want nested JSON", got["structured_body"])
	}
}
