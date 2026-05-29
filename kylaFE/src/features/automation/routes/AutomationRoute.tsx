import { useParams } from "react-router-dom"
import { WorkflowsList } from "../components/WorkflowsList"
import { WorkflowDetail } from "../components/WorkflowDetail"
import { AIStudio } from "../components/AIStudio"

export function AutomationListRoute() {
  return <WorkflowsList />
}

export function WorkflowDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return <WorkflowDetail workflowId={id} />
}

export function AIStudioRoute() {
  return <AIStudio />
}
