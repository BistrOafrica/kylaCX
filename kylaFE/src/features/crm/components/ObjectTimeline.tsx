import { useMemo } from "react"
import {
  IconHistory,
  IconPlus,
  IconEdit,
  IconLink,
  IconTrash,
} from "@tabler/icons-react"
import { EmptyState, CardSkeleton } from "@/design-system"
import { cn } from "@/lib/utils"
import { relativeShort } from "../utils/relative"
import { useObjectTimeline } from "../hooks/queries"

/**
 * ObjectTimeline — reusable activity feed for any Object Core record.
 *
 * Renders chronological events with type-specific icons. Used in the
 * CRM object detail view; will be reused in F1.x for conversation
 * detail timelines and F3 for ticket activity feeds.
 */
export function ObjectTimeline({ objectId }: { objectId: string }) {
  const timeline = useObjectTimeline(objectId)

  const events = useMemo(() => timeline.data ?? [], [timeline.data])

  if (timeline.isPending) {
    return (
      <div className="space-y-2">
        <CardSkeleton lines={2} />
        <CardSkeleton lines={2} />
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <EmptyState
        icon={<IconHistory className="size-4" />}
        title="No activity yet"
        description="Changes to this record will appear here."
        size="sm"
      />
    )
  }

  return (
    <ol className="relative space-y-3 ps-4 before:absolute before:start-1.5 before:top-1.5 before:bottom-1.5 before:w-px before:bg-border-subtle">
      {events.map((event) => (
        <li key={event.id} className="relative">
          <span
            aria-hidden
            className={cn(
              "absolute -start-3.5 top-0.5 size-3 rounded-full flex items-center justify-center",
              "bg-surface border-2 border-border text-fg-muted",
            )}
          >
            <EventIcon type={event.eventType} />
          </span>
          <div className="space-y-0.5">
            <div className="text-sm text-fg-secondary">
              <span className="font-medium text-fg">
                {humanEventType(event.eventType)}
              </span>
              {event.payload && event.eventType !== "created" && (
                <PayloadPreview payload={event.payload} />
              )}
            </div>
            <div className="text-xs font-mono text-fg-muted">
              {relativeShort(event.createdAt)} · {event.actorType}{" "}
              {event.actorId ? `· ${event.actorId.slice(0, 8)}` : ""}
            </div>
          </div>
        </li>
      ))}
    </ol>
  )
}

function EventIcon({ type }: { type: string }) {
  switch (type) {
    case "created":
      return <IconPlus className="size-2" />
    case "updated":
      return <IconEdit className="size-2" />
    case "linked":
    case "unlinked":
      return <IconLink className="size-2" />
    case "deleted":
      return <IconTrash className="size-2" />
    default:
      return <IconHistory className="size-2" />
  }
}

function humanEventType(t: string): string {
  switch (t) {
    case "created":  return "Record created"
    case "updated":  return "Record updated"
    case "linked":   return "Linked"
    case "unlinked": return "Unlinked"
    case "deleted":  return "Deleted"
    default:         return t.replaceAll("_", " ")
  }
}

function PayloadPreview({ payload }: { payload: string }) {
  // Best-effort summary: try to parse JSON and pull out a sensible shape.
  let summary = ""
  try {
    const parsed = JSON.parse(payload) as Record<string, unknown>
    const keys = Object.keys(parsed).slice(0, 3)
    if (keys.length) {
      summary = ` — ${keys.join(", ")}`
    }
  } catch {
    /* not JSON */
  }
  return <span className="text-fg-muted">{summary}</span>
}
