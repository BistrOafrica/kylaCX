import { useState } from "react"
import { Link } from "react-router-dom"
import { IconArrowLeft, IconPlus, IconTicket } from "@tabler/icons-react"
import { ErrorState, CardSkeleton, Surface } from "@/design-system"
import { cn } from "@/lib/utils"
import { useObject, useObjectType } from "@/features/crm/hooks/queries"
import { FieldValue } from "@/features/crm/components/FieldValue"
import { ObjectTimeline } from "@/features/crm/components/ObjectTimeline"
import { useTicketRooms, useCreateTicketRoom } from "../hooks/queries"
import { TicketRoomThread } from "./TicketRoomThread"
import { MacroPicker } from "./MacroPicker"
import { RoomType } from "../api/ticketing"

/**
 * TicketDetail — composite view for a single ticket.
 *
 *   ┌──────────────────────┬───────────────────┬───────────┐
 *   │ Ticket header        │ Active room       │ Macros    │
 *   │ Custom fields panel  │ thread + composer │  panel    │
 *   │ Rooms switcher       │                   │           │
 *   └──────────────────────┴───────────────────┴───────────┘
 */
export function TicketDetail({ ticketId }: { ticketId: string }) {
  const obj = useObject(ticketId)
  const type = useObjectType(obj.data?.typeSlug ?? "ticket")
  const rooms = useTicketRooms(ticketId)
  const createRoom = useCreateTicketRoom()
  const [selectedRoomId, setSelectedRoomId] = useState<string | null>(null)

  // Derive the active room from the user's selection or, when none,
  // the first room returned by the server. This avoids a setState +
  // re-render dance for the common "default to first" case.
  const activeRoomId =
    selectedRoomId ?? rooms.data?.[0]?.id ?? null
  const setActiveRoomId = setSelectedRoomId

  if (obj.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
      </div>
    )
  }

  if (obj.isError || !obj.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load ticket"
          description={(obj.error as Error | undefined)?.message}
          onRetry={() => void obj.refetch()}
        />
      </div>
    )
  }

  const record = obj.data
  const fields = type.data?.schema?.fields ?? []
  const activeRoom = rooms.data?.find((r) => r.id === activeRoomId)
  const subject = pickName(record.data)

  return (
    <div className="h-full flex flex-col bg-canvas">
      <header className="flex items-start gap-3 px-6 py-3 border-b border-border">
        <Link
          to="/tickets"
          className={cn(
            "inline-flex items-center justify-center size-7 rounded-sm mt-0.5",
            "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
          )}
          aria-label="Back to tickets"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className={cn(
            "size-10 rounded-md flex items-center justify-center shrink-0",
            "bg-accent-subtle text-fg-secondary",
          )}
          aria-hidden
        >
          <IconTicket className="size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
            {subject}
          </h1>
          <div className="text-base text-fg-muted font-mono">
            #{ticketId.slice(0, 8)}
          </div>
        </div>
      </header>

      <div className="flex-1 min-h-0 grid grid-cols-12 gap-0">
        <aside className="col-span-3 overflow-y-auto border-e border-border bg-surface p-4 space-y-4">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            Fields
          </h2>
          <dl className="space-y-2">
            {fields.map((f) => (
              <div key={f.key} className="flex flex-col">
                <dt className="text-xs font-mono uppercase tracking-wider text-fg-muted">
                  {f.label}
                </dt>
                <dd className="text-base text-fg">
                  <FieldValue field={f} value={record.data[f.key]} />
                </dd>
              </div>
            ))}
            {fields.length === 0 && (
              <div className="text-base text-fg-muted">
                No schema fields yet.
              </div>
            )}
          </dl>

          <div className="border-t border-border pt-3 space-y-2">
            <div className="flex items-center justify-between">
              <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
                Rooms
              </h2>
              <button
                type="button"
                onClick={() =>
                  createRoom.mutate({
                    ticketId,
                    name: `Room ${(rooms.data?.length ?? 0) + 1}`,
                    type: RoomType.INTERNAL,
                  })
                }
                className="inline-flex items-center justify-center size-5 rounded-xs text-fg-muted hover:text-fg hover:bg-subtle"
                aria-label="New room"
              >
                <IconPlus className="size-3" />
              </button>
            </div>
            <ul className="space-y-px">
              {(rooms.data ?? []).map((r) => (
                <li key={r.id}>
                  <button
                    type="button"
                    onClick={() => setActiveRoomId(r.id)}
                    className={cn(
                      "w-full text-start px-2 h-7 rounded-sm text-base",
                      "hover:bg-subtle transition-colors",
                      activeRoomId === r.id && "bg-accent-subtle text-fg",
                    )}
                  >
                    <span className="truncate">{r.name || "Room"}</span>
                    <span className="ms-2 font-mono text-xs text-fg-muted">
                      {r.messageCount}
                    </span>
                  </button>
                </li>
              ))}
              {!rooms.isPending && (rooms.data?.length ?? 0) === 0 && (
                <li className="text-sm text-fg-muted py-1">No rooms yet.</li>
              )}
            </ul>
          </div>

          <div className="border-t border-border pt-3 space-y-2">
            <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
              Timeline
            </h2>
            <ObjectTimeline objectId={ticketId} />
          </div>
        </aside>

        <section className="col-span-6 min-w-0 flex flex-col">
          {activeRoom ? (
            <TicketRoomThread room={activeRoom} />
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <Surface level={1} radius="md" inset className="text-center">
                <div className="text-md text-fg-secondary">
                  Pick or create a room to start a discussion.
                </div>
              </Surface>
            </div>
          )}
        </section>

        <section className="col-span-3 min-w-0 border-s border-border">
          <MacroPicker
            ticketId={ticketId}
            roomId={activeRoomId ?? undefined}
          />
        </section>
      </div>
    </div>
  )
}

function pickName(data: Record<string, unknown>): string {
  for (const key of ["subject", "title", "name"]) {
    const v = data[key]
    if (typeof v === "string" && v) return v
  }
  return "(untitled ticket)"
}
