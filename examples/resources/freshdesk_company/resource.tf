resource "freshdesk_company" "northwind" {
  name         = "Northwind Traders"
  description  = "Wholesale distribution"
  domains      = ["northwind.example", "northwind-traders.example"]
  health_score = "Happy"
  account_tier = "Premium"
  renewal_date = "2027-03-31"

  custom_fields = {
    # Non-string values are given as JSON.
    seat_count = "250"
  }
}
