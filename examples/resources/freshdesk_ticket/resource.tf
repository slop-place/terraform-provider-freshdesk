resource "freshdesk_ticket" "onboarding" {
  subject      = "Welcome aboard"
  description  = "<p>Reach out here with any setup questions.</p>"
  requester_id = freshdesk_contact.avery.id
  priority     = 1
  status       = 2
  tags         = ["onboarding"]
}

resource "freshdesk_contact" "avery" {
  name  = "Avery Stone"
  email = "avery@northwind.example"
}
