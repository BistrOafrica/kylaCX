import { services, unary } from "@/lib/rpc"
import {
  CreateKnowledgeBaseRequest,
  ReadKnowledgeBaseRequest,
  ReadAllKnowledgeBasesRequest,
  UpdateKnowledgeBaseRequest,
  DeleteKnowledgeBaseRequest,
  KnowledgeBaseSearchRequest,
  KnowledgeBase,
  KnowledgeBaseType,
} from "@/pb/knowledge_base"

/**
 * Knowledge base API.
 *
 * Note: the current proto exposes a "knowledge collection" model
 * (text / website / FAQ) rather than the categorised article + draft +
 * publish flow described in the roadmap. The UI mirrors what the
 * backend can deliver today; when the article+category proto lands
 * we add a parallel set of wrappers without breaking callers.
 */

export interface KnowledgePage {
  items: KnowledgeBase[]
  totalCount: number
}

export async function listKnowledgeBases(
  limit = 50,
  offset = 0,
): Promise<KnowledgePage> {
  const res = await unary(
    services.knowledge.readAllKnowledgeBases(
      ReadAllKnowledgeBasesRequest.create({ limit, offset }) as ReadAllKnowledgeBasesRequest,
    ),
  )
  return {
    items: res.knowledgeBases,
    totalCount: Number(res.totalCount),
  }
}

export async function getKnowledgeBase(id: string): Promise<KnowledgeBase | null> {
  const res = await unary(
    services.knowledge.readKnowledgeBase(
      ReadKnowledgeBaseRequest.create({ id }) as ReadKnowledgeBaseRequest,
    ),
  )
  return res.knowledgeBase ?? null
}

export async function createKnowledgeBase(input: {
  name: string
  description?: string
  type?: KnowledgeBaseType
  text?: string
  url?: string
}): Promise<KnowledgeBase | null> {
  const kb = KnowledgeBase.create({
    id: "",
    name: input.name,
    description: input.description ?? "",
    type: input.type ?? KnowledgeBaseType.TEXT,
    text: input.text ?? "",
    url: input.url ?? "",
    createdAt: "",
    updatedAt: "",
  }) as KnowledgeBase
  const res = await unary(
    services.knowledge.createKnowledgeBase(
      CreateKnowledgeBaseRequest.create({ knowledgeBase: kb }) as CreateKnowledgeBaseRequest,
    ),
  )
  return res.knowledgeBase ?? null
}

export async function updateKnowledgeBase(kb: KnowledgeBase): Promise<KnowledgeBase | null> {
  const res = await unary(
    services.knowledge.updateKnowledgeBase(
      UpdateKnowledgeBaseRequest.create({ knowledgeBase: kb }) as UpdateKnowledgeBaseRequest,
    ),
  )
  return res.knowledgeBase ?? null
}

export async function deleteKnowledgeBase(id: string): Promise<void> {
  await unary(
    services.knowledge.deleteKnowledgeBase(
      DeleteKnowledgeBaseRequest.create({ id }) as DeleteKnowledgeBaseRequest,
    ),
  )
}

export async function searchKnowledge(
  query: string,
  limit = 25,
  offset = 0,
): Promise<KnowledgePage> {
  const res = await unary(
    services.knowledge.knowledgeBaseSearch(
      KnowledgeBaseSearchRequest.create({
        query,
        limit,
        offset,
      }) as KnowledgeBaseSearchRequest,
    ),
  )
  return {
    items: res.knowledgeBases,
    totalCount: Number(res.totalCount),
  }
}

export { KnowledgeBaseType, KnowledgeBase }
