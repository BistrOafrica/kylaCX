import { describe, it, expect } from "vitest"
import {
  channelMeta,
  statusMeta,
  priorityMeta,
  CHANNEL_META,
  STATUS_META,
  PRIORITY_META,
  Channel,
  ConversationStatus,
  ConversationPriority,
} from "@/features/inbox/utils/enums"

describe("channelMeta", () => {
  it("returns design-system token for every proto channel", () => {
    for (const m of CHANNEL_META) {
      expect(channelMeta(m.protoEnum).token).toBe(m.token)
    }
  })

  it("falls back to the first channel for unknown values", () => {
    expect(channelMeta(Channel.UNSPECIFIED)).toBe(CHANNEL_META[0])
    expect(channelMeta(999 as Channel)).toBe(CHANNEL_META[0])
  })
})

describe("statusMeta", () => {
  it("returns a tone for every defined status", () => {
    for (const m of STATUS_META) {
      const got = statusMeta(m.protoEnum)
      expect(got.tone).toBe(m.tone)
      expect(got.label).toBe(m.label)
    }
  })

  it("falls back to first status for unknown values", () => {
    expect(statusMeta(ConversationStatus.UNSPECIFIED)).toBe(STATUS_META[0])
  })
})

describe("priorityMeta", () => {
  it("urgent > high > normal > low rank order", () => {
    const urgent = priorityMeta(ConversationPriority.URGENT)
    const high   = priorityMeta(ConversationPriority.HIGH)
    const normal = priorityMeta(ConversationPriority.NORMAL)
    const low    = priorityMeta(ConversationPriority.LOW)
    expect(urgent.rank).toBeGreaterThan(high.rank)
    expect(high.rank).toBeGreaterThan(normal.rank)
    expect(normal.rank).toBeGreaterThan(low.rank)
  })

  it("uses semantic tones (muted/info/warn/danger)", () => {
    for (const m of PRIORITY_META) {
      expect(["muted", "info", "warn", "danger"]).toContain(m.tone)
    }
  })

  it("defaults to NORMAL for unspecified", () => {
    expect(priorityMeta(ConversationPriority.UNSPECIFIED).protoEnum).toBe(
      ConversationPriority.NORMAL,
    )
  })
})
