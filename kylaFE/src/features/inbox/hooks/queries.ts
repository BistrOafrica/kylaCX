import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query"
import { qk } from "@/lib/query"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  listConversations,
  getConversation,
  listMessages,
  sendMessage,
  assignConversation,
  resolveConversation,
  updateConversationStatus,
  getConversationTimeline,
  type InboxFilters,
  type ListPage,
  type SendMessageInput,
} from "../api/conversations"
import { readContact } from "../api/contact"
import { ConversationStatus } from "@/pb/conversations"

const conversationsKey = (wsId: string, filters: InboxFilters) =>
  qk.conversations.list(wsId, filters)

export function useInbox(filters: InboxFilters) {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")

  return useInfiniteQuery<ListPage, Error>({
    queryKey: conversationsKey(workspaceId, filters),
    enabled: Boolean(workspaceId),
    initialPageParam: "",
    queryFn: ({ pageParam }) => listConversations(filters, pageParam as string),
    getNextPageParam: (last) =>
      last.nextPageToken && last.conversations.length > 0
        ? last.nextPageToken
        : undefined,
    staleTime: 15_000,
  })
}

export function useConversation(id: string | null) {
  return useQuery({
    queryKey: id ? qk.conversations.detail(id) : ["conversations", "detail", "none"],
    enabled: Boolean(id),
    queryFn: () => getConversation(id!),
  })
}

export function useMessages(conversationId: string | null) {
  return useQuery({
    queryKey: conversationId
      ? qk.conversations.messages(conversationId)
      : ["conversations", "messages", "none"],
    enabled: Boolean(conversationId),
    queryFn: () => listMessages(conversationId!).then((r) => r.messages),
    staleTime: 5_000,
  })
}

export function useConversationTimeline(conversationId: string | null) {
  return useQuery({
    queryKey: conversationId
      ? qk.conversations.timeline(conversationId)
      : ["conversations", "timeline", "none"],
    enabled: Boolean(conversationId),
    queryFn: () => getConversationTimeline(conversationId!).then((r) => r.events),
  })
}

export function useContact(contactId: string | null | undefined) {
  return useQuery({
    queryKey: contactId ? qk.contacts.detail(contactId) : ["contacts", "detail", "none"],
    enabled: Boolean(contactId),
    queryFn: () => readContact(contactId!),
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: SendMessageInput) => sendMessage(input),
    onSuccess: (_msg, vars) => {
      void qc.invalidateQueries({ queryKey: qk.conversations.messages(vars.conversationId) })
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
    },
  })
}

export function useAssignConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { conversationId: string; assigneeId: string; teamId?: string }) =>
      assignConversation(input.conversationId, input.assigneeId, input.teamId ?? ""),
    onSuccess: (conv) => {
      qc.setQueryData(qk.conversations.detail(conv.id), conv)
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
    },
  })
}

export function useResolveConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { conversationId: string; reason?: string }) =>
      resolveConversation(input.conversationId, input.reason ?? ""),
    onSuccess: (conv) => {
      qc.setQueryData(qk.conversations.detail(conv.id), conv)
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
    },
  })
}

export function useUpdateStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: {
      conversationId: string
      status: ConversationStatus
      snoozedUntil?: string
    }) =>
      updateConversationStatus(
        input.conversationId,
        input.status,
        input.snoozedUntil,
      ),
    onSuccess: (conv) => {
      qc.setQueryData(qk.conversations.detail(conv.id), conv)
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
    },
  })
}
