package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// This file holds one entity descriptor per Freshdesk object. Each descriptor
// drives both the singular and the plural data source for that object, so the
// two can never present different attributes.

func groupEntity() entity[freshdesk.Group] {
	return entity[freshdesk.Group]{
		label: "group",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":               computedString("Name of the group."),
			"description":        computedString("Description of the group."),
			"escalate_to":        computedInt64("ID of the agent emailed about unassigned tickets."),
			"unassigned_for":     computedString("How long a ticket may stay unassigned."),
			"agent_ids":          computedInt64Set("IDs of the agents in the group."),
			"auto_ticket_assign": computedInt64("Automatic ticket assignment mode."),
			"business_hour_id":   computedInt64("ID of the business-hours calendar the group follows."),
		}, timestampDataAttributes("group")),
		types: mergeTypes(map[string]attr.Type{
			"name":               types.StringType,
			"description":        types.StringType,
			"escalate_to":        types.Int64Type,
			"unassigned_for":     types.StringType,
			"agent_ids":          types.SetType{ElemType: types.Int64Type},
			"auto_ticket_assign": types.Int64Type,
			"business_hour_id":   types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(g *freshdesk.Group) types.String { return idString(g.ID) },
		mapFn: func(g *freshdesk.Group) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":               types.StringValue(g.Name),
				"description":        optString(g.Description),
				"escalate_to":        optInt64(g.EscalateTo),
				"unassigned_for":     optString(g.UnassignedFor),
				"agent_ids":          int64Set(sortedInt64(g.AgentIDs)),
				"auto_ticket_assign": types.Int64Value(int64(g.AutoTicketAssign)),
				"business_hour_id":   optInt64(g.BusinessHourID),
			}, timestampValues(g.CreatedAt, g.UpdatedAt))
		},
	}
}

func agentEntity() entity[freshdesk.Agent] {
	return entity[freshdesk.Agent]{
		label: "agent",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"email":                  computedString("Email address of the agent."),
			"name":                   computedString("Display name of the agent."),
			"ticket_scope":           computedInt64("Ticket permission: 1 global, 2 group, 3 restricted."),
			"occasional":             computedBool("Whether the agent is occasional rather than full-time."),
			"signature":              computedString("Email signature of the agent, in HTML."),
			"language":               computedString("Language code of the agent."),
			"time_zone":              computedString("Time zone of the agent."),
			"focus_mode":             computedBool("Whether focus mode is enabled."),
			"type":                   computedString("Agent type in string form."),
			"available":              computedBool("Whether the agent is accepting omniroute assignments."),
			"active":                 computedBool("Whether the agent has activated their account."),
			"job_title":              computedString("Job title of the agent."),
			"mobile":                 computedString("Mobile number of the agent."),
			"phone":                  computedString("Phone number of the agent."),
			"last_login_at":          computedString("When the agent last signed in (RFC 3339)."),
			"role_ids":               computedInt64Set("IDs of the roles granted to the agent."),
			"group_ids":              computedInt64Set("IDs of the groups the agent belongs to."),
			"skill_ids":              computedInt64Set("IDs of the skills assigned to the agent."),
			"contribution_group_ids": computedInt64Set("IDs of groups the agent may only view."),
		}, timestampDataAttributes("agent")),
		types: mergeTypes(map[string]attr.Type{
			"email": types.StringType, "name": types.StringType,
			"ticket_scope": types.Int64Type, "occasional": types.BoolType,
			"signature": types.StringType, "language": types.StringType,
			"time_zone": types.StringType, "focus_mode": types.BoolType,
			"type": types.StringType, "available": types.BoolType,
			"active": types.BoolType, "job_title": types.StringType,
			"mobile": types.StringType, "phone": types.StringType,
			"last_login_at":          types.StringType,
			"role_ids":               types.SetType{ElemType: types.Int64Type},
			"group_ids":              types.SetType{ElemType: types.Int64Type},
			"skill_ids":              types.SetType{ElemType: types.Int64Type},
			"contribution_group_ids": types.SetType{ElemType: types.Int64Type},
		}, timestampDataTypes()),
		idFn: func(a *freshdesk.Agent) types.String { return idString(a.ID) },
		mapFn: func(a *freshdesk.Agent) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"email":                  optString(a.Contact.Email),
				"name":                   optString(a.Contact.Name),
				"ticket_scope":           types.Int64Value(int64(a.TicketScope)),
				"occasional":             types.BoolValue(a.Occasional),
				"signature":              optString(a.Signature),
				"language":               optString(a.Contact.Language),
				"time_zone":              optString(a.Contact.TimeZone),
				"focus_mode":             types.BoolValue(a.FocusMode),
				"type":                   optString(a.Type),
				"available":              types.BoolValue(a.Available),
				"active":                 types.BoolValue(a.Contact.Active),
				"job_title":              optString(a.Contact.JobTitle),
				"mobile":                 optString(a.Contact.Mobile),
				"phone":                  optString(a.Contact.Phone),
				"last_login_at":          timeString(a.Contact.LastLoginAt),
				"role_ids":               int64Set(sortedInt64(a.RoleIDs)),
				"group_ids":              int64Set(sortedInt64(a.GroupIDs)),
				"skill_ids":              int64Set(sortedInt64(a.SkillIDs)),
				"contribution_group_ids": int64Set(sortedInt64(a.ContributionGroupIDs)),
			}, timestampValues(a.CreatedAt, a.UpdatedAt))
		},
	}
}

