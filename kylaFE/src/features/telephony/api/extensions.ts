import { services, unary } from "@/lib/rpc"
import {
  ReadExtensionRequest,
  ReadExtensionsRequest,
  ReadByAgentIdRequest,
  type CallExtension,
} from "@/pb/call_extension"

export async function listExtensions(args: {
  pageNumber?: number
  pageSize?: number
}): Promise<{ items: CallExtension[]; total: number }> {
  const res = await unary(
    services.callExtension.readExtensions(
      ReadExtensionsRequest.create({
        pageNumber: args.pageNumber ?? 1,
        pageSize: args.pageSize ?? 50,
      }) as ReadExtensionsRequest,
    ),
  )
  return {
    items: (res as { items?: CallExtension[] }).items ?? [],
    total: Number((res as { total?: bigint | number }).total ?? 0),
  }
}

export async function getExtension(id: string): Promise<CallExtension> {
  return unary(
    services.callExtension.readExtension(
      ReadExtensionRequest.create({ id }) as ReadExtensionRequest,
    ),
  )
}

export async function getExtensionByAgent(agentId: string): Promise<CallExtension> {
  return unary(
    services.callExtension.readByAgentId(
      ReadByAgentIdRequest.create({ agentId }) as ReadByAgentIdRequest,
    ),
  )
}

export type { CallExtension }
