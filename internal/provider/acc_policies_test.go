package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// TestAccSLAPolicy covers an SLA policy and its per-priority targets.
func TestAccSLAPolicy(t *testing.T) {
	name := accName(t, "sla")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		// No CheckDestroy: Freshdesk answers 405 to a delete on an SLA policy,
		// so the record survives by design and the provider warns about it.
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_sla_policy" "test" {
  name        = %q
  description = "created by the acceptance suite"

  applicable_to_sources = [1, 2, 3]

  sla_target = {
    priority_1 = { respond_within = 3600, resolve_within = 86400, next_respond_within = 3600 }
    priority_2 = { respond_within = 1800, resolve_within = 43200, next_respond_within = 1800 }
    priority_3 = { respond_within = 900,  resolve_within = 21600, next_respond_within = 900 }
    priority_4 = { respond_within = 300,  resolve_within = 7200,  next_respond_within = 300 }
  }
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test",
						"sla_target.priority_4.respond_within", "300"),
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test",
						"sla_target.priority_1.resolve_within", "86400"),
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test",
						"is_default", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_sla_policy" "test" {
  name        = "%s-updated"
  description = "updated by the acceptance suite"
  active      = true

  applicable_to_sources = [1, 2]

  sla_target = {
    priority_1 = { respond_within = 7200, resolve_within = 172800, next_respond_within = 7200 }
    priority_2 = { respond_within = 3600, resolve_within = 86400,  next_respond_within = 3600 }
    priority_3 = { respond_within = 1800, resolve_within = 43200,  next_respond_within = 1800 }
    priority_4 = { respond_within = 600,  resolve_within = 14400,  next_respond_within = 600 }
  }
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test",
						"name", name+"-updated"),
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test",
						"sla_target.priority_4.respond_within", "600"),
					resource.TestCheckResourceAttr("freshdesk_sla_policy.test", "active", "true"),
				),
			},
			{
				ResourceName:      "freshdesk_sla_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Freshdesk renumbers positions as policies come and go, and it
				// fills the escalation table in with empty per-target objects
				// that the create response omits.
				ImportStateVerifyIgnore: []string{"position", "escalation"},
			},
		},
	})
}

// TestAccAutomationRule covers an observer rule, including the JSON-carried
// conditions and actions and the composite import ID.
func TestAccAutomationRule(t *testing.T) {
	name := accName(t, "rule")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_automation_rule",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetAutomationRule(ctx, freshdesk.AutomationTypeTicketCreation, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_automation_rule" "test" {
  rule_type = 1
  name      = %q
  active    = false

  conditions = jsonencode([
    {
      name       = "condition_set_1"
      match_type = "all"
      properties = [
        { resource_type = "ticket", field_name = "priority", operator = "in", value = [4] },
      ]
    },
  ])

  actions = jsonencode([
    { field_name = "priority", value = 4 },
  ])
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_automation_rule.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_automation_rule.test", "rule_type", "1"),
					resource.TestCheckResourceAttr("freshdesk_automation_rule.test", "active", "false"),
					resource.TestCheckResourceAttrSet("freshdesk_automation_rule.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_automation_rule" "test" {
  rule_type = 1
  name      = "%s-updated"
  active    = false

  conditions = jsonencode([
    {
      name       = "condition_set_1"
      match_type = "all"
      properties = [
        { resource_type = "ticket", field_name = "priority", operator = "in", value = [3] },
      ]
    },
  ])

  actions = jsonencode([
    { field_name = "priority", value = 3 },
  ])
}`, name),
				Check: resource.TestCheckResourceAttr("freshdesk_automation_rule.test",
					"name", name+"-updated"),
			},
			{
				ResourceName: "freshdesk_automation_rule.test",
				ImportState:  true,
				// A rule is addressed by its automation type as well as its ID.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["freshdesk_automation_rule.test"]

					return "1:" + rs.Primary.ID, nil
				},
				ImportStateVerify: true,
				// The API normalises the JSON bodies, so they are compared by
				// the checks above rather than byte-for-byte.
				ImportStateVerifyIgnore: []string{"conditions", "actions", "performer", "events"},
			},
		},
	})
}
