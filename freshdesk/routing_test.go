package freshdesk

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// routeCase describes one client call and the request it must produce.
type routeCase struct {
	name       string
	call       func(*Client) error
	wantMethod string
	wantPath   string
	response   string
}

func runRouteCases(t *testing.T, cases []routeCase) {
	t.Helper()
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

// TestRoutingTicketsAndContacts covers the remaining ticket, contact and
// company entry points.
func TestRoutingTicketsAndContacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"ListTickets", func(c *Client) error { _, e := c.ListTickets(ctx, TicketListOptions{}); return e },
			"GET", "/api/v2/tickets", `[]`},
		{"DeleteArchivedTicket", func(c *Client) error { return c.DeleteArchivedTicket(ctx, 5) },
			"DELETE", "/api/v2/tickets/archived/5", ``},
		{"ListArchivedTicketConversations", func(c *Client) error {
			_, e := c.ListArchivedTicketConversations(ctx, 5, ListOptions{})
			return e
		}, "GET", "/api/v2/tickets/archived/5/conversations", `[]`},
		{"ListTicketTimeEntries", func(c *Client) error {
			_, e := c.ListTicketTimeEntries(ctx, 5, ListOptions{})
			return e
		}, "GET", "/api/v2/tickets/5/time_entries", `[]`},
		{"ListTicketSatisfactionRatings", func(c *Client) error {
			_, e := c.ListTicketSatisfactionRatings(ctx, 5, ListOptions{})
			return e
		}, "GET", "/api/v2/tickets/5/satisfaction_ratings", `[]`},

		{"ListContacts", func(c *Client) error { _, e := c.ListContacts(ctx, ContactListOptions{}); return e },
			"GET", "/api/v2/contacts", `[]`},
		{"UpdateContact", func(c *Client) error { _, e := c.UpdateContact(ctx, 1, ContactRequest{}); return e },
			"PUT", "/api/v2/contacts/1", `{}`},
		{"DeleteContact", func(c *Client) error { return c.DeleteContact(ctx, 1) },
			"DELETE", "/api/v2/contacts/1", ``},
		{"ExportContacts", func(c *Client) error { _, e := c.ExportContacts(ctx, []string{"name"}, nil); return e },
			"POST", "/api/v2/contacts/export", `{}`},
		{"GetContactExport", func(c *Client) error { _, e := c.GetContactExport(ctx, "abc"); return e },
			"GET", "/api/v2/contacts/export/abc", `{}`},
		{"GetContactImport", func(c *Client) error { _, e := c.GetContactImport(ctx, 3); return e },
			"GET", "/api/v2/contacts/imports/3", `{}`},

		{"ListCompanies", func(c *Client) error { _, e := c.ListCompanies(ctx, ListOptions{}); return e },
			"GET", "/api/v2/companies", `[]`},
		{"UpdateCompany", func(c *Client) error { _, e := c.UpdateCompany(ctx, 1, CompanyRequest{}); return e },
			"PUT", "/api/v2/companies/1", `{}`},
		{"ExportCompanies", func(c *Client) error { _, e := c.ExportCompanies(ctx, nil, nil); return e },
			"POST", "/api/v2/companies/export", `{}`},
		{"GetCompanyExport", func(c *Client) error { _, e := c.GetCompanyExport(ctx, "xyz"); return e },
			"GET", "/api/v2/companies/export/xyz", `{}`},
		{"GetCompanyImport", func(c *Client) error { _, e := c.GetCompanyImport(ctx, 4); return e },
			"GET", "/api/v2/companies/imports/4", `{}`},
		{"CancelCompanyImport", func(c *Client) error { return c.CancelCompanyImport(ctx, 4) },
			"POST", "/api/v2/companies/imports/4/cancel", ``},
	})
}

