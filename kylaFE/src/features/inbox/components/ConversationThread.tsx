import { useEffect, useRef } from "react"
import { useMessages } from "../hooks/queries"
import { MessageBubble } from "./MessageBubble"
import { ListRowSkeleton, ErrorState } from "@/design-system"

/**
 * ConversationThread — scrollable message list with auto-scroll to
 * bottom on new messages.
 */
export function ConversationThread({ conversationId }: { conversationId: string }) {
  const messages = useMessages(conversationId)
  const scrollRef = useRef<HTMLDivElement>(null)
  const lastMsgId = messages.data?.[messages.data.length - 1]?.id

  useEffect(() => {
    if (!scrollRef.current) return
    scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [lastMsgId])

  if (messages.isPending) {
    return (
      <div className="flex-1 p-4">
        <ListRowSkeleton count={6} />
      </div>
    )
  }

  if (messages.isError) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <ErrorState
          title="Couldn't load messages"
          description={messages.error.message}
          onRetry={() => void messages.refetch()}
        />
      </div>
    )
  }

  return (
    <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
      {messages.data!.map((m) => (
        <MessageBubble key={m.id} message={m} />
      ))}
    </div>
  )
}
