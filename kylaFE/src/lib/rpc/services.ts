/**
 * Service client registry — typed handles to every backend gRPC service
 * the new code uses. Each client shares the single `transport` so
 * metadata interceptors apply uniformly.
 *
 * Add a new service here when its first real call is added. Until then
 * the legacy `kylaFE/src/api/globalClient/GlobalClients.ts` continues
 * to serve the demo pages.
 */

import { transport } from "./client"
import { AuthServiceClient } from "@/pb/auth.client"
import { OrganisationServiceClient } from "@/pb/organisations.client"
import { UserServiceClient } from "@/pb/user.client"
import { ConversationServiceClient } from "@/pb/conversations.client"
import { ContactServiceClient } from "@/pb/contact.client"
import { SentimentAnalysisClient } from "@/pb/text_classification.client"
import { SummarizationClient } from "@/pb/summarization.client"
import { ObjectCoreServiceClient, ViewServiceClient } from "@/pb/object_core.client"
import { CRMServiceClient } from "@/pb/crm.client"
import { TicketingServiceClient } from "@/pb/ticketing.client"
import { KnowledgeBaseServiceClient } from "@/pb/knowledge_base.client"
import { FormsServiceClient } from "@/pb/forms.client"
import { WorkflowServiceClient } from "@/pb/workflow.client"
import { AIServiceClient } from "@/pb/ai.client"
import { BranchServiceClient } from "@/pb/branch.client"
import { DepartmentServiceClient } from "@/pb/department.client"
import { TeamServiceClient } from "@/pb/team.client"
import { RoleServiceClient, PermissionServiceClient } from "@/pb/rbac.client"
import { InvitationServiceClient } from "@/pb/invitation.client"
import { AppServiceClient } from "@/pb/apps.client"
import { AgentStatusServiceClient } from "@/pb/agents.client"
import { TicketAnalyticsServiceClient } from "@/pb/ticket_analytics.client"
import { CallAnalyticsServiceClient } from "@/pb/call_analytics.client"
import { SubscriptionServiceClient } from "@/pb/billing_subscription.client"
import { WalletsServiceClient } from "@/pb/billing_wallets.client"
import { CallLogServiceClient } from "@/pb/call_log.client"
import { CallSessionServiceClient } from "@/pb/call_session.client"
import { CallQueueServiceClient } from "@/pb/call_queue.client"
import { CallExtensionServiceClient } from "@/pb/call_extension.client"
import { CallIvrFlowServiceClient } from "@/pb/call_ivr_flow.client"
import { CallMonitoringServiceClient } from "@/pb/call_monitoring.client"
import { CallNoteServiceClient } from "@/pb/call_note.client"
import { CallNumberServiceClient } from "@/pb/call_number.client"
import { WhatsappServiceClient } from "@/pb/whatsapp_campaigns.client"
import { AutoDialerServiceClient } from "@/pb/autodialer_campaign.client"
// Phase 5 unified telephony + IVR — supersedes the scattered call_*.client.ts
// imports above for new development. Legacy clients stay in place for now.
import { TelephonyServiceClient } from "@/pb/telephony.client"
import { IVRServiceClient } from "@/pb/ivr.client"

export const services = {
  auth: new AuthServiceClient(transport),
  organisation: new OrganisationServiceClient(transport),
  user: new UserServiceClient(transport),
  conversation: new ConversationServiceClient(transport),
  contact: new ContactServiceClient(transport),
  sentiment: new SentimentAnalysisClient(transport),
  summarization: new SummarizationClient(transport),
  objectCore: new ObjectCoreServiceClient(transport),
  views: new ViewServiceClient(transport),
  crm: new CRMServiceClient(transport),
  ticketing: new TicketingServiceClient(transport),
  knowledge: new KnowledgeBaseServiceClient(transport),
  forms: new FormsServiceClient(transport),
  workflow: new WorkflowServiceClient(transport),
  ai: new AIServiceClient(transport),
  branch: new BranchServiceClient(transport),
  department: new DepartmentServiceClient(transport),
  team: new TeamServiceClient(transport),
  role: new RoleServiceClient(transport),
  permission: new PermissionServiceClient(transport),
  invitation: new InvitationServiceClient(transport),
  app: new AppServiceClient(transport),
  agentStatus: new AgentStatusServiceClient(transport),
  ticketAnalytics: new TicketAnalyticsServiceClient(transport),
  callAnalytics: new CallAnalyticsServiceClient(transport),
  subscription: new SubscriptionServiceClient(transport),
  wallets: new WalletsServiceClient(transport),
  callLog: new CallLogServiceClient(transport),
  callSession: new CallSessionServiceClient(transport),
  callQueue: new CallQueueServiceClient(transport),
  callExtension: new CallExtensionServiceClient(transport),
  callIvrFlow: new CallIvrFlowServiceClient(transport),
  callMonitoring: new CallMonitoringServiceClient(transport),
  callNote: new CallNoteServiceClient(transport),
  callNumber: new CallNumberServiceClient(transport),
  whatsappCampaigns: new WhatsappServiceClient(transport),
  autodialerCampaigns: new AutoDialerServiceClient(transport),
  telephony: new TelephonyServiceClient(transport),
  ivr: new IVRServiceClient(transport),
} as const

export type Services = typeof services