// TestRoutingAgentsGroupsSkills covers agents, groups, skills and roles.
func TestRoutingAgentsGroupsSkills(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"GetAgent", func(c *Client) error { _, e := c.GetAgent(ctx, 1); return e },
			"GET", "/api/v2/agents/1", `{}`},
		{"ListAgents", func(c *Client) error { _, e := c.ListAgents(ctx, AgentListOptions{}); return e },
			"GET", "/api/v2/agents", `[]`},
		{"UpdateAgent", func(c *Client) error { _, e := c.UpdateAgent(ctx, 1, AgentRequest{}); return e },
			"PUT", "/api/v2/agents/1", `{}`},
		{"UpdateAgentAvailability", func(c *Client) error {
			_, e := c.UpdateAgentAvailability(ctx, 1, map[string]any{})
			return e
		}, "PUT", "/api/v2/agents/1/availability", `{}`},
		{"ListAgentAvailability", func(c *Client) error {
			_, e := c.ListAgentAvailability(ctx, ListOptions{})
			return e
		}, "GET", "/api/v2/agents", `[]`},
		{"ListRoles", func(c *Client) error { _, e := c.ListRoles(ctx, ListOptions{}); return e },
			"GET", "/api/v2/roles", `[]`},

		{"UpdateGroup", func(c *Client) error { _, e := c.UpdateGroup(ctx, 1, GroupRequest{}); return e },
			"PUT", "/api/v2/groups/1", `{}`},
		{"DeleteGroup", func(c *Client) error { return c.DeleteGroup(ctx, 1) },
			"DELETE", "/api/v2/groups/1", ``},
		{"GetAdminGroup", func(c *Client) error { _, e := c.GetAdminGroup(ctx, 1); return e },
			"GET", "/api/v2/admin/groups/1", `{}`},
		{"ListAdminGroups", func(c *Client) error { _, e := c.ListAdminGroups(ctx, ListOptions{}); return e },
			"GET", "/api/v2/admin/groups", `[]`},
		{"UpdateAdminGroup", func(c *Client) error { _, e := c.UpdateAdminGroup(ctx, 1, AdminGroupRequest{}); return e },
			"PUT", "/api/v2/admin/groups/1", `{}`},
		{"DeleteAdminGroup", func(c *Client) error { return c.DeleteAdminGroup(ctx, 1) },
			"DELETE", "/api/v2/admin/groups/1", ``},

		{"CreateSkill", func(c *Client) error { _, e := c.CreateSkill(ctx, SkillRequest{}); return e },
			"POST", "/api/v2/admin/skills", `{}`},
		{"ListSkills", func(c *Client) error { _, e := c.ListSkills(ctx, ListOptions{}); return e },
			"GET", "/api/v2/admin/skills", `[]`},
		{"UpdateSkill", func(c *Client) error { _, e := c.UpdateSkill(ctx, 1, SkillRequest{}); return e },
			"PATCH", "/api/v2/admin/skills/1", `{}`},
		{"DeleteSkill", func(c *Client) error { return c.DeleteSkill(ctx, 1) },
			"DELETE", "/api/v2/admin/skills/1", ``},
		{"GetSkillLegacy", func(c *Client) error { _, e := c.GetSkillLegacy(ctx, 1); return e },
			"GET", "/api/v2/skills/1", `{}`},
		{"CreateSkillLegacy", func(c *Client) error { _, e := c.CreateSkillLegacy(ctx, SkillRequest{}); return e },
			"POST", "/api/v2/skills", `{}`},
		{"UpdateSkillLegacy", func(c *Client) error { _, e := c.UpdateSkillLegacy(ctx, 1, SkillRequest{}); return e },
			"PATCH", "/api/v2/skills/1", `{}`},
		{"DeleteSkillLegacy", func(c *Client) error { return c.DeleteSkillLegacy(ctx, 1) },
			"DELETE", "/api/v2/skills/1", ``},
	})
}

