import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import {
	ClientReadableStream,
	Metadata,
	Request,
	RpcError,
	StreamInterceptor,
} from "grpc-web";
import { AuthServiceClient } from "../pb/auth.client";
import { OrganisationServiceClient } from "../pb/organisations.client";
import { UserServiceClient } from "../pb/user.client";
import {
	FlowServiceClient,
	PipelineServiceClient,
	PipelineStageServiceClient,
	StageLabelServiceClient,
} from "../pb/pipelines.client";
import { MessageServiceClient } from "../pb/messaging.client";
import { LeadServiceClient } from "../pb/lead.client";
import { ContactServiceClient } from "../pb/contact.client";
import { QAServiceClient } from "../pb/qa.client";
import { QACommentServiceClient } from "../pb/qa_comment.client";
import {
	IntegrationServiceClient,
	WhatsAppEmbeddedSignUpServiceClient,
} from "../pb/integration.client";
import { CallScriptServiceClient } from "../pb/call_script.client";
import { CallNoteServiceClient } from "../pb/call_note.client";
import { CallMacroServiceClient } from "../pb/call_macro.client";
import { CallLogServiceClient } from "../pb/call_log.client";
import { CallDialplanServiceClient } from "../pb/call_dialplan.client";
import { CallTagServiceClient } from "../pb/call_tag.client";
import { CallIvrFlowServiceClient } from "../pb/call_ivr_flow.client";
import { CallIvrMenuServiceClient } from "../pb/call_ivr_menu.client";
import { CallIvrTriggerServiceClient } from "../pb/call_ivr_trigger.client";
import { CallQueueServiceClient } from "../pb/call_queue.client";
import { CallExtensionServiceClient } from "../pb/call_extension.client";
import { CallMonitoringServiceClient } from "../pb/call_monitoring.client";
import { CallAnalyticsServiceClient } from "../pb/call_analytics.client";
import { CallAnalyticsExportServiceClient } from "../pb/call_analytics_export.client";

import { LeaveServiceClient } from "../pb/leave.client";
import { GroupServiceClient } from "../pb/contact_groups.client";
import { AppServiceClient } from "../pb/apps.client";
import { PermissionServiceClient, RoleServiceClient } from "../pb/rbac.client";
import { TeamServiceClient } from "../pb/team.client";
import { DepartmentServiceClient } from "../pb/department.client";
import { BranchServiceClient } from "../pb/branch.client";
import {
	BreakServiceClient,
	ShiftScheduleServiceClient,
	ShiftServiceClient,
} from "../pb/shifts.client";
import { TicketServiceClient } from "../pb/ticket.client";
import { FaqServiceClient } from "../pb/faq.client";
import { TicketMacroServiceClient } from "../pb/ticket_macro.client";
import { TicketScriptServiceClient } from "../pb/ticket_script.client";
import { TicketCategoryServiceClient } from "../pb/category.client";
import { TicketTagServiceClient } from "../pb/ticket_tag.client";
import { TicketRoomServiceClient } from "../pb/ticket_room.client";
import { TicketLogServiceClient } from "../pb/ticket_log.client";
import { TicketCustomFieldDefinitionServiceClient } from "../pb/ticket_custom_field_definition.client";
import { TicketNoteServiceClient } from "../pb/ticket_note.client";
import { AttachmentServiceClient } from "../pb/attachment.client";
import { TicketAssignmentPolicyServiceClient } from "../pb/ticket_assignment_rule.client";
import { VirtualAgentServiceClient } from "../pb/virtual_agents.client";
import { KnowledgeBaseServiceClient } from "../pb/knowledge_base.client";
import { RAGAgentServiceClient } from "../pb/rag_agents.client";
import { InvitationServiceClient } from "../pb/invitation.client";
import { SLAServiceClient } from "../pb/sla.client";
import { ThreadServiceClient } from "../pb/thread.client";
import { EscalationThreadServiceClient } from "../pb/escalation_thread.client";
import { TicketAnalyticsServiceClient } from "../pb/ticket_analytics.client";
import { TicketDashboardServiceClient } from "../pb/ticket_dashboards.client";
import { QAAgentScoreServiceClient } from "../pb/qa_agent_score.client";
import { BusinessRuleServiceClient } from "../pb/business_rule.client";
import { SubscriptionPlansServiceClient } from "../pb/billing_subscription_plans.client";
import { SubscriptionServiceClient } from "../pb/billing_subscription.client";
import { WalletsServiceClient } from "../pb/billing_wallets.client";
import { BillingAccountsServiceClient } from "../pb/billing_accounts.client";
import { PaymentMethodsServiceClient } from "../pb/billing_payment_methods.client";
import { CallAudioServiceClient } from "../../chat-sdk/lib/pb/call_audio.client";
import { SIPNumberServiceClient } from "../../chat-sdk/lib/pb/call_sip_number.client";
import { SIPServerServiceClient } from "../../chat-sdk/lib/pb/call_sip_server.client";
import { SIPServerAssignmentServiceClient } from "../../chat-sdk/lib/pb/call_sip_server_assignment.client";
import { SIPTrunkServiceClient } from "../../chat-sdk/lib/pb/call_sip_trunk.client";

