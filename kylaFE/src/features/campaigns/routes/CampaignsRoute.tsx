import { useParams } from "react-router-dom"
import { CampaignsList } from "../components/CampaignsList"
import { WhatsappCampaignDetail } from "../components/WhatsappCampaignDetail"
import { AutodialerCampaignDetail } from "../components/AutodialerCampaignDetail"
import { TemplatesManager } from "../components/TemplatesManager"

export function CampaignsListRoute() {
  return <CampaignsList />
}

export function WhatsappCampaignRoute() {
  const { id } = useParams()
  if (!id) return null
  return <WhatsappCampaignDetail id={id} />
}

export function AutodialerCampaignRoute() {
  const { id } = useParams()
  if (!id) return null
  return <AutodialerCampaignDetail id={id} />
}

export function TemplatesRoute() {
  return <TemplatesManager />
}
