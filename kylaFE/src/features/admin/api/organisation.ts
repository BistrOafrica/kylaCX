import { services, unary } from "@/lib/rpc"
import {
  Branch,
  CreateBranchRequest,
  ReadBranchesRequest,
  UpdateBranchRequest,
  DeleteBranchRequest,
} from "@/pb/branch"
import {
  Department,
  CreateDepartmentRequest,
  ReadDepartmentsRequest,
  UpdateDepartmentRequest,
  DeleteDepartmentRequest,
  ReadBranchDepartmentsRequest,
} from "@/pb/department"
import {
  Team,
  CreateTeamRequest,
  ReadTeamListRequest,
  UpdateTeamRequest,
  DeleteTeamRequest,
  AddUserToTeamRequest,
  RemoveUserFromTeamRequest,
  ReadTeamUsersRequest,
} from "@/pb/team"
import { branchScope, departmentScope, orgScope } from "../utils/scope"
import { OwnerType } from "@/pb/owner_type"

// ── Branches ─────────────────────────────────────────────────────────────────

export async function listBranches(): Promise<Branch[]> {
  const res = await unary(
    services.branch.readBranches(
      ReadBranchesRequest.create({ scope: orgScope() }) as ReadBranchesRequest,
    ),
  )
  return (res as { branches?: Branch[] }).branches ?? []
}

export async function createBranch(input: {
  name: string
  description?: string
  parentId?: string
}): Promise<Branch | null> {
  const branch = Branch.create({
    name: input.name,
    description: input.description ?? "",
    parentId: input.parentId ?? "",
    ownerType: OwnerType.ORGANISATIONS,
    ownerId: orgScope().ownerId,
  }) as Branch
  const res = await unary(
    services.branch.createBranch(
      CreateBranchRequest.create({ branch }) as CreateBranchRequest,
    ),
  )
  return (res as { branch?: Branch }).branch ?? null
}

export async function updateBranch(branch: Branch): Promise<Branch | null> {
  const res = await unary(
    services.branch.updateBranch(
      UpdateBranchRequest.create({ branch }) as UpdateBranchRequest,
    ),
  )
  return (res as { branch?: Branch }).branch ?? null
}

export async function deleteBranch(id: string): Promise<void> {
  await unary(
    services.branch.deleteBranch(
      DeleteBranchRequest.create({ id }) as DeleteBranchRequest,
    ),
  )
}

// ── Departments ──────────────────────────────────────────────────────────────

export async function listDepartments(): Promise<Department[]> {
  const res = await unary(
    services.department.readDepartments(
      ReadDepartmentsRequest.create({ scope: orgScope() }) as ReadDepartmentsRequest,
    ),
  )
  return (res as { departments?: Department[] }).departments ?? []
}

export async function listBranchDepartments(branchId: string): Promise<Department[]> {
  const res = await unary(
    services.department.readBranchDepartments(
      ReadBranchDepartmentsRequest.create({
        scope: branchScope(branchId),
      }) as ReadBranchDepartmentsRequest,
    ),
  )
  return (res as { departments?: Department[] }).departments ?? []
}

export async function createDepartment(input: {
  departmentName: string
  departmentBio?: string
  ownerType?: OwnerType
  ownerId?: string
}): Promise<Department | null> {
  const department = Department.create({
    departmentName: input.departmentName,
    departmentBio: input.departmentBio ?? "",
    ownerType: input.ownerType ?? OwnerType.ORGANISATIONS,
    ownerId: input.ownerId ?? orgScope().ownerId,
  }) as Department
  const res = await unary(
    services.department.createDepartment(
      CreateDepartmentRequest.create({ department }) as CreateDepartmentRequest,
    ),
  )
  return (res as { department?: Department }).department ?? null
}

export async function updateDepartment(department: Department): Promise<Department | null> {
  const res = await unary(
    services.department.updateDepartment(
      UpdateDepartmentRequest.create({ department }) as UpdateDepartmentRequest,
    ),
  )
  return (res as { department?: Department }).department ?? null
}

export async function deleteDepartment(id: string): Promise<void> {
  await unary(
    services.department.deleteDepartment(
      DeleteDepartmentRequest.create({ id }) as DeleteDepartmentRequest,
    ),
  )
}

// ── Teams ────────────────────────────────────────────────────────────────────

export async function listTeams(scopeOwner?: { ownerType: OwnerType; ownerId: string }): Promise<Team[]> {
  const which = scopeOwner ?? { ownerType: orgScope().ownerType, ownerId: orgScope().ownerId }
  const req = ReadTeamListRequest.create({
    page: "1",
    pageSize: "100",
  }) as ReadTeamListRequest
  // Use the most specific endpoint based on scope.
  let res
  switch (which.ownerType) {
    case OwnerType.BRANCHES:
      res = await unary(services.team.readTeamsByBranchID({
        ...req,
      }))
      break
    case OwnerType.DEPARTMENTS:
      res = await unary(services.team.readTeamsByDepartmentID({
        ...req,
      }))
      break
    case OwnerType.USERS:
      res = await unary(services.team.readTeamsByUserID({
        ...req,
      }))
      break
    case OwnerType.ORGANISATIONS:
    default:
      res = await unary(services.team.readTeamsByorganisationId({
        ...req,
      }))
      break
  }
  return (res as { teams?: Team[] }).teams ?? []
}

export async function listTeamUsers(teamId: string) {
  const res = await unary(
    services.team.readTeamUsers(
      ReadTeamUsersRequest.create({
        teamId,
        scope: departmentScope(""),
      }) as ReadTeamUsersRequest,
    ),
  )
  return (res as { users?: unknown[] }).users ?? []
}

export async function createTeam(input: {
  name: string
  description?: string
  ownerType?: OwnerType
  ownerId?: string
}): Promise<Team | null> {
  const team = Team.create({
    name: input.name,
    description: input.description ?? "",
    ownerType: input.ownerType ?? OwnerType.ORGANISATIONS,
    ownerId: input.ownerId ?? orgScope().ownerId,
  }) as Team
  const res = await unary(
    services.team.createTeam(
      CreateTeamRequest.create({ team }) as CreateTeamRequest,
    ),
  )
  return (res as { team?: Team }).team ?? null
}

export async function updateTeam(team: Team): Promise<Team | null> {
  const res = await unary(
    services.team.updateTeam(
      UpdateTeamRequest.create({ team }) as UpdateTeamRequest,
    ),
  )
  return (res as { team?: Team }).team ?? null
}

export async function deleteTeam(id: string): Promise<void> {
  await unary(
    services.team.deleteTeam(
      DeleteTeamRequest.create({ id }) as DeleteTeamRequest,
    ),
  )
}

export async function addUserToTeam(teamId: string, userId: string) {
  return unary(
    services.team.addUserToTeam(
      AddUserToTeamRequest.create({
        teamId,
        userId,
      }) as AddUserToTeamRequest,
    ),
  )
}

export async function removeUserFromTeam(teamId: string, userId: string) {
  return unary(
    services.team.removeUserFromTeam(
      RemoveUserFromTeamRequest.create({
        teamId,
        userId,
      }) as RemoveUserFromTeamRequest,
    ),
  )
}

export type { Branch, Department, Team }
