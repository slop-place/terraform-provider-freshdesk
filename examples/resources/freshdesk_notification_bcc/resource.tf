# A singleton: only one of these should exist per helpdesk.
resource "freshdesk_notification_bcc" "this" {
  emails = ["archive@example.com"]
}
