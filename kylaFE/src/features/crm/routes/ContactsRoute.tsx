import { useParams } from "react-router-dom"
import { ObjectList } from "../components/ObjectList"
import { ObjectDetail } from "../components/ObjectDetail"
import { TYPE_SLUG } from "../utils/types"

export function ContactsListRoute() {
  return (
    <ObjectList
      typeSlug={TYPE_SLUG.contact}
      title="Contacts"
      description="People you communicate with"
      basePath="/crm/contacts"
    />
  )
}

export function ContactDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return (
    <ObjectDetail
      objectId={id}
      backHref="/crm/contacts"
      backLabel="Contacts"
    />
  )
}
