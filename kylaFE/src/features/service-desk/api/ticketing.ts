import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreateTicketRoomRequest,
  GetTicketRoomRequest,
  ListTicketRoomsRequest,
  AddRoomMessageRequest,
  ListRoomMessagesRequest,
  CreateMacroRequest,
  GetMacroRequest,
  ListMacrosRequest,
  UpdateMacroRequest,
  DeleteMacroRequest,
  ApplyMacroRequest,
  type TicketRoom,
  type TicketRoomMessage,
  type Macro,
  RoomType,
  MacroVisibility,
} from "@/pb/ticketing"

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

// ── Rooms ────────────────────────────────────────────────────────────────────

export async function listTicketRooms(ticketId: string): Promise<TicketRoom[]> {
  const res = await unary(
    services.ticketing.listTicketRooms(
      ListTicketRoomsRequest.create({
        orgId: scope().orgId,
        ticketId,
      }) as ListTicketRoomsRequest,
    ),
  )
  return res.rooms
}

export async function getTicketRoom(id: string): Promise<TicketRoom> {
  return unary(
    services.ticketing.getTicketRoom(
      GetTicketRoomRequest.create({
        id,
        orgId: scope().orgId,
      }) as GetTicketRoomRequest,
    ),
  )
}

export async function createTicketRoom(input: {
  ticketId: string
  name: string
  type?: RoomType
}): Promise<TicketRoom> {
  return unary(
    services.ticketing.createTicketRoom(
      CreateTicketRoomRequest.create({
        orgId: scope().orgId,
        ticketId: input.ticketId,
        name: input.name,
        type: input.type ?? RoomType.INTERNAL,
      }) as CreateTicketRoomRequest,
    ),
  )
}

// ── Room messages ────────────────────────────────────────────────────────────

export async function listRoomMessages(
  roomId: string,
  before = "",
  limit = 100,
): Promise<{ messages: TicketRoomMessage[]; hasMore: boolean }> {
  const res = await unary(
    services.ticketing.listRoomMessages(
      ListRoomMessagesRequest.create({
        orgId: scope().orgId,
        roomId,
        before,
        limit,
      }) as ListRoomMessagesRequest,
    ),
  )
  return { messages: res.messages, hasMore: res.hasMore }
}

export async function addRoomMessage(input: {
  roomId: string
  content: string
  isPrivate?: boolean
}): Promise<TicketRoomMessage> {
  return unary(
    services.ticketing.addRoomMessage(
      AddRoomMessageRequest.create({
        orgId: scope().orgId,
        roomId: input.roomId,
        content: input.content,
        isPrivate: input.isPrivate ?? false,
      }) as AddRoomMessageRequest,
    ),
  )
}

// ── Macros ───────────────────────────────────────────────────────────────────

export async function listMacros(): Promise<Macro[]> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.ticketing.listMacros(
      ListMacrosRequest.create({
        orgId,
        workspaceId,
      }) as ListMacrosRequest,
    ),
  )
  return res.macros
}

export async function getMacro(id: string): Promise<Macro> {
  return unary(
    services.ticketing.getMacro(
      GetMacroRequest.create({
        id,
        orgId: scope().orgId,
      }) as GetMacroRequest,
    ),
  )
}

export async function createMacro(input: {
  name: string
  content: string
  actions?: Record<string, unknown>
  visibility?: MacroVisibility
}): Promise<Macro> {
  const { orgId, workspaceId } = scope()
  return unary(
    services.ticketing.createMacro(
      CreateMacroRequest.create({
        orgId,
        workspaceId,
        name: input.name,
        content: input.content,
        actions: input.actions ? JSON.stringify(input.actions) : "",
        visibility: input.visibility ?? MacroVisibility.TEAM,
      }) as CreateMacroRequest,
    ),
  )
}

export async function updateMacro(input: {
  id: string
  name?: string
  content?: string
  actions?: Record<string, unknown>
  visibility?: MacroVisibility
}): Promise<Macro> {
  return unary(
    services.ticketing.updateMacro(
      UpdateMacroRequest.create({
        id: input.id,
        orgId: scope().orgId,
        name: input.name ?? "",
        content: input.content ?? "",
        actions: input.actions ? JSON.stringify(input.actions) : "",
        visibility: input.visibility ?? MacroVisibility.UNSPECIFIED,
      }) as UpdateMacroRequest,
    ),
  )
}

export async function deleteMacro(id: string): Promise<void> {
  await unary(
    services.ticketing.deleteMacro(
      DeleteMacroRequest.create({
        id,
        orgId: scope().orgId,
      }) as DeleteMacroRequest,
    ),
  )
}

export async function applyMacro(input: {
  macroId: string
  ticketId: string
  roomId?: string
}): Promise<void> {
  await unary(
    services.ticketing.applyMacro(
      ApplyMacroRequest.create({
        orgId: scope().orgId,
        macroId: input.macroId,
        ticketId: input.ticketId,
        roomId: input.roomId ?? "",
      }) as ApplyMacroRequest,
    ),
  )
}

export { RoomType, MacroVisibility }
export type { TicketRoom, TicketRoomMessage, Macro }
