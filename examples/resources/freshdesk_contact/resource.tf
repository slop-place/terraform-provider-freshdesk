resource "freshdesk_contact" "avery" {
  name             = "Avery Stone"
  email            = "avery@northwind.example"
  job_title        = "Operations Lead"
  company_id       = freshdesk_company.northwind.id
  view_all_tickets = true
  tags             = ["priority-customer"]
}

resource "freshdesk_company" "northwind" {
  name = "Northwind Traders"
}
