import { describe, it, expect } from "vitest"
import {
  formatDuration,
  DIRECTION_LABEL,
  STATUS_LABEL,
  statusTone,
} from "@/features/telephony/utils/formatters"
import { CallDirection, CallStatus } from "@/pb/call_session"

describe("formatDuration", () => {
  it("returns '0s' for zero / negative input", () => {
    expect(formatDuration(0)).toBe("0s")
    expect(formatDuration(-1)).toBe("0s")
  })

  it("renders sub-minute durations in seconds", () => {
    expect(formatDuration(45)).toBe("45s")
  })

  it("renders sub-hour durations as Nm Ns", () => {
    expect(formatDuration(72)).toBe("1m 12s")
    expect(formatDuration(120)).toBe("2m")
  })

  it("renders multi-hour durations as Nh Nm", () => {
    expect(formatDuration(3_660)).toBe("1h 1m")
  })

  it("accepts bigint duration values from the proto", () => {
    expect(formatDuration(BigInt(125))).toBe("2m 5s")
  })
})

describe("call direction + status labels", () => {
  it("covers every CallDirection enum value", () => {
    expect(DIRECTION_LABEL[CallDirection.INBOUND]).toBe("Inbound")
    expect(DIRECTION_LABEL[CallDirection.OUTBOUND]).toBe("Outbound")
    expect(DIRECTION_LABEL[CallDirection.LOCAL]).toBe("Local")
  })

  it("covers every CallStatus enum value", () => {
    expect(STATUS_LABEL[CallStatus.ONGOING]).toBe("Ongoing")
    expect(STATUS_LABEL[CallStatus.COMPLETED]).toBe("Completed")
    expect(STATUS_LABEL[CallStatus.DROPPED]).toBe("Dropped")
    expect(STATUS_LABEL[CallStatus.REJECTED]).toBe("Rejected")
    expect(STATUS_LABEL[CallStatus.ON_HOLD]).toBe("On hold")
  })

  it("statusTone maps to semantic tones", () => {
    expect(statusTone(CallStatus.COMPLETED)).toBe("success")
    expect(statusTone(CallStatus.ONGOING)).toBe("info")
    expect(statusTone(CallStatus.ON_HOLD)).toBe("warn")
    expect(statusTone(CallStatus.DROPPED)).toBe("danger")
    expect(statusTone(CallStatus.REJECTED)).toBe("danger")
  })
})