// TestRoutingFieldsAndForms covers field and form management.
func TestRoutingFieldsAndForms(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"CreateTicketField", func(c *Client) error { _, e := c.CreateTicketField(ctx, TicketFieldRequest{}); return e },
			"POST", "/api/v2/admin/ticket_fields", `{}`},
		{"ListTicketFields", func(c *Client) error { _, e := c.ListTicketFields(ctx, ListOptions{}); return e },
			"GET", "/api/v2/admin/ticket_fields", `[]`},
		{"UpdateTicketField", func(c *Client) error { _, e := c.UpdateTicketField(ctx, 1, TicketFieldRequest{}); return e },
			"PUT", "/api/v2/admin/ticket_fields/1", `{}`},
		{"DeleteTicketField", func(c *Client) error { return c.DeleteTicketField(ctx, 1) },
			"DELETE", "/api/v2/admin/ticket_fields/1", ``},
		{"GetSection", func(c *Client) error { _, e := c.GetSection(ctx, 22, 3); return e },
			"GET", "/api/v2/admin/ticket_fields/22/sections/3", `{}`},
		{"DeleteSection", func(c *Client) error { return c.DeleteSection(ctx, 22, 3) },
			"DELETE", "/api/v2/admin/ticket_fields/22/sections/3", ``},

		{"GetContactField", func(c *Client) error { _, e := c.GetContactField(ctx, 1); return e },
			"GET", "/api/v2/contact_fields/1", `{}`},
		{"ListContactFields", func(c *Client) error { _, e := c.ListContactFields(ctx, ListOptions{}); return e },
			"GET", "/api/v2/contact_fields", `[]`},
		{"CreateContactField", func(c *Client) error { _, e := c.CreateContactField(ctx, ContactFieldRequest{}); return e },
			"POST", "/api/v2/contact_fields", `{}`},
		{"UpdateContactField", func(c *Client) error {
			_, e := c.UpdateContactField(ctx, 1, ContactFieldRequest{})
			return e
		},
			"PUT", "/api/v2/contact_fields/1", `{}`},

		{"GetCompanyField", func(c *Client) error { _, e := c.GetCompanyField(ctx, 1); return e },
			"GET", "/api/v2/company_fields/1", `{}`},
		{"ListCompanyFields", func(c *Client) error { _, e := c.ListCompanyFields(ctx, ListOptions{}); return e },
			"GET", "/api/v2/company_fields", `[]`},
		{"CreateCompanyField", func(c *Client) error { _, e := c.CreateCompanyField(ctx, CompanyFieldRequest{}); return e },
			"POST", "/api/v2/company_fields", `{}`},
		{"DeleteCompanyField", func(c *Client) error { return c.DeleteCompanyField(ctx, 1) },
			"DELETE", "/api/v2/company_fields/1", ``},

		{"GetTicketForm", func(c *Client) error { _, e := c.GetTicketForm(ctx, 1); return e },
			"GET", "/api/v2/ticket-forms/1", `{}`},
		{"ListTicketForms", func(c *Client) error { _, e := c.ListTicketForms(ctx, ListOptions{}); return e },
			"GET", "/api/v2/ticket-forms", `[]`},
		{"CreateTicketForm", func(c *Client) error { _, e := c.CreateTicketForm(ctx, TicketFormRequest{}); return e },
			"POST", "/api/v2/ticket-forms", `{}`},
		{"UpdateTicketForm", func(c *Client) error { _, e := c.UpdateTicketForm(ctx, 1, TicketFormRequest{}); return e },
			"PUT", "/api/v2/ticket-forms/1", `{}`},
		{"DeleteTicketForm", func(c *Client) error { return c.DeleteTicketForm(ctx, 1) },
			"DELETE", "/api/v2/ticket-forms/1", ``},
		{"GetTicketFormField", func(c *Client) error { _, e := c.GetTicketFormField(ctx, 1, 5); return e },
			"GET", "/api/v2/ticket-forms/1/fields/5", `{}`},
		{"DeleteTicketFormField", func(c *Client) error { return c.DeleteTicketFormField(ctx, 1, 5) },
			"DELETE", "/api/v2/ticket-forms/1/fields/5", ``},
	})
}

