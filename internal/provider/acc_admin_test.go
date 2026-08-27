package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// TestAccAdminGroup covers the admin groups API, which adds group types on top
// of what freshdesk_group offers.
func TestAccAdminGroup(t *testing.T) {
	name := accName(t, "admin-grp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_admin_group",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetAdminGroup(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_admin_group" "test" {
  name           = %q
  description    = "created by the acceptance suite"
  type           = "support_agent_group"
  unassigned_for = "2h"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_admin_group.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_admin_group.test",
						"type", "support_agent_group"),
					resource.TestCheckResourceAttr("freshdesk_admin_group.test",
						"unassigned_for", "2h"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_admin_group" "test" {
  name           = "%s-updated"
  description    = "updated by the acceptance suite"
  type           = "support_agent_group"
  unassigned_for = "8h"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_admin_group.test",
						"name", name+"-updated"),
					resource.TestCheckResourceAttr("freshdesk_admin_group.test",
						"unassigned_for", "8h"),
				),
			},
			{
				ResourceName:      "freshdesk_admin_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccTicketFieldSection covers a dynamic section and its composite import
// ID, which pairs the section with the field whose choices reveal it.
func TestAccTicketFieldSection(t *testing.T) {
	name := accName(t, "section")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_field" "parent" {
  label                  = "%[1]s-parent"
  label_for_customers    = "%[1]s-parent"
  type                   = "custom_dropdown"
  displayed_to_customers = true

  choices = [
    { value = "Mobile" },
    { value = "Web" },
  ]
}

resource "freshdesk_ticket_field_section" "test" {
  ticket_field_id = freshdesk_ticket_field.parent.id
  label           = "%[1]s-mobile"
  choice_ids      = [freshdesk_ticket_field.parent.choices[0].id]
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket_field_section.test",
						"label", name+"-mobile"),
					resource.TestCheckResourceAttrPair("freshdesk_ticket_field_section.test",
						"ticket_field_id", "freshdesk_ticket_field.parent", "id"),
					resource.TestCheckResourceAttrSet("freshdesk_ticket_field_section.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_field" "parent" {
  label                  = "%[1]s-parent"
  label_for_customers    = "%[1]s-parent"
  type                   = "custom_dropdown"
  displayed_to_customers = true

  choices = [
    { value = "Mobile" },
    { value = "Web" },
  ]
}

resource "freshdesk_ticket_field_section" "test" {
  ticket_field_id = freshdesk_ticket_field.parent.id
  label           = "%[1]s-renamed"
  choice_ids      = [freshdesk_ticket_field.parent.choices[0].id]
}`, name),
				Check: resource.TestCheckResourceAttr("freshdesk_ticket_field_section.test",
					"label", name+"-renamed"),
			},
		},
	})
}

// TestAccEmailSettings covers the account-wide email toggles. It restores the
// values it found so the sandbox is left as it was.
func TestAccEmailSettings(t *testing.T) {
	client := testAccClient(t)

	original, err := client.GetEmailSettings(context.Background())
	if err != nil {
		t.Skipf("cannot read the email settings: %v", err)
	}

	t.Cleanup(func() {
		_, restoreErr := client.UpdateEmailSettings(context.Background(), map[string]bool{
			"email_subject_match":  original.EmailSubjectMatch,
			"extended_quoted_text": original.ExtendedQuotedText,
		})
		if restoreErr != nil {
			t.Logf("restoring the email settings: %v", restoreErr)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: `
resource "freshdesk_email_settings" "test" {
  email_subject_match  = true
  extended_quoted_text = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_email_settings.test",
						"email_subject_match", "true"),
					resource.TestCheckResourceAttr("freshdesk_email_settings.test",
						"extended_quoted_text", "true"),
					resource.TestCheckResourceAttr("freshdesk_email_settings.test", "id", "freshdesk"),
				),
			},
			{
				Config: `
resource "freshdesk_email_settings" "test" {
  email_subject_match  = false
  extended_quoted_text = true
}`,
				Check: resource.TestCheckResourceAttr("freshdesk_email_settings.test",
					"email_subject_match", "false"),
			},
			{
				ResourceName:      "freshdesk_email_settings.test",
				ImportState:       true,
				ImportStateId:     "freshdesk",
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationBCC covers the automatic BCC list, restoring whatever the
// sandbox had configured.
func TestAccNotificationBCC(t *testing.T) {
	client := testAccClient(t)

	original, err := client.GetNotificationBCC(context.Background())
	if err != nil {
		t.Skipf("cannot read the BCC addresses: %v", err)
	}

	t.Cleanup(func() {
		if _, restoreErr := client.UpdateNotificationBCC(context.Background(), original); restoreErr != nil {
			t.Logf("restoring the BCC addresses: %v", restoreErr)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		Steps: []resource.TestStep{
			{
				Config: `
resource "freshdesk_notification_bcc" "test" {
  emails = ["tfacc-archive@example.com"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_notification_bcc.test",
						"emails.#", "1"),
					resource.TestCheckResourceAttr("freshdesk_notification_bcc.test",
						"emails.0", "tfacc-archive@example.com"),
				),
			},
			{
				Config: `
resource "freshdesk_notification_bcc" "test" {
  emails = []
}`,
				Check: resource.TestCheckResourceAttr("freshdesk_notification_bcc.test",
					"emails.#", "0"),
			},
		},
	})
}
