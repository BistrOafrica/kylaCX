import { services, unary } from "@/lib/rpc"
import {
  CreateIvrFlowRequest,
  ReadIvrFlowRequest,
  ReadIvrFlowsRequest,
  UpdateIvrFlowRequest,
  IvrNodeType,
  type CallIvrFlow,
  type IvrFlowNode,
} from "@/pb/call_ivr_flow"

export async function listIvrFlows(args: {
  pageNumber?: number
  pageSize?: number
}): Promise<{ items: CallIvrFlow[]; total: number }> {
  const res = await unary(
    services.callIvrFlow.readIvrFlows(
      ReadIvrFlowsRequest.create({
        pageNumber: args.pageNumber ?? 1,
        pageSize: args.pageSize ?? 50,
      }) as ReadIvrFlowsRequest,
    ),
  )
  return {
    items: (res as { items?: CallIvrFlow[] }).items ?? [],
    total: Number((res as { total?: bigint | number }).total ?? 0),
  }
}

export async function getIvrFlow(id: string): Promise<CallIvrFlow> {
  return unary(
    services.callIvrFlow.readIvrFlow(
      ReadIvrFlowRequest.create({ id }) as ReadIvrFlowRequest,
    ),
  )
}

export async function createIvrFlow(input: {
  name: string
  description?: string
  nodes?: IvrFlowNode[]
}): Promise<void> {
  await unary(
    services.callIvrFlow.createIvrFlow(
      CreateIvrFlowRequest.create({
        name: input.name,
        description: input.description ?? "",
        nodes: input.nodes ?? [],
      }) as CreateIvrFlowRequest,
    ),
  )
}

export async function updateIvrFlow(input: {
  id: string
  name?: string
  description?: string
  nodes?: IvrFlowNode[]
}): Promise<void> {
  await unary(
    services.callIvrFlow.updateIvrFlow(
      UpdateIvrFlowRequest.create({
        id: input.id,
        name: input.name ?? "",
        description: input.description ?? "",
        nodes: input.nodes ?? [],
      }) as UpdateIvrFlowRequest,
    ),
  )
}

export { IvrNodeType }
export type { CallIvrFlow, IvrFlowNode }