export const orgHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_ORG_HOSTNAME as string;
};

export const crmHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_CRM_HOSTNAME as string;
	//   return import.meta.env.VITE_CRM_HOSTNAME_LOCAL as string;
};

export const qaHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_QA_HOSTNAME as string;
};

export const integrationHttpsHostName = (): string => {
	// @ts-ignore

	return import.meta.env.VITE_INTEGRATION_HTTPS_HOSTNAME as string;
};

export const integrationGrpcHostName = (): string => {
	// @ts-ignore

	return import.meta.env.VITE_INTEGRATION_GRPC_HOSTNAME as string;
};

export const chatdeskHttpsHostName = (): string => {
	// @ts-ignore

	return import.meta.env.VITE_CHATDESK_HTTPS_HOSTNAME as string;
};

export const chatdeskHostName = (): string => {
	// @ts-ignore

	return import.meta.env.VITE_CHATDESK_HOSTNAME as string;
};

export const virtualAgentsHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_NIA_HOSTNAME;
};

export const contactServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.ContactService/";
};

export const groupServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.GroupService/";
};

export const integrationServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.IntegrationService/";
};

export const leaveServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.LeaveService/";
};

export const leadServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.LeadService/";
};
export const departmentServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.DepartmentService/";
};
export const orgServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.OrganizationService/";
};
export const userServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.UserService/";
};
export const branchServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.BranchService/";
};
export const roleServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.RoleService/";
};
export const permissionServicePath = (): string => {
	return "/da.proto.PermissionService/";
};
export const appServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.AppService/";
};

export const chatHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_ORG_CHAT_HOSTNAME as string;
};
export const chatdeskServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.ChatdeskService/";
};
export const ticketServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketService/";
};
export const faqServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.FaqService/";
};
export const ticketTagServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketTagService/";
};
export const ticketCategoryServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketCategoryService/";
};
export const ticketScriptServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketScriptService/";
};
export const ticketNoteServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketNoteService/";
};
export const attachmentServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.AttachmentService/";
};
export const ticketMacroServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketMacroService/";
};
export const ticketRoomServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketRoomService/";
};
export const ticketLogServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketLogService/";
};
export const ticketCustomFieldDefinitionServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketCustomFieldDefinitionService/";
};
export const ticketAssignmentRuleServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.TicketAssignmentRuleService/";
};
export const threadServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.ThreadService/";
};

export const callHostName = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_CALL_HOSTNAME as string;
};

export const callScriptServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallScriptService/";
};

export const callMacroServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallMacroService/";
};

export const callNoteServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallNoteService/";
};

export const callLogServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallLogService/";
};

export const callDialplanServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallDialplanService/";
};

export const callTagServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallTagService/";
};

export const callIvrFlowServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallIvrFlowService/";
};

export const callIvrMenuServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallIvrMenuService/";
};

export const callIvrTriggerServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallIvrTriggerService/";
};

export const callQueueServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallQueueService/";
};

export const callExtensionServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallExtensionService/";
};

export const callSipServerServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.SIPServerService/";
};

export const callSipServerAssignmentServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.SIPServerAssignmentService/";
};

export const callAudioServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallAudioService/";
};

export const callMonitoringServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.CallMonitoringService/";
};

export const sipTrunkServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.SIPTrunkService/";
};

export const sipNumberServicePath = (): string => {
	// @ts-ignore
	return "/da.proto.SIPNumberService/";
};

export const billingHostname = (): string => {
	// @ts-ignore
	return import.meta.env.VITE_BILLING_URL as string;
};



export class GRPCServiceClient<T> {
	private hostname: string;
	public Service: T;

