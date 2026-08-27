package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

func ticketEntity() entity[freshdesk.Ticket] {
	return entity[freshdesk.Ticket]{
		label: "ticket",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"subject":          computedString("Subject line of the ticket."),
			"description":      computedString("Opening message, in HTML."),
			"description_text": computedString("Plain-text rendering of the opening message."),
			"type":             computedString("Ticket type."),
			"status":           computedInt64("Status: 2 open, 3 pending, 4 resolved, 5 closed."),
			"priority":         computedInt64("Priority: 1 low, 2 medium, 3 high, 4 urgent."),
			"source":           computedInt64("Channel the ticket arrived on."),
			"requester_id":     computedInt64("ID of the contact who raised the ticket."),
			"responder_id":     computedInt64("ID of the agent the ticket is assigned to."),
			"group_id":         computedInt64("ID of the group the ticket is assigned to."),
			"company_id":       computedInt64("ID of the company the ticket belongs to."),
			"product_id":       computedInt64("ID of the product the ticket concerns."),
			"email_config_id":  computedInt64("ID of the email config the ticket came through."),
			"tags":             computedStringSet("Tags applied to the ticket."),
			"cc_emails":        computedStringSet("Addresses copied on the correspondence."),
			"custom_fields":    computedStringMap("Custom field values keyed by field name."),
			"due_by":           computedString("When the ticket must be resolved (RFC 3339)."),
			"fr_due_by":        computedString("When the first response is due (RFC 3339)."),
			"is_escalated":     computedBool("Whether the resolution SLA has been breached."),
			"spam":             computedBool("Whether the ticket is marked as spam."),
			"parent_id":        computedInt64("ID of the parent ticket, for a child ticket."),
		}, timestampDataAttributes("ticket")),
		types: mergeTypes(map[string]attr.Type{
			"subject": types.StringType, "description": types.StringType,
			"description_text": types.StringType, "type": types.StringType,
			"status": types.Int64Type, "priority": types.Int64Type, "source": types.Int64Type,
			"requester_id": types.Int64Type, "responder_id": types.Int64Type,
			"group_id": types.Int64Type, "company_id": types.Int64Type,
			"product_id": types.Int64Type, "email_config_id": types.Int64Type,
			"tags":          types.SetType{ElemType: types.StringType},
			"cc_emails":     types.SetType{ElemType: types.StringType},
			"custom_fields": types.MapType{ElemType: types.StringType},
			"due_by":        types.StringType, "fr_due_by": types.StringType,
			"is_escalated": types.BoolType, "spam": types.BoolType,
			"parent_id": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(t *freshdesk.Ticket) types.String { return idString(t.ID) },
		mapFn: func(t *freshdesk.Ticket) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"subject":          optString(t.Subject),
				"description":      optString(t.Description),
				"description_text": optString(t.DescriptionText),
				"type":             optString(t.Type),
				"status":           types.Int64Value(int64(t.Status)),
				"priority":         types.Int64Value(int64(t.Priority)),
				"source":           types.Int64Value(int64(t.Source)),
				"requester_id":     optInt64(t.RequesterID),
				"responder_id":     optInt64(t.ResponderID),
				"group_id":         optInt64(t.GroupID),
				"company_id":       optInt64(t.CompanyID),
				"product_id":       optInt64(t.ProductID),
				"email_config_id":  optInt64(t.EmailConfigID),
				"tags":             stringSet(t.Tags),
				"cc_emails":        stringSet(t.CCEmails),
				"custom_fields":    customFieldsToMap(t.CustomFields),
				"due_by":           timeString(t.DueBy),
				"fr_due_by":        timeString(t.FrDueBy),
				"is_escalated":     types.BoolValue(t.IsEscalated),
				"spam":             types.BoolValue(t.Spam),
				"parent_id":        optInt64(t.ParentID),
			}, timestampValues(t.CreatedAt, t.UpdatedAt))
		},
	}
}

