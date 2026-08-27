#!/bin/bash
# Remove records left behind by an interrupted acceptance run.
# Everything the suite creates is prefixed "tfacc-".
set -u
H="${FRESHDESK_DOMAIN:?set FRESHDESK_DOMAIN}.freshdesk.com"
K="${FRESHDESK_API_KEY:?set FRESHDESK_API_KEY}"

api() { curl -s -u "$K:X" -H 'Content-Type: application/json' "$@"; }

purge() { # purge <collection-path> <name-field>
  local path="$1" field="${2:-name}"
  api "https://$H/api/v2/$path?per_page=100" \
    | python3 -c "
import json,sys
try: items=json.load(sys.stdin)
except Exception: sys.exit()
if not isinstance(items,list): sys.exit()
for i in items:
    n=str(i.get('$field') or '')
    if n.startswith('tfacc-'):
        print(i['id'])
" | while read -r id; do
    [ -n "$id" ] && api -X DELETE "https://$H/api/v2/$path/$id" -o /dev/null -w "deleted $path/$id (%{http_code})\n"
  done
}

# A soft-deleted contact keeps its email address reserved, so contacts are
# removed permanently rather than merely trashed.
purge_contacts() {
  api "https://$H/api/v2/contacts?per_page=100&state=all" \
    | python3 -c "
import json,sys
try: items=json.load(sys.stdin)
except Exception: sys.exit()
if not isinstance(items,list): sys.exit()
for i in items:
    if str(i.get('name') or '').startswith('tfacc-'):
        print(i['id'])
" | while read -r id; do
    [ -n "$id" ] && api -X DELETE "https://$H/api/v2/contacts/$id/hard_delete?force=true" \
      -o /dev/null -w "hard-deleted contacts/$id (%{http_code})\n"
  done
}

purge groups
purge admin/groups
purge admin/skills
purge companies
purge_contacts
purge solutions/categories
purge discussions/categories
# Canned responses and their folders reject DELETE (405), so leftovers cannot
# be removed here; they must be deleted in the portal.
purge ticket-forms title
purge admin/ticket_fields label
purge contact_fields label
purge company_fields label
# SLA policies and canned responses cannot be deleted through the API, so
# leftovers are deactivated rather than removed.
api "https://$H/api/v2/sla_policies?per_page=100" \
  | python3 -c "
import json,sys
try: items=json.load(sys.stdin)
except Exception: sys.exit()
for i in items if isinstance(items,list) else []:
    if str(i.get('name') or '').startswith('tfacc-'):
        print(i['id'])
" | while read -r id; do
  [ -n "$id" ] && api -X PUT -d '{"active":false}' "https://$H/api/v2/sla_policies/$id" \
    -o /dev/null -w "deactivated sla_policies/$id (%{http_code})\n"
done
