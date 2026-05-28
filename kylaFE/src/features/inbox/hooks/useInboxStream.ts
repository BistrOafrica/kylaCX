import { useEffect } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { qk } from "@/lib/query"
import { useWorkspaceStore } from "@/lib/workspace"
import { streamInboxUpdates } from "../api/conversations"
import type { Conversation, Message, ConversationUpdate } from "@/pb/conversations"

/**
 * Subscribe to live inbox updates for the current workspace.
 *
 * Maps each `ConversationUpdate.eventType` to a TanStack-Query mutation:
 *   - conversation.updated / assigned / resolved → setQueryData on the
 *       detail key + invalidate the list
 *   - message.received → invalidate messages key
 *   - typing / presence_update → no cache mutation; consumed via store hooks
 *
 * Returning the cleanup is the responsibility of the consumer; mount
 * exactly once at the AppShell level for the inbox surface.
 */
export function useInboxStream(enabled = true) {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  const qc = useQueryClient()

  useEffect(() => {
    if (!enabled || !workspaceId) return

    const sub = streamInboxUpdates(workspaceId, {
      onUpdate: (update: ConversationUpdate) => {
        handleUpdate(update, qc)
      },
      onError: (err) => {
        // Stream is best-effort. Surface to the console for now; F1.x
        // adds a tiny reconnect banner above the inbox list once we
        // observe drop patterns in real usage.
        console.warn("[inbox-stream] disconnected:", err.message)
      },
    })

    return () => sub.cancel()
  }, [enabled, workspaceId, qc])
}

function handleUpdate(
  update: ConversationUpdate,
  qc: ReturnType<typeof useQueryClient>,
) {
  switch (update.eventType) {
    case "conversation.updated":
    case "assigned":
    case "resolved":
      if (update.conversation) {
        const conv = update.conversation as Conversation
        qc.setQueryData(qk.conversations.detail(conv.id), conv)
      }
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
      break

    case "message.received":
      if (update.message) {
        const msg = update.message as Message
        qc.setQueryData<Message[]>(
          qk.conversations.messages(msg.conversationId),
          (prev) => (prev ? [...prev, msg] : [msg]),
        )
      }
      void qc.invalidateQueries({ queryKey: qk.conversations.all() })
      break

    // typing + presence are handled outside the cache.
    default:
      break
  }
}