// TestRoutingSolutionsAndDiscussions covers the knowledge base and forums.
func TestRoutingSolutionsAndDiscussions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"CreateSolutionCategory", func(c *Client) error {
			_, e := c.CreateSolutionCategory(ctx, SolutionCategoryRequest{})
			return e
		}, "POST", "/api/v2/solutions/categories", `{}`},
		{"ListSolutionCategories", func(c *Client) error {
			_, e := c.ListSolutionCategories(ctx, "", ListOptions{})
			return e
		}, "GET", "/api/v2/solutions/categories", `[]`},
		{"ListSolutionCategoriesTranslated", func(c *Client) error {
			_, e := c.ListSolutionCategories(ctx, "es", ListOptions{})
			return e
		}, "GET", "/api/v2/solutions/categories/es", `[]`},
		{"UpdateSolutionCategory", func(c *Client) error {
			_, e := c.UpdateSolutionCategory(ctx, 3, "", SolutionCategoryRequest{})
			return e
		}, "PUT", "/api/v2/solutions/categories/3", `{}`},
		{"CreateSolutionCategoryTranslation", func(c *Client) error {
			_, e := c.CreateSolutionCategoryTranslation(ctx, 3, "es", SolutionCategoryRequest{})
			return e
		}, "POST", "/api/v2/solutions/categories/3/es", `{}`},
		{"DeleteSolutionCategory", func(c *Client) error { return c.DeleteSolutionCategory(ctx, 3) },
			"DELETE", "/api/v2/solutions/categories/3", ``},

		{"GetSolutionFolder", func(c *Client) error { _, e := c.GetSolutionFolder(ctx, 2, ""); return e },
			"GET", "/api/v2/solutions/folders/2", `{}`},
		{"CreateSolutionFolder", func(c *Client) error {
			_, e := c.CreateSolutionFolder(ctx, 3, SolutionFolderRequest{})
			return e
		}, "POST", "/api/v2/solutions/categories/3/folders", `{}`},
		{"CreateSolutionFolderTranslation", func(c *Client) error {
			_, e := c.CreateSolutionFolderTranslation(ctx, 2, "es", SolutionFolderRequest{})
			return e
		}, "POST", "/api/v2/solutions/folders/2/es", `{}`},
		{"UpdateSolutionFolder", func(c *Client) error {
			_, e := c.UpdateSolutionFolder(ctx, 2, "", SolutionFolderRequest{})
			return e
		}, "PUT", "/api/v2/solutions/folders/2", `{}`},
		{"DeleteSolutionFolder", func(c *Client) error { return c.DeleteSolutionFolder(ctx, 2) },
			"DELETE", "/api/v2/solutions/folders/2", ``},

		{"GetSolutionArticle", func(c *Client) error { _, e := c.GetSolutionArticle(ctx, 7, ""); return e },
			"GET", "/api/v2/solutions/articles/7", `{}`},
		{"ListSolutionArticles", func(c *Client) error {
			_, e := c.ListSolutionArticles(ctx, 4, "", ListOptions{})
			return e
		}, "GET", "/api/v2/solutions/folders/4/articles", `[]`},
		{"CreateSolutionArticleTranslation", func(c *Client) error {
			_, e := c.CreateSolutionArticleTranslation(ctx, 7, "es", SolutionArticleRequest{})
			return e
		}, "POST", "/api/v2/solutions/articles/7/es", `{}`},
		{"UpdateSolutionArticle", func(c *Client) error {
			_, e := c.UpdateSolutionArticle(ctx, 7, "", SolutionArticleRequest{})
			return e
		}, "PUT", "/api/v2/solutions/articles/7", `{}`},
		{"DeleteSolutionArticle", func(c *Client) error { return c.DeleteSolutionArticle(ctx, 7) },
			"DELETE", "/api/v2/solutions/articles/7", ``},

		{"GetForumCategory", func(c *Client) error { _, e := c.GetForumCategory(ctx, 1); return e },
			"GET", "/api/v2/discussions/categories/1", `{}`},
		{"ListForumCategories", func(c *Client) error { _, e := c.ListForumCategories(ctx, ListOptions{}); return e },
			"GET", "/api/v2/discussions/categories", `[]`},
		{"CreateForumCategory", func(c *Client) error {
			_, e := c.CreateForumCategory(ctx, ForumCategoryRequest{})
			return e
		}, "POST", "/api/v2/discussions/categories", `{}`},
		{"UpdateForumCategory", func(c *Client) error {
			_, e := c.UpdateForumCategory(ctx, 1, ForumCategoryRequest{})
			return e
		}, "PUT", "/api/v2/discussions/categories/1", `{}`},
		{"DeleteForumCategory", func(c *Client) error { return c.DeleteForumCategory(ctx, 1) },
			"DELETE", "/api/v2/discussions/categories/1", ``},

		{"GetForum", func(c *Client) error { _, e := c.GetForum(ctx, 5); return e },
			"GET", "/api/v2/discussions/forums/5", `{}`},
		{"ListForums", func(c *Client) error { _, e := c.ListForums(ctx, 1, ListOptions{}); return e },
			"GET", "/api/v2/discussions/categories/1/forums", `[]`},
		{"UpdateForum", func(c *Client) error { _, e := c.UpdateForum(ctx, 5, ForumRequest{}); return e },
			"PUT", "/api/v2/discussions/forums/5", `{}`},
		{"DeleteForum", func(c *Client) error { return c.DeleteForum(ctx, 5) },
			"DELETE", "/api/v2/discussions/forums/5", ``},
		{"FollowForum", func(c *Client) error { return c.FollowForum(ctx, 5, 2) },
			"POST", "/api/v2/discussions/forums/5/follow", ``},
		{"UnfollowForum", func(c *Client) error { return c.UnfollowForum(ctx, 5, 2) },
			"DELETE", "/api/v2/discussions/forums/5/follow", ``},

		{"GetTopic", func(c *Client) error { _, e := c.GetTopic(ctx, 9); return e },
			"GET", "/api/v2/discussions/topics/9", `{}`},
		{"CreateTopic", func(c *Client) error { _, e := c.CreateTopic(ctx, 5, TopicRequest{}); return e },
			"POST", "/api/v2/discussions/forums/5/topics", `{}`},
		{"UpdateTopic", func(c *Client) error { _, e := c.UpdateTopic(ctx, 9, TopicRequest{}); return e },
			"PUT", "/api/v2/discussions/topics/9", `{}`},
		{"DeleteTopic", func(c *Client) error { return c.DeleteTopic(ctx, 9) },
			"DELETE", "/api/v2/discussions/topics/9", ``},
		{"UnfollowTopic", func(c *Client) error { return c.UnfollowTopic(ctx, 9, 2) },
			"DELETE", "/api/v2/discussions/topics/9/follow", ``},
		{"ListTopicsParticipatedBy", func(c *Client) error {
			_, e := c.ListTopicsParticipatedBy(ctx, 2, ListOptions{})
			return e
		}, "GET", "/api/v2/discussions/topics/participated_by", `[]`},

		{"CreateComment", func(c *Client) error { _, e := c.CreateComment(ctx, 9, CommentRequest{}); return e },
			"POST", "/api/v2/discussions/topics/9/comments", `{}`},
		{"UpdateComment", func(c *Client) error { _, e := c.UpdateComment(ctx, 4, CommentRequest{}); return e },
			"PUT", "/api/v2/discussions/comments/4", `{}`},
		{"DeleteComment", func(c *Client) error { return c.DeleteComment(ctx, 4) },
			"DELETE", "/api/v2/discussions/comments/4", ``},
	})
}

