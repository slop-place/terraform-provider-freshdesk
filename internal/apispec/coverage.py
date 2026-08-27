#!/usr/bin/env python3
"""Audit client coverage of the documented Freshdesk API surface.

Reads the operation index scraped from developers.freshdesk.com and checks that
the client package references a matching request path for each one.
"""
import re, sys, os, glob

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SPEC = os.path.join(ROOT, "internal", "apispec", "index.txt")
ENDPOINTS = os.path.join(ROOT, "internal", "apispec", "endpoints.txt")

# Endpoints intentionally not modelled, with the reason.
SKIP = {
    "#intro", "#ratelimit", "#authentication", "#error", "#pagination",
}

# Literal paths that appear in the docs' prose rather than naming an endpoint.
SKIP_PATHS = {"resource_name"}

def norm(u):
    u = u.strip().split("?")[0].rstrip("/")
    u = re.sub(r"\.json$", "", u)
    u = re.sub(r"^/?api/v2/", "", u)
    u = re.sub(r"\[[^\]]*\]|\{[^}]*\}|schema[-_]id|\bid\b", "*", u)
    u = re.sub(r"/\*+", "/*", u)
    return u.strip("/")

def code_paths():
    """Collect the literal API paths the client builds."""
    out = set()
    for f in glob.glob(os.path.join(ROOT, "freshdesk", "*.go")):
        src = open(f).read()
        for m in re.finditer(r'"([a-z][a-zA-Z0-9_/.-]*)"', src):
            p = m.group(1)
            if "/" in p or p in ("account", "agents", "roles", "groups", "products",
                                 "tickets", "contacts", "companies", "jobs", "skills",
                                 "surveys", "attachments", "conversations", "time_entries",
                                 "sla_policies", "business_hours", "email_configs",
                                 "scenario_automations", "ticket_fields", "contact_fields",
                                 "company_fields", "contact_field", "company_field",
                                 "canned_responses", "canned_response_folder",
                                 "canned_response_folders", "ticket-forms",
                                 "contact-activities"):
                out.add(p.strip("/"))
    return out

def main():
    paths = code_paths()
    # Expand: a path is covered if some code path is a prefix-compatible match.
    def covered(target):
        t = norm(target)
        if not t:
            return True
        for p in paths:
            pn = norm(p)
            if pn == t:
                return True
            # code builds "tickets" + "/{id}/notes" style suffixes
            if t.startswith(pn + "/") or t.replace("/*", "") == pn:
                return True
            if pn.startswith(t.replace("/*", "")):
                return True
        return False

    missing, total = [], 0
    for line in open(SPEC):
        parts = line.rstrip("\n").split("\t")
        if len(parts) != 4:
            continue
        section, title, url, anchor = parts
        if anchor in SKIP or not url or not url.startswith(("/", "api")):
            continue
        total += 1
        if not covered(url):
            missing.append((section, title, url))

    # Also audit every distinct path the docs mention, which catches the
    # sections (Solutions, Discussions) whose nav entries carry no URL.
    for line in open(ENDPOINTS):
        url = line.strip()
        if not url or "[" in url and url.endswith("["):
            continue
        if norm(url) in SKIP_PATHS:
            continue
        total += 1
        if not covered(url):
            missing.append(("(scraped)", "", url))

    print(f"documented operations with a URL: {total}")
    print(f"covered: {total - len(missing)}")
    print(f"missing: {len(missing)}")
    for s, t, u in missing:
        print(f"  MISSING  {s:<22} {t:<42} {u}")
    return 1 if missing else 0

if __name__ == "__main__":
    sys.exit(main())