	constructor(
		hostname: string,
		ServiceClient: new (transport: GrpcWebFetchTransport) => T
	) {
		this.hostname = hostname;

		let transport = new GrpcWebFetchTransport({
			baseUrl: this.hostname,
			format: "binary",
		});

		this.Service = new ServiceClient(transport);
	}

	getService(): T {
		return this.Service;
	}

	getMethod(methodName: string) {
		return (this.Service as unknown as { [key: string]: any })[methodName];
	}
}

export const orgServiceClient =
	new GRPCServiceClient<OrganisationServiceClient>(
		orgHostName(),
		OrganisationServiceClient
	);
export const authServiceClient = new GRPCServiceClient<AuthServiceClient>(
	orgHostName(),
	AuthServiceClient
);

export const userServiceClient = new GRPCServiceClient<UserServiceClient>(
	orgHostName(),
	UserServiceClient
);

export const invitationServiceClient =
	new GRPCServiceClient<InvitationServiceClient>(
		orgHostName(),
		InvitationServiceClient
	);

export const appServiceClient = new GRPCServiceClient<AppServiceClient>(
	orgHostName(),
	AppServiceClient
);

export const roleServiceClient = new GRPCServiceClient<RoleServiceClient>(
	orgHostName(),
	RoleServiceClient
);

export const teamServiceClient = new GRPCServiceClient<TeamServiceClient>(
	orgHostName(),
	TeamServiceClient
);

export const departmentServiceClient =
	new GRPCServiceClient<DepartmentServiceClient>(
		orgHostName(),
		DepartmentServiceClient
	);

export const branchServiceClient = new GRPCServiceClient<BranchServiceClient>(
	orgHostName(),
	BranchServiceClient
);

export const permissionServiceClient =
	new GRPCServiceClient<PermissionServiceClient>(
		orgHostName(),
		PermissionServiceClient
	);

export const shiftServiceClient = new GRPCServiceClient<ShiftServiceClient>(
	orgHostName(),
	ShiftServiceClient
);

export const shiftScheduleServiceClient =
	new GRPCServiceClient<ShiftScheduleServiceClient>(
		orgHostName(),
		ShiftScheduleServiceClient
	);

export const scheduleBreakServiceClient =
	new GRPCServiceClient<BreakServiceClient>(orgHostName(), BreakServiceClient);

export const contactServiceClient = new GRPCServiceClient<ContactServiceClient>(
	orgHostName(),
	ContactServiceClient
);

export const groupServiceClient = new GRPCServiceClient<GroupServiceClient>(
	orgHostName(),
	GroupServiceClient
);

export const pipelineServiceClient =
	new GRPCServiceClient<PipelineServiceClient>(
		crmHostName(),
		PipelineServiceClient
	);

export const leadServiceClient = new GRPCServiceClient<LeadServiceClient>(
	crmHostName(),
	LeadServiceClient
);

export const pipelineStageServiceClient =
	new GRPCServiceClient<PipelineStageServiceClient>(
		crmHostName(),
		PipelineStageServiceClient
	);

export const stageLabelServiceClient =
	new GRPCServiceClient<StageLabelServiceClient>(
		crmHostName(),
		StageLabelServiceClient
	);
export const flowServiceClient = new GRPCServiceClient<FlowServiceClient>(
	crmHostName(),
	FlowServiceClient
);
export const messagingServiceClient =
	new GRPCServiceClient<MessageServiceClient>(
		"http://localhost:8082",
		MessageServiceClient
	);

export const sendMessageServiceClient =
	new GRPCServiceClient<MessageServiceClient>(
		"http://localhost:8083",
		MessageServiceClient
	);

export const qaServiceClient = new GRPCServiceClient<QAServiceClient>(
	qaHostName(),
	QAServiceClient
);

export const qaServiceCommentsClient =
	new GRPCServiceClient<QACommentServiceClient>(
		qaHostName(),
		QACommentServiceClient
	);

export const whatsappEmbeddedSignUpServiceClient =
	new GRPCServiceClient<WhatsAppEmbeddedSignUpServiceClient>(
		integrationGrpcHostName(),
		WhatsAppEmbeddedSignUpServiceClient
	);

export const integrationServiceClient =
	new GRPCServiceClient<IntegrationServiceClient>(
		integrationGrpcHostName(),
		IntegrationServiceClient
	);

export const businessRuleClient =
	new GRPCServiceClient<BusinessRuleServiceClient>(
		crmHostName(),
		BusinessRuleServiceClient
	);

export const ticketServiceClient = new GRPCServiceClient<TicketServiceClient>(
	chatdeskHostName(),
	TicketServiceClient
);