// TestRoutingOperations covers SLA, automations, canned responses, time
// entries, email, surveys, custom objects and collaboration.
func TestRoutingOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"ListSLAPolicies", func(c *Client) error { _, e := c.ListSLAPolicies(ctx, ListOptions{}); return e },
			"GET", "/api/v2/sla_policies", `[]`},
		{"UpdateSLAPolicy", func(c *Client) error { _, e := c.UpdateSLAPolicy(ctx, 1, SLAPolicyRequest{}); return e },
			"PUT", "/api/v2/sla_policies/1", `{}`},
		{"DeleteSLAPolicy", func(c *Client) error { return c.DeleteSLAPolicy(ctx, 1) },
			"DELETE", "/api/v2/sla_policies/1", ``},

		{"GetAutomationRule", func(c *Client) error { _, e := c.GetAutomationRule(ctx, 4, 8); return e },
			"GET", "/api/v2/automations/4/rules/8", `{}`},
		{"UpdateAutomationRule", func(c *Client) error {
			_, e := c.UpdateAutomationRule(ctx, 4, 8, AutomationRuleRequest{})
			return e
		}, "PUT", "/api/v2/automations/4/rules/8", `{}`},
		{"DeleteAutomationRule", func(c *Client) error { return c.DeleteAutomationRule(ctx, 4, 8) },
			"DELETE", "/api/v2/automations/4/rules/8", ``},
		{"GetScenarioAutomation", func(c *Client) error { _, e := c.GetScenarioAutomation(ctx, 2); return e },
			"GET", "/api/v2/scenario_automations/2", `{}`},

		{"GetCannedResponse", func(c *Client) error { _, e := c.GetCannedResponse(ctx, 1); return e },
			"GET", "/api/v2/canned_responses/1", `{}`},
		{"CreateCannedResponse", func(c *Client) error {
			_, e := c.CreateCannedResponse(ctx, CannedResponseRequest{})
			return e
		}, "POST", "/api/v2/canned_responses", `{}`},
		{"UpdateCannedResponse", func(c *Client) error {
			_, e := c.UpdateCannedResponse(ctx, 1, CannedResponseRequest{})
			return e
		}, "PUT", "/api/v2/canned_responses/1", `{}`},
		{"DeleteCannedResponse", func(c *Client) error { return c.DeleteCannedResponse(ctx, 1) },
			"DELETE", "/api/v2/canned_responses/1", ``},
		{"GetCannedResponseFolder", func(c *Client) error { _, e := c.GetCannedResponseFolder(ctx, 1); return e },
			"GET", "/api/v2/canned_response_folders/1", `{}`},
		{"UpdateCannedResponseFolder", func(c *Client) error {
			_, e := c.UpdateCannedResponseFolder(ctx, 1, CannedResponseFolderRequest{})
			return e
		}, "PUT", "/api/v2/canned_response_folders/1", `{}`},
		{"DeleteCannedResponseFolder", func(c *Client) error { return c.DeleteCannedResponseFolder(ctx, 1) },
			"DELETE", "/api/v2/canned_response_folders/1", ``},

		{"ListTimeEntries", func(c *Client) error { _, e := c.ListTimeEntries(ctx, TimeEntryListOptions{}); return e },
			"GET", "/api/v2/time_entries", `[]`},
		{"UpdateTimeEntry", func(c *Client) error { _, e := c.UpdateTimeEntry(ctx, 3, TimeEntryRequest{}); return e },
			"PUT", "/api/v2/time_entries/3", `{}`},
		{"DeleteTimeEntry", func(c *Client) error { return c.DeleteTimeEntry(ctx, 3) },
			"DELETE", "/api/v2/time_entries/3", ``},

		{"GetEmailConfig", func(c *Client) error { _, e := c.GetEmailConfig(ctx, 1); return e },
			"GET", "/api/v2/email_configs/1", `{}`},
		{"GetEmailMailbox", func(c *Client) error { _, e := c.GetEmailMailbox(ctx, 1); return e },
			"GET", "/api/v2/email/mailboxes/1", `{}`},
		{"ListEmailMailboxes", func(c *Client) error { _, e := c.ListEmailMailboxes(ctx, ListOptions{}); return e },
			"GET", "/api/v2/email/mailboxes", `[]`},
		{"UpdateEmailMailbox", func(c *Client) error {
			_, e := c.UpdateEmailMailbox(ctx, 1, EmailMailboxRequest{})
			return e
		}, "PUT", "/api/v2/email/mailboxes/1", `{}`},
		{"DeleteEmailMailbox", func(c *Client) error { return c.DeleteEmailMailbox(ctx, 1) },
			"DELETE", "/api/v2/email/mailboxes/1", ``},
		{"UpdateEmailSettings", func(c *Client) error {
			_, e := c.UpdateEmailSettings(ctx, map[string]bool{"multiple_to": true})
			return e
		}, "PUT", "/api/v2/email/settings", `{}`},

		{"GetProduct", func(c *Client) error { _, e := c.GetProduct(ctx, 1); return e },
			"GET", "/api/v2/products/1", `{}`},
		{"ListProducts", func(c *Client) error { _, e := c.ListProducts(ctx, ListOptions{}); return e },
			"GET", "/api/v2/products", `[]`},
		{"GetBusinessHours", func(c *Client) error { _, e := c.GetBusinessHours(ctx, 1); return e },
			"GET", "/api/v2/business_hours/1", `{}`},
		{"ListBusinessHours", func(c *Client) error { _, e := c.ListBusinessHours(ctx, ListOptions{}); return e },
			"GET", "/api/v2/business_hours", `[]`},

		{"CreateSurveyResponse", func(c *Client) error {
			_, e := c.CreateSurveyResponse(ctx, "uuid-1", map[string]any{})
			return e
		}, "POST", "/api/v2/customer-satisfaction/surveys/uuid-1/responses", `{}`},

		{"GetCustomObjectSchema", func(c *Client) error { _, e := c.GetCustomObjectSchema(ctx, "sch-1"); return e },
			"GET", "/api/v2/custom_objects/schemas/sch-1", `{}`},
		{"GetCustomObjectRecord", func(c *Client) error {
			_, e := c.GetCustomObjectRecord(ctx, "sch-1", "BKG-1")
			return e
		}, "GET", "/api/v2/custom_objects/schemas/sch-1/records/BKG-1", `{}`},
		{"CreateCustomObjectRecord", func(c *Client) error {
			_, e := c.CreateCustomObjectRecord(ctx, "sch-1", map[string]any{"a": 1})
			return e
		}, "POST", "/api/v2/custom_objects/schemas/sch-1/records", `{}`},
		{"UpdateCustomObjectRecord", func(c *Client) error {
			_, e := c.UpdateCustomObjectRecord(ctx, "sch-1", "BKG-1", map[string]any{"a": 2})
			return e
		}, "PUT", "/api/v2/custom_objects/schemas/sch-1/records/BKG-1", `{}`},

		{"GetThread", func(c *Client) error { _, e := c.GetThread(ctx, 2); return e },
			"GET", "/api/v2/collaboration/threads/2", `{}`},
		{"UpdateThread", func(c *Client) error { _, e := c.UpdateThread(ctx, 2, ThreadRequest{}); return e },
			"PUT", "/api/v2/collaboration/threads/2", `{}`},
		{"DeleteThread", func(c *Client) error { return c.DeleteThread(ctx, 2) },
			"DELETE", "/api/v2/collaboration/threads/2", ``},
		{"GetThreadMessage", func(c *Client) error { _, e := c.GetThreadMessage(ctx, 3); return e },
			"GET", "/api/v2/collaboration/messages/3", `{}`},
		{"UpdateThreadMessage", func(c *Client) error {
			_, e := c.UpdateThreadMessage(ctx, 3, ThreadMessageRequest{})
			return e
		}, "PUT", "/api/v2/collaboration/messages/3", `{}`},
		{"DeleteThreadMessage", func(c *Client) error { return c.DeleteThreadMessage(ctx, 3) },
			"DELETE", "/api/v2/collaboration/messages/3", ``},
		{"GetOutboundMessage", func(c *Client) error { _, e := c.GetOutboundMessage(ctx, "01K99"); return e },
			"GET", "/api/v2/channels/outbound-messages/01K99", `{}`},

		{"ExportAccount", func(c *Client) error { _, e := c.ExportAccount(ctx, map[string]any{}); return e },
			"POST", "/api/v2/account/export", `{}`},
	})
}

