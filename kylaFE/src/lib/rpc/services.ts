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
} as const

export type Services = typeof services
