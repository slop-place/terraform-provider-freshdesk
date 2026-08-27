package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

func cannedResponseFolderEntity() entity[freshdesk.CannedResponseFolder] {
	return entity[freshdesk.CannedResponseFolder]{
		label: "canned response folder",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":            computedString("Name of the folder."),
			"personal":        computedBool("Whether this is an agent's private folder."),
			"responses_count": computedInt64("Number of canned responses in the folder."),
		}, timestampDataAttributes("folder")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "personal": types.BoolType,
			"responses_count": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.CannedResponseFolder) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.CannedResponseFolder) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":            types.StringValue(f.Name),
				"personal":        types.BoolValue(f.Personal),
				"responses_count": types.Int64Value(int64(f.ResponsesCount)),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func solutionCategoryEntity() entity[freshdesk.SolutionCategory] {
	return entity[freshdesk.SolutionCategory]{
		label: "solution category",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":               computedString("Name of the category."),
			"description":        computedString("Description of the category."),
			"visible_in_portals": computedInt64Set("IDs of the portals the category appears in."),
		}, timestampDataAttributes("category")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"visible_in_portals": types.SetType{ElemType: types.Int64Type},
		}, timestampDataTypes()),
		idFn: func(c *freshdesk.SolutionCategory) types.String { return idString(c.ID) },
		mapFn: func(c *freshdesk.SolutionCategory) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":               types.StringValue(c.Name),
				"description":        optString(c.Description),
				"visible_in_portals": int64Set(sortedInt64(c.VisibleInPortals)),
			}, timestampValues(c.CreatedAt, c.UpdatedAt))
		},
	}
}

func solutionFolderEntity() entity[freshdesk.SolutionFolder] {
	return entity[freshdesk.SolutionFolder]{
		label: "solution folder",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":             computedString("Name of the folder."),
			"description":      computedString("Description of the folder."),
			"visibility":       computedInt64("Who may read the folder."),
			"company_ids":      computedInt64Set("IDs of companies that may read the folder."),
			"category_id":      computedInt64("ID of the parent category."),
			"parent_folder_id": computedInt64("ID of the parent folder, for a subfolder."),
			"articles_count":   computedInt64("Number of articles in the folder."),
		}, timestampDataAttributes("folder")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"visibility":  types.Int64Type,
			"company_ids": types.SetType{ElemType: types.Int64Type},
			"category_id": types.Int64Type, "parent_folder_id": types.Int64Type,
			"articles_count": types.Int64Type,
		}, timestampDataTypes()),
		idFn: func(f *freshdesk.SolutionFolder) types.String { return idString(f.ID) },
		mapFn: func(f *freshdesk.SolutionFolder) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":             types.StringValue(f.Name),
				"description":      optString(f.Description),
				"visibility":       types.Int64Value(int64(f.Visibility)),
				"company_ids":      int64Set(sortedInt64(f.CompanyIDs)),
				"category_id":      optInt64(f.CategoryID),
				"parent_folder_id": optInt64(f.ParentFolderID),
				"articles_count":   types.Int64Value(int64(f.ArticlesCount)),
			}, timestampValues(f.CreatedAt, f.UpdatedAt))
		},
	}
}

func solutionArticleEntity() entity[freshdesk.SolutionArticle] {
	return entity[freshdesk.SolutionArticle]{
		label: "solution article",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"title":            computedString("Title of the article."),
			"description":      computedString("Body of the article, in HTML."),
			"description_text": computedString("Plain-text rendering of the body."),
			"status":           computedInt64("Publication state: 1 draft, 2 published."),
			"type":             computedInt64("Article type: 1 permanent, 2 workaround."),
			"agent_id":         computedInt64("ID of the agent credited as the author."),
			"category_id":      computedInt64("ID of the article's category."),
			"folder_id":        computedInt64("ID of the article's folder."),
			"tags":             computedStringSet("Tags applied to the article."),
			"hits":             computedInt64("How many times the article has been viewed."),
			"thumbs_up":        computedInt64("How many readers found the article helpful."),
			"thumbs_down":      computedInt64("How many readers found the article unhelpful."),
			"seo_data":         computedString("Search-engine metadata, as JSON."),
		}, timestampDataAttributes("article")),
		types: mergeTypes(map[string]attr.Type{
			"title": types.StringType, "description": types.StringType,
			"description_text": types.StringType, "status": types.Int64Type,
			"type": types.Int64Type, "agent_id": types.Int64Type,
			"category_id": types.Int64Type, "folder_id": types.Int64Type,
			"tags": types.SetType{ElemType: types.StringType},
			"hits": types.Int64Type, "thumbs_up": types.Int64Type,
			"thumbs_down": types.Int64Type, "seo_data": types.StringType,
		}, timestampDataTypes()),
		idFn: func(a *freshdesk.SolutionArticle) types.String { return idString(a.ID) },
		mapFn: func(a *freshdesk.SolutionArticle) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"title":            types.StringValue(a.Title),
				"description":      optString(a.Description),
				"description_text": optString(a.DescriptionText),
				"status":           types.Int64Value(int64(a.Status)),
				"type":             types.Int64Value(int64(a.Type)),
				"agent_id":         optInt64(a.AgentID),
				"category_id":      optInt64(a.CategoryID),
				"folder_id":        optInt64(a.FolderID),
				"tags":             stringSet(a.Tags),
				"hits":             types.Int64Value(int64(a.Hits)),
				"thumbs_up":        types.Int64Value(int64(a.ThumbsUp)),
				"thumbs_down":      types.Int64Value(int64(a.ThumbsDown)),
				"seo_data":         jsonString(a.SEOData),
			}, timestampValues(a.CreatedAt, a.UpdatedAt))
		},
	}
}

