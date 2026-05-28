import { useParams } from "react-router-dom"
import { KnowledgeList } from "../components/KnowledgeList"
import { KnowledgeEditor } from "../components/KnowledgeEditor"

export function KnowledgeListRoute() {
  return <KnowledgeList />
}

export function KnowledgeDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return <KnowledgeEditor id={id} />
}
