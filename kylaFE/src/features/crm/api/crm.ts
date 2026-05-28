import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreatePipelineRequest,
  GetPipelineRequest,
  ListPipelinesRequest,
  UpdatePipelineRequest,
  DeletePipelineRequest,
  CreateStageRequest,
  ListStagesRequest,
  UpdateStageRequest,
  DeleteStageRequest,
  ReorderStagesRequest,
  MoveDealRequest,
  GetPipelineBoardRequest,
  type Pipeline,
  type PipelineStage,
  type PipelineBoardColumn,
  PipelineType,
} from "@/pb/crm"

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

export async function listPipelines(): Promise<Pipeline[]> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.crm.listPipelines(
      ListPipelinesRequest.create({ orgId, workspaceId }) as ListPipelinesRequest,
    ),
  )
  return res.pipelines
}

export async function getPipeline(id: string): Promise<Pipeline> {
  return unary(
    services.crm.getPipeline(
      GetPipelineRequest.create({
        id,
        orgId: scope().orgId,
      }) as GetPipelineRequest,
    ),
  )
}

export async function createPipeline(input: {
  name: string
  description?: string
  type?: PipelineType
  color?: string
}): Promise<Pipeline> {
  const { orgId, workspaceId } = scope()
  return unary(
    services.crm.createPipeline(
      CreatePipelineRequest.create({
        orgId,
        workspaceId,
        name: input.name,
        description: input.description ?? "",
        type: input.type ?? PipelineType.SALES,
        color: input.color ?? "#10B981",
      }) as CreatePipelineRequest,
    ),
  )
}

export async function updatePipeline(input: {
  id: string
  name?: string
  description?: string
  color?: string
}): Promise<Pipeline> {
  return unary(
    services.crm.updatePipeline(
      UpdatePipelineRequest.create({
        id: input.id,
        orgId: scope().orgId,
        name: input.name ?? "",
        description: input.description ?? "",
        color: input.color ?? "",
      }) as UpdatePipelineRequest,
    ),
  )
}

export async function deletePipeline(id: string): Promise<void> {
  await unary(
    services.crm.deletePipeline(
      DeletePipelineRequest.create({
        id,
        orgId: scope().orgId,
      }) as DeletePipelineRequest,
    ),
  )
}

// ── Stages ───────────────────────────────────────────────────────────────────

export async function listStages(pipelineId: string): Promise<PipelineStage[]> {
  const res = await unary(
    services.crm.listStages(
      ListStagesRequest.create({
        pipelineId,
        orgId: scope().orgId,
      }) as ListStagesRequest,
    ),
  )
  return res.stages
}

export async function createStage(input: {
  pipelineId: string
  name: string
  color?: string
  probability?: number
}): Promise<PipelineStage> {
  return unary(
    services.crm.createStage(
      CreateStageRequest.create({
        pipelineId: input.pipelineId,
        orgId: scope().orgId,
        name: input.name,
        color: input.color ?? "#A1A1AA",
        probability: input.probability ?? 0,
      }) as CreateStageRequest,
    ),
  )
}

export async function updateStage(input: {
  id: string
  name?: string
  color?: string
  probability?: number
}): Promise<PipelineStage> {
  return unary(
    services.crm.updateStage(
      UpdateStageRequest.create({
        id: input.id,
        orgId: scope().orgId,
        name: input.name ?? "",
        color: input.color ?? "",
        probability: input.probability ?? 0,
      }) as UpdateStageRequest,
    ),
  )
}

export async function deleteStage(id: string): Promise<void> {
  await unary(
    services.crm.deleteStage(
      DeleteStageRequest.create({
        id,
        orgId: scope().orgId,
      }) as DeleteStageRequest,
    ),
  )
}

export async function reorderStages(
  pipelineId: string,
  stageIds: string[],
): Promise<PipelineStage[]> {
  const res = await unary(
    services.crm.reorderStages(
      ReorderStagesRequest.create({
        pipelineId,
        orgId: scope().orgId,
        stageIds,
      }) as ReorderStagesRequest,
    ),
  )
  return res.stages
}

// ── Deals (pipeline operations) ──────────────────────────────────────────────

export async function moveDeal(
  dealObjectId: string,
  stageId: string,
): Promise<void> {
  await unary(
    services.crm.moveDeal(
      MoveDealRequest.create({
        dealObjectId,
        stageId,
        orgId: scope().orgId,
      }) as MoveDealRequest,
    ),
  )
}

export interface PipelineBoard {
  pipeline: Pipeline | null
  columns: PipelineBoardColumn[]
}

export async function getPipelineBoard(
  pipelineId: string,
  pageSize = 50,
): Promise<PipelineBoard> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.crm.getPipelineBoard(
      GetPipelineBoardRequest.create({
        pipelineId,
        orgId,
        workspaceId,
        pageSize,
      }) as GetPipelineBoardRequest,
    ),
  )
  return {
    pipeline: res.pipeline ?? null,
    columns: res.columns,
  }
}

export { PipelineType }
export type { Pipeline, PipelineStage, PipelineBoardColumn }