// TestGetTimeEntryFiltersCollection covers the client-side lookup used because
// Freshdesk offers no single-entry read.
func TestGetTimeEntryFiltersCollection(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":3,"note":"a"},{"id":9,"note":"b"}]`))
	})
	got, err := c.GetTimeEntry(context.Background(), 9)
	if err != nil {
		t.Fatalf("GetTimeEntry: %v", err)
	}
	if got.Note != "b" {
		t.Errorf("Note = %q, want b", got.Note)
	}
	if _, err := c.GetTimeEntry(context.Background(), 404); !NotFound(err) {
		t.Errorf("missing entry: err = %v, want NotFound", err)
	}
}

// TestImportsUploadCSV covers the multipart import entry points.
func TestImportsUploadCSV(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	csv := filepath.Join(dir, "contacts.csv")
	if err := os.WriteFile(csv, []byte("name,email\n"), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"ImportContacts", func(c *Client) error {
			_, e := c.ImportContacts(ctx, csv, map[string]string{"name": "Name"})
			return e
		}, "POST", "/api/v2/contacts/imports", `{}`},
		{"ImportCompanies", func(c *Client) error {
			_, e := c.ImportCompanies(ctx, csv, map[string]string{"name": "Name"})
			return e
		}, "POST", "/api/v2/companies/imports", `{}`},
	})
}

