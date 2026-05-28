package casbin

import "strings"

// RoutePolicy describes the Casbin resource+action for a gRPC full method.
// If Open is true, the route requires only a valid JWT (no Casbin enforcement).
type RoutePolicy struct {
	Resource string
	Action   string
	Open     bool
}

// methodPolicies is the canonical map of every gRPC full method to its RoutePolicy.
// Methods NOT present in this map are denied by default.
var methodPolicies = func() map[string]RoutePolicy {
	const contactServicePath = "/da.proto.ContactService/"
	const groupServicePath = "/da.proto.GroupService/"
	const orgServicePath = "/da.proto.OrganisationService/"
	const userServicePath = "/da.proto.UserService/"
	const branchServicePath = "/da.proto.BranchService/"
	const departmentServicePath = "/da.proto.DepartmentService/"
	const roleServicePath = "/da.proto.RoleService/"
	const permissionServicePath = "/da.proto.PermissionService/"
	const authServicePath = "/da.proto.AuthService/"
	const tagServicePath = "/da.proto.TagService/"
	const labelServicePath = "/da.proto.LabelService/"
	const leadServicePath = "/da.proto.LeadService/"
	const pipelineServicePath = "/da.proto.PipelineService/"
	const pipelineStageServicePath = "/da.proto.PipelineStageService/"
	const stageLabelServicePath = "/da.proto.StageLabelService/"
	const chatdeskServicePath = "/da.proto.ChatdeskService/"
	const productServicePath = "/da.proto.ProductService/"
	const appServicePath = "/da.proto.AppService/"
	const callServicePath = "/da.proto.CallService/"
	const flowServicePath = "/da.proto.FlowService/"
	const integrationServicePath = "/da.proto.IntegrationService/"
	const leaveServicePath = "/da.proto.LeaveService/"
	const botFlowServicePath = "/da.proto.BotFlowService/"
	const shiftServicePath = "/da.proto.ShiftService/"
	const virtualAgentsPath = "/da.proto.VirtualAgentService/"
	const knowledgeBaseServicePath = "/da.proto.KnowledgeBaseService/"
	const invitationServicePath = "/da.proto.InvitationService/"
	const automationServicePath = "/da.proto.AutomationService/"
	const subscriptionPlansServicePath = "/da.proto.SubscriptionPlansService/"
	const subscriptionServicePath = "/da.proto.SubscriptionService/"
	const walletsServicePath = "/da.proto.WalletsService/"
	const billingAccountsServicePath = "/da.proto.BillingAccountsService/"
	const paymentMethodsServicePath = "/da.proto.PaymentMethodsService/"
	const workspaceServicePath = "/da.proto.WorkspaceService/"
	const resourceSharingServicePath = "/da.proto.ResourceSharing/"
	const encryptionServicePath = "/da.proto.EncryptionService/"
	const streamingServicePath = "/da.proto.StreamingService/"
	const objectCoreServicePath = "/da.proto.ObjectCoreService/"
	const viewServicePath = "/da.proto.ViewService/"
	const conversationServicePath = "/da.proto.ConversationService/"

	open := func(r, a string) RoutePolicy { return RoutePolicy{Resource: r, Action: a, Open: true} }
	p := func(r, a string) RoutePolicy { return RoutePolicy{Resource: r, Action: a} }

	return map[string]RoutePolicy{
		// ── Open / unauthenticated routes ──────────────────────────────────────
		resourceSharingServicePath + "GetRequests": open("sharing", "read"),

		encryptionServicePath + "GenerateIdentity":        open("encryption", "create"),
		encryptionServicePath + "GetPreKeyBundle":         open("encryption", "read"),
		encryptionServicePath + "RotateKeys":              open("encryption", "update"),
		encryptionServicePath + "CreateSession":           open("encryption", "create"),
		encryptionServicePath + "GetSession":              open("encryption", "read"),
		encryptionServicePath + "GetSessions":             open("encryption", "read"),
		encryptionServicePath + "GetAllIdentities":        open("encryption", "read"),
		encryptionServicePath + "GetIdentity":             open("encryption", "read"),
		encryptionServicePath + "SendMessage":             open("encryption", "create"),
		encryptionServicePath + "DecryptMessage":          open("encryption", "read"),
		encryptionServicePath + "ForwardMessage":          open("encryption", "create"),
		encryptionServicePath + "UpdateMessageStatus":     open("encryption", "update"),
		encryptionServicePath + "SendGroupMessage":        open("encryption", "create"),
		encryptionServicePath + "DecryptGroupMessage":     open("encryption", "read"),
		encryptionServicePath + "AddReaction":             open("encryption", "update"),
		encryptionServicePath + "AddAttachment":           open("encryption", "create"),
		encryptionServicePath + "UpdateTypingIndicator":   open("encryption", "update"),
		encryptionServicePath + "SearchMessages":          open("encryption", "read"),
		encryptionServicePath + "BackupMessages":          open("encryption", "read"),
		encryptionServicePath + "RestoreMessages":         open("encryption", "create"),
		encryptionServicePath + "EditMessage":             open("encryption", "update"),
		encryptionServicePath + "PinMessage":              open("encryption", "update"),
		encryptionServicePath + "UnpinMessage":            open("encryption", "update"),
		encryptionServicePath + "CreateThread":            open("encryption", "create"),
		encryptionServicePath + "AddMessageToThread":      open("encryption", "create"),
		encryptionServicePath + "ReplyToMessage":          open("encryption", "create"),
		encryptionServicePath + "SetMessageReminder":      open("encryption", "update"),
		encryptionServicePath + "ScheduleMessage":         open("encryption", "create"),
		encryptionServicePath + "CreateTemplate":          open("encryption", "create"),
		encryptionServicePath + "SendMessageFromTemplate": open("encryption", "create"),
		encryptionServicePath + "CreateCategory":          open("encryption", "create"),
		encryptionServicePath + "AssignMessageToCategory": open("encryption", "update"),
		encryptionServicePath + "ArchiveMessage":          open("encryption", "archive"),
		encryptionServicePath + "UnarchiveMessage":        open("encryption", "update"),

		streamingServicePath + "StreamMessages":         open("streaming", "read"),
		streamingServicePath + "StreamPresence":         open("streaming", "read"),
		streamingServicePath + "UpdatePresence":         open("streaming", "update"),
		streamingServicePath + "StreamTypingIndicators": open("streaming", "read"),
		streamingServicePath + "UpdateTypingStatus":     open("streaming", "update"),

		invitationServicePath + "AcceptInvitation": open("invitation", "update"),
		invitationServicePath + "RejectInvitation": open("invitation", "update"),

		// ── ContactService ─────────────────────────────────────────────────────
		contactServicePath + "CreateContact":              p("contact", "create"),
		contactServicePath + "ReadContact":                p("contact", "read"),
		contactServicePath + "ReadContacts":               p("contact", "read"),
		contactServicePath + "UpdateContact":              p("contact", "update"),
		contactServicePath + "DeleteContact":              p("contact", "delete"),
		contactServicePath + "ReadCustomFieldDefinitions": p("contact", "read"),
		contactServicePath + "BulkContactsImport":         p("contact", "create"),
		contactServicePath + "SearchContacts":             p("contact", "read"),
		contactServicePath + "ReadContactChildren":        p("contact", "read"),
		contactServicePath + "ReadAccountContacts":        p("contact", "read"),
		contactServicePath + "BulkContactsExport":         p("contact", "read"),
		contactServicePath + "FindContactDuplicates":      p("contact", "read"),
		contactServicePath + "MergeDuplicateContacts":     p("contact", "update"),

		// ── GroupService ───────────────────────────────────────────────────────
		groupServicePath + "CreateGroup":       p("group", "create"),
		groupServicePath + "ReadGroup":         p("group", "read"),
		groupServicePath + "ReadGroups":        p("group", "read"),
		groupServicePath + "UpdateGroup":       p("group", "update"),
		groupServicePath + "DeleteGroup":       p("group", "delete"),
		groupServicePath + "ReadGroupContacts": p("group", "read"),

		// ── OrganisationService ────────────────────────────────────────────────
		orgServicePath + "CreateOrganisation":          p("organisation", "create"),
		orgServicePath + "ReadOrganisation":            p("organisation", "read"),
		orgServicePath + "ReadOrganisations":           p("organisation", "read"),
		orgServicePath + "UpdateOrganisation":          p("organisation", "update"),
		orgServicePath + "DeleteOrganisation":          p("organisation", "delete"),
		orgServicePath + "CreateDefaultStructure":      p("organisation", "create"),
		orgServicePath + "CreateStructureWithTemplate": p("organisation", "create"),
		orgServicePath + "ReadMe":                      p("organisation", "read"),

		// ── UserService ────────────────────────────────────────────────────────
		userServicePath + "CreateUser":         p("user", "create"),
		userServicePath + "ReadUser":           p("user", "read"),
		userServicePath + "ReadMe":             p("user", "read"),
		userServicePath + "ReadUsers":          p("user", "read"),
		userServicePath + "UpdateUser":         p("user", "update"),
		userServicePath + "DeleteUser":         p("user", "delete"),
		userServicePath + "AddRoleToUser":      p("user", "update"),
		userServicePath + "RemoveRoleFromUser": p("user", "update"),
		userServicePath + "AddUsersToRole":     p("user", "update"),
		userServicePath + "SearchUsers":        p("user", "read"),

		// ── BranchService ──────────────────────────────────────────────────────
		branchServicePath + "CreateBranch": p("branch", "create"),
		branchServicePath + "ReadBranch":   p("branch", "read"),
		branchServicePath + "ReadBranches": p("branch", "read"),
		branchServicePath + "UpdateBranch": p("branch", "update"),
		branchServicePath + "DeleteBranch": p("branch", "delete"),

		// ── DepartmentService ──────────────────────────────────────────────────
		departmentServicePath + "CreateDepartment":      p("department", "create"),
		departmentServicePath + "ReadDepartment":        p("department", "read"),
		departmentServicePath + "ReadDepartments":       p("department", "read"),
		departmentServicePath + "UpdateDepartment":      p("department", "update"),
		departmentServicePath + "DeleteDepartment":      p("department", "delete"),
		departmentServicePath + "ReadBranchDepartments": p("department", "read"),

		// ── RoleService ────────────────────────────────────────────────────────
		roleServicePath + "CreateRole":               p("role", "create"),
		roleServicePath + "ReadRole":                 p("role", "read"),
		roleServicePath + "ReadRoles":                p("role", "read"),
		roleServicePath + "UpdateRole":               p("role", "update"),
		roleServicePath + "DeleteRole":               p("role", "delete"),
		roleServicePath + "AddPermissionToRole":      p("role", "update"),
		roleServicePath + "RemovePermissionFromRole": p("role", "update"),
		roleServicePath + "ReadUserRoles":            p("role", "read"),

		// ── PermissionService ──────────────────────────────────────────────────
		permissionServicePath + "ReadPermissions": p("permission", "read"),

		// ── AuthService ────────────────────────────────────────────────────────
		authServicePath + "Logout":              p("auth", "update"),
		authServicePath + "ChangePassword":      p("auth", "update"),
		authServicePath + "ReadUserPermissions": p("auth", "read"),

		// ── TagService ─────────────────────────────────────────────────────────
		tagServicePath + "CreateTag": p("tag", "create"),
		tagServicePath + "ReadTag":   p("tag", "read"),
		tagServicePath + "ReadTags":  p("tag", "read"),
		tagServicePath + "UpdateTag": p("tag", "update"),
		tagServicePath + "DeleteTag": p("tag", "delete"),

		// ── LabelService ───────────────────────────────────────────────────────
		labelServicePath + "CreateLabel": p("label", "create"),
		labelServicePath + "ReadLabel":   p("label", "read"),
		labelServicePath + "ReadLabels":  p("label", "read"),
		labelServicePath + "UpdateLabel": p("label", "update"),
		labelServicePath + "DeleteLabel": p("label", "delete"),

		// ── LeadService ────────────────────────────────────────────────────────
		leadServicePath + "OnContactCreate":    p("lead", "create"),
		leadServicePath + "OnUpdateContact":    p("lead", "update"),
		leadServicePath + "OnDeleteContact":    p("lead", "delete"),
		leadServicePath + "CreateLead":         p("lead", "create"),
		leadServicePath + "ReadLead":           p("lead", "read"),
		leadServicePath + "ReadLeads":          p("lead", "read"),
		leadServicePath + "ReadLeadKeys":       p("lead", "read"),
		leadServicePath + "UpdateLead":         p("lead", "update"),
		leadServicePath + "DeleteLead":         p("lead", "delete"),
		leadServicePath + "ReadAvailableLeads": p("lead", "read"),
		leadServicePath + "BulkLeadsImport":    p("lead", "create"),
		leadServicePath + "BulkLeadsExport":    p("lead", "read"),

		// ── PipelineService ────────────────────────────────────────────────────
		pipelineServicePath + "CreatePipeline":  p("pipeline", "create"),
		pipelineServicePath + "ReadPipeline":    p("pipeline", "read"),
		pipelineServicePath + "ReadPipelines":   p("pipeline", "read"),
		pipelineServicePath + "UpdatePipeline":  p("pipeline", "update"),
		pipelineServicePath + "DeletePipeline":  p("pipeline", "delete"),
		pipelineServicePath + "ReadAppMetaData": p("pipeline", "read"),

		// ── PipelineStageService ───────────────────────────────────────────────
		pipelineStageServicePath + "CreatePipelineStage": p("pipeline_stage", "create"),
		pipelineStageServicePath + "ReadPipelineStage":   p("pipeline_stage", "read"),
		pipelineStageServicePath + "ReadPipelineStages":  p("pipeline_stage", "read"),
		pipelineStageServicePath + "UpdatePipelineStage": p("pipeline_stage", "update"),
		pipelineStageServicePath + "DeletePipelineStage": p("pipeline_stage", "delete"),

		// ── StageLabelService ──────────────────────────────────────────────────
		stageLabelServicePath + "CreateStageLabel": p("stage_label", "create"),
		stageLabelServicePath + "ReadStageLabel":   p("stage_label", "read"),
		stageLabelServicePath + "ReadStageLabels":  p("stage_label", "read"),
		stageLabelServicePath + "UpdateStageLabel": p("stage_label", "update"),
		stageLabelServicePath + "DeleteStageLabel": p("stage_label", "delete"),

		// ── ChatdeskService ────────────────────────────────────────────────────
		chatdeskServicePath + "CreateTicket":                         p("ticket", "create"),
		chatdeskServicePath + "CanBeAssignedTickets":                 p("ticket", "read"),
		chatdeskServicePath + "ReadAllTickets":                       p("ticket", "read"),
		chatdeskServicePath + "UpdateTicket":                         p("ticket", "update"),
		chatdeskServicePath + "DeleteTicket":                         p("ticket", "delete"),
		chatdeskServicePath + "ChatdeskAnalytics":                    p("ticket_analytics", "read"),
		chatdeskServicePath + "CreateChatdeskSLAs":                   p("sla", "create"),
		chatdeskServicePath + "UpdateChatdeskSLAs":                   p("sla", "update"),
		chatdeskServicePath + "DeleteChatdeskSLAs":                   p("sla", "delete"),
		chatdeskServicePath + "ReadChatdeskSLAs":                     p("sla", "read"),
		chatdeskServicePath + "CreateChatdeskCustomizations":         p("chatdesk_customization", "create"),
		chatdeskServicePath + "UpdateChatdeskCustomizations":         p("chatdesk_customization", "update"),
		chatdeskServicePath + "DeleteChatdeskCustomizations":         p("chatdesk_customization", "delete"),
		chatdeskServicePath + "ReadChatdeskCustomizations":           p("chatdesk_customization", "read"),
		chatdeskServicePath + "CreateChatdeskGroups":                 p("chatdesk_group", "create"),
		chatdeskServicePath + "UpdateChatdeskGroups":                 p("chatdesk_group", "update"),
		chatdeskServicePath + "DeleteChatdeskGroups":                 p("chatdesk_group", "delete"),
		chatdeskServicePath + "ReadChatdeskGroups":                   p("chatdesk_group", "read"),
		chatdeskServicePath + "CreateChatdeskTags":                   p("chatdesk_tag", "create"),
		chatdeskServicePath + "UpdateChatdeskTags":                   p("chatdesk_tag", "update"),
		chatdeskServicePath + "DeleteChatdeskTags":                   p("chatdesk_tag", "delete"),
		chatdeskServicePath + "ReadChatdeskTags":                     p("chatdesk_tag", "read"),
		chatdeskServicePath + "CreateChatdeskMacros":                 p("chatdesk_macro", "create"),
		chatdeskServicePath + "UpdateChatdeskMacros":                 p("chatdesk_macro", "update"),
		chatdeskServicePath + "DeleteChatdeskMacros":                 p("chatdesk_macro", "delete"),
		chatdeskServicePath + "ReadChatdeskMacros":                   p("chatdesk_macro", "read"),
		chatdeskServicePath + "CreateChatdeskFAQs":                   p("faq", "create"),
		chatdeskServicePath + "UpdateChatdeskFAQs":                   p("faq", "update"),
		chatdeskServicePath + "DeleteChatdeskFAQs":                   p("faq", "delete"),
		chatdeskServicePath + "ReadChatdeskFAQs":                     p("faq", "read"),
		chatdeskServicePath + "CreateChatdeskCategories":             p("chatdesk_category", "create"),
		chatdeskServicePath + "UpdateChatdeskCategories":             p("chatdesk_category", "update"),
		chatdeskServicePath + "DeleteChatdeskCategories":             p("chatdesk_category", "delete"),
		chatdeskServicePath + "ReadChatdeskCategories":               p("chatdesk_category", "read"),
		chatdeskServicePath + "CreateChatdeskReports":                p("chatdesk_report", "create"),
		chatdeskServicePath + "UpdateChatdeskReports":                p("chatdesk_report", "update"),
		chatdeskServicePath + "DeleteChatdeskReports":                p("chatdesk_report", "delete"),
		chatdeskServicePath + "ReadChatdeskReports":                  p("chatdesk_report", "read"),
		chatdeskServicePath + "CreateChatdeskRooms":                  p("chatdesk_room", "create"),
		chatdeskServicePath + "UpdateChatdeskRooms":                  p("chatdesk_room", "update"),
		chatdeskServicePath + "DeleteChatdeskRooms":                  p("chatdesk_room", "delete"),
		chatdeskServicePath + "ReadChatdeskRooms":                    p("chatdesk_room", "read"),
		chatdeskServicePath + "CreateChatdeskCustomFieldDefinitions": p("chatdesk_custom_field", "create"),
		chatdeskServicePath + "UpdateChatdeskCustomFieldDefinitions": p("chatdesk_custom_field", "update"),
		chatdeskServicePath + "DeleteChatdeskCustomFieldDefinitions": p("chatdesk_custom_field", "delete"),
		chatdeskServicePath + "ReadChatdeskCustomFieldDefinitions":   p("chatdesk_custom_field", "read"),
		chatdeskServicePath + "CreateChatdeskAssignmentRules":        p("chatdesk_assignment_rule", "create"),
		chatdeskServicePath + "UpdateChatdeskAssignmentRules":        p("chatdesk_assignment_rule", "update"),
		chatdeskServicePath + "DeleteChatdeskAssignmentRules":        p("chatdesk_assignment_rule", "delete"),
		chatdeskServicePath + "ReadChatdeskAssignmentRules":          p("chatdesk_assignment_rule", "read"),
		chatdeskServicePath + "CreateChatdeskSLAPolicies":            p("sla_policy", "create"),
		chatdeskServicePath + "UpdateChatdeskSLAPolicies":            p("sla_policy", "update"),
		chatdeskServicePath + "DeleteChatdeskSLAPolicies":            p("sla_policy", "delete"),
		chatdeskServicePath + "ListChatdeskSLAPolicies":              p("sla_policy", "read"),
		chatdeskServicePath + "CreateChatdeskDashboards":             p("chatdesk_dashboard", "create"),
		chatdeskServicePath + "UpdateChatdeskDashboards":             p("chatdesk_dashboard", "update"),
		chatdeskServicePath + "DeleteChatdeskDashboards":             p("chatdesk_dashboard", "delete"),
		chatdeskServicePath + "ReadChatdeskDashboards":               p("chatdesk_dashboard", "read"),
		chatdeskServicePath + "GetKPIMetrics":                        p("ticket_analytics", "read"),
		chatdeskServicePath + "GetTicketVolume":                      p("ticket_analytics", "read"),
		chatdeskServicePath + "GetChannelDistribution":               p("ticket_analytics", "read"),
		chatdeskServicePath + "GetIntegrationDistribution":           p("ticket_analytics", "read"),
		chatdeskServicePath + "GetStatusDistribution":                p("ticket_analytics", "read"),
		chatdeskServicePath + "GetPriorityDistribution":              p("ticket_analytics", "read"),
		chatdeskServicePath + "GetAgentPerformance":                  p("ticket_analytics", "read"),
		chatdeskServicePath + "GetSLACompliance":                     p("ticket_analytics", "read"),
		chatdeskServicePath + "GetActivityHeatmap":                   p("ticket_analytics", "read"),
		chatdeskServicePath + "GetResolutionTimeAnalytics":           p("ticket_analytics", "read"),
		chatdeskServicePath + "GetEscalationMetrics":                 p("ticket_analytics", "read"),
		chatdeskServicePath + "GetLiveTicketTracking":                p("ticket_analytics", "read"),
		chatdeskServicePath + "GetTrendAnalysis":                     p("ticket_analytics", "read"),
		chatdeskServicePath + "GetSentimentAnalytics":                p("ticket_analytics", "read"),
		chatdeskServicePath + "CreateChatdeskAssignmentPolicies":     p("chatdesk_assignment_policy", "create"),
		chatdeskServicePath + "UpdateChatdeskAssignmentPolicies":     p("chatdesk_assignment_policy", "update"),
		chatdeskServicePath + "DeleteChatdeskAssignmentPolicies":     p("chatdesk_assignment_policy", "delete"),
		chatdeskServicePath + "ReadChatdeskAssignmentPolicies":       p("chatdesk_assignment_policy", "read"),

		// ── ProductService ─────────────────────────────────────────────────────
		productServicePath + "CreateProduct": p("product", "create"),
		productServicePath + "ReadProduct":   p("product", "read"),
		productServicePath + "ReadProducts":  p("product", "read"),
		productServicePath + "UpdateProduct": p("product", "update"),
		productServicePath + "DeleteProduct": p("product", "delete"),

		// ── AppService ─────────────────────────────────────────────────────────
		appServicePath + "CreateApp":  p("app", "create"),
		appServicePath + "ReadApp":    p("app", "read"),
		appServicePath + "ReadApps":   p("app", "read"),
		appServicePath + "UpdateApp":  p("app", "update"),
		appServicePath + "DeleteApp":  p("app", "delete"),
		appServicePath + "ApproveApp": p("app", "update"),
		appServicePath + "RejectApp":  p("app", "update"),

		// ── CallService ────────────────────────────────────────────────────────
		callServicePath + "ReadOverallReports":     p("call_report", "read"),
		callServicePath + "ReadOverallAnalytics":   p("call_analytics", "read"),
		callServicePath + "ReadPersonalReports":    p("call_report", "read"),
		callServicePath + "ReadPersonalAnalytics":  p("call_analytics", "read"),
		callServicePath + "ReadExtensionReports":   p("call_report", "read"),
		callServicePath + "ReadExtensionAnalytics": p("call_analytics", "read"),
		callServicePath + "ReadQueues":             p("call_queue", "read"),
		callServicePath + "ReadQueue":              p("call_queue", "read"),
		callServicePath + "ReadQueueCalls":         p("call_queue", "read"),
		callServicePath + "CreateQueue":            p("call_queue", "create"),
		callServicePath + "UpdateQueue":            p("call_queue", "update"),
		callServicePath + "DeleteQueue":            p("call_queue", "delete"),
		callServicePath + "ReadQueueMembers":       p("call_queue", "read"),
		callServicePath + "AddQueueMember":         p("call_queue", "update"),
		callServicePath + "RemoveQueueMember":      p("call_queue", "update"),
		callServicePath + "ReadIVRMenus":           p("call_ivr", "read"),
		callServicePath + "ReadIVRMenu":            p("call_ivr", "read"),
		callServicePath + "CreateIVRMenu":          p("call_ivr", "create"),
		callServicePath + "UpdateIVRMenu":          p("call_ivr", "update"),
		callServicePath + "DeleteIVRMenu":          p("call_ivr", "delete"),
		callServicePath + "ReadIVRMenuOptions":     p("call_ivr", "read"),
		callServicePath + "CreateIVRMenuOption":    p("call_ivr", "create"),
		callServicePath + "DeleteIVRMenuOption":    p("call_ivr", "delete"),
		callServicePath + "ReadRecordings":         p("call_recording", "read"),
		callServicePath + "UploadRecording":        p("call_recording", "create"),
		callServicePath + "DeleteRecording":        p("call_recording", "delete"),
		callServicePath + "ReadDestinations":       p("call_destination", "read"),
		callServicePath + "CreateDestination":      p("call_destination", "create"),
		callServicePath + "UpdateDestination":      p("call_destination", "update"),
		callServicePath + "DeleteDestination":      p("call_destination", "delete"),
		callServicePath + "ReadCsat":               p("call_csat", "read"),
		callServicePath + "ReadCsatReports":        p("call_csat", "read"),
		callServicePath + "CreateCsat":             p("call_csat", "create"),
		callServicePath + "UpdateCsat":             p("call_csat", "update"),
		callServicePath + "DeleteCsat":             p("call_csat", "delete"),
		callServicePath + "ReadMusicOnHold":        p("call_music", "read"),
		callServicePath + "CreateMusicOnHold":      p("call_music", "create"),
		callServicePath + "UpdateMusicOnHold":      p("call_music", "update"),
		callServicePath + "DeleteMusicOnHold":      p("call_music", "delete"),
		callServicePath + "MonitorCalls":           p("call_monitoring", "read"),
		callServicePath + "ViewOnlineAgents":       p("call_monitoring", "read"),
		callServicePath + "TransferCalls":          p("call_session", "update"),
		callServicePath + "ReadExtensions":         p("call_extension", "read"),
		callServicePath + "ReadExtension":          p("call_extension", "read"),
		callServicePath + "CreateExtension":        p("call_extension", "create"),
		callServicePath + "UpdateExtension":        p("call_extension", "update"),
		callServicePath + "DeleteExtension":        p("call_extension", "delete"),
		callServicePath + "AssignExtension":        p("call_extension", "update"),
		callServicePath + "MakeCall":               p("call_session", "create"),

		// ── FlowService ────────────────────────────────────────────────────────
		flowServicePath + "CreateFlow":     p("flow", "create"),
		flowServicePath + "ReadFlow":       p("flow", "read"),
		flowServicePath + "ReadFlows":      p("flow", "read"),
		flowServicePath + "UpdateFlow":     p("flow", "update"),
		flowServicePath + "DeleteFlow":     p("flow", "delete"),
		flowServicePath + "CreateFlowNode": p("flow_node", "create"),
		flowServicePath + "ReadFlowNode":   p("flow_node", "read"),
		flowServicePath + "ReadFlowNodes":  p("flow_node", "read"),
		flowServicePath + "UpdateFlowNode": p("flow_node", "update"),
		flowServicePath + "DeleteFlowNode": p("flow_node", "delete"),
		flowServicePath + "CreateFlowEdge": p("flow_edge", "create"),
		flowServicePath + "ReadFlowEdge":   p("flow_edge", "read"),
		flowServicePath + "ReadFlowEdges":  p("flow_edge", "read"),
		flowServicePath + "UpdateFlowEdge": p("flow_edge", "update"),
		flowServicePath + "DeleteFlowEdge": p("flow_edge", "delete"),

		// ── IntegrationService ─────────────────────────────────────────────────
		integrationServicePath + "CreateIntegrations": p("integration", "create"),
		integrationServicePath + "UpdateIntegrations": p("integration", "update"),
		integrationServicePath + "DeleteIntegrations": p("integration", "delete"),
		integrationServicePath + "ReadIntegrations":   p("integration", "read"),

		// ── LeaveService ───────────────────────────────────────────────────────
		leaveServicePath + "CreateLeaveType":          p("leave_type", "create"),
		leaveServicePath + "UpdateLeaveType":          p("leave_type", "update"),
		leaveServicePath + "DeleteLeaveType":          p("leave_type", "delete"),
		leaveServicePath + "ReadLeaveType":            p("leave_type", "read"),
		leaveServicePath + "ReadLeaveTypes":           p("leave_type", "read"),
		leaveServicePath + "CreateLeaveRequest":       p("leave_request", "create"),
		leaveServicePath + "ReadLeaveRequest":         p("leave_request", "read"),
		leaveServicePath + "ReadLeaveRequests":        p("leave_request", "read"),
		leaveServicePath + "ReadMyLeaveRequests":      p("leave_request", "read"),
		leaveServicePath + "ApproveLeaveRequest":      p("leave_request", "update"),
		leaveServicePath + "RejectLeaveRequest":       p("leave_request", "update"),
		leaveServicePath + "CancelLeaveRequest":       p("leave_request", "update"),
		leaveServicePath + "EndLeaveRequest":          p("leave_request", "update"),
		leaveServicePath + "AppealLeaveRequest":       p("leave_request", "create"),
		leaveServicePath + "ReadLeaveBalance":         p("leave_balance", "read"),
		leaveServicePath + "ReadLeaveBalances":        p("leave_balance", "read"),
		leaveServicePath + "ReadLeaveRequestsMetrics": p("leave_request", "read"),

		// ── BotFlowService ─────────────────────────────────────────────────────
		botFlowServicePath + "CreateBotFlow":           p("bot_flow", "create"),
		botFlowServicePath + "ReadBotFlow":             p("bot_flow", "read"),
		botFlowServicePath + "ReadBotFlows":            p("bot_flow", "read"),
		botFlowServicePath + "UpdateBotFlow":           p("bot_flow", "update"),
		botFlowServicePath + "DeleteBotFlow":           p("bot_flow", "delete"),
		botFlowServicePath + "CopyBotFlow":             p("bot_flow", "create"),
		botFlowServicePath + "AssignIntegrationToFlow": p("bot_flow", "update"),

		// ── ShiftService ───────────────────────────────────────────────────────
		shiftServicePath + "CreateShift": p("shift", "create"),
		shiftServicePath + "ReadShift":   p("shift", "read"),
		shiftServicePath + "UpdateShift": p("shift", "update"),
		shiftServicePath + "DeleteShift": p("shift", "delete"),
		shiftServicePath + "ReadShifts":  p("shift", "read"),

		// ── VirtualAgentService ────────────────────────────────────────────────
		virtualAgentsPath + "CreateVirtualAgent":           p("virtual_agent", "create"),
		virtualAgentsPath + "ReadVirtualAgent":             p("virtual_agent", "read"),
		virtualAgentsPath + "ReadAllVirtualAgents":         p("virtual_agent", "read"),
		virtualAgentsPath + "UpdateVirtualAgent":           p("virtual_agent", "update"),
		virtualAgentsPath + "DeleteVirtualAgent":           p("virtual_agent", "delete"),
		virtualAgentsPath + "ReadAllToolDefinitions":       p("virtual_agent", "read"),
		virtualAgentsPath + "ReadAllToolActionDefinitions": p("virtual_agent", "read"),
		virtualAgentsPath + "AssignToolsToVirtualAgent":    p("virtual_agent", "update"),
		virtualAgentsPath + "PromptVirtualAgent":           p("virtual_agent", "read"),
		virtualAgentsPath + "ReadThreadHistory":            p("virtual_agent", "read"),
		virtualAgentsPath + "ReadThreadMessages":           p("virtual_agent", "read"),

		// ── KnowledgeBaseService ───────────────────────────────────────────────
		knowledgeBaseServicePath + "CreateKnowledgeBase": p("knowledge_base", "create"),
		knowledgeBaseServicePath + "ReadKnowledgeBase":   p("knowledge_base", "read"),
		knowledgeBaseServicePath + "ReadKnowledgeBases":  p("knowledge_base", "read"),
		knowledgeBaseServicePath + "UpdateKnowledgeBase": p("knowledge_base", "update"),
		knowledgeBaseServicePath + "DeleteKnowledgeBase": p("knowledge_base", "delete"),
		knowledgeBaseServicePath + "KnowledgeBaseSearch": p("knowledge_base", "read"),

		// ── InvitationService (protected) ──────────────────────────────────────
		invitationServicePath + "CreateInvitation": p("invitation", "create"),
		invitationServicePath + "ReadInvitation":   p("invitation", "read"),
		invitationServicePath + "ReadInvitations":  p("invitation", "read"),
		invitationServicePath + "UpdateInvitation": p("invitation", "update"),
		invitationServicePath + "DeleteInvitation": p("invitation", "delete"),
		invitationServicePath + "CancelInvitation": p("invitation", "update"),

		// ── AutomationService ──────────────────────────────────────────────────
		automationServicePath + "CreateAutomation": p("automation", "create"),
		automationServicePath + "ReadAutomation":   p("automation", "read"),
		automationServicePath + "ReadAutomations":  p("automation", "read"),
		automationServicePath + "UpdateAutomation": p("automation", "update"),
		automationServicePath + "DeleteAutomation": p("automation", "delete"),
		automationServicePath + "SaveTrigger":      p("automation", "update"),
		automationServicePath + "DeleteTrigger":    p("automation", "update"),
		automationServicePath + "ReadTriggers":     p("automation", "read"),
		automationServicePath + "SaveAction":       p("automation", "update"),
		automationServicePath + "DeleteAction":     p("automation", "update"),
		automationServicePath + "ReadActions":      p("automation", "read"),

		// ── Billing ────────────────────────────────────────────────────────────
		subscriptionPlansServicePath + "ReadSubscriptionPlans": p("billing_plan", "read"),
		subscriptionPlansServicePath + "SyncSubscriptionPlans": p("billing_plan", "update"),
		subscriptionServicePath + "SubscribeToPlan":            p("billing_subscription", "create"),
		subscriptionServicePath + "CancelSubscription":         p("billing_subscription", "update"),
		walletsServicePath + "CreateWallet":                    p("billing_wallet", "create"),
		walletsServicePath + "TopUpWallet":                     p("billing_wallet", "update"),
		walletsServicePath + "ChargeWallet":                    p("billing_wallet", "update"),
		walletsServicePath + "ReadWallet":                      p("billing_wallet", "read"),
		walletsServicePath + "ReadWalletTransactions":          p("billing_wallet", "read"),
		billingAccountsServicePath + "CreateBillingAccount":    p("billing_account", "create"),
		billingAccountsServicePath + "ReadBillingAccount":      p("billing_account", "read"),
		paymentMethodsServicePath + "ReadAllPaymentMethods":    p("billing_payment_method", "read"),

		// ── WorkspaceService (Phase 1) ─────────────────────────────────────────
		workspaceServicePath + "CreateWorkspace":  p("workspace", "create"),
		workspaceServicePath + "GetWorkspace":     p("workspace", "read"),
		workspaceServicePath + "ListWorkspaces":   p("workspace", "read"),
		workspaceServicePath + "UpdateWorkspace":  p("workspace", "update"),
		workspaceServicePath + "ArchiveWorkspace": p("workspace", "archive"),
		workspaceServicePath + "AddMember":        p("member", "create"),
		workspaceServicePath + "RemoveMember":     p("member", "delete"),
		workspaceServicePath + "UpdateMemberRole": p("member", "update"),
		workspaceServicePath + "ListMembers":      p("member", "read"),

		// ── ObjectCoreService (Phase 2) ────────────────────────────────────────
		objectCoreServicePath + "CreateObjectType":   p("object_type", "create"),
		objectCoreServicePath + "GetObjectType":      p("object_type", "read"),
		objectCoreServicePath + "ListObjectTypes":    p("object_type", "read"),
		objectCoreServicePath + "UpdateObjectSchema": p("object_type", "update"),
		objectCoreServicePath + "DeleteObjectType":   p("object_type", "delete"),
		objectCoreServicePath + "CreateObject":       p("object", "create"),
		objectCoreServicePath + "GetObject":          p("object", "read"),
		objectCoreServicePath + "ListObjects":        p("object", "read"),
		objectCoreServicePath + "SearchObjects":      p("object", "read"),
		objectCoreServicePath + "UpdateObject":       p("object", "update"),
		objectCoreServicePath + "DeleteObject":       p("object", "delete"),
		objectCoreServicePath + "LinkObjects":        p("object_relation", "create"),
		objectCoreServicePath + "UnlinkObjects":      p("object_relation", "delete"),
		objectCoreServicePath + "GetObjectRelations": p("object_relation", "read"),
		objectCoreServicePath + "GetObjectTimeline":  p("object", "read"),

		// ── ViewService (Phase 2) ──────────────────────────────────────────────
		viewServicePath + "CreateSavedView": p("saved_view", "create"),
		viewServicePath + "GetSavedView":    p("saved_view", "read"),
		viewServicePath + "ListSavedViews":  p("saved_view", "read"),
		viewServicePath + "UpdateSavedView": p("saved_view", "update"),
		viewServicePath + "DeleteSavedView": p("saved_view", "delete"),

		// ── ConversationService (Phase 3) ──────────────────────────────────────
		conversationServicePath + "CreateConversation":         p("conversation", "create"),
		conversationServicePath + "GetConversation":            p("conversation", "read"),
		conversationServicePath + "ListConversations":          p("conversation", "read"),
		conversationServicePath + "AssignConversation":         p("conversation", "update"),
		conversationServicePath + "UpdateConversationStatus":   p("conversation", "update"),
		conversationServicePath + "ResolveConversation":        p("conversation", "update"),
		conversationServicePath + "SendMessage":                p("message", "create"),
		conversationServicePath + "ListMessages":               p("message", "read"),
		conversationServicePath + "StreamConversationUpdates":  p("conversation", "read"),
		conversationServicePath + "GetConversationTimeline":    p("conversation", "read"),
	}
}()

