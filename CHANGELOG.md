# Changelog

## 0.1.2 (2026-08-27)

### Fixed

- Every JSON attribute the practitioner writes is now passed to the API
  verbatim. Each was being decoded into a struct that enumerated a handful of
  fields and silently discarded the rest, so a `trigger_webhook` automation
  action reached Freshdesk carrying only its `field_name` — no `url`,
  `request_type`, `content_type`, `content_layout`, `content` or
  `custom_headers` — and the API answered `500` rather than naming what was
  missing. The same fault affected `freshdesk_email_mailbox.custom_mailbox`,
  whose accepted keys differ by provider and authentication method, and
  `freshdesk_admin_group.automatic_agent_assignment`, whose options depend on
  the plan and the assignment type. Scenario automation actions were losing
  keys on read for the same reason.
- An automation rule created without `conditions` no longer reports drift.
  Freshdesk answers such a rule with a single empty condition set, which turned
  an unset attribute into a value and failed Terraform's consistency check. An
  empty set now leaves the attribute unset; a real one still round-trips.

Reported from a live apply. Thanks to @nutgood for the diagnosis and the fix.

## 0.1.1 (2026-08-27)

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
