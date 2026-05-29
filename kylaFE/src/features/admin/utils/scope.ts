import { Scope, OwnerType } from "@/pb/owner_type"
import { useWorkspaceStore } from "@/lib/workspace"
import { useAuthStore } from "@/lib/auth"

/**
 * Admin-side Scope helpers.
 *
 * Almost every admin RPC takes a `Scope` (ownerType + ownerId).
 * The workspace + auth stores carry everything we need; these helpers
 * keep the per-call boilerplate to one line.
 */

export function orgScope(): Scope {
  const orgId = useWorkspaceStore.getState().organisation?.id ?? ""
  return Scope.create({
    ownerType: OwnerType.ORGANISATIONS,
    ownerId: orgId,
  }) as Scope
}

export function branchScope(branchId: string): Scope {
  return Scope.create({
    ownerType: OwnerType.BRANCHES,
    ownerId: branchId,
  }) as Scope
}

export function departmentScope(deptId: string): Scope {
  return Scope.create({
    ownerType: OwnerType.DEPARTMENTS,
    ownerId: deptId,
  }) as Scope
}

export function teamScope(teamId: string): Scope {
  return Scope.create({
    ownerType: OwnerType.TEAMS,
    ownerId: teamId,
  }) as Scope
}

export function userScope(): Scope {
  const userId = useAuthStore.getState().identity?.userId ?? ""
  return Scope.create({
    ownerType: OwnerType.USERS,
    ownerId: userId,
  }) as Scope
}

export { OwnerType }
