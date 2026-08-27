data "freshdesk_account" "current" {}

output "helpdesk" {
  value = "${data.freshdesk_account.current.account_name} on ${data.freshdesk_account.current.tier_type}"
}
