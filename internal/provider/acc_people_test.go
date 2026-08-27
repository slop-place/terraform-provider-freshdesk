package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// TestAccSkill covers a routing skill and its JSON-carried conditions.
func TestAccSkill(t *testing.T) {
	name := accName(t, "skill")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_skill",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetSkill(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_skill" "test" {
  name = %q

  conditions = jsonencode([
    {
      channel = "ticket"
      channel_conditions = [
        {
          name       = "condition_set_1"
          match_type = "all"
          properties = [
            { resource_type = "ticket", field_name = "priority", operator = "in", value = [4] },
          ]
        },
      ]
    },
  ])
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_skill.test", "name", name),
					resource.TestCheckResourceAttrSet("freshdesk_skill.test", "rank"),
					resource.TestCheckResourceAttrSet("freshdesk_skill.test", "conditions"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_skill" "test" {
  name = "%s-updated"

  conditions = jsonencode([
    {
      channel = "ticket"
      channel_conditions = [
        {
          name       = "condition_set_1"
          match_type = "all"
          properties = [
            { resource_type = "ticket", field_name = "priority", operator = "in", value = [3, 4] },
          ]
        },
      ]
    },
  ])
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_skill.test", "name", name+"-updated"),
				),
			},
			{
				ResourceName:      "freshdesk_skill.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API normalises the condition body, and its updated_at
				// ticks a second between the write response and the next read.
				ImportStateVerifyIgnore: []string{"conditions", "updated_at"},
			},
		},
	})
}

// TestAccTicket covers a ticket's lifecycle, including its custom fields and
// the requester link.
func TestAccTicket(t *testing.T) {
	name := accName(t, "ticket")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_ticket",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				ticket, err := c.GetTicket(ctx, id, freshdesk.TicketIncludes{})
				if err != nil {
					return err
				}
				// A trashed ticket is still readable but flagged deleted.
				if ticket.Deleted {
					return &freshdesk.Error{StatusCode: 404, Code: "not_found",
						Message: "ticket is in the trash"}
				}

				return nil
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_contact" "requester" {
  name  = "%s-requester"
  email = "%s-requester@tfacc-example.test"
}

resource "freshdesk_ticket" "test" {
  subject      = %q
  description  = "<p>Raised by the acceptance suite.</p>"
  requester_id = freshdesk_contact.requester.id
  priority     = 2
  status       = 2
  tags         = ["tfacc"]
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "subject", name),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "priority", "2"),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "status", "2"),
					resource.TestCheckResourceAttrPair("freshdesk_ticket.test",
						"requester_id", "freshdesk_contact.requester", "id"),
					resource.TestCheckResourceAttrSet("freshdesk_ticket.test", "description_text"),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "spam", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_contact" "requester" {
  name  = "%s-requester"
  email = "%s-requester@tfacc-example.test"
}

resource "freshdesk_ticket" "test" {
  subject      = "%s-updated"
  description  = "<p>Raised by the acceptance suite.</p>"
  requester_id = freshdesk_contact.requester.id
  priority     = 4
  status       = 3
  tags         = ["tfacc", "escalated"]
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "subject", name+"-updated"),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "priority", "4"),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "status", "3"),
					resource.TestCheckResourceAttr("freshdesk_ticket.test", "tags.#", "2"),
				),
			},
		},
	})
}

// TestAccTimeEntry covers time logged against a ticket.
func TestAccTimeEntry(t *testing.T) {
	name := accName(t, "time")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_contact" "requester" {
  name  = "%s-requester"
  email = "%s-requester@tfacc-example.test"
}

resource "freshdesk_ticket" "test" {
  subject      = %q
  description  = "<p>Raised by the acceptance suite.</p>"
  requester_id = freshdesk_contact.requester.id
}

resource "freshdesk_time_entry" "test" {
  ticket_id  = freshdesk_ticket.test.id
  note       = "investigating"
  time_spent = "01:30"
  billable   = true
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_time_entry.test",
						"note", "investigating"),
					resource.TestCheckResourceAttr("freshdesk_time_entry.test",
						"time_spent", "01:30"),
					resource.TestCheckResourceAttr("freshdesk_time_entry.test", "billable", "true"),
					resource.TestCheckResourceAttrPair("freshdesk_time_entry.test",
						"ticket_id", "freshdesk_ticket.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_contact" "requester" {
  name  = "%s-requester"
  email = "%s-requester@tfacc-example.test"
}

resource "freshdesk_ticket" "test" {
  subject      = %q
  description  = "<p>Raised by the acceptance suite.</p>"
  requester_id = freshdesk_contact.requester.id
}

resource "freshdesk_time_entry" "test" {
  ticket_id  = freshdesk_ticket.test.id
  note       = "root cause found"
  time_spent = "02:15"
  billable   = false
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_time_entry.test",
						"note", "root cause found"),
					resource.TestCheckResourceAttr("freshdesk_time_entry.test",
						"time_spent", "02:15"),
					resource.TestCheckResourceAttr("freshdesk_time_entry.test", "billable", "false"),
				),
			},
		},
	})
}
