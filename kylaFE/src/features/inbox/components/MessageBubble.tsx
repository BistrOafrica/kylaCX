import { memo } from "react"
import { IconCheck, IconChecks, IconClock, IconAlertCircle } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { SenderType, MessageStatus, channelMeta } from "../utils/enums"
import { readContentText, relativeTime } from "../utils/format"
import type { Message } from "@/pb/conversations"

/**
 * Channel-aware message bubble.
 *
 * Agent messages align end, contact messages start. System/bot messages
 * render centered with a muted dashed border (no bubble).
 */
function MessageBubbleImpl({ message }: { message: Message }) {
  const fromAgent = message.senderType === SenderType.AGENT
  const fromBot = message.senderType === SenderType.BOT
  const fromSystem = message.senderType === SenderType.SYSTEM
  const ch = channelMeta(message.channel)
  const body = readContentText(message.content)

  if (fromSystem) {
    return (
      <div className="flex justify-center my-2">
        <span className="text-xs font-mono uppercase tracking-wider text-fg-muted px-2 py-0.5 border border-border-subtle rounded-xs">
          {body || message.externalId || "System event"}
        </span>
      </div>
    )
  }

  return (
    <div
      className={cn(
        "flex flex-col gap-1 max-w-[78%]",
        fromAgent ? "ms-auto items-end" : "items-start",
      )}
    >
      <div
        className={cn(
          "px-3 py-2 rounded-md text-base whitespace-pre-wrap break-words",
          fromAgent
            ? "bg-accent text-accent-fg rounded-ee-xs"
            : "bg-surface border border-border text-fg rounded-es-xs",
          fromBot && "bg-accent-subtle text-fg border border-accent/30",
        )}
      >
        {body}
      </div>
      <div className="flex items-center gap-1.5 text-xs text-fg-muted font-mono">
        <span>{ch.label}</span>
        <span aria-hidden>·</span>
        <span>{relativeTime(message.createdAt)}</span>
        {fromAgent && <MessageStatusIcon status={message.status} />}
      </div>
    </div>
  )
}

export const MessageBubble = memo(MessageBubbleImpl, (a, b) =>
  a.message.id === b.message.id && a.message.status === b.message.status,
)

function MessageStatusIcon({ status }: { status: MessageStatus }) {
  switch (status) {
    case MessageStatus.PENDING:
      return <IconClock className="size-3" aria-label="Pending" />
    case MessageStatus.SENT:
      return <IconCheck className="size-3" aria-label="Sent" />
    case MessageStatus.DELIVERED:
      return <IconChecks className="size-3" aria-label="Delivered" />
    case MessageStatus.READ:
      return <IconChecks className="size-3 text-info" aria-label="Read" />
    case MessageStatus.FAILED:
      return <IconAlertCircle className="size-3 text-danger" aria-label="Failed" />
    default:
      return null
  }
}
