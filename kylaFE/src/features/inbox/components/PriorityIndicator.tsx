import { cn } from "@/lib/utils"
import { priorityMeta, type ConversationPriority } from "../utils/enums"

const TONE_CLASS = {
  muted:  "bg-fg-muted",
  info:   "bg-info",
  warn:   "bg-warn",
  danger: "bg-danger",
} as const

/**
 * 2px vertical bar at the row's leading edge — Linear-style priority
 * indicator. Cleaner than a label; reads at a glance.
 */
export function PriorityIndicator({ priority }: { priority: ConversationPriority }) {
  const meta = priorityMeta(priority)
  return (
    <span
      aria-label={`Priority: ${meta.label}`}
      className={cn(
        "block w-[2px] h-full shrink-0",
        TONE_CLASS[meta.tone],
      )}
    />
  )
}
