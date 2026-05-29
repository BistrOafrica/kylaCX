import { services, unary } from "@/lib/rpc"
import {
  CreateUserRequest,
  ReadUserRequest,
  ReadUsersRequest,
  UpdateUserRequest,
  DeactivateUserRequest,
  ActivateUserRequest,
  DeleteUserRequest,
  SearchUsersRequest,
  RoleToUserRequest,
  User,
} from "@/pb/user"
import { orgScope } from "../utils/scope"

// ── Users ────────────────────────────────────────────────────────────────────

export async function listUsers(): Promise<User[]> {
  const res = await unary(
    services.user.readUsers(
      ReadUsersRequest.create({ scope: orgScope() }) as ReadUsersRequest,
    ),
  )
  return (res as { users?: User[] }).users ?? []
}

export async function getUser(id: string): Promise<User | null> {
  const res = await unary(
    services.user.readUser(
      ReadUserRequest.create({ user: id }) as ReadUserRequest,
    ),
  )
  return (res as { user?: User }).user ?? null
}

export async function searchUsers(query: string): Promise<User[]> {
  const res = await unary(
    services.user.searchUsers(
      SearchUsersRequest.create({
        query,
        scope: orgScope(),
      }) as SearchUsersRequest,
    ),
  )
  return (res as { users?: User[] }).users ?? []
}

export async function createUser(input: Partial<User>): Promise<User | null> {
  const user = User.create({
    firstName: input.firstName ?? "",
    lastName: input.lastName ?? "",
    email: input.email ?? "",
    phone: input.phone ?? "",
    username: input.username ?? input.email ?? "",
    ownerId: orgScope().ownerId,
    ownerType: orgScope().ownerType,
  }) as User
  const res = await unary(
    services.user.createUser(
      CreateUserRequest.create({ user }) as CreateUserRequest,
    ),
  )
  return (res as { user?: User }).user ?? null
}

export async function updateUser(user: User): Promise<User | null> {
  const res = await unary(
    services.user.updateUser(
      UpdateUserRequest.create({ user }) as UpdateUserRequest,
    ),
  )
  return (res as { user?: User }).user ?? null
}

export async function activateUser(userId: string): Promise<void> {
  await unary(
    services.user.activateUser(
      ActivateUserRequest.create({ userId }) as ActivateUserRequest,
    ),
  )
}

export async function deactivateUser(userId: string): Promise<void> {
  await unary(
    services.user.deactivateUser(
      DeactivateUserRequest.create({ userId }) as DeactivateUserRequest,
    ),
  )
}

export async function deleteUser(id: string): Promise<void> {
  await unary(
    services.user.deleteUser(
      DeleteUserRequest.create({ id }) as DeleteUserRequest,
    ),
  )
}

export async function addRoleToUser(userId: string, roleId: string): Promise<void> {
  await unary(
    services.user.addRoleToUser(
      RoleToUserRequest.create({
        userId,
        roleId,
        scope: orgScope(),
      }) as RoleToUserRequest,
    ),
  )
}

export async function removeRoleFromUser(userId: string, roleId: string): Promise<void> {
  await unary(
    services.user.removeRoleFromUser(
      RoleToUserRequest.create({
        userId,
        roleId,
        scope: orgScope(),
      }) as RoleToUserRequest,
    ),
  )
}

export type { User }
