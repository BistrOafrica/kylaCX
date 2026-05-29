import { services, unary } from "@/lib/rpc"
import {
  Role,
  CreateRoleRequest,
  ReadRoleRequest,
  ReadRolesRequest,
  UpdateRoleRequest,
  DeleteRoleRequest,
  PermissionToRoleRequest,
  ReadPermissionsRequest,
  type Permission,
} from "@/pb/rbac"
import { orgScope } from "../utils/scope"

// ── Roles ────────────────────────────────────────────────────────────────────

export async function readRoles(): Promise<Role[]> {
  const res = await unary(
    services.role.readRoles(
      ReadRolesRequest.create({ scope: orgScope() }) as ReadRolesRequest,
    ),
  )
  return (res as { roles?: Role[] }).roles ?? []
}

export async function readRole(id: string): Promise<Role | null> {
  const res = await unary(
    services.role.readRole(
      ReadRoleRequest.create({ id }) as ReadRoleRequest,
    ),
  )
  return (res as { role?: Role }).role ?? null
}

export async function createRole(input: {
  name: string
  description?: string
  permissions?: string[]
}): Promise<Role | null> {
  const role = Role.create({
    name: input.name,
    description: input.description ?? "",
    permissions: input.permissions ?? [],
    ownerType: orgScope().ownerType,
    ownerId: orgScope().ownerId,
  }) as Role
  const res = await unary(
    services.role.createRole(
      CreateRoleRequest.create({ role }) as CreateRoleRequest,
    ),
  )
  return (res as { role?: Role }).role ?? null
}

export async function updateRole(role: Role): Promise<Role | null> {
  const res = await unary(
    services.role.updateRole(
      UpdateRoleRequest.create({ role }) as UpdateRoleRequest,
    ),
  )
  return (res as { role?: Role }).role ?? null
}

export async function deleteRole(id: string): Promise<void> {
  await unary(
    services.role.deleteRole(
      DeleteRoleRequest.create({ id }) as DeleteRoleRequest,
    ),
  )
}

export async function addPermissionToRole(roleId: string, codename: string): Promise<void> {
  await unary(
    services.role.addPermissionToRole(
      PermissionToRoleRequest.create({
        roleId,
        permissionCodenames: [codename],
      }) as PermissionToRoleRequest,
    ),
  )
}

export async function removePermissionFromRole(roleId: string, codename: string): Promise<void> {
  await unary(
    services.role.removePermissionFromRole(
      PermissionToRoleRequest.create({
        roleId,
        permissionCodenames: [codename],
      }) as PermissionToRoleRequest,
    ),
  )
}

export async function listPermissions(): Promise<Permission[]> {
  const res = await unary(
    services.permission.readPermissions(
      ReadPermissionsRequest.create({}) as ReadPermissionsRequest,
    ),
  )
  return (res as { permissions?: Permission[] }).permissions ?? []
}

export type { Role, Permission }