// MethodPolicy returns the RoutePolicy for the given gRPC full method name.
// Returns (policy, true) if found. If not found, returns (zero, false) — caller should deny.
func MethodPolicy(fullMethod string) (RoutePolicy, bool) {
	rp, ok := methodPolicies[fullMethod]
	return rp, ok
}

// codeSingulars maps common plural resource names from permission codenames to their
// Casbin singular forms, ensuring consistent policy matching.
var codeSingulars = map[string]string{
	"contacts":      "contact",
	"organisations": "organisation",
	"workspaces":    "workspace",
	"tickets":       "ticket",
	"conversations": "conversation",
	"members":       "member",
	"roles":         "role",
	"permissions":   "permission",
	"apps":          "app",
	"branches":      "branch",
	"teams":         "team",
	"users":         "user",
	"departments":   "department",
	"tags":          "tag",
	"labels":        "label",
	"leads":         "lead",
	"deals":         "deal",
	"pipelines":     "pipeline",
	"flows":         "flow",
	"integrations":  "integration",
	"shifts":        "shift",
	"leaves":        "leave",
	"groups":        "group",
	"notes":         "note",
	"queues":        "queue",
	"sessions":      "session",
	"logs":          "log",
	"reports":       "report",
	"agents":        "agent",
	"campaigns":     "campaign",
	"auths":         "auth",
}

// CodeNameToPolicy parses a permission codename (e.g. "CREATE_WORKSPACE") into
// Casbin (resource, action) strings. The codename format is {ACTION}_{RESOURCE}.
// Returns ("", "", false) if the codename cannot be parsed.
func CodeNameToPolicy(codeName string) (resource, action string, ok bool) {
	lower := strings.ToLower(codeName)
	idx := strings.Index(lower, "_")
	if idx < 0 || idx == len(lower)-1 {
		return "", "", false
	}
	action = lower[:idx]
	resource = lower[idx+1:]
	// Normalise plural resource names used in codenames to singular forms.
	if singular, found := codeSingulars[resource]; found {
		resource = singular
	}
	return resource, action, true
}
