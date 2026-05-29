import { services, unary } from "@/lib/rpc"
import {
  AutoDialerCampaign,
  CreateAutoDialerCampaignRequest,
  ReadAutoDialerCampaignRequest,
  ReadAutoDialerCampaignsRequest,
  UpdateAutoDialerCampaignRequest,
  DeleteAutoDialerCampaignRequest,
} from "@/pb/autodialer_campaign"

export async function listAutoDialerCampaigns(): Promise<AutoDialerCampaign[]> {
  const res = await unary(
    services.autodialerCampaigns.readAutoDialerCampaigns(
      ReadAutoDialerCampaignsRequest.create({}) as ReadAutoDialerCampaignsRequest,
    ),
  )
  return (res as { campaigns?: AutoDialerCampaign[] }).campaigns ?? []
}

export async function getAutoDialerCampaign(id: string): Promise<AutoDialerCampaign | null> {
  const res = await unary(
    services.autodialerCampaigns.readAutoDialerCampaign(
      ReadAutoDialerCampaignRequest.create({ id }) as ReadAutoDialerCampaignRequest,
    ),
  )
  return (res as { campaign?: AutoDialerCampaign }).campaign ?? null
}

export async function createAutoDialerCampaign(input: {
  name: string
  contacts: string                // CSV
  startDate?: string
  endDate?: string
  targetType?: string
  target?: string
  frequency?: string
}): Promise<AutoDialerCampaign | null> {
  const campaign = AutoDialerCampaign.create({
    id: "",
    name: input.name,
    startDate: input.startDate ?? new Date().toISOString(),
    endDate: input.endDate ?? "",
    targetType: input.targetType ?? "agent",
    target: input.target ?? "",
    frequency: input.frequency ?? "once",
    contacts: input.contacts,
  }) as AutoDialerCampaign
  const res = await unary(
    services.autodialerCampaigns.createAutoDialerCampaign(
      CreateAutoDialerCampaignRequest.create({ campaign }) as CreateAutoDialerCampaignRequest,
    ),
  )
  return (res as { campaign?: AutoDialerCampaign }).campaign ?? null
}

export async function updateAutoDialerCampaign(
  campaign: AutoDialerCampaign,
): Promise<AutoDialerCampaign | null> {
  const res = await unary(
    services.autodialerCampaigns.updateAutoDialerCampaign(
      UpdateAutoDialerCampaignRequest.create({ campaign }) as UpdateAutoDialerCampaignRequest,
    ),
  )
  return (res as { campaign?: AutoDialerCampaign }).campaign ?? null
}

export async function deleteAutoDialerCampaign(id: string): Promise<void> {
  await unary(
    services.autodialerCampaigns.deleteAutoDialerCampaign(
      DeleteAutoDialerCampaignRequest.create({ id }) as DeleteAutoDialerCampaignRequest,
    ),
  )
}

export type { AutoDialerCampaign }
