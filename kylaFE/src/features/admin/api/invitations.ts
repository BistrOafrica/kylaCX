import { services, unary } from "@/lib/rpc"
import {
  CreateInviteRequest,
  GetInviteRequest,
  ListInvitationsRequest,
  AcceptInviteRequest,
  RejectInviteRequest,
  CancelInviteRequest,
  DeleteInviteRequest,
  type Invitation,
  InvitationType,
  InvitationStatus,
} from "@/pb/invitation"
import { useAuthStore } from "@/lib/auth"
import { useWorkspaceStore } from "@/lib/workspace"

function ids() {
  return {
    orgId: useWorkspaceStore.getState().organisation?.id ?? "",
    userId: useAuthStore.getState().identity?.userId ?? "",
  }
}

export async function listInvitations(filters?: {
  status?: InvitationStatus
  branchId?: string
  departmentId?: string
  teamId?: string
}): Promise<Invitation[]> {
  const { orgId } = ids()
  const res = await unary(
    services.invitation.listInvitations(
      ListInvitationsRequest.create({
        email: "",
        invitedBy: "",
        type: InvitationType.UNSPECIFIED,
        status: filters?.status ?? InvitationStatus.UNSPECIFIED,
        organisationId: orgId,
        branchId: filters?.branchId ?? "",
        departmentId: filters?.departmentId ?? "",
        teamId: filters?.teamId ?? "",
        userId: "",
        pageSize: 100,
        pageToken: "",
      }) as ListInvitationsRequest,
    ),
  )
  return res.invitations ?? []
}

export async function getInvitation(id: string): Promise<Invitation> {
  return unary(
    services.invitation.getInvitation(
      GetInviteRequest.create({ id }) as GetInviteRequest,
    ),
  )
}

export async function createInvitation(input: {
  email: string
  type?: InvitationType
  branchId?: string
  departmentId?: string
  teamId?: string
  userId?: string
  roleIds?: string[]
  expirationHours?: number
}): Promise<Invitation> {
  const { orgId, userId } = ids()
  return unary(
    services.invitation.createInvitation(
      CreateInviteRequest.create({
        email: input.email,
        invitedBy: userId,
        type: input.type ?? InvitationType.NEW_USER,
        organisationId: orgId,
        branchId: input.branchId ?? "",
        departmentId: input.departmentId ?? "",
        teamId: input.teamId ?? "",
        userId: input.userId ?? "",
        roleIds: input.roleIds ?? [],
        expirationHours: input.expirationHours ?? 72,
      }) as CreateInviteRequest,
    ),
  )
}

export async function acceptInvitation(id: string): Promise<Invitation> {
  return unary(
    services.invitation.acceptInvitation(
      AcceptInviteRequest.create({ id }) as AcceptInviteRequest,
    ),
  )
}

export async function rejectInvitation(id: string): Promise<Invitation> {
  return unary(
    services.invitation.rejectInvitation(
      RejectInviteRequest.create({ id }) as RejectInviteRequest,
    ),
  )
}

export async function cancelInvitation(id: string): Promise<Invitation> {
  return unary(
    services.invitation.cancelInvitation(
      CancelInviteRequest.create({ id }) as CancelInviteRequest,
    ),
  )
}

export async function deleteInvitation(id: string): Promise<void> {
  await unary(
    services.invitation.deleteInvitation(
      DeleteInviteRequest.create({ id }) as DeleteInviteRequest,
    ),
  )
}

export { InvitationType, InvitationStatus }
export type { Invitation }
