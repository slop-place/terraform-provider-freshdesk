resource "freshdesk_group" "billing" {
  name           = "Billing"
  description    = "Invoices, refunds and payment questions"
  unassigned_for = "1h"
  escalate_to    = data.freshdesk_agents.all.agents[0].id

  agent_ids = [
    for a in data.freshdesk_agents.all.agents : a.id
    if a.active
  ]
}

data "freshdesk_agents" "all" {}
