package provider

import (
	"context"
	"fmt"
	"regexp"
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

// TestAccAutomationRuleWebhook reproduces the shape that made Freshdesk answer
// 500: a trigger_webhook action carries request_type, content_type,
// content_layout, url, content and custom_headers, and dropping any of them
// leaves the API with an action it cannot process.
func TestAccAutomationRuleWebhook(t *testing.T) {
	name := accName(t, "webhook")

	// example.test is reserved by RFC 6761 and never resolves, so the rule
	// cannot reach anything even if Freshdesk fires it.
	const action = `[{
    field_name     = "trigger_webhook"
    request_type   = "POST"
    content_type   = "JSON"
    content_layout = "2"
    url            = "https://hooks.example.test/tfacc"
    content        = { ticket_id = "{{ticket.id}}" }
    custom_headers = { "x-acceptance-secret" = "not-a-real-secret" }
  }]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_automation_rule",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetAutomationRule(ctx, freshdesk.AutomationTypeTicketUpdate, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_automation_rule" "webhook" {
  rule_type = 4
  name      = %q
  active    = false
  performer = jsonencode({ type = 1 })

  events = jsonencode([
    { field_name = "status", from = "--", to = "--" },
  ])

  actions = jsonencode(%s)
}`, name, action),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_automation_rule.webhook", "name", name),
					// The whole action must survive the round trip, not just
					// field_name.
					resource.TestMatchResourceAttr("freshdesk_automation_rule.webhook",
						"actions", regexp.MustCompile(`hooks\.example\.test`)),
					resource.TestMatchResourceAttr("freshdesk_automation_rule.webhook",
						"actions", regexp.MustCompile(`custom_headers`)),
					resource.TestMatchResourceAttr("freshdesk_automation_rule.webhook",
						"actions", regexp.MustCompile(`"content_type":"JSON"`)),
				),
			},
			{
				// Activating it exercises the update path with the same body.
				Config: fmt.Sprintf(`
resource "freshdesk_automation_rule" "webhook" {
  rule_type = 4
  name      = "%s-live"
  active    = true
  performer = jsonencode({ type = 1 })

  events = jsonencode([
    { field_name = "status", from = "--", to = "--" },
  ])

  actions = jsonencode(%s)
}`, name, action),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_automation_rule.webhook",
						"active", "true"),
					resource.TestMatchResourceAttr("freshdesk_automation_rule.webhook",
						"actions", regexp.MustCompile(`hooks\.example\.test`)),
				),
			},
		},
	})
}
