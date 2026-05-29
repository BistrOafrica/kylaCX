import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import {
  PRESETS,
  resolveRange,
  granularityFor,
  type TimeRangePreset,
} from "@/features/analytics/utils/time-range"

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date("2026-05-28T14:00:00Z"))
})
afterEach(() => vi.useRealTimers())

const day = 24 * 60 * 60 * 1000

function span(preset: TimeRangePreset): number {
  const r = resolveRange(preset)
  return new Date(r.endDate).getTime() - new Date(r.startDate).getTime()
}

describe("time-range presets", () => {
  it("exposes the documented preset list", () => {
    expect(PRESETS.map((p) => p.id)).toEqual([
      "today",
      "yesterday",
      "last_7d",
      "last_30d",
      "last_90d",
      "this_quarter",
      "ytd",
    ])
  })
})

describe("resolveRange spans", () => {
  it("today spans less than a full day", () => {
    expect(span("today")).toBeLessThan(day)
    expect(span("today")).toBeGreaterThan(0)
  })

  it("yesterday spans roughly one day", () => {
    expect(span("yesterday")).toBeGreaterThan(0)
    expect(span("yesterday")).toBeLessThan(2 * day)
  })

  it("last_7d spans about a week (allow TZ slack)", () => {
    const s = span("last_7d")
    expect(s).toBeGreaterThan(6 * day)
    expect(s).toBeLessThan(8 * day)
  })

  it("last_30d spans about a month (allow TZ slack)", () => {
    const s = span("last_30d")
    expect(s).toBeGreaterThan(29 * day)
    expect(s).toBeLessThan(31 * day)
  })

  it("ytd starts before this_quarter", () => {
    const ytd = new Date(resolveRange("ytd").startDate).getTime()
    const q = new Date(resolveRange("this_quarter").startDate).getTime()
    expect(ytd).toBeLessThanOrEqual(q)
  })

  it("every preset returns an end > start", () => {
    for (const p of PRESETS) {
      const r = resolveRange(p.id as TimeRangePreset)
      expect(new Date(r.endDate).getTime()).toBeGreaterThan(
        new Date(r.startDate).getTime(),
      )
    }
  })
})

describe("granularityFor", () => {
  it("returns DAILY for short windows", () => {
    expect(granularityFor("today")).toBe("DAILY")
    expect(granularityFor("yesterday")).toBe("DAILY")
    expect(granularityFor("last_7d")).toBe("DAILY")
  })

  it("returns WEEKLY for medium windows", () => {
    expect(granularityFor("last_30d")).toBe("WEEKLY")
    expect(granularityFor("last_90d")).toBe("WEEKLY")
    expect(granularityFor("this_quarter")).toBe("WEEKLY")
  })

  it("returns MONTHLY for year-scale windows", () => {
    expect(granularityFor("ytd")).toBe("MONTHLY")
  })

  it("covers every preset", () => {
    for (const p of PRESETS) {
      expect(granularityFor(p.id as TimeRangePreset)).toMatch(
        /DAILY|WEEKLY|MONTHLY/,
      )
    }
  })
})
