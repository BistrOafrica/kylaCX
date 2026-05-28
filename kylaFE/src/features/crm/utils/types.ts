import type {
  Object as CoreObject,
  ObjectType,
  FieldDefinition,
  ObjectEvent,
  ObjectRelation,
} from "@/pb/object_core"

/**
 * Domain-friendly aliases over the generated proto types.
 *
 * The proto's `Object` collides with the global JavaScript `Object`,
 * so the feature module talks in `CoreObject` / `CoreObjectType`
 * throughout. Field-definition + event + relation are re-exported
 * directly for component prop typing.
 */

export type { CoreObject, ObjectType, FieldDefinition, ObjectEvent, ObjectRelation }

/**
 * Canonical type slugs for first-class CRM entities. These are seeded
 * on every workspace by the backend, so the UI can hard-code them.
 */
export const TYPE_SLUG = {
  contact: "contact",
  company: "company",
  lead:    "lead",
  deal:    "deal",
  ticket:  "ticket",
  task:    "task",
} as const

export type TypeSlug = (typeof TYPE_SLUG)[keyof typeof TYPE_SLUG]

/**
 * Parsed view of an Object — JSON `data` blob deserialized + merged
 * with system fields. Components consume `ObjectRecord`, never the raw
 * `Object` shape.
 */
export interface ObjectRecord {
  id: string
  orgId: string
  workspaceId: string
  typeSlug: string
  createdBy: string
  createdAt: string
  updatedAt: string
  /** Custom field values keyed by `FieldDefinition.key`. */
  data: Record<string, unknown>
  /** Source object for fields we don't surface explicitly. */
  raw: CoreObject
}

export function toRecord(obj: CoreObject): ObjectRecord {
  let parsed: Record<string, unknown> = {}
  try {
    parsed = JSON.parse(obj.data || "{}") as Record<string, unknown>
  } catch {
    parsed = {}
  }
  return {
    id: obj.id,
    orgId: obj.orgId,
    workspaceId: obj.workspaceId,
    typeSlug: obj.typeSlug,
    createdBy: obj.createdBy,
    createdAt: obj.createdAt,
    updatedAt: obj.updatedAt,
    data: parsed,
    raw: obj,
  }
}
