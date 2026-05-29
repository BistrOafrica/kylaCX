import { services } from "@/lib/rpc/services"
import type { IVRFlow, IVRDefinition, IVRNode } from "@/pb/ivr"
import { Struct } from "@/pb/google/protobuf/struct"

/**
 * Helpers for the Phase 5c IVRService (the new unified surface).
 *
 * Lives alongside the legacy api/ivr.ts so existing call_ivr_flow callers
 * keep working. New components should import from here.
 */

export async function listIvrV2Flows(workspaceId: string, activeOnly = false): Promise<IVRFlow[]> {
  const res = await services.ivr.listIVRFlows({ workspaceId, activeOnly })
  return res.response.flows
}

export async function getIvrV2Flow(id: string): Promise<IVRFlow> {
  const res = await services.ivr.getIVRFlow({ id })
  return res.response
}

export async function createIvrV2Flow(flow: Partial<IVRFlow>): Promise<IVRFlow> {
  const res = await services.ivr.createIVRFlow({
    flow: emptyIvrFlow(flow),
  })
  return res.response
}

export async function updateIvrV2Flow(flow: IVRFlow): Promise<IVRFlow> {
  const res = await services.ivr.updateIVRFlow({ flow })
  return res.response
}

export async function deleteIvrV2Flow(id: string): Promise<void> {
  await services.ivr.deleteIVRFlow({ id })
}

/**
 * Dry-run validation of a flow definition. Sends the inline definition
 * rather than the saved id so the builder can validate unsaved edits.
 */
export async function testRunIvrV2Flow(flow: IVRFlow) {
  const res = await services.ivr.testRunIVRFlow({
    id: "",
    definition: flow.definition,
  })
  return res.response
}

/**
 * Builds a fresh IVRFlow message with sane defaults so consumers don't have
 * to fill in every required field of the protobuf-ts type. The pb-ts types
 * require every field (including timestamps); we set them to empty values
 * the backend ignores at write time.
 */
export function emptyIvrFlow(seed: Partial<IVRFlow> = {}): IVRFlow {
  return {
    id: "",
    orgId: "",
    workspaceId: "",
    name: "",
    description: "",
    isActive: false,
    version: 0,
    createdBy: "",
    definition: emptyDefinition(),
    ...seed,
  } as IVRFlow
}

export function emptyDefinition(seed: Partial<IVRDefinition> = {}): IVRDefinition {
  return {
    startNodeId: "",
    nodes: [],
    ...seed,
  } as IVRDefinition
}

/**
 * nodeConfig pulls a typed JS object out of an IVRNode's google.protobuf.Struct
 * config. The protobuf-ts type for Struct decodes to { fields: { [k]: Value } }
 * so we collapse it to a plain map for editing.
 */
export function nodeConfig(node: IVRNode): Record<string, unknown> {
  if (!node.config) return {}
  const json = Struct.toJson(node.config)
  return (json ?? {}) as Record<string, unknown>
}

/**
 * setNodeConfig encodes a plain JS object back into the Struct shape.
 * Numbers, strings, bools, and nested objects/arrays all round-trip cleanly
 * via Struct's JSON encoder.
 */
export function setNodeConfig(node: IVRNode, value: Record<string, unknown>): IVRNode {
  // Struct.fromJson accepts a JsonValue; cast through unknown to satisfy
  // strict typing without leaking implementation detail.
  node.config = Struct.fromJson(value as unknown as never)
  return node
}
