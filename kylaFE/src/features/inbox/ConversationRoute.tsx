import { useParams, useNavigate } from "react-router-dom"
import { useHotkeys } from "react-hotkeys-hook"
import { IconInbox } from "@tabler/icons-react"
import { EmptyState, ErrorState, ListRowSkeleton } from "@/design-system"
import { useConversation, useResolveConversation, useUpdateStatus } from "./hooks/queries"
import { ConversationHeader } from "./components/ConversationHeader"
import { ConversationThread } from "./components/ConversationThread"
import { Composer } from "./components/Composer"
import { ContactSidePanel } from "./components/ContactSidePanel"
import { InboxCopilot } from "./components/InboxCopilot"
import { ConversationStatus } from "./utils/enums"

/**
 * Right-pane router for the inbox.
 *
 * /inbox            → empty state
 * /inbox/:id        → header + thread + composer + side panel + copilot
 *
 * Owns the per-conversation keyboard shortcuts (E = resolve, S = snooze)
 * via react-hotkeys-hook scoped to this surface.
 */
export function ConversationRoute() {
  const { id } = useParams()
  const navigate = useNavigate()
  const conv = useConversation(id ?? null)
  const resolve = useResolveConversation()
  const updateStatus = useUpdateStatus()

  useHotkeys(
    "e",
    () => {
      if (!id || resolve.isPending) return
      resolve.mutate({ conversationId: id })
    },
    { enabled: Boolean(id), enableOnFormTags: false },
    [id, resolve],
  )

  useHotkeys(
    "s",
    () => {
      if (!id || updateStatus.isPending) return
      const until = new Date(Date.now() + 4 * 60 * 60 * 1000).toISOString()
      updateStatus.mutate({
        conversationId: id,
        status: ConversationStatus.SNOOZED,
        snoozedUntil: until,
      })
    },
    { enabled: Boolean(id), enableOnFormTags: false },
    [id, updateStatus],
  )

  useHotkeys(
    "esc",
    () => navigate("/inbox"),
    { enabled: Boolean(id), enableOnFormTags: false },
  )

  if (!id) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <EmptyState
          icon={<IconInbox className="size-5" />}
          title="Pick a conversation"
          description="Choose a thread on the left, or use j / k to move between rows."
        />
      </div>
    )
  }

  if (conv.isPending) {
    return (
      <div className="flex-1 p-4">
        <ListRowSkeleton count={6} />
      </div>
    )
  }

  if (conv.isError || !conv.data) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <ErrorState
          title="Couldn't load conversation"
          description={conv.error?.message}
          onRetry={() => void conv.refetch()}
        />
      </div>
    )
  }

  return (
    <div className="flex-1 flex min-w-0">
      <div className="flex-1 min-w-0 flex flex-col">
        <ConversationHeader conv={conv.data} />
        <ConversationThread conversationId={id} />
        <Composer conversationId={id} />
      </div>
      <ContactSidePanel conv={conv.data} />
      <InboxCopilot conv={conv.data} />
    </div>
  )
}
