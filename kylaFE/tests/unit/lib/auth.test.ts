import { describe, it, expect, beforeEach } from "vitest"
import { useAuthStore } from "@/lib/auth/store"
import { readJwtExpiry } from "@/lib/auth/storage"

/**
 * Auth store + token-decode smoke tests.
 *
 * These don't hit the real backend — they verify the store transitions
 * and the JWT exp-claim reader. End-to-end login is covered by the
 * Playwright spec.
 */

beforeEach(() => {
  useAuthStore.getState().signOut()
})

describe("auth store", () => {
  it("starts in idle state and transitions to unauthenticated on hydrate without persisted session", () => {
    useAuthStore.getState().hydrate()
    expect(useAuthStore.getState().status).toBe("unauthenticated")
  })

  it("setSession marks authenticated and stores tokens", () => {
    useAuthStore.getState().setSession({
      accessToken: "a.b.c",
      refreshToken: "r.t",
      identity: { userId: "u-1", email: "u@k.io" },
    })
    const state = useAuthStore.getState()
    expect(state.status).toBe("authenticated")
    expect(state.accessToken).toBe("a.b.c")
    expect(state.identity?.email).toBe("u@k.io")
  })

  it("signOut clears identity + tokens", () => {
    useAuthStore.getState().setSession({
      accessToken: "x",
      refreshToken: "y",
      identity: { userId: "u-1" },
    })
    useAuthStore.getState().signOut()
    const state = useAuthStore.getState()
    expect(state.status).toBe("unauthenticated")
    expect(state.accessToken).toBeNull()
    expect(state.identity).toBeNull()
  })

  it("beginMfaChallenge moves to mfa_required and stashes pending userId", () => {
    useAuthStore.getState().beginMfaChallenge("u-42")
    const state = useAuthStore.getState()
    expect(state.status).toBe("mfa_required")
    expect(state.pendingMfaUserId).toBe("u-42")
  })
})

describe("readJwtExpiry", () => {
  it("returns undefined for malformed tokens", () => {
    expect(readJwtExpiry("not-a-jwt")).toBeUndefined()
  })

  it("decodes the exp claim and converts to ms", () => {
    const payload = { exp: 1_700_000_000 }
    const b64 = btoa(JSON.stringify(payload)).replace(/=+$/, "")
    const token = `header.${b64}.sig`
    expect(readJwtExpiry(token)).toBe(1_700_000_000 * 1000)
  })
})
