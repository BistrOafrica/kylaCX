import { services, unary } from "@/lib/rpc"
import {
  App,
  CreateAppRequest,
  ReadAppRequest,
  ReadAppsRequest,
  UpdateAppRequest,
  DeleteAppRequest,
  ApproveAppRequest,
} from "@/pb/apps"
import { orgScope, OwnerType } from "../utils/scope"

export async function listApps(): Promise<App[]> {
  const res = await unary(
    services.app.readApps(
      ReadAppsRequest.create({ scope: orgScope() }) as ReadAppsRequest,
    ),
  )
  return (res as { apps?: App[] }).apps ?? []
}

export async function getApp(id: string): Promise<App | null> {
  const res = await unary(
    services.app.readApp(
      ReadAppRequest.create({ id }) as ReadAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export async function createApp(input: {
  name: string
  description?: string
  permissionsCodenames?: string[]
}): Promise<App | null> {
  const app = App.create({
    name: input.name,
    description: input.description ?? "",
    permissionsCodenames: input.permissionsCodenames ?? [],
    ownerType: OwnerType.ORGANISATIONS,
    ownerId: orgScope().ownerId,
    status: "PENDING",
  }) as App
  const res = await unary(
    services.app.createApp(
      CreateAppRequest.create({ app }) as CreateAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export async function updateApp(app: App): Promise<App | null> {
  const res = await unary(
    services.app.updateApp(
      UpdateAppRequest.create({ app }) as UpdateAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export async function regenerateAppSecret(app: App): Promise<App | null> {
  const res = await unary(
    services.app.regenerateAppKeyAndSecret(
      UpdateAppRequest.create({ app }) as UpdateAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export async function deleteApp(id: string): Promise<void> {
  await unary(
    services.app.deleteApp(
      DeleteAppRequest.create({ id }) as DeleteAppRequest,
    ),
  )
}

export async function approveApp(app: App): Promise<App | null> {
  const res = await unary(
    services.app.approveApp(
      ApproveAppRequest.create({ app }) as ApproveAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export async function rejectApp(app: App): Promise<App | null> {
  const res = await unary(
    services.app.rejectApp(
      ApproveAppRequest.create({ app }) as ApproveAppRequest,
    ),
  )
  return (res as { app?: App }).app ?? null
}

export type { App }
