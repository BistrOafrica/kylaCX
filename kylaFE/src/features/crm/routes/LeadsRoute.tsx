import { useParams } from "react-router-dom"
import { ObjectList } from "../components/ObjectList"
import { ObjectDetail } from "../components/ObjectDetail"
import { TYPE_SLUG } from "../utils/types"

export function LeadsListRoute() {
  return (
    <ObjectList
      typeSlug={TYPE_SLUG.lead}
      title="Leads"
      description="Prospective contacts not yet qualified"
      basePath="/crm/leads"
    />
  )
}

export function LeadDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return (
    <ObjectDetail
      objectId={id}
      backHref="/crm/leads"
      backLabel="Leads"
    />
  )
}
