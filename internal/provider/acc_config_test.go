package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// TestAccTicketField covers a custom dropdown field and its choices.
func TestAccTicketField(t *testing.T) {
	label := accName(t, "field")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_ticket_field",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetTicketField(ctx, id, false)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_field" "test" {
  label                  = %q
  label_for_customers    = "Issue Type"
  type                   = "custom_dropdown"
  displayed_to_customers = true

  choices = [
    { value = "Refund" },
    { value = "Faulty Product" },
  ]
}`, label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test", "label", label),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test",
						"type", "custom_dropdown"),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test", "choices.#", "2"),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test",
						"choices.0.value", "Refund"),
					resource.TestCheckResourceAttrSet("freshdesk_ticket_field.test", "name"),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test", "default", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_field" "test" {
  label                  = "%s-updated"
  label_for_customers    = "Issue Category"
  type                   = "custom_dropdown"
  displayed_to_customers = true
  required_for_agents    = true

  choices = [
    { value = "Refund" },
    { value = "Faulty Product" },
    { value = "Not Delivered" },
  ]
}`, label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test",
						"label", label+"-updated"),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test", "choices.#", "3"),
					resource.TestCheckResourceAttr("freshdesk_ticket_field.test",
						"required_for_agents", "true"),
				),
			},
			{
				ResourceName:      "freshdesk_ticket_field.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccContactField covers a custom field on the contact form.
func TestAccContactField(t *testing.T) {
	label := accName(t, "cf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_contact_field",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetContactField(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_contact_field" "test" {
  label               = %q
  label_for_customers = %q
  type                = "custom_text"
}`, label, label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_contact_field.test", "label", label),
					resource.TestCheckResourceAttr("freshdesk_contact_field.test",
						"type", "custom_text"),
					resource.TestCheckResourceAttrSet("freshdesk_contact_field.test", "name"),
				),
			},
			{
				ResourceName:      "freshdesk_contact_field.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccCompanyField covers a custom field on the company form.
func TestAccCompanyField(t *testing.T) {
	label := accName(t, "cof")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_company_field",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetCompanyField(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_company_field" "test" {
  label = %q
  type  = "custom_text"
}`, label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_company_field.test", "label", label),
					resource.TestCheckResourceAttrSet("freshdesk_company_field.test", "name"),
				),
			},
			{
				ResourceName:      "freshdesk_company_field.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccCannedResponse covers a folder and a response inside it.
func TestAccCannedResponse(t *testing.T) {
	name := accName(t, "cr")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		// No CheckDestroy: Freshdesk answers 405 to a delete on both canned
		// response folders and canned responses, so the records survive by
		// design and the provider warns about it.
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_canned_response_folder" "test" {
  name = %q
}

resource "freshdesk_canned_response" "test" {
  folder_id    = freshdesk_canned_response_folder.test.id
  title        = "%s-response"
  content_html = "<p>Thanks for getting in touch.</p>"
  visibility   = 0
}`, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_canned_response_folder.test",
						"name", name),
					resource.TestCheckResourceAttr("freshdesk_canned_response.test",
						"title", name+"-response"),
					resource.TestCheckResourceAttrPair("freshdesk_canned_response.test",
						"folder_id", "freshdesk_canned_response_folder.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_canned_response_folder" "test" {
  name = "%s-renamed"
}

resource "freshdesk_canned_response" "test" {
  folder_id    = freshdesk_canned_response_folder.test.id
  title        = "%s-response-updated"
  content_html = "<p>Thanks very much for getting in touch.</p>"
  visibility   = 0
}`, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_canned_response_folder.test",
						"name", name+"-renamed"),
					resource.TestCheckResourceAttr("freshdesk_canned_response.test",
						"title", name+"-response-updated"),
				),
			},
		},
	})
}

// TestAccForumTree covers the community hierarchy: a category holding a forum
// holding a topic holding a comment.
func TestAccForumTree(t *testing.T) {
	name := accName(t, "forum")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_forum_category",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetForumCategory(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_forum_category" "test" {
  name        = %q
  description = "created by the acceptance suite"
}

resource "freshdesk_forum" "test" {
  forum_category_id = freshdesk_forum_category.test.id
  name              = "%s-forum"
  description       = "created by the acceptance suite"
  forum_type        = 1
  forum_visibility  = 1
}

resource "freshdesk_topic" "test" {
  forum_id = freshdesk_forum.test.id
  title    = "%s-topic"
  message  = "<p>Opened by the acceptance suite.</p>"
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_forum_category.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_forum.test", "forum_type", "1"),
					resource.TestCheckResourceAttr("freshdesk_topic.test", "title", name+"-topic"),
					resource.TestCheckResourceAttrPair("freshdesk_forum.test",
						"forum_category_id", "freshdesk_forum_category.test", "id"),
					resource.TestCheckResourceAttrPair("freshdesk_topic.test",
						"forum_id", "freshdesk_forum.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_forum_category" "test" {
  name        = %q
  description = "updated by the acceptance suite"
}

resource "freshdesk_forum" "test" {
  forum_category_id = freshdesk_forum_category.test.id
  name              = "%s-forum-updated"
  description       = "updated by the acceptance suite"
  forum_type        = 1
  forum_visibility  = 1
}

resource "freshdesk_topic" "test" {
  forum_id = freshdesk_forum.test.id
  title    = "%s-topic-updated"
  message  = "<p>Updated by the acceptance suite.</p>"
  sticky   = true
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_forum.test",
						"name", name+"-forum-updated"),
					resource.TestCheckResourceAttr("freshdesk_topic.test", "sticky", "true"),
				),
			},
			{
				ResourceName:      "freshdesk_forum_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccTicketForm covers a portal ticket form.
func TestAccTicketForm(t *testing.T) {
	name := accName(t, "form")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_ticket_form",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetTicketForm(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_form" "test" {
  title       = %q
  description = "created by the acceptance suite"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_ticket_form.test", "title", name),
					resource.TestCheckResourceAttrSet("freshdesk_ticket_form.test", "name"),
					resource.TestCheckResourceAttr("freshdesk_ticket_form.test", "default", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_ticket_form" "test" {
  title       = "%s-updated"
  description = "updated by the acceptance suite"
}`, name),
				Check: resource.TestCheckResourceAttr("freshdesk_ticket_form.test",
					"title", name+"-updated"),
			},
			{
				ResourceName:      "freshdesk_ticket_form.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