export const faqServiceClient = new GRPCServiceClient<FaqServiceClient>(
	chatdeskHostName(),
	FaqServiceClient
);

export const ticketMacroServiceClient =
	new GRPCServiceClient<TicketMacroServiceClient>(
		chatdeskHostName(),
		TicketMacroServiceClient
	);

export const ticketScriptServiceClient =
	new GRPCServiceClient<TicketScriptServiceClient>(
		chatdeskHostName(),
		TicketScriptServiceClient
	);

export const ticketCategoryServiceClient =
	new GRPCServiceClient<TicketCategoryServiceClient>(
		chatdeskHostName(),
		TicketCategoryServiceClient
	);

export const ticketTagServiceClient =
	new GRPCServiceClient<TicketTagServiceClient>(
		chatdeskHostName(),
		TicketTagServiceClient
	);

export const ticketRoomServiceClient =
	new GRPCServiceClient<TicketRoomServiceClient>(
		chatdeskHostName(),
		TicketRoomServiceClient
	);

export const ticketLogServiceClient =
	new GRPCServiceClient<TicketLogServiceClient>(
		chatdeskHostName(),
		TicketLogServiceClient
	);

export const ticketCustomFieldDefinitionServiceClient =
	new GRPCServiceClient<TicketCustomFieldDefinitionServiceClient>(
		chatdeskHostName(),
		TicketCustomFieldDefinitionServiceClient
	);

export const ticketNoteServiceClient =
	new GRPCServiceClient<TicketNoteServiceClient>(
		chatdeskHostName(),
		TicketNoteServiceClient
	);

export const attachmentServiceClient =
	new GRPCServiceClient<AttachmentServiceClient>(
		chatdeskHostName(),
		AttachmentServiceClient
	);

export const ticketAssignmentPolicyServiceClient =
	new GRPCServiceClient<TicketAssignmentPolicyServiceClient>(
		chatdeskHostName(),
		TicketAssignmentPolicyServiceClient
	);

export const slaServiceClient = new GRPCServiceClient<SLAServiceClient>(
	chatdeskHostName(),
	SLAServiceClient
);

export const threadServiceClient = new GRPCServiceClient<ThreadServiceClient>(
	chatdeskHostName(),
	ThreadServiceClient
);

export const escalationThreadServiceClient =
	new GRPCServiceClient<EscalationThreadServiceClient>(
		chatdeskHostName(),
		EscalationThreadServiceClient
	);

export const ticketAnalyticsServiceClient =
	new GRPCServiceClient<TicketAnalyticsServiceClient>(
		chatdeskHostName(),
		TicketAnalyticsServiceClient
	);

export const ticketDashboardServiceClient =
	new GRPCServiceClient<TicketDashboardServiceClient>(
		chatdeskHostName(),
		TicketDashboardServiceClient
	);

export const callScriptClient = new GRPCServiceClient<CallScriptServiceClient>(
	callHostName(),
	CallScriptServiceClient
);

export const callMacroClient = new GRPCServiceClient<CallMacroServiceClient>(
	callHostName(),
	CallMacroServiceClient
);

export const callNoteClient = new GRPCServiceClient<CallNoteServiceClient>(
	callHostName(),
	CallNoteServiceClient
);

export const callLogClient = new GRPCServiceClient<CallLogServiceClient>(
	callHostName(),
	CallLogServiceClient
);

export const callDialplanClient =
	new GRPCServiceClient<CallDialplanServiceClient>(
		callHostName(),
		CallDialplanServiceClient
	);

export const callTagClient = new GRPCServiceClient<CallTagServiceClient>(
	callHostName(),
	CallTagServiceClient
);

export const callIvrFlowClient =
	new GRPCServiceClient<CallIvrFlowServiceClient>(
		callHostName(),
		CallIvrFlowServiceClient
	);

export const callIvrMenuClient =
	new GRPCServiceClient<CallIvrMenuServiceClient>(
		callHostName(),
		CallIvrMenuServiceClient
	);

export const callIvrTriggerClient =
	new GRPCServiceClient<CallIvrTriggerServiceClient>(
		callHostName(),
		CallIvrTriggerServiceClient
	);

export const callQueueClient = new GRPCServiceClient<CallQueueServiceClient>(
	callHostName(),
	CallQueueServiceClient
);

export const callExtensionClient =
	new GRPCServiceClient<CallExtensionServiceClient>(
		callHostName(),
		CallExtensionServiceClient
	);
export const callSipServerClient =
	new GRPCServiceClient<SIPServerServiceClient>(
		callHostName(),
		SIPServerServiceClient
	);

