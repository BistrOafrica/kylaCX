import { services, unary } from "@/lib/rpc"
import {
  CreateWhatsappCampaignRequest,
  ViewWhatsappCampaignRequest,
  ListWhatsappCampaignRequest,
  CreateWhatsappTemplateRequest,
  ListWhatsappTemplateRequest,
  UpdateWhatsappTemplateRequest,
  DeleteWhatsappTemplateRequest,
  ViewWhatsappCampaignAnalyticsRequest,
  WhatsappCampaign,
  WhatsappTemplate,
  type WhatsappAnalytics,
  type AnalyticsCount,
} from "@/pb/whatsapp_campaigns"

// ── Campaigns ────────────────────────────────────────────────────────────────

export async function listWhatsappCampaigns(): Promise<WhatsappCampaign[]> {
  const res = await unary(
    services.whatsappCampaigns.listWhatsappCampaigns(
      ListWhatsappCampaignRequest.create({}) as ListWhatsappCampaignRequest,
    ),
  )
  return res.campaigns
}

export async function getWhatsappCampaign(id: string): Promise<WhatsappCampaign | null> {
  const res = await unary(
    services.whatsappCampaigns.viewWhatsappCampaign(
      ViewWhatsappCampaignRequest.create({ id }) as ViewWhatsappCampaignRequest,
    ),
  )
  return res.campaign ?? null
}

export async function createWhatsappCampaign(input: {
  name: string
  description?: string
  templateId: string
  contacts: string                // CSV of phone numbers or contact_ids
  startDate?: string              // ISO 8601
  endDate?: string                // ISO 8601
  frequency?: string              // "once" | "daily" | "weekly"
}): Promise<WhatsappCampaign | null> {
  const campaign = WhatsappCampaign.create({
    id: "",
    name: input.name,
    description: input.description ?? "",
    templateId: input.templateId,
    contacts: input.contacts,
    frequency: input.frequency ?? "once",
    startDate: input.startDate ?? new Date().toISOString(),
    endDate: input.endDate ?? "",
    status: "draft",
  }) as WhatsappCampaign
  const res = await unary(
    services.whatsappCampaigns.createWhatsappCampaign(
      CreateWhatsappCampaignRequest.create({ campaign }) as CreateWhatsappCampaignRequest,
    ),
  )
  return res.campaign ?? null
}

// ── Templates ────────────────────────────────────────────────────────────────

export async function listWhatsappTemplates(): Promise<WhatsappTemplate[]> {
  const res = await unary(
    services.whatsappCampaigns.listWhatsappTemplates(
      ListWhatsappTemplateRequest.create({}) as ListWhatsappTemplateRequest,
    ),
  )
  return res.templates
}

export async function createWhatsappTemplate(input: {
  name: string
  body: string
  language?: string
  category?: string
  header?: string
  footer?: string
  integrationId?: string
}): Promise<WhatsappTemplate | null> {
  const template = WhatsappTemplate.create({
    id: "",
    name: input.name,
    body: input.body,
    language: input.language ?? "en",
    category: input.category ?? "MARKETING",
    header: input.header ?? "",
    footer: input.footer ?? "",
    status: "PENDING",
    integrationId: input.integrationId ?? "",
    templateId: "",
    reason: "",
    phoneNumberId: "",
    wabaId: "",
  }) as WhatsappTemplate
  const res = await unary(
    services.whatsappCampaigns.createWhatsappTemplate(
      CreateWhatsappTemplateRequest.create({ template }) as CreateWhatsappTemplateRequest,
    ),
  )
  return res.template ?? null
}

export async function updateWhatsappTemplate(template: WhatsappTemplate): Promise<WhatsappTemplate | null> {
  const res = await unary(
    services.whatsappCampaigns.updateWhatsappTemplate(
      UpdateWhatsappTemplateRequest.create({ template }) as UpdateWhatsappTemplateRequest,
    ),
  )
  return res.template ?? null
}

export async function deleteWhatsappTemplate(id: string): Promise<void> {
  await unary(
    services.whatsappCampaigns.deleteWhatsappTemplate(
      DeleteWhatsappTemplateRequest.create({ id }) as DeleteWhatsappTemplateRequest,
    ),
  )
}

// ── Analytics ────────────────────────────────────────────────────────────────

export interface CampaignAnalyticsResult {
  rows: WhatsappAnalytics[]
  count: AnalyticsCount | null
}

export async function getCampaignAnalytics(
  campaignId: string,
  status = "",
): Promise<CampaignAnalyticsResult> {
  const res = await unary(
    services.whatsappCampaigns.viewWhatsappCampaignAnalytics(
      ViewWhatsappCampaignAnalyticsRequest.create({
        campaignId,
        status,
      }) as ViewWhatsappCampaignAnalyticsRequest,
    ),
  )
  return {
    rows: res.analytics,
    count: res.count ?? null,
  }
}

export type { WhatsappCampaign, WhatsappTemplate, WhatsappAnalytics, AnalyticsCount }