func forumCategoryEntity() entity[freshdesk.ForumCategory] {
	return entity[freshdesk.ForumCategory]{
		label: "forum category",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Name of the category."),
			"description": computedString("Description of the category."),
		}, timestampDataAttributes("category")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
		}, timestampDataTypes()),
		idFn: func(c *freshdesk.ForumCategory) types.String { return idString(c.ID) },
		mapFn: func(c *freshdesk.ForumCategory) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":        types.StringValue(c.Name),
				"description": optString(c.Description),
			}, timestampValues(c.CreatedAt, c.UpdatedAt))
		},
	}
}

func surveyEntity() entity[freshdesk.Survey] {
	return entity[freshdesk.Survey]{
		label: "survey",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"title":          computedString("Title of the survey."),
			"description":    computedString("Description of the survey."),
			"header_message": computedString("Message introducing the survey to the customer."),
			"state":          computedString("Whether the survey is `ACTIVE` or `INACTIVE`."),
			"active":         computedBool("Whether the survey is being sent."),
			"language":       computedString("Language the survey is written in."),
			"questions":      computedString("The survey's questions, as JSON."),
		}, timestampDataAttributes("survey")),
		types: mergeTypes(map[string]attr.Type{
			"title": types.StringType, "description": types.StringType,
			"header_message": types.StringType, "state": types.StringType,
			"active": types.BoolType, "language": types.StringType,
			"questions": types.StringType,
		}, timestampDataTypes()),
		// Surveys are keyed by a UUID rather than a number.
		idFn: func(s *freshdesk.Survey) types.String { return types.StringValue(s.ID) },
		mapFn: func(s *freshdesk.Survey) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"title":          types.StringValue(s.Title),
				"description":    optString(s.Description),
				"header_message": optString(s.HeaderMessage),
				"state":          optString(s.State),
				"active":         types.BoolValue(s.Active()),
				"language":       optString(s.Language),
				"questions":      jsonString(s.Questions),
			}, timestampValues(s.CreatedAt, s.UpdatedAt))
		},
	}
}

func satisfactionRatingEntity() entity[freshdesk.SatisfactionRating] {
	return entity[freshdesk.SatisfactionRating]{
		label: "satisfaction rating",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"ticket_id": computedInt64("ID of the ticket that was rated."),
			"survey_id": computedInt64("ID of the survey that was answered."),
			"agent_id":  computedInt64("ID of the agent who handled the ticket."),
			"group_id":  computedInt64("ID of the group that handled the ticket."),
			"user_id":   computedInt64("ID of the contact who left the rating."),
			"ratings":   computedString("Ratings keyed by question, as JSON."),
			"feedback":  computedString("Free-text comment the customer left."),
		}, timestampDataAttributes("rating")),
		types: mergeTypes(map[string]attr.Type{
			"ticket_id": types.Int64Type, "survey_id": types.Int64Type,
			"agent_id": types.Int64Type, "group_id": types.Int64Type,
			"user_id": types.Int64Type, "ratings": types.StringType,
			"feedback": types.StringType,
		}, timestampDataTypes()),
		idFn: func(r *freshdesk.SatisfactionRating) types.String { return idString(r.ID) },
		mapFn: func(r *freshdesk.SatisfactionRating) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"ticket_id": optInt64(r.TicketID),
				"survey_id": optInt64(r.SurveyID),
				"agent_id":  optInt64(r.AgentID),
				"group_id":  optInt64(r.GroupID),
				"user_id":   optInt64(r.UserID),
				"ratings":   jsonString(r.Ratings),
				"feedback":  optString(r.Feedback),
			}, timestampValues(r.CreatedAt, r.UpdatedAt))
		},
	}
}

