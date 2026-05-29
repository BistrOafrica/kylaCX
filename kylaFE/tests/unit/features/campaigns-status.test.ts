import { describe, it, expect } from "vitest"
import {
  bucketCampaignStatus,
  STATUS_LABEL,
  STATUS_TONE,
  parseContacts,
  deliveryRate,
  readRate,
} from "@/features/campaigns/utils/status"

describe("bucketCampaignStatus", () => {
  it("maps assorted backend strings into stable buckets", () => {
    expect(bucketCampaignStatus("draft")).toBe("draft")
    expect(bucketCampaignStatus("DRAFT")).toBe("draft")
    expect(bucketCampaignStatus("scheduled")).toBe("scheduled")
    expect(bucketCampaignStatus("Scheduling")).toBe("scheduled")
    expect(bucketCampaignStatus("running")).toBe("running")
    expect(bucketCampaignStatus("active")).toBe("running")
    expect(bucketCampaignStatus("completed")).toBe("completed")
    expect(bucketCampaignStatus("done")).toBe("completed")
    expect(bucketCampaignStatus("paused")).toBe("paused")
    expect(bucketCampaignStatus("failed")).toBe("failed")
    expect(bucketCampaignStatus("error: rate-limited")).toBe("failed")
  })

  it("defaults to draft for unknown values", () => {
    expect(bucketCampaignStatus("xyz")).toBe("draft")
    expect(bucketCampaignStatus("")).toBe("draft")
  })

  it("every bucket has a label and tone", () => {
    for (const bucket of ["draft", "scheduled", "running", "completed", "paused", "failed"] as const) {
      expect(STATUS_LABEL[bucket]).toBeTruthy()
      expect(STATUS_TONE[bucket]).toBeTruthy()
    }
  })
})

describe("parseContacts", () => {
  it("returns [] for empty input", () => {
    expect(parseContacts("")).toEqual([])
    expect(parseContacts("   ")).toEqual([])
  })

  it("parses CSV / line-separated lists", () => {
    expect(parseContacts("+15550101,+15550102")).toEqual([
      "+15550101",
      "+15550102",
    ])
    expect(parseContacts("+15550101\n+15550102")).toEqual([
      "+15550101",
      "+15550102",
    ])
    expect(parseContacts("+15550101; +15550102 ;+15550103")).toEqual([
      "+15550101",
      "+15550102",
      "+15550103",
    ])
  })

  it("parses JSON-encoded arrays", () => {
    expect(parseContacts('["+15550101","+15550102"]')).toEqual([
      "+15550101",
      "+15550102",
    ])
  })

  it("falls back to delimiter split on malformed JSON", () => {
    expect(parseContacts("[bad,+15550101]")).toEqual(["[bad", "+15550101]"])
  })
})

describe("deliveryRate + readRate", () => {
  it("returns 0 when sent is zero", () => {
    expect(deliveryRate(0n, 0n)).toBe(0)
    expect(readRate(0n, 0n)).toBe(0)
  })

  it("computes integer-percent with one decimal", () => {
    expect(deliveryRate(100n, 95n)).toBe(95)
    expect(deliveryRate(1000, 850)).toBe(85)
    expect(readRate(200n, 50n)).toBe(25)
  })

  it("rounds to 1 decimal", () => {
    expect(deliveryRate(7, 5)).toBe(71.4)
  })
})
