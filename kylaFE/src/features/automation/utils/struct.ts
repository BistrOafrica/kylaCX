import { Struct } from "@/pb/google/protobuf/struct"

/**
 * google.protobuf.Struct ⇄ plain JS object.
 *
 * The Workflow proto stores trigger / conditions / actions as Struct
 * fields. Components should never touch Struct directly — they read
 * and write plain JS objects via these helpers.
 *
 * `Struct.fromJson` / `toJson` from the protobuf-ts runtime handles
 * the heavy lifting; we just type the result so the rest of the code
 * has a clean Record<string, unknown> surface.
 */

export function fromStruct(s: Struct | undefined | null): Record<string, unknown> {
  if (!s) return {}
  return Struct.toJson(s) as Record<string, unknown>
}

export function toStruct(obj: Record<string, unknown>): Struct {
  return Struct.fromJson(obj as never) as Struct
}

export function fromStructArray(items: Struct[] | undefined): Record<string, unknown>[] {
  if (!items) return []
  return items.map((s) => fromStruct(s))
}

export function toStructArray(items: Record<string, unknown>[]): Struct[] {
  return items.map((o) => toStruct(o))
}
