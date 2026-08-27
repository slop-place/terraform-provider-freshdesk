# Roles are defined by Freshdesk and cannot be created through the API, so this
# is how you find the IDs an agent needs.
data "freshdesk_roles" "all" {}

output "agent_role_id" {
  value = one([for r in data.freshdesk_roles.all.roles : r.id if r.name == "Agent"])
}
