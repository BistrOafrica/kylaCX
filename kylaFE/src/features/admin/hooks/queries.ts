import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import { useWorkspaceStore } from "@/lib/workspace"
import * as orgApi from "../api/organisation"
import * as userApi from "../api/users"
import * as rbacApi from "../api/rbac"
import * as inviteApi from "../api/invitations"
import * as appsApi from "../api/apps"
import * as agentApi from "../api/agent-status"

const orgId = () => useWorkspaceStore.getState().organisation?.id ?? ""

// ── Org tree ─────────────────────────────────────────────────────────────────

export function useBranches() {
  return useQuery({
    queryKey: ["admin", "branches", orgId()],
    enabled: Boolean(orgId()),
    queryFn: orgApi.listBranches,
  })
}

export function useDepartments() {
  return useQuery({
    queryKey: ["admin", "departments", orgId()],
    enabled: Boolean(orgId()),
    queryFn: orgApi.listDepartments,
  })
}

export function useBranchDepartments(branchId: string | null) {
  return useQuery({
    queryKey: branchId ? ["admin", "branch", branchId, "departments"] : ["admin", "branch-departments", "none"],
    enabled: Boolean(branchId),
    queryFn: () => orgApi.listBranchDepartments(branchId!),
  })
}

export function useTeams() {
  return useQuery({
    queryKey: ["admin", "teams", orgId()],
    enabled: Boolean(orgId()),
    queryFn: () => orgApi.listTeams(),
  })
}

function adminInvalidator(qc: ReturnType<typeof useQueryClient>) {
  return () => void qc.invalidateQueries({ queryKey: ["admin"] })
}

export function useCreateBranch() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.createBranch, onSuccess: adminInvalidator(qc) })
}

export function useDeleteBranch() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.deleteBranch, onSuccess: adminInvalidator(qc) })
}

export function useCreateDepartment() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.createDepartment, onSuccess: adminInvalidator(qc) })
}

export function useDeleteDepartment() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.deleteDepartment, onSuccess: adminInvalidator(qc) })
}

export function useCreateTeam() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.createTeam, onSuccess: adminInvalidator(qc) })
}

export function useDeleteTeam() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: orgApi.deleteTeam, onSuccess: adminInvalidator(qc) })
}

// ── Users ────────────────────────────────────────────────────────────────────

export function useUsers() {
  return useQuery({
    queryKey: ["admin", "users", orgId()],
    enabled: Boolean(orgId()),
    queryFn: userApi.listUsers,
  })
}

export function useUser(id: string | null) {
  return useQuery({
    queryKey: id ? ["admin", "user", id] : ["admin", "user", "none"],
    enabled: Boolean(id),
    queryFn: () => userApi.getUser(id!),
  })
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: userApi.updateUser, onSuccess: adminInvalidator(qc) })
}

export function useActivateUser() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: userApi.activateUser, onSuccess: adminInvalidator(qc) })
}

export function useDeactivateUser() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: userApi.deactivateUser, onSuccess: adminInvalidator(qc) })
}

export function useAddRoleToUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { userId: string; roleId: string }) =>
      userApi.addRoleToUser(input.userId, input.roleId),
    onSuccess: adminInvalidator(qc),
  })
}

export function useRemoveRoleFromUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { userId: string; roleId: string }) =>
      userApi.removeRoleFromUser(input.userId, input.roleId),
    onSuccess: adminInvalidator(qc),
  })
}

// ── RBAC ─────────────────────────────────────────────────────────────────────

export function useRoles() {
  return useQuery({
    queryKey: ["admin", "roles", orgId()],
    enabled: Boolean(orgId()),
    queryFn: rbacApi.readRoles,
  })
}

export function usePermissions() {
  return useQuery({
    queryKey: ["admin", "permissions"],
    queryFn: rbacApi.listPermissions,
    staleTime: 30 * 60_000,
  })
}

export function useCreateRole() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: rbacApi.createRole, onSuccess: adminInvalidator(qc) })
}

export function useUpdateRole() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: rbacApi.updateRole, onSuccess: adminInvalidator(qc) })
}

export function useDeleteRole() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: rbacApi.deleteRole, onSuccess: adminInvalidator(qc) })
}

export function useTogglePermission() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { roleId: string; codename: string; on: boolean }) =>
      input.on
        ? rbacApi.addPermissionToRole(input.roleId, input.codename)
        : rbacApi.removePermissionFromRole(input.roleId, input.codename),
    onSuccess: adminInvalidator(qc),
  })
}

// ── Invitations ──────────────────────────────────────────────────────────────

export function useInvitations(status?: inviteApi.InvitationStatus) {
  return useQuery({
    queryKey: ["admin", "invitations", orgId(), status ?? "all"],
    enabled: Boolean(orgId()),
    queryFn: () => inviteApi.listInvitations({ status }),
  })
}

export function useCreateInvitation() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: inviteApi.createInvitation, onSuccess: adminInvalidator(qc) })
}

export function useCancelInvitation() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: inviteApi.cancelInvitation, onSuccess: adminInvalidator(qc) })
}

// ── Apps ─────────────────────────────────────────────────────────────────────

export function useApps() {
  return useQuery({
    queryKey: ["admin", "apps", orgId()],
    enabled: Boolean(orgId()),
    queryFn: appsApi.listApps,
  })
}

export function useCreateApp() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: appsApi.createApp, onSuccess: adminInvalidator(qc) })
}

export function useRegenerateAppSecret() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: appsApi.regenerateAppSecret,
    onSuccess: adminInvalidator(qc),
  })
}

export function useDeleteApp() {
  const qc = useQueryClient()
  return useMutation({ mutationFn: appsApi.deleteApp, onSuccess: adminInvalidator(qc) })
}

// ── Agent status ─────────────────────────────────────────────────────────────

export function useAgentStatus() {
  return useQuery({
    queryKey: ["admin", "agent-status", "me"],
    queryFn: () => agentApi.getLatestStatus(),
    refetchInterval: 60_000,
  })
}

export function useSetAgentStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: agentApi.setStatus,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["admin", "agent-status"] }),
  })
}
