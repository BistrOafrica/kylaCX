import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import { relativeShort } from "@/features/crm/utils/relative"

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date("2026-05-28T12:00:00Z"))
})
afterEach(() => vi.useRealTimers())

describe("relativeShort", () => {
  it("returns 'now' for <60s", () => {
    expect(relativeShort("2026-05-28T11:59:50Z")).toBe("now")
  })

  it("returns minutes for sub-hour timestamps", () => {
    expect(relativeShort("2026-05-28T11:30:00Z")).toBe("30m")
  })

  it("returns hours for sub-day timestamps", () => {
    expect(relativeShort("2026-05-28T05:00:00Z")).toBe("7h")
  })

  it("returns days for sub-month timestamps", () => {
    expect(relativeShort("2026-05-20T12:00:00Z")).toBe("8d")
  })

  it("returns months for sub-year timestamps", () => {
    expect(relativeShort("2026-01-01T12:00:00Z")).toBe("4mo")
  })

  it("returns years for very old timestamps", () => {
    expect(relativeShort("2023-05-28T12:00:00Z")).toBe("3y")
  })

  it("returns empty string for missing input", () => {
    expect(relativeShort("")).toBe("")
    expect(relativeShort("not-a-date")).toBe("")
  })
})
