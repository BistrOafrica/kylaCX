import { describe, it, expect } from "vitest"
import { RpcError } from "@/lib/rpc/errors"

describe("RpcError", () => {
  it("normalizes numeric gRPC codes to enum names", () => {
    const err = RpcError.from({ code: 16, message: "no" })
    expect(err.code).toBe("UNAUTHENTICATED")
    expect(err.isUnauthenticated).toBe(true)
  })

  it("normalizes string codes when already enum-named", () => {
    const err = RpcError.from({ code: "INVALID_ARGUMENT", message: "x" })
    expect(err.code).toBe("INVALID_ARGUMENT")
    expect(err.isUserFacing).toBe(true)
  })

  it("UNKNOWN for unmapped numeric codes", () => {
    const err = RpcError.from({ code: 99 })
    expect(err.code).toBe("UNKNOWN")
  })

  it("returns the original RpcError unchanged when fed one", () => {
    const original = new RpcError("boom", "INTERNAL")
    expect(RpcError.from(original)).toBe(original)
  })
})
