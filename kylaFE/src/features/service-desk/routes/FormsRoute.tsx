import { useParams } from "react-router-dom"
import { FormsList } from "../components/FormsList"
import { FormDetail } from "../components/FormDetail"

export function FormsListRoute() {
  return <FormsList />
}

export function FormDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return <FormDetail formId={id} />
}
