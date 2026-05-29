import type { Workflow as ProtoWorkflow, WorkflowRun as ProtoWorkflowRun } from "@/pb/workflow"
import type { ActionType, TriggerType, Condition } from "./actions"
import { fromStruct, fromStructArray, toStruct, toStructArray } from "./struct"

/**
 * Domain view of a Workflow: trigger / conditions / actions decoded
 * into typed plain objects the canvas can render.
 *
 * Components consume `WorkflowSpec`; the proto's `Workflow` is only
 * touched by the API layer.
 */

export interface TriggerSpecValue {
  type: TriggerType
  /** Optional event-pattern filters (e.g. `{ object_type: "ticket" }`). */
  config?: Record<string, unknown>
}

export interface ActionNode {
  /** Stable per-node id used by React Flow + serialization. */
  id: string
  type: ActionType
  params: Record<string, unknown>
}

export interface WorkflowSpec {
  id: string
  orgId: string
  workspaceId: string
  name: string
  description: string
  status: "draft" | "active" | "inactive"
  trigger: TriggerSpecValue | null
  conditions: Condition[]
  actions: ActionNode[]
  createdAt?: string
  updatedAt?: string
  createdBy: string
}

export function toWorkflowSpec(w: ProtoWorkflow): WorkflowSpec {
  const triggerRaw = fromStruct(w.trigger)
  const actionsRaw = fromStructArray(w.actions)
  const conditionsRaw = fromStructArray(w.conditions)

  return {
    id: w.id,
    orgId: w.orgId,
    workspaceId: w.workspaceId,
    name: w.name,
    description: w.description,
    status: (w.status as "draft" | "active" | "inactive") || "draft",
    trigger: triggerRaw.type
      ? { type: triggerRaw.type as TriggerType, config: omitKey(triggerRaw, "type") }
      : null,
    conditions: conditionsRaw.map((c) => ({
      field: String(c.field ?? ""),
      op: (c.op ?? "eq") as Condition["op"],
      value: c.value,
    })),
    actions: actionsRaw.map((a, i) => ({
      id: String(a.id ?? `node-${i}`),
      type: (a.type ?? "delay") as ActionType,
      params: (a.params as Record<string, unknown>) ?? {},
    })),
    createdAt: w.createdAt ? isoFromTimestamp(w.createdAt as { seconds: bigint | string; nanos: number }) : undefined,
    updatedAt: w.updatedAt ? isoFromTimestamp(w.updatedAt as { seconds: bigint | string; nanos: number }) : undefined,
    createdBy: w.createdBy,
  }
}

export function toWorkflowJson(spec: Partial<WorkflowSpec>): {
  trigger: ReturnType<typeof toStruct> | undefined
  conditions: ReturnType<typeof toStructArray>
  actions: ReturnType<typeof toStructArray>
} {
  const trigger = spec.trigger
    ? toStruct({
        type: spec.trigger.type,
        ...(spec.trigger.config ?? {}),
      })
    : undefined

  const conditions = toStructArray(
    (spec.conditions ?? []).map((c) => ({
      field: c.field,
      op: c.op,
      value: c.value as never,
    })),
  )

  const actions = toStructArray(
    (spec.actions ?? []).map((a) => ({
      id: a.id,
      type: a.type,
      params: a.params as never,
    })),
  )

  return { trigger, conditions, actions }
}

export interface WorkflowRunRow {
  id: string
  workflowId: string
  temporalRunId: string
  triggerEventId: string
  status: "pending" | "running" | "success" | "failed" | string
  context: Record<string, unknown>
  error: string
  startedAt?: string
  finishedAt?: string
}

export function toWorkflowRun(r: ProtoWorkflowRun): WorkflowRunRow {
  return {
    id: r.id,
    workflowId: r.workflowId,
    temporalRunId: r.temporalRunId,
    triggerEventId: r.triggerEventId,
    status: r.status,
    context: fromStruct(r.context),
    error: r.error,
    startedAt: r.startedAt
      ? isoFromTimestamp(r.startedAt as { seconds: bigint | string; nanos: number })
      : undefined,
    finishedAt: r.finishedAt
      ? isoFromTimestamp(r.finishedAt as { seconds: bigint | string; nanos: number })
      : undefined,
  }
}

function isoFromTimestamp(ts: { seconds: bigint | string; nanos: number }): string {
  const s = typeof ts.seconds === "bigint" ? Number(ts.seconds) : Number(ts.seconds)
  if (!Number.isFinite(s)) return ""
  return new Date(s * 1000).toISOString()
}

function omitKey<T extends Record<string, unknown>>(o: T, key: string): Record<string, unknown> {
  const copy: Record<string, unknown> = {}
  for (const k of Object.keys(o)) {
    if (k !== key) copy[k] = o[k]
  }
  return copy
}
