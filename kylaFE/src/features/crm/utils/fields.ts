import { FieldType, type FieldDefinition } from "@/pb/object_core"

/**
 * Field-type formatting + parsing helpers used by the schema-aware
 * list and detail views.
 *
 * The proto's FieldType enum has 14 values; we surface them through
 * stable string keys so component switches stay readable.
 */

export const FIELD_TYPE_NAME: Record<FieldType, string> = {
  [FieldType.TEXT]:     "text",
  [FieldType.NUMBER]:   "number",
  [FieldType.DATE]:     "date",
  [FieldType.DATETIME]: "datetime",
  [FieldType.SELECT]:   "select",
  [FieldType.MULTI]:    "multi",
  [FieldType.BOOLEAN]:  "boolean",
  [FieldType.USER]:     "user",
  [FieldType.RELATION]: "relation",
  [FieldType.FILE]:     "file",
  [FieldType.EMAIL]:    "email",
  [FieldType.PHONE]:    "phone",
  [FieldType.URL]:      "url",
  [FieldType.CURRENCY]: "currency",
}

/**
 * Render a field value as a human-readable string for list / table
 * cells. Components that need richer rendering (color chips, links,
 * relation popovers) should switch on `field.type` directly and
 * compose their own JSX.
 */
export function formatFieldValue(
  field: FieldDefinition,
  value: unknown,
): string {
  if (value === undefined || value === null || value === "") return "—"

  switch (field.type) {
    case FieldType.NUMBER:
      return Number(value).toLocaleString()

    case FieldType.CURRENCY: {
      const n = Number(value)
      if (Number.isNaN(n)) return String(value)
      return new Intl.NumberFormat(undefined, {
        style: "currency",
        currency: "USD",
        maximumFractionDigits: 0,
      }).format(n)
    }

    case FieldType.DATE:
    case FieldType.DATETIME: {
      const d = new Date(String(value))
      if (Number.isNaN(d.getTime())) return String(value)
      return field.type === FieldType.DATETIME
        ? d.toLocaleString()
        : d.toLocaleDateString()
    }

    case FieldType.BOOLEAN:
      return value ? "Yes" : "No"

    case FieldType.SELECT: {
      const opt = field.options?.find((o) => o.value === value)
      return opt?.label ?? String(value)
    }

    case FieldType.MULTI: {
      if (!Array.isArray(value)) return String(value)
      return value
        .map((v) => field.options?.find((o) => o.value === v)?.label ?? String(v))
        .join(", ")
    }

    case FieldType.USER:
    case FieldType.RELATION:
      return String(value)

    default:
      return String(value)
  }
}

/**
 * Pick the best handful of fields to show as columns in a list view.
 * Used when no SavedView is selected. Prefers searchable + first
 * non-system fields up to `limit`.
 */
export function defaultColumns(
  fields: FieldDefinition[],
  limit = 5,
): FieldDefinition[] {
  return fields
    .filter((f) => !f.key.startsWith("_"))
    .slice(0, limit)
}

/**
 * Build a JSON filter blob from a friendly key/value map. Backend
 * expects `{"key":"value"}` for equality filters; we use this from
 * the CRM list view's pill selectors.
 */
export function buildFilterJson(values: Record<string, unknown>): string {
  const entries = Object.entries(values).filter(
    ([, v]) => v !== undefined && v !== "" && v !== null,
  )
  if (!entries.length) return ""
  return JSON.stringify(Object.fromEntries(entries))
}
