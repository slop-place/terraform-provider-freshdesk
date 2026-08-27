package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// TestAccGroup covers the full lifecycle of an agent group: create, refresh,
// update in place, import, and destroy.
func TestAccGroup(t *testing.T) {
	name := accName(t, "grp")
	updated := name + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_group",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetGroup(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_group" "test" {
  name            = %q
  description     = "created by the acceptance suite"
  unassigned_for  = "1h"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_group.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_group.test", "description",
						"created by the acceptance suite"),
					resource.TestCheckResourceAttr("freshdesk_group.test", "unassigned_for", "1h"),
					resource.TestCheckResourceAttrSet("freshdesk_group.test", "id"),
					resource.TestCheckResourceAttrSet("freshdesk_group.test", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_group" "test" {
  name            = %q
  description     = "updated by the acceptance suite"
  unassigned_for  = "4h"
}`, updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_group.test", "name", updated),
					resource.TestCheckResourceAttr("freshdesk_group.test", "description",
						"updated by the acceptance suite"),
					resource.TestCheckResourceAttr("freshdesk_group.test", "unassigned_for", "4h"),
				),
			},
			{
				ResourceName:      "freshdesk_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccCompany covers a company's lifecycle, including the set attributes
// and the date handling on renewal_date.
func TestAccCompany(t *testing.T) {
	name := accName(t, "co")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_company",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetCompany(ctx, id)

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_company" "test" {
  name         = %q
  description  = "created by the acceptance suite"
  domains      = ["tfacc-example.test", "tfacc-example.invalid"]
  renewal_date = "2030-12-31"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_company.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_company.test", "domains.#", "2"),
					resource.TestCheckResourceAttr("freshdesk_company.test",
						"renewal_date", "2030-12-31"),
				),
			},
			{
				// Removing the domains must clear them, not leave them behind.
				Config: fmt.Sprintf(`
resource "freshdesk_company" "test" {
  name        = %q
  description = "updated by the acceptance suite"
  note        = "a note"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_company.test", "note", "a note"),
					resource.TestCheckResourceAttr("freshdesk_company.test", "domains.#", "0"),
				),
			},
			{
				ResourceName:      "freshdesk_company.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API echoes the renewal date as a timestamp, and note is
				// only returned on a direct read, so both are checked above.
				ImportStateVerifyIgnore: []string{"renewal_date"},
			},
		},
	})
}

// TestAccContact covers a contact's lifecycle and its link to a company.
func TestAccContact(t *testing.T) {
	name := accName(t, "contact")
	email := name + "@tfacc-example.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_contact",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				contact, err := c.GetContact(ctx, id)
				if err != nil {
					return err
				}
				// A soft-deleted contact is still readable but flagged.
				if contact.Deleted {
					return &freshdesk.Error{StatusCode: 404, Code: "not_found",
						Message: "contact is soft-deleted"}
				}

				return nil
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_company" "test" {
  name = "%s-co"
}

resource "freshdesk_contact" "test" {
  name       = %q
  email      = %q
  job_title  = "Tester"
  company_id = freshdesk_company.test.id
  tags       = ["tfacc"]
}`, name, name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_contact.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_contact.test", "email", email),
					resource.TestCheckResourceAttr("freshdesk_contact.test", "job_title", "Tester"),
					resource.TestCheckResourceAttrPair(
						"freshdesk_contact.test", "company_id", "freshdesk_company.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "freshdesk_company" "test" {
  name = "%s-co"
}

resource "freshdesk_contact" "test" {
  name       = "%s-renamed"
  email      = %q
  job_title  = "Senior Tester"
  company_id = freshdesk_company.test.id
  tags       = ["tfacc", "updated"]
}`, name, name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_contact.test",
						"name", name+"-renamed"),
					resource.TestCheckResourceAttr("freshdesk_contact.test",
						"job_title", "Senior Tester"),
					resource.TestCheckResourceAttr("freshdesk_contact.test", "tags.#", "2"),
				),
			},
			{
				ResourceName:      "freshdesk_contact.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccSolutionTree covers the knowledge-base hierarchy end to end: a
// category holding a folder holding an article.
func TestAccSolutionTree(t *testing.T) {
	name := accName(t, "kb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6Providers,
		CheckDestroy: checkDestroyed(t, "freshdesk_solution_category",
			func(ctx context.Context, c *freshdesk.Client, id int64) error {
				_, err := c.GetSolutionCategory(ctx, id, "")

				return err
			}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "freshdesk_solution_category" "test" {
  name        = %q
  description = "created by the acceptance suite"
}

resource "freshdesk_solution_folder" "test" {
  category_id = freshdesk_solution_category.test.id
  name        = "%s-folder"
  description = "created by the acceptance suite"
  visibility  = 2
}

resource "freshdesk_solution_article" "test" {
  folder_id   = freshdesk_solution_folder.test.id
  title       = "%s-article"
  description = "<p>Written by the acceptance suite.</p>"
  status      = 1
  tags        = ["tfacc"]
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_solution_category.test", "name", name),
					resource.TestCheckResourceAttr("freshdesk_solution_folder.test",
						"visibility", "2"),
					resource.TestCheckResourceAttr("freshdesk_solution_article.test",
						"status", "1"),
					resource.TestCheckResourceAttrPair("freshdesk_solution_folder.test",
						"category_id", "freshdesk_solution_category.test", "id"),
					resource.TestCheckResourceAttrPair("freshdesk_solution_article.test",
						"folder_id", "freshdesk_solution_folder.test", "id"),
					resource.TestCheckResourceAttrSet("freshdesk_solution_article.test",
						"description_text"),
				),
			},
			{
				// Publishing the article and widening the folder must both take.
				Config: fmt.Sprintf(`
resource "freshdesk_solution_category" "test" {
  name        = %q
  description = "updated by the acceptance suite"
}

resource "freshdesk_solution_folder" "test" {
  category_id = freshdesk_solution_category.test.id
  name        = "%s-folder"
  visibility  = 1
}

resource "freshdesk_solution_article" "test" {
  folder_id   = freshdesk_solution_folder.test.id
  title       = "%s-article-updated"
  description = "<p>Updated by the acceptance suite.</p>"
  status      = 2
}`, name, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freshdesk_solution_folder.test",
						"visibility", "1"),
					resource.TestCheckResourceAttr("freshdesk_solution_article.test",
						"status", "2"),
					resource.TestCheckResourceAttr("freshdesk_solution_article.test",
						"title", name+"-article-updated"),
				),
			},
			{
				ResourceName:      "freshdesk_solution_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "freshdesk_solution_article.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API does not echo tags or SEO data on a plain read.
				ImportStateVerifyIgnore: []string{"tags"},
			},
		},
	})
}
