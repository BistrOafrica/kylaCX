import { memo } from "react"
import { Link } from "react-router-dom"
import { ChannelBadge, StatusDot } from "@/design-system"
import { cn } from "@/lib/utils"
import { channelMeta, type ConversationStatus } from "../utils/enums"
import { relativeTime } from "../utils/format"
import { SlaBadge } from "./SlaBadge"
import { PriorityIndicator } from "./PriorityIndicator"
import type { Conversation } from "@/pb/conversations"

/**
 * Single inbox row. Renders inside a virtualizer so we keep the
 * height stable at 32px (dense default).
 *
 * Click → navigate to /inbox/:id. Active state styled by NavLink
 * via the `selected` prop (parent owns selection so keyboard nav
 * works without router churn).
 */
function InboxRowImpl({
  conv,
  selected,
  isUnread,
}: {
  conv: Conversation
  selected: boolean
  isUnread?: boolean
}) {
  const channel = channelMeta(conv.channel)

  return (
    <Link
      to={`/inbox/${conv.id}`}
      data-conversation-id={conv.id}
      aria-current={selected ? "true" : undefined}
      className={cn(
        "group flex items-stretch gap-0 h-8 px-0 border-b border-border-subtle",
        "text-fg-secondary",
        "hover:bg-subtle transition-colors",
        selected && "bg-accent-subtle text-fg",
      )}
    >
      <PriorityIndicator priority={conv.priority} />
      <div className="flex-1 min-w-0 flex items-center gap-2 px-2">
        <StatusDot
          tone={isUnread ? "unread" : "read"}
          size={6}
          className={cn(!isUnread && "opacity-0 group-hover:opacity-100")}
        />
        <span
          className={cn(
            "shrink-0 truncate text-base font-medium",
            "w-32 sm:w-40",
            selected ? "text-fg" : "text-fg-secondary",
            isUnread && "text-fg",
          )}
        >
          {contactLabel(conv)}
        </span>
        <ChannelBadge channel={channel.token} />
        <span
          className={cn(
            "flex-1 min-w-0 truncate text-base",
            isUnread ? "text-fg" : "text-fg-muted",
          )}
        >
          {conv.subject || "(no subject)"}
        </span>
        <SlaBadge deadline={conv.slaDeadline ?? null} size="sm" />
        <span className="font-mono text-xs text-fg-muted shrink-0">
          {relativeTime(conv.updatedAt)}
        </span>
      </div>
    </Link>
  )
}

export const InboxRow = memo(InboxRowImpl, (a, b) =>
  a.conv.id === b.conv.id &&
  a.conv.updatedAt === b.conv.updatedAt &&
  a.conv.status === b.conv.status &&
  a.conv.priority === b.conv.priority &&
  a.conv.assignedTo === b.conv.assignedTo &&
  a.conv.slaDeadline === b.conv.slaDeadline &&
  a.selected === b.selected &&
  a.isUnread === b.isUnread,
)

function contactLabel(conv: Conversation): string {
  if (conv.contactId) return conv.contactId.slice(0, 12)
  return conv.channelRef || conv.id.slice(0, 8)
}

// Re-export the type-erased status hint for filter UI.
export type { ConversationStatus }
