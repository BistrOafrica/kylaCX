import { describe, it, expect } from "vitest"
import { FieldType, type FieldDefinition } from "@/pb/object_core"
import {
  formatFieldValue,
  defaultColumns,
  buildFilterJson,
  FIELD_TYPE_NAME,
} from "@/features/crm/utils/fields"

function field(partial: Partial<FieldDefinition>): FieldDefinition {
  return {
    key: "f",
    label: "Field",
    type: FieldType.TEXT,
    required: false,
    unique: false,
    searchable: false,
    options: [],
    relatesTo: "",
    ...partial,
  }
}

describe("formatFieldValue", () => {
  it("renders em-dash for empty/null", () => {
    expect(formatFieldValue(field({}), "")).toBe("—")
    expect(formatFieldValue(field({}), null)).toBe("—")
    expect(formatFieldValue(field({}), undefined)).toBe("—")
  })

  it("formats numbers with locale separators", () => {
    expect(formatFieldValue(field({ type: FieldType.NUMBER }), 1500000)).toMatch(/1,500,000|1\.500\.000/)
  })

  it("formats booleans as Yes/No", () => {
    expect(formatFieldValue(field({ type: FieldType.BOOLEAN }), true)).toBe("Yes")
    expect(formatFieldValue(field({ type: FieldType.BOOLEAN }), false)).toBe("No")
  })

  it("maps SELECT to option label when defined", () => {
    const f = field({
      type: FieldType.SELECT,
      options: [{ value: "new", label: "New customer", color: "#0EA5E9" }],
    })
    expect(formatFieldValue(f, "new")).toBe("New customer")
  })

  it("joins MULTI labels with commas", () => {
    const f = field({
      type: FieldType.MULTI,
      options: [
        { value: "a", label: "Alpha", color: "" },
        { value: "b", label: "Beta", color: "" },
      ],
    })
    expect(formatFieldValue(f, ["a", "b"])).toBe("Alpha, Beta")
  })

  it("falls back to raw string for unknown SELECT values", () => {
    const f = field({ type: FieldType.SELECT, options: [] })
    expect(formatFieldValue(f, "unknown")).toBe("unknown")
  })
})

describe("defaultColumns", () => {
  it("returns up to 5 non-system fields by default", () => {
    const fields = [
      field({ key: "a", label: "A" }),
      field({ key: "b", label: "B" }),
      field({ key: "c", label: "C" }),
      field({ key: "d", label: "D" }),
      field({ key: "e", label: "E" }),
      field({ key: "f", label: "F" }),
      field({ key: "_internal", label: "System" }),
    ]
    expect(defaultColumns(fields).map((f) => f.key)).toEqual([
      "a", "b", "c", "d", "e",
    ])
  })
})

describe("buildFilterJson", () => {
  it("returns empty string when no filters", () => {
    expect(buildFilterJson({})).toBe("")
    expect(buildFilterJson({ status: "" })).toBe("")
  })

  it("serializes truthy values", () => {
    expect(buildFilterJson({ status: "open", stage_id: "abc" }))
      .toBe('{"status":"open","stage_id":"abc"}')
  })
})

describe("FIELD_TYPE_NAME", () => {
  it("covers every proto FieldType", () => {
    expect(FIELD_TYPE_NAME[FieldType.TEXT]).toBe("text")
    expect(FIELD_TYPE_NAME[FieldType.CURRENCY]).toBe("currency")
    expect(FIELD_TYPE_NAME[FieldType.RELATION]).toBe("relation")
  })
})
