import { describe, it, expect, beforeEach } from "vitest"
import {
  orgScope,
  branchScope,
  departmentScope,
  teamScope,
  userScope,
  OwnerType,
} from "@/features/admin/utils/scope"
import { useWorkspaceStore } from "@/lib/workspace"
import { useAuthStore } from "@/lib/auth"

describe("admin scope builders", () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      organisation: { id: "org-1", name: "Acme" },
      workspace: { id: "ws-1", name: "Support" },
    })
    useAuthStore.getState().setSession({
      accessToken: "a",
      refreshToken: "r",
      identity: { userId: "user-1", email: "u@k.io" },
    })
  })

  it("orgScope reads the active organisation", () => {
    const s = orgScope()
    expect(s.ownerType).toBe(OwnerType.ORGANISATIONS)
    expect(s.ownerId).toBe("org-1")
  })

  it("branchScope returns BRANCHES owner type with the given id", () => {
    const s = branchScope("b-1")
    expect(s.ownerType).toBe(OwnerType.BRANCHES)
    expect(s.ownerId).toBe("b-1")
  })

  it("departmentScope returns DEPARTMENTS owner type with the given id", () => {
    const s = departmentScope("d-1")
    expect(s.ownerType).toBe(OwnerType.DEPARTMENTS)
    expect(s.ownerId).toBe("d-1")
  })

  it("teamScope returns TEAMS owner type with the given id", () => {
    const s = teamScope("t-1")
    expect(s.ownerType).toBe(OwnerType.TEAMS)
    expect(s.ownerId).toBe("t-1")
  })

  it("userScope reads the authenticated identity", () => {
    const s = userScope()
    expect(s.ownerType).toBe(OwnerType.USERS)
    expect(s.ownerId).toBe("user-1")
  })

  it("OwnerType enum covers the documented values", () => {
    expect(OwnerType.USERS).toBe(0)
    expect(OwnerType.TEAMS).toBe(1)
    expect(OwnerType.DEPARTMENTS).toBe(2)
    expect(OwnerType.BRANCHES).toBe(3)
    expect(OwnerType.ORGANISATIONS).toBe(4)
  })
})