func roleEntity() entity[freshdesk.Role] {
	return entity[freshdesk.Role]{
		label: "role",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Name of the role."),
			"description": computedString("What the role permits."),
			"default":     computedBool("Whether this is a built-in role."),
			"agent_type":  computedInt64("Agent kind the role applies to."),
		}, timestampDataAttributes("role")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"default": types.BoolType, "agent_type": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(r *freshdesk.Role) types.String { return idString(r.ID) },
		mapFn: func(r *freshdesk.Role) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":        types.StringValue(r.Name),
				"description": optString(r.Description),
				"default":     types.BoolValue(r.Default),
				"agent_type":  types.Int64Value(int64(r.AgentType)),
			}, timestampValues(r.CreatedAt, r.UpdatedAt))
		},
	}
}

func skillEntity() entity[freshdesk.Skill] {
	return entity[freshdesk.Skill]{
		label: "skill",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":       computedString("Name of the skill."),
			"match_type": computedString("Whether all or any conditions must match."),
			"agent_ids":  computedInt64Set("IDs of the agents who hold the skill."),
			"conditions": computedString("Matching rules, as JSON."),
			"rank":       computedInt64("Evaluation order of the skill."),
		}, timestampDataAttributes("skill")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "match_type": types.StringType,
			"agent_ids":  types.SetType{ElemType: types.Int64Type},
			"conditions": types.StringType, "rank": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(s *freshdesk.Skill) types.String { return idString(s.ID) },
		mapFn: func(s *freshdesk.Skill) map[string]attr.Value {
			ids := make([]int64, 0, len(s.Agents))
			for _, a := range s.Agents {
				ids = append(ids, a.ID)
			}

			return mergeValues(map[string]attr.Value{
				"name":       types.StringValue(s.Name),
				"match_type": optString(s.MatchType),
				"agent_ids":  int64Set(sortedInt64(ids)),
				"conditions": jsonString(s.Conditions),
				"rank":       types.Int64Value(int64(s.Rank)),
			}, timestampValues(s.CreatedAt, s.UpdatedAt))
		},
	}
}

func productEntity() entity[freshdesk.Product] {
	return entity[freshdesk.Product]{
		label: "product",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":          computedString("Name of the product."),
			"description":   computedString("Description of the product."),
			"primary_email": computedString("Support address for the product."),
			"default":       computedBool("Whether new tickets fall back to this product."),
		}, timestampDataAttributes("product")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"primary_email": types.StringType, "default": types.BoolType,
		}, timestampDataTypes()),
		idFn: func(p *freshdesk.Product) types.String { return idString(p.ID) },
		mapFn: func(p *freshdesk.Product) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":          types.StringValue(p.Name),
				"description":   optString(p.Description),
				"primary_email": optString(p.PrimaryEmail),
				"default":       types.BoolValue(p.Default),
			}, timestampValues(p.CreatedAt, p.UpdatedAt))
		},
	}
}

