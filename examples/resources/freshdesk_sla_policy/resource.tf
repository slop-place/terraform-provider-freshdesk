resource "freshdesk_sla_policy" "premium" {
  name        = "Premium customers"
  description = "Tighter targets for customers on the premium tier"
  active      = true

  applicable_to_company_ids = [freshdesk_company.northwind.id]

  # Every duration is in seconds.
  sla_target = {
    priority_1 = { respond_within = 14400, resolve_within = 172800, next_respond_within = 14400 }
    priority_2 = { respond_within = 7200, resolve_within = 86400, next_respond_within = 7200 }
    priority_3 = { respond_within = 3600, resolve_within = 43200, next_respond_within = 3600 }
    priority_4 = { respond_within = 900, resolve_within = 14400, next_respond_within = 900 }
  }
}

resource "freshdesk_company" "northwind" {
  name = "Northwind Traders"
}