export const callSipServerAssignmentClient =
	new GRPCServiceClient<SIPServerAssignmentServiceClient>(
		callHostName(),
		SIPServerAssignmentServiceClient
	);

export const callAudioClient = new GRPCServiceClient<CallAudioServiceClient>(
	callHostName(),
	CallAudioServiceClient
);

export const callMonitoringClient: GRPCServiceClient<CallMonitoringServiceClient> =
	new GRPCServiceClient<CallMonitoringServiceClient>(
		callHostName(),
		CallMonitoringServiceClient
	);

export const callAnalyticsClient: GRPCServiceClient<CallAnalyticsServiceClient> =
	new GRPCServiceClient<CallAnalyticsServiceClient>(
		callHostName(),
		CallAnalyticsServiceClient
	);

export const callAnalyticsExportClient: GRPCServiceClient<CallAnalyticsExportServiceClient> =
	new GRPCServiceClient<CallAnalyticsExportServiceClient>(
		callHostName(),
		CallAnalyticsExportServiceClient
	);

export const callSipTrunkClient = new GRPCServiceClient<SIPTrunkServiceClient>(
	callHostName(),
	SIPTrunkServiceClient
);

export const callSipNumberClient =
	new GRPCServiceClient<SIPNumberServiceClient>(
		callHostName(),
		SIPNumberServiceClient
	);

export const leaveServiceClient = new GRPCServiceClient<LeaveServiceClient>(
	orgHostName(),
	LeaveServiceClient
);

export const virtualAgentServiceClient =
	new GRPCServiceClient<VirtualAgentServiceClient>(
		virtualAgentsHostName(),
		VirtualAgentServiceClient
	);

export const knowledgeBaseServiceClient =
	new GRPCServiceClient<KnowledgeBaseServiceClient>(
		virtualAgentsHostName(),
		KnowledgeBaseServiceClient
	);

export const ragAgentServiceClient =
	new GRPCServiceClient<RAGAgentServiceClient>(
		virtualAgentsHostName(),
		RAGAgentServiceClient
	);

export const qaAgentScoreServiceClient =
	new GRPCServiceClient<QAAgentScoreServiceClient>(
		qaHostName(),
		QAAgentScoreServiceClient
	);


export const subscriptionPlansServiceClient =
	new GRPCServiceClient<SubscriptionPlansServiceClient>(
		billingHostname(),
		SubscriptionPlansServiceClient
	);

export const subscriptionServiceClient =
	new GRPCServiceClient<SubscriptionServiceClient>(
		billingHostname(),
		SubscriptionServiceClient
	)

export const walletsServiceClient =
	new GRPCServiceClient<WalletsServiceClient>(
		billingHostname(),
		WalletsServiceClient
	)

export const billingAccountsServiceClient =
	new GRPCServiceClient<BillingAccountsServiceClient>(
		billingHostname(),
		BillingAccountsServiceClient
	)


export const paymentMethodsServiceClient =
	new GRPCServiceClient<PaymentMethodsServiceClient>(
		billingHostname(),
		PaymentMethodsServiceClient
	)

class AuthInterceptorStream<REQ extends Request<any, any>, RES = any>
	implements StreamInterceptor<REQ, RES> {
	intercept(
		request: REQ,
		invoker: (request: REQ, metadata?: Metadata) => ClientReadableStream<RES>
	) {
		const metadata = request.getMetadata();
		metadata["Content-Type"] = "application/grpc-web-text";
		const stream = invoker(request);
		console.log("Intercepted stream", stream);
		return new AuthReadableStreamWrapper<RES>(stream);
	}
}

class AuthReadableStreamWrapper<RES = any> {
	private stream: ClientReadableStream<RES>;

	constructor(stream: ClientReadableStream<RES>) {
		this.stream = stream;
	}
	on<F extends Function>(eventType: any, callback: F) {
		if (eventType === "error") {
			this.stream.on("error", (err: RpcError) => {
				callback(err);
			});
		} else if (eventType === "data") {
			this.stream.on("data", (resp) => {
				console.log("interceptor data", resp);
				callback(resp);
			});
		} else if (eventType === "status") {
			this.stream.on("status", (status) => {
				callback(status);
			});
		} else if (eventType === "end") {
			this.stream.on("end", callback as any);
		}
		return this;
	}

	removeListener(eventType: any, callback: any) {
		this.stream.removeListener(eventType, callback);
	}

	cancel() {
		this.stream.cancel();
		return this;
	}
}

export { AuthInterceptorStream };
