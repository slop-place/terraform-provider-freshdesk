data "freshdesk_roles" "all" {}

locals {
  agent_role = one([
    for r in data.freshdesk_roles.all.roles : r.id if r.name == "Agent"
  ])
}

resource "freshdesk_agent" "sam" {
  email        = "sam@example.com"
  name         = "Sam Okafor"
  ticket_scope = 2
  occasional   = false
  role_ids     = [local.agent_role]
  group_ids    = [freshdesk_group.billing.id]
  signature    = "<p>Sam &mdash; Billing</p>"
}

resource "freshdesk_group" "billing" {
  name = "Billing"
}
