# A singleton: only one of these should exist per helpdesk.
resource "freshdesk_email_settings" "this" {
  personalized_email_replies      = true
  create_requester_using_reply_to = false
  email_subject_match             = true
  auto_response_detector_toggle   = true
}
