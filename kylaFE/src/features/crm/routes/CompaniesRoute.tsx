import { useParams } from "react-router-dom"
import { ObjectList } from "../components/ObjectList"
import { ObjectDetail } from "../components/ObjectDetail"
import { TYPE_SLUG } from "../utils/types"

export function CompaniesListRoute() {
  return (
    <ObjectList
      typeSlug={TYPE_SLUG.company}
      title="Companies"
      description="Accounts your contacts belong to"
      basePath="/crm/companies"
    />
  )
}

export function CompanyDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return (
    <ObjectDetail
      objectId={id}
      backHref="/crm/companies"
      backLabel="Companies"
    />
  )
}
