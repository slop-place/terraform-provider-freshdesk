package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReadOnlyDataSources exercises every data source that needs no
// fixture, checking that each one reads and populates its collection.
func TestAccReadOnlyDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: `
data "freshdesk_account" "test" {}
data "freshdesk_helpdesk_settings" "test" {}
data "freshdesk_email_settings" "test" {}

data "freshdesk_roles" "test" {}
data "freshdesk_products" "test" {}
data "freshdesk_business_hours_list" "test" {}
data "freshdesk_email_configs" "test" {}
data "freshdesk_email_mailboxes" "test" {}
data "freshdesk_groups" "test" {}
data "freshdesk_agents" "test" {}
data "freshdesk_skills" "test" {}
data "freshdesk_companies" "test" {}
data "freshdesk_ticket_fields" "test" {}
data "freshdesk_contact_fields" "test" {}
data "freshdesk_company_fields" "test" {}
data "freshdesk_ticket_forms" "test" {}
data "freshdesk_sla_policies" "test" {}
data "freshdesk_scenario_automations" "test" {}
data "freshdesk_canned_response_folders" "test" {}
data "freshdesk_solution_categories" "test" {}
data "freshdesk_forum_categories" "test" {}
data "freshdesk_surveys" "test" {}
data "freshdesk_satisfaction_ratings" "test" {}
data "freshdesk_time_entries" "test" {}

data "freshdesk_automation_rules" "dispatcher" { rule_type = 1 }
data "freshdesk_automation_rules" "time_triggers" { rule_type = 3 }
data "freshdesk_automation_rules" "observer" { rule_type = 4 }
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The account and its settings must resolve.
					resource.TestCheckResourceAttrSet("data.freshdesk_account.test", "account_name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_account.test", "tier_type"),
					resource.TestCheckResourceAttrSet("data.freshdesk_account.test", "id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_helpdesk_settings.test",
						"primary_language"),
					resource.TestCheckResourceAttrSet("data.freshdesk_email_settings.test", "id"),

					// Every helpdesk has these, so the lists must be non-empty.
					resource.TestCheckResourceAttrSet("data.freshdesk_roles.test", "roles.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_products.test",
						"products.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_business_hours_list.test",
						"business_hours.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_email_configs.test",
						"email_configs.0.to_email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_email_mailboxes.test",
						"email_mailboxes.0.support_email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_agents.test",
						"agents.0.email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_ticket_fields.test",
						"ticket_fields.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_contact_fields.test",
						"contact_fields.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_company_fields.test",
						"company_fields.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_ticket_forms.test",
						"ticket_forms.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_sla_policies.test",
						"sla_policies.0.name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_surveys.test",
						"surveys.0.title"),

					// These may legitimately be empty, so only the envelope is
					// asserted.
					resource.TestCheckResourceAttrSet("data.freshdesk_groups.test", "id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_skills.test", "id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_companies.test", "id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_scenario_automations.test",
						"id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_canned_response_folders.test",
						"id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_solution_categories.test",
						"id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_forum_categories.test", "id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_satisfaction_ratings.test",
						"id"),
					resource.TestCheckResourceAttrSet("data.freshdesk_time_entries.test", "id"),

					resource.TestCheckResourceAttr("data.freshdesk_automation_rules.dispatcher",
						"rule_type", "1"),
					resource.TestCheckResourceAttr("data.freshdesk_automation_rules.time_triggers",
						"rule_type", "3"),
					resource.TestCheckResourceAttrSet("data.freshdesk_automation_rules.observer",
						"automation_rules.0.name"),
				),
			},
		},
	})
}

// TestAccLookupDataSources creates a fixture for each singular data source and
// checks that looking it up by ID returns the same values.
func TestAccLookupDataSources(t *testing.T) {
	name := accName(t, "ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_group" "test" {
  name = "%[1]s-group"
}

resource "freshdesk_company" "test" {
  name = "%[1]s-company"
}

resource "freshdesk_contact" "test" {
  name  = "%[1]s-contact"
  email = "%[1]s-contact@tfacc-example.test"
}

resource "freshdesk_solution_category" "test" {
  name = "%[1]s-category"
}

resource "freshdesk_solution_folder" "test" {
  category_id = freshdesk_solution_category.test.id
  name        = "%[1]s-folder"
  visibility  = 2
}

resource "freshdesk_solution_article" "test" {
  folder_id   = freshdesk_solution_folder.test.id
  title       = "%[1]s-article"
  description = "<p>Read by a data source.</p>"
  status      = 1
}

resource "freshdesk_ticket_form" "test" {
  title = "%[1]s-form"
}

data "freshdesk_group" "test"             { id = freshdesk_group.test.id }
data "freshdesk_company" "test"           { id = freshdesk_company.test.id }
data "freshdesk_contact" "test"           { id = freshdesk_contact.test.id }
data "freshdesk_solution_category" "test" { id = freshdesk_solution_category.test.id }
data "freshdesk_solution_folder" "test"   { id = freshdesk_solution_folder.test.id }
data "freshdesk_solution_article" "test"  { id = freshdesk_solution_article.test.id }
data "freshdesk_ticket_form" "test"       { id = freshdesk_ticket_form.test.id }

data "freshdesk_agent" "me" {
  id = data.freshdesk_agents.all.agents[0].id
}

data "freshdesk_agents" "all" {}

data "freshdesk_role" "first" {
  id = data.freshdesk_roles.all.roles[0].id
}

data "freshdesk_roles" "all" {}

data "freshdesk_product" "first" {
  id = data.freshdesk_products.all.products[0].id
}

data "freshdesk_products" "all" {}

data "freshdesk_business_hours" "first" {
  id = data.freshdesk_business_hours_list.all.business_hours[0].id
}

data "freshdesk_business_hours_list" "all" {}

data "freshdesk_email_config" "first" {
  id = data.freshdesk_email_configs.all.email_configs[0].id
}

data "freshdesk_email_configs" "all" {}

data "freshdesk_email_mailbox" "first" {
  id = data.freshdesk_email_mailboxes.all.email_mailboxes[0].id
}

data "freshdesk_email_mailboxes" "all" {}

data "freshdesk_ticket_field" "first" {
  id = data.freshdesk_ticket_fields.all.ticket_fields[0].id
}

data "freshdesk_ticket_fields" "all" {}

data "freshdesk_sla_policy" "first" {
  id = data.freshdesk_sla_policies.all.sla_policies[0].id
}

data "freshdesk_sla_policies" "all" {}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Each lookup must agree with the resource it read.
					resource.TestCheckResourceAttrPair(
						"data.freshdesk_group.test", "name", "freshdesk_group.test", "name"),
					resource.TestCheckResourceAttrPair(
						"data.freshdesk_company.test", "name", "freshdesk_company.test", "name"),
					resource.TestCheckResourceAttrPair(
						"data.freshdesk_contact.test", "name", "freshdesk_contact.test", "name"),
					resource.TestCheckResourceAttrPair("data.freshdesk_solution_category.test",
						"name", "freshdesk_solution_category.test", "name"),
					resource.TestCheckResourceAttrPair("data.freshdesk_solution_folder.test",
						"name", "freshdesk_solution_folder.test", "name"),
					resource.TestCheckResourceAttrPair("data.freshdesk_solution_article.test",
						"title", "freshdesk_solution_article.test", "title"),
					resource.TestCheckResourceAttrPair("data.freshdesk_ticket_form.test",
						"title", "freshdesk_ticket_form.test", "title"),

					// The lookups driven off a collection must resolve too.
					resource.TestCheckResourceAttrSet("data.freshdesk_agent.me", "email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_role.first", "name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_product.first", "name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_business_hours.first",
						"time_zone"),
					resource.TestCheckResourceAttrSet("data.freshdesk_email_config.first",
						"to_email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_email_mailbox.first",
						"support_email"),
					resource.TestCheckResourceAttrSet("data.freshdesk_ticket_field.first", "name"),
					resource.TestCheckResourceAttrSet("data.freshdesk_sla_policy.first", "name"),
				),
			},
		},
	})
}
