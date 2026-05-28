import { useMemo, useRef, useEffect, useState } from "react"
import { useParams } from "react-router-dom"
import { useVirtualizer } from "@tanstack/react-virtual"
import { IconInbox, IconRefresh } from "@tabler/icons-react"
import { ListRowSkeleton, EmptyState, ErrorState } from "@/design-system"
import { cn } from "@/lib/utils"
import { useInbox } from "../hooks/queries"
import { useInboxStream } from "../hooks/useInboxStream"
import { InboxRow } from "./InboxRow"
import { InboxFiltersBar } from "./InboxFilters"
import type { InboxFilters } from "../api/conversations"
import type { Conversation } from "@/pb/conversations"

const ROW_HEIGHT = 32

/**
 * InboxList — virtualized list of conversations + filters + realtime.
 *
 * The list owns its filter state (held in a small URL-less store; F1.x
 * can persist via search params if agents start sharing links). Rows
 * are virtualized via TanStack Virtual so 10k+ conversations stay
 * smooth.
 */
export function InboxList() {
  const { id: selectedId } = useParams()
  const [filters, setFilters] = useState<InboxFilters>({ activeOnly: true })
  useInboxStream(true)

  const inbox = useInbox(filters)

  const allConversations = useMemo<Conversation[]>(() => {
    return inbox.data?.pages.flatMap((p) => p.conversations) ?? []
  }, [inbox.data])

  const total = inbox.data?.pages[inbox.data.pages.length - 1]?.total

  const parentRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: allConversations.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 12,
  })

  // Infinite scroll: fetch next page when the last virtual item is in view.
  useEffect(() => {
    const items = virtualizer.getVirtualItems()
    const last = items[items.length - 1]
    if (!last) return
    if (
      last.index >= allConversations.length - 4 &&
      inbox.hasNextPage &&
      !inbox.isFetchingNextPage
    ) {
      void inbox.fetchNextPage()
    }
  }, [virtualizer, allConversations.length, inbox])

  return (
    <div className="flex flex-col h-full bg-canvas">
      <InboxFiltersBar
        value={filters}
        onChange={setFilters}
        total={total}
        loaded={allConversations.length}
      />

      {inbox.isPending ? (
        <div className="flex-1 overflow-hidden p-2">
          <ListRowSkeleton count={12} />
        </div>
      ) : inbox.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load inbox"
            description={inbox.error.message}
            onRetry={() => void inbox.refetch()}
          />
        </div>
      ) : allConversations.length === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconInbox className="size-5" />}
            title="Inbox zero"
            description="No conversations match these filters. Adjust filters or wait for a new message."
            action={
              <button
                type="button"
                onClick={() => void inbox.refetch()}
                className="inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md border border-border hover:bg-subtle"
              >
                <IconRefresh className="size-3.5" />
                Refresh
              </button>
            }
          />
        </div>
      ) : (
        <div
          ref={parentRef}
          className={cn("flex-1 overflow-y-auto", "outline-none")}
          tabIndex={-1}
          role="listbox"
          aria-label="Inbox"
        >
          <div
            style={{
              height: virtualizer.getTotalSize(),
              position: "relative",
            }}
          >
            {virtualizer.getVirtualItems().map((item) => {
              const conv = allConversations[item.index]
              if (!conv) return null
              return (
                <div
                  key={conv.id}
                  data-index={item.index}
                  style={{
                    position: "absolute",
                    top: 0,
                    insetInlineStart: 0,
                    insetInlineEnd: 0,
                    transform: `translateY(${item.start}px)`,
                    height: item.size,
                  }}
                >
                  <InboxRow
                    conv={conv}
                    selected={selectedId === conv.id}
                  />
                </div>
              )
            })}
          </div>
          {inbox.isFetchingNextPage && (
            <div className="p-2">
              <ListRowSkeleton count={3} />
            </div>
          )}
        </div>
      )}
    </div>
  )
}