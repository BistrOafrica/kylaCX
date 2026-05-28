import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import {
  readContentText,
  formatSlaWindow,
  secondsUntil,
  relativeTime,
} from "@/features/inbox/utils/format"

describe("readContentText", () => {
  it("extracts text from {text: ...} JSON", () => {
    expect(readContentText('{"text":"hello"}')).toBe("hello")
  })

  it("returns raw string when not JSON", () => {
    expect(readContentText("plain")).toBe("plain")
  })

  it("returns raw string when JSON has no .text key", () => {
    expect(readContentText('{"other":1}')).toBe('{"other":1}')
  })
})

describe("formatSlaWindow", () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date("2026-05-28T12:00:00Z"))
  })
  afterEach(() => vi.useRealTimers())

  it("returns null for missing deadline", () => {
    expect(formatSlaWindow(null)).toBeNull()
    expect(formatSlaWindow(undefined)).toBeNull()
  })

  it("formats sub-hour deadlines as minutes", () => {
    const deadline = new Date("2026-05-28T12:28:00Z").toISOString()
    expect(formatSlaWindow(deadline)).toEqual({ text: "28m", breached: false })
  })

  it("formats hour deadlines as h+m", () => {
    const deadline = new Date("2026-05-28T13:12:00Z").toISOString()
    expect(formatSlaWindow(deadline)).toEqual({ text: "1h 12m", breached: false })
  })

  it("marks past deadlines as breached", () => {
    const deadline = new Date("2026-05-28T11:59:00Z").toISOString()
    expect(formatSlaWindow(deadline)).toEqual({ text: "Breached", breached: true })
  })
})

describe("secondsUntil", () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date("2026-05-28T12:00:00Z"))
  })
  afterEach(() => vi.useRealTimers())

  it("returns positive for future timestamps", () => {
    const future = new Date("2026-05-28T12:00:30Z").toISOString()
    expect(secondsUntil(future)).toBe(30)
  })

  it("returns negative for past timestamps", () => {
    const past = new Date("2026-05-28T11:59:30Z").toISOString()
    expect(secondsUntil(past)).toBe(-30)
  })

  it("returns null for malformed input", () => {
    expect(secondsUntil("not-a-date")).toBeNull()
    expect(secondsUntil(null)).toBeNull()
  })
})

describe("relativeTime", () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date("2026-05-28T12:00:00Z"))
  })
  afterEach(() => vi.useRealTimers())

  it("returns 'now' for very recent timestamps", () => {
    const past = new Date("2026-05-28T11:59:55Z").toISOString()
    expect(relativeTime(past)).toBe("now")
  })

  it("returns minute-based string for minutes-ago timestamps", () => {
    const past = new Date("2026-05-28T11:58:00Z").toISOString()
    expect(relativeTime(past)).toMatch(/^\d+m$/)
  })

  it("returns empty string for missing input", () => {
    expect(relativeTime(null)).toBe("")
    expect(relativeTime("")).toBe("")
  })
})
