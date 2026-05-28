import { useDraggable } from "@dnd-kit/core"
import { Link } from "react-router-dom"
import { IconUser } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import type { DealCard as DealCardModel } from "@/pb/crm"
import { relativeShort } from "../utils/relative"

/**
 * DealCardTile — single card on the pipeline kanban.
 *
 * Draggable via dnd-kit (drag-to-stage handled by the parent board).
 * Click navigates to the deal detail page (Object Core record).
 */
export function DealCardTile({
  deal,
  basePath,
}: {
  deal: DealCardModel
  basePath: string
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({ id: deal.objectId })

  const style = transform
    ? { transform: `translate3d(${transform.x}px, ${transform.y}px, 0)` }
    : undefined

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        "select-none",
        isDragging && "opacity-50 cursor-grabbing",
        !isDragging && "cursor-grab",
      )}
    >
      <Link
        to={`${basePath}/${deal.objectId}`}
        onClick={(e) => {
          // Don't open detail while a drag is starting.
          if (isDragging) e.preventDefault()
        }}
        className={cn(
          "block rounded-md border border-border bg-surface px-2.5 py-2",
          "hover:border-border-strong transition-colors",
        )}
      >
        <div className="text-md font-medium text-fg truncate">
          {deal.name || deal.objectId.slice(0, 8)}
        </div>
        <div className="mt-1.5 flex items-center justify-between gap-2 text-xs">
          <span className="font-mono tabular-nums text-fg-secondary">
            {formatValue(deal.value)}
          </span>
          {deal.closeDate && (
            <span className="font-mono text-fg-muted">
              {relativeShort(deal.closeDate)}
            </span>
          )}
        </div>
        {deal.assigneeId && (
          <div className="mt-1.5 flex items-center gap-1 text-xs text-fg-muted">
            <IconUser className="size-3" aria-hidden />
            <span className="font-mono truncate">
              {deal.assigneeId.slice(0, 8)}
            </span>
          </div>
        )}
      </Link>
    </div>
  )
}

function formatValue(v: string): string {
  if (!v) return "—"
  const n = Number(v)
  if (Number.isNaN(n)) return v
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(n)
}
