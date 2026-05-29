import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  GetWorkflowRequest,
  ListWorkflowsRequest,
  DeleteWorkflowRequest,
  GetRunHistoryRequest,
  TestRunWorkflowRequest,
} from "@/pb/workflow"
import {
  toWorkflowSpec,
  toWorkflowJson,
  toWorkflowRun,
  type WorkflowSpec,
  type WorkflowRunRow,
} from "../utils/types"
import { toStruct } from "../utils/struct"

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

export interface WorkflowsPage {
  items: WorkflowSpec[]
  total: number
}

export async function listWorkflows(
  pageNumber = 1,
  pageSize = 50,
): Promise<WorkflowsPage> {
  const { workspaceId } = scope()
  const res = await unary(
    services.workflow.listWorkflows(
      ListWorkflowsRequest.create({
        workspaceId,
        pageNumber,
        pageSize,
      }) as ListWorkflowsRequest,
    ),
  )
  return {
    items: res.items.map(toWorkflowSpec),
    total: Number(res.total),
  }
}

export async function getWorkflow(id: string): Promise<WorkflowSpec> {
  const res = await unary(
    services.workflow.getWorkflow(
      GetWorkflowRequest.create({ id }) as GetWorkflowRequest,
    ),
  )
  return toWorkflowSpec(res)
}

export async function createWorkflow(
  spec: Omit<WorkflowSpec, "id" | "orgId" | "workspaceId" | "createdAt" | "updatedAt" | "createdBy">,
): Promise<WorkflowSpec> {
  const { workspaceId } = scope()
  const { trigger, conditions, actions } = toWorkflowJson(spec)
  const res = await unary(
    services.workflow.createWorkflow(
      CreateWorkflowRequest.create({
        workspaceId,
        name: spec.name,
        description: spec.description,
        trigger,
        conditions,
        actions,
        status: spec.status,
      }) as CreateWorkflowRequest,
    ),
  )
  return toWorkflowSpec(res)
}

export async function updateWorkflow(
  spec: WorkflowSpec,
): Promise<WorkflowSpec> {
  const { trigger, conditions, actions } = toWorkflowJson(spec)
  const res = await unary(
    services.workflow.updateWorkflow(
      UpdateWorkflowRequest.create({
        id: spec.id,
        name: spec.name,
        description: spec.description,
        trigger,
        conditions,
        actions,
        status: spec.status,
      }) as UpdateWorkflowRequest,
    ),
  )
  return toWorkflowSpec(res)
}

export async function deleteWorkflow(id: string): Promise<void> {
  await unary(
    services.workflow.deleteWorkflow(
      DeleteWorkflowRequest.create({ id }) as DeleteWorkflowRequest,
    ),
  )
}

export interface RunHistoryPage {
  runs: WorkflowRunRow[]
  total: number
}

export async function getRunHistory(
  workflowId: string,
  pageNumber = 1,
  pageSize = 25,
): Promise<RunHistoryPage> {
  const res = await unary(
    services.workflow.getRunHistory(
      GetRunHistoryRequest.create({
        workflowId,
        pageNumber,
        pageSize,
      }) as GetRunHistoryRequest,
    ),
  )
  return {
    runs: res.runs.map(toWorkflowRun),
    total: Number(res.total),
  }
}

export async function testRunWorkflow(
  workflowId: string,
  sampleEvent: Record<string, unknown> = {},
): Promise<{ temporalRunId: string }> {
  const res = await unary(
    services.workflow.testRunWorkflow(
      TestRunWorkflowRequest.create({
        workflowId,
        sampleEvent: toStruct(sampleEvent),
      }) as TestRunWorkflowRequest,
    ),
  )
  return { temporalRunId: res.temporalRunId }
}
