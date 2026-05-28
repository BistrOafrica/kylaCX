import { useEffect, useRef, useState, type KeyboardEvent } from "react"
import { IconLock, IconSend, IconLoader2 } from "@tabler/icons-react"
import { ListRowSkeleton, ErrorState, Kbd } from "@/design-system"
import { cn } from "@/lib/utils"
import { useRoomMessages, useAddRoomMessage } from "../hooks/queries"
import { relativeShort } from "@/features/crm/utils/relative"
import type { TicketRoom, TicketRoomMessage } from "../api/ticketing"

/**
 * TicketRoomThread — threaded conversation inside a ticket room.
 *
 * Mirrors the inbox conversation thread but is room-scoped and uses the
 * Ticketing service's lighter message model (no channel, no SLA).
 * Internal rooms are dimmed and badged "Private".
 */
export function TicketRoomThread({ room }: { room: TicketRoom }) {
  const messages = useRoomMessages(room.id)
  const add = useAddRoomMessage()
  const [text, setText] = useState("")
  const [privateNote, setPrivateNote] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)
  const lastId = messages.data?.[messages.data.length - 1]?.id

  useEffect(() => {
    if (!scrollRef.current) return
    scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [lastId])

  const onSend = async () => {
    const body = text.trim()
    if (!body) return
    try {
      await add.mutateAsync({
        roomId: room.id,
        content: body,
        isPrivate: privateNote,
      })
      setText("")
    } catch {
      /* mutation cache toasts */
    }
  }

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault()
      void onSend()
    }
  }

  return (
    <div className="flex flex-col h-full bg-canvas">
      <header className="flex items-center gap-2 h-9 shrink-0 px-3 border-b border-border">
        <span className="text-md font-medium text-fg truncate flex-1">
          {room.name || "Discussion"}
        </span>
        <span
          className={cn(
            "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
            room.type === 2 /* CUSTOMER_REPLY */
              ? "bg-info-subtle text-info"
              : "bg-subtle text-fg-muted",
          )}
        >
          {room.type === 2 ? "Customer-facing" : "Internal"}
        </span>
        <span className="font-mono text-xs text-fg-muted">
          {room.messageCount} msg
        </span>
      </header>

      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-3 space-y-2">
        {messages.isPending ? (
          <ListRowSkeleton count={4} />
        ) : messages.isError ? (
          <ErrorState
            title="Couldn't load messages"
            description={(messages.error as Error).message}
            onRetry={() => void messages.refetch()}
          />
        ) : messages.data!.length === 0 ? (
          <div className="text-center text-base text-fg-muted py-8">
            No messages yet.
          </div>
        ) : (
          messages.data!.map((m) => <Bubble key={m.id} message={m} />)
        )}
      </div>

      <div className="border-t border-border bg-surface">
        <div className="flex items-center gap-1 px-3 pt-2">
          <button
            type="button"
            onClick={() => setPrivateNote(false)}
            aria-pressed={!privateNote}
            className={cn(
              "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm",
              "text-fg-muted hover:text-fg hover:bg-subtle",
              !privateNote && "bg-accent-subtle text-fg",
            )}
          >
            Reply
          </button>
          <button
            type="button"
            onClick={() => setPrivateNote(true)}
            aria-pressed={privateNote}
            className={cn(
              "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm",
              "text-fg-muted hover:text-fg hover:bg-subtle",
              privateNote && "bg-warn-subtle text-warn",
            )}
          >
            <IconLock className="size-3" />
            Private
          </button>
        </div>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={onKeyDown}
          rows={3}
          placeholder={
            privateNote
              ? "Private note. Not visible to the customer."
              : "Write a message…  ⌘+Enter to send"
          }
          className={cn(
            "w-full resize-none px-3 py-2 bg-transparent outline-none",
            "text-base text-fg placeholder:text-fg-muted",
            privateNote && "bg-warn-subtle",
          )}
        />
        <div className="flex items-center gap-2 px-3 pb-2">
          <button
            type="button"
            onClick={() => void onSend()}
            disabled={add.isPending || !text.trim()}
            className={cn(
              "ms-auto inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md font-medium",
              "bg-accent text-accent-fg hover:bg-accent-hover",
              "disabled:opacity-40 disabled:pointer-events-none transition-colors",
            )}
          >
            {add.isPending ? (
              <IconLoader2 className="size-3.5 animate-spin" />
            ) : (
              <IconSend className="size-3.5" />
            )}
            Send
            <Kbd keys={["Cmd", "Enter"]} size="sm" />
          </button>
        </div>
      </div>
    </div>
  )
}

function Bubble({ message }: { message: TicketRoomMessage }) {
  return (
    <div
      className={cn(
        "rounded-md px-3 py-2 max-w-2xl",
        message.isPrivate
          ? "bg-warn-subtle border border-warn/20"
          : "bg-surface border border-border",
      )}
    >
      <div className="text-base text-fg whitespace-pre-wrap break-words">
        {message.content}
      </div>
      <div className="mt-1 flex items-center gap-1.5 text-xs text-fg-muted font-mono">
        <span className="truncate">{message.authorId.slice(0, 8)}</span>
        <span aria-hidden>·</span>
        <span>{relativeShort(message.createdAt)}</span>
        {message.isPrivate && (
          <>
            <span aria-hidden>·</span>
            <span className="inline-flex items-center gap-0.5">
              <IconLock className="size-2.5" />
              Private
            </span>
          </>
        )}
      </div>
    </div>
  )
}
