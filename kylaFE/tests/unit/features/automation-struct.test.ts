import { describe, it, expect } from "vitest"
import {
  fromStruct,
  toStruct,
  fromStructArray,
  toStructArray,
} from "@/features/automation/utils/struct"

describe("Struct ⇄ plain object round-trip", () => {
  it("preserves primitives and nesting", () => {
    const obj = {
      type: "object.created",
      filter: { object_type: "ticket", priority: "high" },
      labels: ["billing", "urgent"],
      retry: 3,
      enabled: true,
    }
    const round = fromStruct(toStruct(obj))
    expect(round).toEqual(obj)
  })

  it("empty object survives the round-trip", () => {
    expect(fromStruct(toStruct({}))).toEqual({})
  })

  it("fromStruct handles undefined", () => {
    expect(fromStruct(undefined)).toEqual({})
  })

  it("array helpers preserve order + content", () => {
    const items = [
      { type: "delay", params: { duration: 60 } },
      { type: "send_message", params: { channel: "whatsapp" } },
    ]
    const round = fromStructArray(toStructArray(items))
    expect(round).toEqual(items)
  })
})
