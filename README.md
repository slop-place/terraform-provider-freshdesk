# Terraform Provider for Freshdesk

Manage a [Freshdesk](https://freshdesk.com) helpdesk as code: agents and groups,
contacts and companies, ticket and contact fields, SLA policies, automation
rules, the knowledge base, community forums, mailboxes and tickets.

The provider covers the whole of the documented [Freshdesk API
v2](https://developers.freshdesk.com/api/) — **28 resources**, **45 data
sources**, and a standalone Go client with **245 methods** spanning every
endpoint in the reference.

## Usage

```hcl
terraform {
  required_providers {
    freshdesk = {
      source  = "slop-place/freshdesk"
      version = "~> 0.1"
    }
  }
}

provider "freshdesk" {
  domain  = "acme" # or "acme.freshdesk.com"
  api_key = var.freshdesk_api_key
}

resource "freshdesk_group" "billing" {
  name           = "Billing"
  description    = "Invoices, refunds and payment questions"
  unassigned_for = "1h"
}

resource "freshdesk_ticket_field" "issue_type" {
  label                  = "Issue Type"
  type                   = "custom_dropdown"
  displayed_to_customers = true

  choices = [
    { value = "Refund" },
    { value = "Faulty product" },
  ]
}
```

Credentials can come from the environment instead of the provider block:

| Argument  | Environment variable                    |
|-----------|-----------------------------------------|
| `domain`  | `FRESHDESK_DOMAIN` (or `FRESHDESK_URL`) |
| `api_key` | `FRESHDESK_API_KEY`                     |

Find your API key under **Profile picture → Profile Settings** in the agent
portal. It inherits that agent's permissions, so use an administrator's key to
manage configuration.

Full reference documentation lives in [`docs/`](docs/) and on the Terraform
Registry.

## Using the API client on its own

The Freshdesk client is a separate package with no dependency on Terraform, so
it can be imported directly:

```go
import "github.com/slop-place/terraform-provider-freshdesk/freshdesk"

client, err := freshdesk.New(freshdesk.Config{
    Domain: "acme",
    APIKey: os.Getenv("FRESHDESK_API_KEY"),
})
if err != nil {
    return err
}

groups, err := client.ListGroups(ctx, freshdesk.ListOptions{})
```

It handles pagination, rate-limit retries with `Retry-After`, multipart
attachment uploads, and the several places where Freshdesk's wire format
disagrees with its own documentation. `Client.Do` is the escape hatch for
anything the typed methods do not cover.

## Where Freshdesk's API differs from its documentation

These were found by exercising the provider against a live account, and are
handled by the provider rather than left to trip you up:

| Behaviour | Handling |
|---|---|
| `business_hours` and `sla_policies` quote their IDs; everything else sends numbers | `freshdesk.ID` accepts both |
| `org_company_id` is a number on create, a string on list | `freshdesk.FlexString` accepts both |
| `choices` is an array of objects on custom fields but one of four map shapes on built-in ones | `freshdesk.RawJSON`, with a `Choices()` accessor |
| Contact and company field writes 404 on the documented singular paths | The plural paths are used |
| Canned response folders create on the plural path, not the documented singular one | The plural path is used |
| Skills take PATCH, not PUT | `Client.Patch` |
| `admin/ticket_fields` rejects `page` and `per_page`; `ticket-forms` caps `per_page` at 50 | Per-endpoint page caps |
| CSAT surveys wrap their collection in a `data` envelope | Decoded explicitly |
| Skills use a channel-scoped condition shape on omniroute accounts | Conditions pass through as opaque JSON |
| Emails are folded to lower case | Compared case-insensitively |
| HTML bodies are rewritten by the server | Your configured value is kept; the stored rendering is exposed separately |
| SLA policies and canned responses reject DELETE | Destroy warns and drops from state |

## Development

```sh
make tools    # install golangci-lint, tfplugindocs, goimports
make check    # fmt, vet, lint, unit tests
make docs     # regenerate docs/ from the schemas and examples
```

Unit tests run against an in-process mock and need no credentials:

```sh
make test
```

Acceptance tests create and destroy real records, so point them at a sandbox:

```sh
export FRESHDESK_DOMAIN=your-sandbox
export FRESHDESK_API_KEY=...
make testacc
```

Everything the suite creates is named `tfacc-*`. If a run is interrupted,
`internal/apispec/cleanup.sh` removes the leftovers.

`make coverage-api` re-checks the client against the endpoint inventory scraped
from Freshdesk's own reference, so a gap in coverage fails the build.

## Licence

Mozilla Public License 2.0. See [LICENSE](LICENSE).