func businessHoursEntity() entity[freshdesk.BusinessHours] {
	return entity[freshdesk.BusinessHours]{
		label: "business hours calendar",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Name of the calendar."),
			"description": computedString("Description of the calendar."),
			"time_zone":   computedString("Time zone the calendar is expressed in."),
			"is_default":  computedBool("Whether this is the account's default calendar."),
			"business_hours": computedStringMap(
				"Working window per weekday, as a `start_time`/`end_time` JSON object."),
			"holidays": computedStringMap("Holiday dates keyed by name."),
		}, timestampDataAttributes("calendar")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"time_zone": types.StringType, "is_default": types.BoolType,
			"business_hours": types.MapType{ElemType: types.StringType},
			"holidays":       types.MapType{ElemType: types.StringType},
		}, timestampDataTypes()),
		idFn: func(b *freshdesk.BusinessHours) types.String { return idString(b.ID.Int64()) },
		mapFn: func(b *freshdesk.BusinessHours) map[string]attr.Value {
			windows := map[string]any{}
			for day, w := range b.BusinessHours {
				windows[day] = w
			}

			holidays := map[string]any{}
			for name, date := range b.Holidays {
				holidays[name] = date
			}

			return mergeValues(map[string]attr.Value{
				"name":           types.StringValue(b.Name),
				"description":    optString(b.Description),
				"time_zone":      optString(b.TimeZone),
				"is_default":     types.BoolValue(b.IsDefault),
				"business_hours": customFieldsToMap(windows),
				"holidays":       customFieldsToMap(holidays),
			}, timestampValues(b.CreatedAt, b.UpdatedAt))
		},
	}
}

func emailConfigEntity() entity[freshdesk.EmailConfig] {
	return entity[freshdesk.EmailConfig]{
		label: "email config",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":         computedString("Name of the email config."),
			"to_email":     computedString("Address customers write to."),
			"reply_email":  computedString("Address replies are sent from."),
			"product_id":   computedInt64("ID of the product the config belongs to."),
			"group_id":     computedInt64("ID of the group tickets are assigned to."),
			"primary_role": computedBool("Whether this is the account's default support address."),
			"active":       computedBool("Whether the config is in use."),
			"workspace_id": computedInt64("ID of the workspace the config belongs to."),
		}, timestampDataAttributes("config")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "to_email": types.StringType,
			"reply_email": types.StringType, "product_id": types.Int64Type,
			"group_id": types.Int64Type, "primary_role": types.BoolType,
			"active": types.BoolType, "workspace_id": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(e *freshdesk.EmailConfig) types.String { return idString(e.ID) },
		mapFn: func(e *freshdesk.EmailConfig) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":         types.StringValue(e.Name),
				"to_email":     optString(e.ToEmail),
				"reply_email":  optString(e.ReplyEmail),
				"product_id":   optInt64(e.ProductID),
				"group_id":     optInt64(e.GroupID),
				"primary_role": types.BoolValue(e.PrimaryRole),
				"active":       types.BoolValue(e.Active),
				"workspace_id": optInt64(e.WorkspaceID),
			}, timestampValues(e.CreatedAt, e.UpdatedAt))
		},
	}
}