func timeEntryEntity() entity[freshdesk.TimeEntry] {
	return entity[freshdesk.TimeEntry]{
		label: "time entry",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"ticket_id":     computedInt64("ID of the ticket the time is logged against."),
			"agent_id":      computedInt64("ID of the agent who logged the time."),
			"note":          computedString("Description of the work done."),
			"time_spent":    computedString("Duration as HH:MM."),
			"billable":      computedBool("Whether the time is chargeable."),
			"timer_running": computedBool("Whether the entry's timer is counting."),
			"start_time":    computedString("When the timer was started (RFC 3339)."),
			"executed_at":   computedString("When the work was done (RFC 3339)."),
		}, timestampDataAttributes("entry")),
		types: mergeTypes(map[string]attr.Type{
			"ticket_id": types.Int64Type, "agent_id": types.Int64Type,
			"note": types.StringType, "time_spent": types.StringType,
			"billable": types.BoolType, "timer_running": types.BoolType,
			"start_time": types.StringType, "executed_at": types.StringType,
		}, timestampDataTypes()),
		idFn: func(e *freshdesk.TimeEntry) types.String { return idString(e.ID) },
		mapFn: func(e *freshdesk.TimeEntry) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"ticket_id":     optInt64(e.TicketID),
				"agent_id":      optInt64(e.AgentID),
				"note":          optString(e.Note),
				"time_spent":    optString(e.TimeSpent),
				"billable":      types.BoolValue(e.Billable),
				"timer_running": types.BoolValue(e.TimerRunning),
				"start_time":    timeString(e.StartTime),
				"executed_at":   timeString(e.ExecutedAt),
			}, timestampValues(e.CreatedAt, e.UpdatedAt))
		},
	}
}

func emailMailboxEntity() entity[freshdesk.EmailMailbox] {
	return entity[freshdesk.EmailMailbox]{
		label: "email mailbox",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":                computedString("Name of the mailbox."),
			"support_email":       computedString("Address customers write to."),
			"group_id":            computedInt64("ID of the group tickets are assigned to."),
			"product_id":          computedInt64("ID of the product the mailbox belongs to."),
			"default_reply_email": computedBool("Whether replies are sent from this address."),
			"active":              computedBool("Whether the mailbox is accepting mail."),
			"mailbox_type":        computedString("Whether the mailbox is Freshdesk-hosted or custom."),
			"forward_email":       computedString("Address to forward mail to, if Freshdesk-hosted."),
			"csat_settings":       computedString("Satisfaction survey settings, as JSON."),
		}, timestampDataAttributes("mailbox")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "support_email": types.StringType,
			"group_id": types.Int64Type, "product_id": types.Int64Type,
			"default_reply_email": types.BoolType, "active": types.BoolType,
			"mailbox_type": types.StringType, "forward_email": types.StringType,
			"csat_settings": types.StringType,
		}, timestampDataTypes()),
		idFn: func(b *freshdesk.EmailMailbox) types.String { return idString(b.ID) },
		mapFn: func(b *freshdesk.EmailMailbox) map[string]attr.Value {
			forward := types.StringNull()
			if b.FreshdeskMailbox != nil {
				forward = optString(b.FreshdeskMailbox.ForwardEmail)
			}

			return mergeValues(map[string]attr.Value{
				"name":                types.StringValue(b.Name),
				"support_email":       optString(b.SupportEmail),
				"group_id":            optInt64(b.GroupID),
				"product_id":          optInt64(b.ProductID),
				"default_reply_email": types.BoolValue(b.DefaultReplyEmail),
				"active":              types.BoolValue(b.Active),
				"mailbox_type":        optString(b.MailboxType),
				"forward_email":       forward,
				"csat_settings":       jsonString(b.CSATSettings),
			}, timestampValues(b.CreatedAt, b.UpdatedAt))
		},
	}
}

func customObjectSchemaEntity() entity[freshdesk.CustomObjectSchema] {
	return entity[freshdesk.CustomObjectSchema]{
		label: "custom object schema",
		attrs: mergeAttrs(map[string]dschema.Attribute{
			"name":        computedString("Name of the custom object."),
			"description": computedString("Description of the custom object."),
			"prefix":      computedString("Prefix Freshdesk gives the object's record IDs."),
			"fields":      computedString("The object's fields, as JSON."),
		}, timestampDataAttributes("schema")),
		types: mergeTypes(map[string]attr.Type{
			"name": types.StringType, "description": types.StringType,
			"prefix": types.StringType, "fields": types.StringType,
		}, timestampDataTypes()),
		// Custom object schemas are keyed by a string rather than a number.
		idFn: func(s *freshdesk.CustomObjectSchema) types.String { return types.StringValue(s.ID) },
		mapFn: func(s *freshdesk.CustomObjectSchema) map[string]attr.Value {
			return mergeValues(map[string]attr.Value{
				"name":        types.StringValue(s.Name),
				"description": optString(s.Description),
				"prefix":      optString(s.Prefix),
				"fields":      jsonString(s.Fields),
			}, timestampValues(s.CreatedAt, s.UpdatedAt))
		},
	}
}
