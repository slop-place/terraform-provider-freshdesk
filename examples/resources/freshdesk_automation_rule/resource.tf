# A dispatcher rule (rule_type 1) that routes urgent tickets to a group.
resource "freshdesk_automation_rule" "urgent_to_billing" {
  rule_type = 1
  name      = "Route urgent billing tickets"
  active    = true

  conditions = jsonencode([
    {
      name       = "condition_set_1"
      match_type = "all"
      properties = [
        {
          resource_type = "ticket"
          field_name    = "priority"
          operator      = "in"
          value         = [4]
        },
      ]
    },
  ])

  actions = jsonencode([
    { field_name = "group_id", value = freshdesk_group.billing.id },
  ])
}

resource "freshdesk_group" "billing" {
  name = "Billing"
}