func ticketFieldEntity() entity[freshdesk.TicketField] {
	return entity[freshdesk.TicketField]{
		label: "ticket field",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":                   computedString("Internal name of the field."),
			"label":                  computedString("Label shown to agents."),
			"label_for_customers":    computedString("Label shown in the customer portal."),
			"description":            computedString("Help text shown beneath the field."),
			"type":                   computedString("Field type."),
			"position":               computedInt64("Position of the field on the form."),
			"default":                computedBool("Whether this is a built-in field."),
			"customers_can_edit":     computedBool("Whether customers may edit the field."),
			"displayed_to_customers": computedBool("Whether the field appears in the portal."),
			"required_for_customers": computedBool("Whether customers must fill the field in."),
			"required_for_agents":    computedBool("Whether agents must fill the field in."),
			"required_for_closure":   computedBool("Whether the field is needed to close a ticket."),
			"choices":                computedString("Dropdown options, as JSON."),
		}, timestampDataAttributes("field")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "label": types.StringType,
			"label_for_customers": types.StringType, "description": types.StringType,
			"type": types.StringType, "position": types.Int64Type,
			"default": types.BoolType, "customers_can_edit": types.BoolType,
			"displayed_to_customers": types.BoolType, "required_for_customers": types.BoolType,
			"required_for_agents": types.BoolType, "required_for_closure": types.BoolType,
			"choices": types.StringType,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.TicketField) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.TicketField) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":                   types.StringValue(f.Name),
				"label":                  types.StringValue(f.Label),
				"label_for_customers":    optString(f.LabelForCustomers),
				"description":            optString(f.Description),
				"type":                   types.StringValue(f.Type),
				"position":               types.Int64Value(int64(f.Position)),
				"default":                types.BoolValue(f.Default),
				"customers_can_edit":     types.BoolValue(f.CustomersCanEdit),
				"displayed_to_customers": types.BoolValue(f.DisplayedToCustomers),
				"required_for_customers": types.BoolValue(f.RequiredForCustomers),
				"required_for_agents":    types.BoolValue(f.RequiredForAgents),
				"required_for_closure":   types.BoolValue(f.RequiredForClosure),
				"choices":                rawJSONString(f.Choices),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func contactFieldEntity() entity[freshdesk.ContactField] {
	return entity[freshdesk.ContactField]{
		label: "contact field",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":                    computedString("Internal name of the field."),
			"label":                   computedString("Label shown to agents."),
			"label_for_customers":     computedString("Label shown in the customer portal."),
			"type":                    computedString("Field type."),
			"position":                computedInt64("Position of the field on the form."),
			"default":                 computedBool("Whether this is a built-in field."),
			"editable_in_signup":      computedBool("Whether customers may set it at signup."),
			"required_for_agents":     computedBool("Whether agents must fill the field in."),
			"agents_can_edit":         computedBool("Whether agents may edit the field."),
			"displayed_for_agents":    computedBool("Whether agents see the field."),
			"quick_add_for_agent":     computedBool("Whether it appears in the quick-add form."),
			"unique":                  computedBool("Whether values must be unique."),
			"customers_can_edit":      computedBool("Whether customers may edit the field."),
			"required_for_customers":  computedBool("Whether customers must fill the field in."),
			"displayed_for_customers": computedBool("Whether customers see the field."),
			"choices":                 computedString("Dropdown options, as JSON."),
		}, timestampDataAttributes("field")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "label": types.StringType,
			"label_for_customers": types.StringType, "type": types.StringType,
			"position": types.Int64Type, "default": types.BoolType,
			"editable_in_signup": types.BoolType, "required_for_agents": types.BoolType,
			"agents_can_edit": types.BoolType, "displayed_for_agents": types.BoolType,
			"quick_add_for_agent": types.BoolType, "unique": types.BoolType,
			"customers_can_edit": types.BoolType, "required_for_customers": types.BoolType,
			"displayed_for_customers": types.BoolType, "choices": types.StringType,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.ContactField) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.ContactField) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":                    types.StringValue(f.Name),
				"label":                   types.StringValue(f.Label),
				"label_for_customers":     optString(f.LabelForCustomers),
				"type":                    types.StringValue(f.Type),
				"position":                types.Int64Value(int64(f.Position)),
				"default":                 types.BoolValue(f.Default),
				"editable_in_signup":      types.BoolValue(f.EditableInSignup),
				"required_for_agents":     types.BoolValue(f.RequiredForAgents),
				"agents_can_edit":         types.BoolValue(f.AgentsCanEdit),
				"displayed_for_agents":    types.BoolValue(f.DisplayedForAgents),
				"quick_add_for_agent":     types.BoolValue(f.QuickAddForAgent),
				"unique":                  types.BoolValue(f.Unique),
				"customers_can_edit":      types.BoolValue(f.CustomersCanEdit),
				"required_for_customers":  types.BoolValue(f.RequiredForCustomers),
				"displayed_for_customers": types.BoolValue(f.DisplayedForCustomers),
				"choices":                 rawJSONString(f.Choices),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func companyFieldEntity() entity[freshdesk.CompanyField] {
	return entity[freshdesk.CompanyField]{
		label: "company field",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":                 computedString("Internal name of the field."),
			"label":                computedString("Label shown to agents."),
			"type":                 computedString("Field type."),
			"position":             computedInt64("Position of the field on the form."),
			"default":              computedBool("Whether this is a built-in field."),
			"required_for_agents":  computedBool("Whether agents must fill the field in."),
			"agents_can_edit":      computedBool("Whether agents may edit the field."),
			"displayed_for_agents": computedBool("Whether agents see the field."),
			"quick_add_for_agent":  computedBool("Whether it appears in the quick-add form."),
			"unique":               computedBool("Whether values must be unique."),
			"choices":              computedString("Dropdown options, as JSON."),
		}, timestampDataAttributes("field")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "label": types.StringType, "type": types.StringType,
			"position": types.Int64Type, "default": types.BoolType,
			"required_for_agents": types.BoolType, "agents_can_edit": types.BoolType,
			"displayed_for_agents": types.BoolType, "quick_add_for_agent": types.BoolType,
			"unique": types.BoolType, "choices": types.StringType,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.CompanyField) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.CompanyField) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":                 types.StringValue(f.Name),
				"label":                types.StringValue(f.Label),
				"type":                 types.StringValue(f.Type),
				"position":             types.Int64Value(int64(f.Position)),
				"default":              types.BoolValue(f.Default),
				"required_for_agents":  types.BoolValue(f.RequiredForAgents),
				"agents_can_edit":      types.BoolValue(f.AgentsCanEdit),
				"displayed_for_agents": types.BoolValue(f.DisplayedForAgents),
				"quick_add_for_agent":  types.BoolValue(f.QuickAddForAgent),
				"unique":               types.BoolValue(f.Unique),
				"choices":              rawJSONString(f.Choices),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func ticketFormEntity() entity[freshdesk.TicketForm] {
	return entity[freshdesk.TicketForm]{
		label: "ticket form",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Internal name of the form."),
			"title":       computedString("Heading shown above the form."),
			"description": computedString("Text shown beneath the title."),
			"default":     computedBool("Whether this is the fallback form."),
			"fields":      computedString("The form's fields, as JSON."),
		}, timestampDataAttributes("form")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "title": types.StringType,
			"description": types.StringType, "default": types.BoolType,
			"fields": types.StringType,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.TicketForm) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.TicketForm) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":        types.StringValue(f.Name),
				"title":       optString(f.Title),
				"description": optString(f.Description),
				"default":     types.BoolValue(f.Default),
				"fields":      jsonString(f.Fields),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func slaPolicyEntity() entity[freshdesk.SLAPolicy] {
	return entity[freshdesk.SLAPolicy]{
		label: "SLA policy",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":          computedString("Name of the policy."),
			"description":   computedString("Description of the policy."),
			"active":        computedBool("Whether the policy is in force."),
			"is_default":    computedBool("Whether this is the fallback policy."),
			"position":      computedInt64("Evaluation order of the policy."),
			"sla_target":    computedString("Targets per priority, as JSON."),
			"applicable_to": computedString("What the policy is scoped to, as JSON."),
			"escalation":    computedString("Escalation recipients, as JSON."),
		}, timestampDataAttributes("policy")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"active": types.BoolType, "is_default": types.BoolType,
			"position": types.Int64Type, "sla_target": types.StringType,
			"applicable_to": types.StringType, "escalation": types.StringType,
		}, timestampDataTypes()),
		idFn: func(p *freshdesk.SLAPolicy) types.String { return idString(p.ID.Int64()) },
		mapFn: func(p *freshdesk.SLAPolicy) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":          types.StringValue(p.Name),
				"description":   optString(p.Description),
				"active":        types.BoolValue(p.Active),
				"is_default":    types.BoolValue(p.IsDefault),
				"position":      types.Int64Value(int64(p.Position)),
				"sla_target":    jsonString(p.SLATarget),
				"applicable_to": jsonString(p.ApplicableTo),
				"escalation":    jsonString(p.Escalation),
			}, timestampValues(p.CreatedAt, p.UpdatedAt))
		},
	}
}

