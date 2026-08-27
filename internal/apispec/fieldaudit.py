#!/usr/bin/env python3
"""Find every JSON key the Freshdesk API returns that the client does not map.

The provider's worst bug so far was a struct that enumerated a few fields of an
open-ended object and silently discarded the rest. This walks the live API,
collects every key each endpoint actually returns, and compares that against the
`json:"..."` tags declared in the Go client. Anything unmapped is a field the
client is throwing away.

    FRESHDESK_DOMAIN=acme FRESHDESK_API_KEY=... python3 fieldaudit.py
"""
import base64
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DOMAIN = os.environ.get("FRESHDESK_DOMAIN") or os.environ.get("FRESHDESK_URL", "")
KEY = os.environ.get("FRESHDESK_API_KEY", "")

if not DOMAIN or not KEY:
    sys.exit("set FRESHDESK_DOMAIN and FRESHDESK_API_KEY")

HOST = DOMAIN if "." in DOMAIN else DOMAIN + ".freshdesk.com"
AUTH = base64.b64encode(f"{KEY}:X".encode()).decode()

# endpoint -> the Go type its elements decode into
ENDPOINTS = {
    "account": "Account",
    "settings/helpdesk": "HelpdeskSettings",
    "roles": "Role",
    "groups": "Group",
    "admin/groups": "AdminGroup",
    "agents": "Agent",
    "admin/skills": "Skill",
    "products": "Product",
    "business_hours": "BusinessHours",
    "email_configs": "EmailConfig",
    "email/mailboxes": "EmailMailbox",
    "email/settings": "EmailSettings",
    "sla_policies": "SLAPolicy",
    "scenario_automations": "ScenarioAutomation",
    "contacts": "Contact",
    "companies": "Company",
    "tickets": "Ticket",
    "admin/ticket_fields": "TicketField",
    "contact_fields": "ContactField",
    "company_fields": "CompanyField",
    "ticket-forms": "TicketForm",
    "solutions/categories": "SolutionCategory",
    "discussions/categories": "ForumCategory",
    "time_entries": "TimeEntry",
    "surveys/satisfaction_ratings": "SatisfactionRating",
    "automations/1/rules": "AutomationRule",
    "automations/3/rules": "AutomationRule",
    "automations/4/rules": "AutomationRule",
}

# Keys deliberately not surfaced, with the reason.
IGNORED = {
    # Freshdesk echoes the request's own paging envelope on some endpoints.
    "paging",
}


def get(path, attempts=6):
    """Fetch a collection, waiting out the account's per-minute rate limit."""
    req = urllib.request.Request(
        f"https://{HOST}/api/v2/{path}",
        headers={"Authorization": "Basic " + AUTH, "Accept": "application/json"},
    )
    for attempt in range(attempts):
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                return json.loads(resp.read() or "null"), None
        except urllib.error.HTTPError as e:
            if e.code == 429 and attempt < attempts - 1:
                wait = int(e.headers.get("Retry-After") or 0) or 20
                print(f"  ...   rate limited, waiting {wait}s", flush=True)
                time.sleep(wait)
                continue
            return None, f"HTTP {e.code}"
        except Exception as e:  # noqa: BLE001 - the audit reports, it does not raise
            return None, str(e)

    return None, "gave up after rate limiting"


def top_level_keys(payload):
    """Collect the keys of the objects an endpoint returns."""
    if isinstance(payload, dict):
        # Unwrap the envelopes a few endpoints use.
        for env in ("data", "schemas", "records", "sections", "companies"):
            if env in payload and isinstance(payload[env], list):
                payload = payload[env]
                break
    if isinstance(payload, list):
        keys = set()
        for item in payload[:25]:
            if isinstance(item, dict):
                keys |= set(item)
        return keys
    if isinstance(payload, dict):
        return set(payload)
    return set()


def struct_tags(type_name):
    """Read the json tags declared on a Go struct, following embedded types."""
    tags = set()
    for path in (ROOT / "freshdesk").glob("*.go"):
        src = path.read_text()
        m = re.search(r"^type %s struct \{(.*?)^\}" % re.escape(type_name), src,
                      re.S | re.M)
        if not m:
            continue
        for tag in re.finditer(r'json:"([^",]+)', m.group(1)):
            tags.add(tag.group(1))
    return tags


def main():
    findings = 0
    checked = 0

    for path, type_name in ENDPOINTS.items():
        time.sleep(1.4)  # stay inside the account's 50 requests a minute
        payload, err = get(path)
        if err:
            print(f"  skip  {path:<32} ({err})")
            continue

        live = top_level_keys(payload)
        if not live:
            print(f"  empty {path:<32} (nothing to compare)")
            continue

        declared = struct_tags(type_name)
        if not declared:
            print(f"  !!    {path:<32} no Go type named {type_name}")
            findings += 1
            continue

        missing = sorted(live - declared - IGNORED)
        checked += 1

        if missing:
            findings += 1
            print(f"  DROP  {path:<32} {type_name} does not map: {', '.join(missing)}")
        else:
            print(f"  ok    {path:<32} {type_name} ({len(live)} keys)")

    print(f"\nendpoints compared: {checked}   with unmapped keys: {findings}")
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
