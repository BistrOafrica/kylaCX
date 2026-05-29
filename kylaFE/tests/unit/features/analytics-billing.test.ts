import { describe, it, expect } from "vitest"
import { formatMoney } from "@/features/analytics/api/billing"

describe("formatMoney", () => {
  it("converts bigint minor-units into a currency string", () => {
    const out = formatMoney(150_000n, "USD")
    // $1,500 — exact string varies by locale but should contain 1,500
    expect(out).toMatch(/1[,.]500/)
  })

  it("accepts plain numbers too", () => {
    expect(formatMoney(2500, "USD")).toMatch(/25/)
  })

  it("uses provided currency code", () => {
    const out = formatMoney(500_00n, "EUR")
    expect(out).toMatch(/€|EUR/)
  })

  it("zero balance renders as a zero amount", () => {
    expect(formatMoney(0n, "USD")).toMatch(/0/)
  })
})