// TestLowLevelVerbs exercises the exported escape hatch used for endpoints the
// typed API does not model.
func TestLowLevelVerbs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runRouteCases(t, []routeCase{
		{"Get", func(c *Client) error {
			var out map[string]any
			return c.Get(ctx, "anything", url.Values{"a": {"b"}}, &out)
		}, "GET", "/api/v2/anything", `{}`},
		{"Post", func(c *Client) error { return c.Post(ctx, "anything", map[string]any{"x": 1}, nil) },
			"POST", "/api/v2/anything", ``},
		{"Put", func(c *Client) error { return c.Put(ctx, "anything", map[string]any{"x": 1}, nil) },
			"PUT", "/api/v2/anything", ``},
		{"Delete", func(c *Client) error { return c.Delete(ctx, "anything") },
			"DELETE", "/api/v2/anything", ``},
		{"Do", func(c *Client) error {
			_, err := c.Do(ctx, Request{Method: "PATCH", Path: "anything"}, nil)
			return err
		}, "PATCH", "/api/v2/anything", ``},
	})
}

func TestBaseURL(t *testing.T) {
	t.Parallel()

	c, err := New(Config{Domain: "acme", APIKey: "k"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.BaseURL(); got != "https://acme.freshdesk.com/api/v2/" {
		t.Errorf("BaseURL() = %q", got)
	}
}

func TestPostMultipartDirect(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	c, rec := newRecorder(t, `{}`)
	err := c.PostMultipart(context.Background(), "uploads",
		map[string][]string{"note": {"hi"}}, map[string][]string{"file": {f}}, nil)
	if err != nil {
		t.Fatalf("PostMultipart: %v", err)
	}
	if rec.path != "/api/v2/uploads" || rec.method != "POST" {
		t.Errorf("%s %s", rec.method, rec.path)
	}
}

func TestResolveRejectsInvalidPath(t *testing.T) {
	t.Parallel()

	c, err := New(Config{Domain: "acme", APIKey: "k"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.Do(context.Background(), Request{Method: "GET", Path: "://bad"}, nil); err == nil {
		t.Fatal("want error for an unparseable path")
	}
}