func automationRuleEntity() entity[freshdesk.AutomationRule] {
	return entity[freshdesk.AutomationRule]{
		label: "automation rule",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":                   computedString("Name of the rule."),
			"description":            computedString("Description of the rule."),
			"position":               computedInt64("Evaluation order within the automation."),
			"active":                 computedBool("Whether the rule is running."),
			"outdated":               computedBool("Whether the rule references a missing field."),
			"operator":               computedString("How the condition sets combine."),
			"performer":              computedString("Who must act for the rule to fire, as JSON."),
			"events":                 computedString("Field changes the rule watches, as JSON."),
			"conditions":             computedString("Condition sets, as JSON."),
			"actions":                computedString("Actions the rule applies, as JSON."),
			"affected_tickets_count": computedInt64("How many tickets the rule has acted on."),
		}, timestampDataAttributes("rule")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"position": types.Int64Type, "active": types.BoolType,
			"outdated": types.BoolType, "operator": types.StringType,
			"performer": types.StringType, "events": types.StringType,
			"conditions": types.StringType, "actions": types.StringType,
			"affected_tickets_count": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(r *freshdesk.AutomationRule) types.String { return idString(r.ID) },
		mapFn: func(r *freshdesk.AutomationRule) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":                   types.StringValue(r.Name),
				"description":            optString(r.Description),
				"position":               types.Int64Value(int64(r.Position)),
				"active":                 types.BoolValue(r.Active),
				"outdated":               types.BoolValue(r.Outdated),
				"operator":               optString(r.Operator),
				"performer":              jsonString(r.Performer),
				"events":                 jsonString(r.Events),
				"conditions":             jsonString(r.Conditions),
				"actions":                jsonString(r.Actions),
				"affected_tickets_count": types.Int64Value(int64(r.AffectedTicketsCount)),
			}, timestampValues(r.CreatedAt, r.UpdatedAt))
		},
	}
}

func scenarioAutomationEntity() entity[freshdesk.ScenarioAutomation] {
	return entity[freshdesk.ScenarioAutomation]{
		label: "scenario automation",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Name of the scenario."),
			"description": computedString("Description of the scenario."),
			"private":     computedBool("Whether the scenario is private to its owner."),
			"actions":     computedString("Actions the scenario applies, as JSON."),
		}, timestampDataAttributes("scenario")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"private": types.BoolType, "actions": types.StringType,
		}, timestampDataTypes()),
		idFn: func(s *freshdesk.ScenarioAutomation) types.String { return idString(s.ID) },
		mapFn: func(s *freshdesk.ScenarioAutomation) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":        types.StringValue(s.Name),
				"description": optString(s.Description),
				"private":     types.BoolValue(s.Private),
				"actions":     jsonString(s.Actions),
			}, timestampValues(s.CreatedAt, s.UpdatedAt))
		},
	}
}
