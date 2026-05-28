import { services, unary, stream, type Subscription } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  ListConversationsRequest,
  GetConversationRequest,
  ListMessagesRequest,
  AssignConversationRequest,
  UpdateConversationStatusRequest,
  ResolveConversationRequest,
  SendMessageRequest,
  StreamConversationUpdatesRequest,
  GetConversationTimelineRequest,
  type Conversation,
  type Message,
  type ConversationUpdate,
  ConversationStatus,
  ConversationPriority,
  Channel,
  ContentType,
} from "@/pb/conversations"

/**
 * Inbox API — typed wrappers around `ConversationServiceClient`.
 *
 * The feature module imports only from here, so the rest of the UI
 * is decoupled from the proto layer (and future regenerations don't
 * ripple through every screen). The wrappers also inject the active
 * org + workspace IDs from the workspace store so callers don't have
 * to thread them through every prop chain.
 */

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

export interface InboxFilters {
  channels?: Channel[]
  statuses?: ConversationStatus[]
  priorities?: ConversationPriority[]
  assignedTo?: string         // user_id
  teamId?: string
  unassignedOnly?: boolean
  query?: string              // free-text search (client-side)
  activeOnly?: boolean        // backend-supported flag
}

export interface ListPage {
  conversations: Conversation[]
  nextPageToken: string
  total: number
}

export async function listConversations(
  filters: InboxFilters,
  pageToken = "",
  pageSize = 50,
): Promise<ListPage> {
  const { orgId, workspaceId } = scope()

  // The backend ListConversations accepts at most one channel +
  // one status at a time. When the UI has multiple selections we
  // request unfiltered and narrow client-side. (F1.x can replace
  // this with a server-side multi-value filter once exposed.)
  const channelArg =
    filters.channels?.length === 1 ? filters.channels[0]! : Channel.UNSPECIFIED
  const statusArg =
    filters.statuses?.length === 1
      ? filters.statuses[0]!
      : ConversationStatus.UNSPECIFIED

  const req = ListConversationsRequest.create({
    orgId,
    workspaceId,
    channel: channelArg,
    status: statusArg,
    assignedTo: filters.assignedTo ?? "",
    pageSize,
    pageToken,
    activeOnly: filters.activeOnly ?? false,
  }) as ListConversationsRequest

  const res = await unary(services.conversation.listConversations(req))
  let conversations = res.conversations

  if (filters.channels && filters.channels.length > 1) {
    conversations = conversations.filter((c) => filters.channels!.includes(c.channel))
  }
  if (filters.statuses && filters.statuses.length > 1) {
    conversations = conversations.filter((c) => filters.statuses!.includes(c.status))
  }
  if (filters.priorities?.length) {
    conversations = conversations.filter((c) => filters.priorities!.includes(c.priority))
  }
  if (filters.unassignedOnly) {
    conversations = conversations.filter((c) => !c.assignedTo)
  }
  if (filters.query?.trim()) {
    const q = filters.query.trim().toLowerCase()
    conversations = conversations.filter(
      (c) =>
        c.subject.toLowerCase().includes(q) ||
        c.contactId.toLowerCase().includes(q),
    )
  }

  return {
    conversations,
    nextPageToken: res.nextPageToken,
    total: Number(res.total),
  }
}

export async function getConversation(id: string): Promise<Conversation> {
  return unary(
    services.conversation.getConversation(
      GetConversationRequest.create({ id, orgId: scope().orgId }) as GetConversationRequest,
    ),
  )
}

export async function listMessages(
  conversationId: string,
  before = "",
  limit = 100,
): Promise<{ messages: Message[]; hasMore: boolean }> {
  const res = await unary(
    services.conversation.listMessages(
      ListMessagesRequest.create({
        conversationId,
        orgId: scope().orgId,
        limit,
        before,
      }) as ListMessagesRequest,
    ),
  )
  return { messages: res.messages, hasMore: res.hasMore }
}

export async function assignConversation(
  conversationId: string,
  assignedTo: string,
  teamId = "",
): Promise<Conversation> {
  return unary(
    services.conversation.assignConversation(
      AssignConversationRequest.create({
        id: conversationId,
        orgId: scope().orgId,
        assignedTo,
        teamId,
      }) as AssignConversationRequest,
    ),
  )
}

export async function updateConversationStatus(
  conversationId: string,
  status: ConversationStatus,
  snoozedUntil?: string,
): Promise<Conversation> {
  return unary(
    services.conversation.updateConversationStatus(
      UpdateConversationStatusRequest.create({
        id: conversationId,
        orgId: scope().orgId,
        status,
        snoozedUntil: snoozedUntil ?? "",
      }) as UpdateConversationStatusRequest,
    ),
  )
}

export async function resolveConversation(
  conversationId: string,
  _reason = "",
): Promise<Conversation> {
  void _reason
  return unary(
    services.conversation.resolveConversation(
      ResolveConversationRequest.create({
        id: conversationId,
        orgId: scope().orgId,
      }) as ResolveConversationRequest,
    ),
  )
}

export interface SendMessageInput {
  conversationId: string
  body: string
  contentType?: ContentType
}

export async function sendMessage(input: SendMessageInput): Promise<Message> {
  return unary(
    services.conversation.sendMessage(
      SendMessageRequest.create({
        conversationId: input.conversationId,
        orgId: scope().orgId,
        contentType: input.contentType ?? ContentType.TEXT,
        content: JSON.stringify({ text: input.body }),
      }) as SendMessageRequest,
    ),
  )
}

export async function getConversationTimeline(
  conversationId: string,
): Promise<{
  events: { id: string; eventType: string; payload: string; createdAt: string; actorId: string }[]
}> {
  const res = await unary(
    services.conversation.getConversationTimeline(
      GetConversationTimelineRequest.create({
        id: conversationId,
        orgId: scope().orgId,
      }) as GetConversationTimelineRequest,
    ),
  )
  return {
    events: res.events.map((e) => ({
      id: e.id,
      eventType: e.eventType,
      payload: e.payload,
      createdAt: e.createdAt,
      actorId: e.actorId,
    })),
  }
}

export interface StreamHandlers {
  onUpdate: (update: ConversationUpdate) => void
  onError?: (err: Error) => void
}

export function streamInboxUpdates(
  workspaceId: string,
  handlers: StreamHandlers,
): Subscription {
  return stream(
    (opts) =>
      services.conversation.streamConversationUpdates(
        StreamConversationUpdatesRequest.create({
          workspaceId,
          orgId: scope().orgId,
        }) as StreamConversationUpdatesRequest,
        opts,
      ),
    {
      onMessage: handlers.onUpdate,
      onError: handlers.onError,
    },
  )
}
