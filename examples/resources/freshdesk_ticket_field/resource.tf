resource "freshdesk_ticket_field" "issue_type" {
  label                  = "Issue Type"
  label_for_customers    = "What went wrong?"
  type                   = "custom_dropdown"
  displayed_to_customers = true
  required_for_customers = true

  choices = [
    { value = "Refund" },
    { value = "Faulty product" },
    { value = "Not delivered" },
  ]
}
