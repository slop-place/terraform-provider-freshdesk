# Useful for finding the generated `name` of a custom field, which is what
# `custom_fields` on a ticket or contact is keyed by.
data "freshdesk_ticket_fields" "all" {}

output "custom_field_names" {
  value = [for f in data.freshdesk_ticket_fields.all.ticket_fields : f.name if !f.default]
}
