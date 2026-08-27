#!/bin/bash
# Create one record of each kind the field audit would otherwise skip, so the
# audit has something to compare. Everything is named tfaudit-*; cleanup.sh
# removes it afterwards.
set -u
H="${FRESHDESK_DOMAIN:?}.freshdesk.com"
K="${FRESHDESK_API_KEY:?}"
api() { curl -s -u "$K:X" -H 'Content-Type: application/json' "$@"; }
id_of() { python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("id",""))' 2>/dev/null; }

post() { # post <path> <json> <label>
  local out; out=$(api -X POST -d "$2" "https://$H/api/v2/$1")
  local id; id=$(echo "$out" | id_of)
  printf "  %-26s %s\n" "$3" "${id:-FAILED: $(echo "$out" | head -c 120)}"
  echo "$id"
  sleep 1.5
}

post groups '{"name":"tfaudit-group"}' group >/dev/null
post admin/groups '{"name":"tfaudit-admin-group","type":"support_agent_group"}' "admin group" >/dev/null
post admin/skills '{"name":"tfaudit-skill","conditions":[{"channel":"ticket","channel_conditions":[{"name":"condition_set_1","match_type":"all","properties":[{"resource_type":"ticket","field_name":"priority","operator":"in","value":[4]}]}]}]}' skill >/dev/null
post solutions/categories '{"name":"tfaudit-category"}' "solution category" >/dev/null
post discussions/categories '{"name":"tfaudit-forum-category"}' "forum category" >/dev/null
# A dispatcher rule is rejected without conditions.
post automations/1/rules '{"name":"tfaudit-rule","active":false,"conditions":[{"name":"condition_set_1","match_type":"all","properties":[{"resource_type":"ticket","field_name":"priority","operator":"in","value":[4]}]}],"actions":[{"field_name":"priority","value":4}]}' "automation rule" >/dev/null

# A time entry needs a ticket to hang off.
TICKET=$(post tickets '{"subject":"tfaudit ticket","description":"<p>audit fixture</p>","email":"tfaudit@example.test","status":2,"priority":1}' ticket)
if [ -n "$TICKET" ]; then
  post "tickets/$TICKET/time_entries" '{"note":"tfaudit","time_spent":"00:30"}' "time entry" >/dev/null
fi
