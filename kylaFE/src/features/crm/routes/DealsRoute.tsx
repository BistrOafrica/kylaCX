import { useParams } from "react-router-dom"
import { PipelineBoard } from "../components/PipelineBoard"
import { ObjectDetail } from "../components/ObjectDetail"

export function DealsBoardRoute() {
  return <PipelineBoard basePath="/crm/deals" />
}

export function DealDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return (
    <ObjectDetail
      objectId={id}
      backHref="/crm/deals"
      backLabel="Deals"
    />
  )
}
