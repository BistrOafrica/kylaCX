import { describe, it, expect } from "vitest"
import {
  parseFields,
  serializeFields,
  type FormFieldSpec,
} from "@/features/service-desk/api/forms"

describe("forms field schema", () => {
  it("parseFields returns [] for empty input", () => {
    expect(parseFields("")).toEqual([])
  })

  it("parseFields returns [] for malformed JSON", () => {
    expect(parseFields("not-json")).toEqual([])
  })

  it("parseFields returns [] when JSON isn't an array", () => {
    expect(parseFields('{"k":"v"}')).toEqual([])
  })

  it("round-trips a typed schema through serialize → parse", () => {
    const fields: FormFieldSpec[] = [
      { key: "name",  label: "Full name", type: "text",   required: true },
      { key: "email", label: "Email",     type: "email",  required: true },
      {
        key: "tier",
        label: "Tier",
        type: "select",
        options: [
          { value: "free", label: "Free" },
          { value: "pro",  label: "Pro" },
        ],
      },
    ]
    const round = parseFields(serializeFields(fields))
    expect(round).toEqual(fields)
  })
})
