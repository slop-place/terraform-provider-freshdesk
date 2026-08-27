# Changelog

## 0.1.0 (2026-08-27)

First release.

### Features

- **28 resources** and **45 data sources** covering the whole of the documented
  Freshdesk API v2: agents, groups and admin groups, skills, contacts and
  companies, tickets and time entries, ticket/contact/company fields and their
  dynamic sections, ticket forms, SLA policies, automation rules, canned
  responses, the knowledge base, community forums, mailboxes, custom object
  records, and the account-wide email settings.
- A standalone Go client (`github.com/slop-place/terraform-provider-freshdesk/freshdesk`)
  with 245 methods, usable without Terraform. It handles pagination, rate-limit
  retries honouring `Retry-After`, multipart attachment uploads, and an escape
  hatch (`Client.Do`) for anything not yet modelled.
- Semantic comparison for the attributes Freshdesk normalises: email addresses
  compare case-insensitively, and JSON bodies compare by value so a re-ordered
  response is not a diff.
- Import support on every resource, including composite IDs for ticket field
  sections (`<field_id>:<section_id>`), automation rules
  (`<rule_type>:<rule_id>`) and custom object records (`<schema_id>:<record_id>`).

### Known API limitations

These are Freshdesk's, not the provider's, and are surfaced rather than hidden:

- SLA policies, canned responses and canned response folders cannot be deleted
  through the API. Destroying one removes it from state and warns; the record
  itself must be deleted in the portal.
- Ticket descriptions and canned response bodies are rewritten by the server, so
  the provider keeps the configured markup and exposes the stored rendering
  separately. An edit made in the portal will not show as drift.
- A forum topic's opening message is never returned by the API and so cannot be
  refreshed.
- Custom objects require a plan that enables the feature; other accounts answer
  `403`.