func companyEntity() entity[freshdesk.Company] {
	return entity[freshdesk.Company]{
		label: "company",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":          computedString("Name of the company."),
			"description":   computedString("Description of the company."),
			"note":          computedString("Free-text note about the company."),
			"domains":       computedStringSet("Email domains belonging to the company."),
			"health_score":  computedString("Relationship strength."),
			"account_tier":  computedString("Account tier of the company."),
			"industry":      computedString("Industry the company serves."),
			"renewal_date":  computedString("Contract renewal date (RFC 3339)."),
			"custom_fields": computedStringMap("Custom field values keyed by field name."),
		}, timestampDataAttributes("company")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"note": types.StringType, "domains": types.SetType{ElemType: types.StringType},
			"health_score": types.StringType, "account_tier": types.StringType,
			"industry": types.StringType, "renewal_date": types.StringType,
			"custom_fields": types.MapType{ElemType: types.StringType},
		}, timestampDataTypes()),
		idFn: func(c *freshdesk.Company) types.String { return idString(c.ID) },
		mapFn: func(c *freshdesk.Company) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":          types.StringValue(c.Name),
				"description":   optString(c.Description),
				"note":          optString(c.Note),
				"domains":       stringSet(c.Domains),
				"health_score":  optString(c.HealthScore),
				"account_tier":  optString(c.AccountTier),
				"industry":      optString(c.Industry),
				"renewal_date":  timeString(c.RenewalDate),
				"custom_fields": customFieldsToMap(c.CustomFields),
			}, timestampValues(c.CreatedAt, c.UpdatedAt))
		},
	}
}

func contactEntity() entity[freshdesk.Contact] {
	return entity[freshdesk.Contact]{
		label: "contact",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":               computedString("Name of the contact."),
			"email":              computedString("Primary email address."),
			"phone":              computedString("Telephone number."),
			"mobile":             computedString("Mobile number."),
			"twitter_id":         computedString("Twitter handle."),
			"unique_external_id": computedString("Identifier in an external system."),
			"address":            computedString("Postal address."),
			"description":        computedString("Free-text note about the contact."),
			"job_title":          computedString("Job title."),
			"language":           computedString("Language code."),
			"time_zone":          computedString("Time zone."),
			"other_emails":       computedStringSet("Additional email addresses."),
			"tags":               computedStringSet("Tags applied to the contact."),
			"company_id":         computedInt64("ID of the company the contact belongs to."),
			"view_all_tickets":   computedBool("Whether the contact sees their company's tickets."),
			"active":             computedBool("Whether the contact activated their portal account."),
			"contact_type":       computedString("Contact type reported by the API."),
			"custom_fields":      computedStringMap("Custom field values keyed by field name."),
		}, timestampDataAttributes("contact")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "email": types.StringType, "phone": types.StringType,
			"mobile": types.StringType, "twitter_id": types.StringType,
			"unique_external_id": types.StringType, "address": types.StringType,
			"description": types.StringType, "job_title": types.StringType,
			"language": types.StringType, "time_zone": types.StringType,
			"other_emails": types.SetType{ElemType: types.StringType},
			"tags":         types.SetType{ElemType: types.StringType},
			"company_id":   types.Int64Type, "view_all_tickets": types.BoolType,
			"active": types.BoolType, "contact_type": types.StringType,
			"custom_fields": types.MapType{ElemType: types.StringType},
		}, timestampDataTypes()),
		idFn: func(c *freshdesk.Contact) types.String { return idString(c.ID) },
		mapFn: func(c *freshdesk.Contact) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":               types.StringValue(c.Name),
				"email":              optString(c.Email),
				"phone":              optString(c.Phone),
				"mobile":             optString(c.Mobile),
				"twitter_id":         optString(c.TwitterID),
				"unique_external_id": optString(c.UniqueExternalID),
				"address":            optString(c.Address),
				"description":        optString(c.Description),
				"job_title":          optString(c.JobTitle),
				"language":           optString(c.Language),
				"time_zone":          optString(c.TimeZone),
				"other_emails":       stringSet(c.OtherEmails),
				"tags":               stringSet(c.Tags),
				"company_id":         optInt64(c.CompanyID),
				"view_all_tickets":   types.BoolValue(c.ViewAllTickets),
				"active":             types.BoolValue(c.Active),
				"contact_type":       optString(c.ContactType),
				"custom_fields":      customFieldsToMap(c.CustomFields),
			}, timestampValues(c.CreatedAt, c.UpdatedAt))
		},
	}
}
