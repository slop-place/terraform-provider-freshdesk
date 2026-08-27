package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// unassignedForValidator restricts a group's escalation delay to the values
// Freshdesk accepts.
func unassignedForValidator() validatorString {
	return stringvalidator.OneOf("30m", "1h", "2h", "4h", "8h", "12h", "1d", "2d", "3d")
}

// ticketPriorityValidator restricts a ticket priority to Low..Urgent.
func ticketPriorityValidator() validatorInt64 {
	return int64validator.Between(freshdesk.TicketPriorityLow, freshdesk.TicketPriorityUrgent)
}

// ticketSourceValidator restricts a ticket source to the documented channels.
func ticketSourceValidator() validatorInt64 {
	return int64validator.OneOf(
		freshdesk.TicketSourceEmail,
		freshdesk.TicketSourcePortal,
		freshdesk.TicketSourcePhone,
		freshdesk.TicketSourceChat,
		freshdesk.TicketSourceFeedbackWidget,
		freshdesk.TicketSourceOutboundEmail,
		freshdesk.TicketSourceEcommerce,
	)
}

// agentScopeValidator restricts an agent's ticket scope.
func agentScopeValidator() validatorInt64 {
	return int64validator.Between(freshdesk.AgentScopeGlobal, freshdesk.AgentScopeRestricted)
}

// agentTypeValidator restricts an agent's type.
func agentTypeValidator() validatorInt64 {
	return int64validator.Between(freshdesk.AgentTypeSupport, freshdesk.AgentTypeCollaborator)
}

// matchTypeValidator restricts a rule's condition joiner.
func matchTypeValidator() validatorString {
	return stringvalidator.OneOf("all", "any")
}

// articleStatusValidator restricts a solution article's status.
func articleStatusValidator() validatorInt64 {
	return int64validator.OneOf(freshdesk.ArticleStatusDraft, freshdesk.ArticleStatusPublished)
}

// folderVisibilityValidator restricts a solution folder's visibility.
func folderVisibilityValidator() validatorInt64 {
	return int64validator.Between(freshdesk.FolderVisibilityAllUsers, freshdesk.FolderVisibilityBots)
}

// forumTypeValidator restricts a forum's type.
func forumTypeValidator() validatorInt64 {
	return int64validator.Between(freshdesk.ForumTypeHowTo, freshdesk.ForumTypeAnnouncement)
}

// forumVisibilityValidator restricts a forum's visibility.
func forumVisibilityValidator() validatorInt64 {
	return int64validator.Between(
		freshdesk.ForumVisibilityAll, freshdesk.ForumVisibilitySelectCompanies)
}

// cannedResponseVisibilityValidator restricts a canned response's visibility.
func cannedResponseVisibilityValidator() validatorInt64 {
	return int64validator.Between(
		freshdesk.CannedResponseVisibilityAllAgents, freshdesk.CannedResponseVisibilityGroups)
}

// automationTypeValidator restricts an automation rule to a documented type.
func automationTypeValidator() validatorInt64 {
	return int64validator.OneOf(
		freshdesk.AutomationTypeTicketCreation,
		freshdesk.AutomationTypeTimeTriggered,
		freshdesk.AutomationTypeTicketUpdate,
	)
}

// groupTypeValidator restricts an admin group's type.
func groupTypeValidator() validatorString {
	return stringvalidator.OneOf(freshdesk.GroupTypeSupportAgent, freshdesk.GroupTypeFieldAgent)
}

// mailboxTypeValidator restricts a mailbox's type.
func mailboxTypeValidator() validatorString {
	return stringvalidator.OneOf(freshdesk.MailboxTypeFreshdesk, freshdesk.MailboxTypeCustom)
}
