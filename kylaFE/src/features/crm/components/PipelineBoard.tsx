import { useMemo, useState } from "react"
import {
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  useDroppable,
  type DragEndEvent,
  type DragStartEvent,
  DragOverlay,
} from "@dnd-kit/core"
import { IconLayoutKanban, IconPlus, IconChevronDown } from "@tabler/icons-react"
import { toast } from "sonner"
import { EmptyState, ErrorState, PageHeader, CardSkeleton } from "@/design-system"
import { cn } from "@/lib/utils"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@/components/ui/dropdown-menu"
import {
  usePipelineBoard,
  usePipelines,
  useMoveDeal,
} from "../hooks/queries"
import { DealCardTile } from "./DealCard"
import type { PipelineBoardColumn } from "@/pb/crm"
import type { DealCard as DealCardModel } from "@/pb/crm"

/**
 * PipelineBoard — drag-to-stage kanban over CRM pipelines.
 *
 *   <PipelineBoard basePath="/crm/deals" />
 *
 * The active pipeline is held in local state; the picker lets the
 * user switch between pipelines listed for the workspace. dnd-kit
 * powers the drag layer; on drop we fire MoveDeal and rely on cache
 * invalidation to refetch the board.
 */
export function PipelineBoard({ basePath }: { basePath: string }) {
  const pipelines = usePipelines()
  const [activeId, setActiveId] = useState<string | null>(null)
  const [draggingDealId, setDraggingDealId] = useState<string | null>(null)
  const moveDeal = useMoveDeal()

  // Lock the first pipeline once we have data.
  const effectiveId = activeId ?? pipelines.data?.[0]?.id ?? null
  const board = usePipelineBoard(effectiveId)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
  )

  const dealById = useMemo(() => {
    const map = new Map<string, DealCardModel>()
    for (const col of board.data?.columns ?? []) {
      for (const d of col.deals) map.set(d.objectId, d)
    }
    return map
  }, [board.data])

  if (pipelines.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={2} />
      </div>
    )
  }

  if (pipelines.isError) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load pipelines"
          description={(pipelines.error as Error).message}
          onRetry={() => void pipelines.refetch()}
        />
      </div>
    )
  }

  if (!pipelines.data?.length) {
    return (
      <div className="h-full flex flex-col bg-canvas">
        <PageHeader title="Deals" description="No pipelines yet" />
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconLayoutKanban className="size-5" />}
            title="Create a pipeline"
            description="Pipelines hold the stages that deals move through (e.g. New → Qualified → Won)."
            action={
              <button className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg">
                <IconPlus className="size-3.5" />
                New pipeline
              </button>
            }
          />
        </div>
      </div>
    )
  }

  const onDragStart = (e: DragStartEvent) =>
    setDraggingDealId(String(e.active.id))

  const onDragEnd = (e: DragEndEvent) => {
    setDraggingDealId(null)
    if (!e.over) return
    const dealObjectId = String(e.active.id)
    const stageId = String(e.over.id).replace(/^stage:/, "")
    if (!stageId || stageId === e.active.data.current?.stageId) return
    moveDeal.mutate(
      { dealObjectId, stageId },
      {
        onError: (err) => toast.error((err as Error).message),
      },
    )
  }

  return (
    <div className="h-full flex flex-col bg-canvas">
      <PageHeader
        title="Deals"
        description={
          board.data?.pipeline?.description ||
          board.data?.pipeline?.name ||
          "Pipeline kanban"
        }
        actions={
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger
                className={cn(
                  "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
                  "border border-border hover:bg-subtle",
                )}
              >
                {board.data?.pipeline?.name ?? "Pipeline"}
                <IconChevronDown className="size-3.5 text-fg-muted" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuRadioGroup
                  value={effectiveId ?? ""}
                  onValueChange={(v) => setActiveId(v)}
                >
                  {pipelines.data.map((p) => (
                    <DropdownMenuRadioItem key={p.id} value={p.id}>
                      {p.name}
                    </DropdownMenuRadioItem>
                  ))}
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            <button className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover">
              <IconPlus className="size-3.5" />
              New deal
            </button>
          </div>
        }
      />

      <div className="flex-1 overflow-hidden">
        {board.isPending ? (
          <div className="p-6 grid grid-cols-3 gap-3">
            <CardSkeleton />
            <CardSkeleton />
            <CardSkeleton />
          </div>
        ) : board.isError ? (
          <div className="p-12">
            <ErrorState
              title="Couldn't load board"
              description={(board.error as Error).message}
              onRetry={() => void board.refetch()}
            />
          </div>
        ) : (
          <DndContext
            sensors={sensors}
            onDragStart={onDragStart}
            onDragEnd={onDragEnd}
          >
            <div className="h-full overflow-x-auto p-3">
              <div className="flex gap-3 h-full">
                {board.data?.columns.map((col) => (
                  <Column key={col.stage?.id} column={col} basePath={basePath} />
                ))}
              </div>
            </div>
            <DragOverlay>
              {draggingDealId && dealById.get(draggingDealId) ? (
                <div className="w-72 rotate-1">
                  <DealCardTile
                    deal={dealById.get(draggingDealId)!}
                    basePath={basePath}
                  />
                </div>
              ) : null}
            </DragOverlay>
          </DndContext>
        )}
      </div>
    </div>
  )
}

function Column({
  column,
  basePath,
}: {
  column: PipelineBoardColumn
  basePath: string
}) {
  const id = `stage:${column.stage?.id ?? "unknown"}`
  const { isOver, setNodeRef } = useDroppable({ id })
  const stage = column.stage

  const colorStripe: React.CSSProperties = stage?.color
    ? { background: stage.color }
    : { background: "var(--accent-solid)" }

  return (
    <section
      ref={setNodeRef}
      className={cn(
        "w-72 shrink-0 flex flex-col bg-surface border border-border rounded-md overflow-hidden",
        isOver && "border-accent shadow-elev-2",
      )}
      aria-label={`Stage: ${stage?.name ?? "Unknown"}`}
    >
      <header className="flex items-center gap-2 px-3 h-9 border-b border-border">
        <span className="block w-1.5 h-3.5 rounded-xs" style={colorStripe} aria-hidden />
        <span className="text-md font-medium text-fg truncate flex-1">
          {stage?.name ?? "Unknown"}
        </span>
        <span className="font-mono text-xs text-fg-muted">{column.total}</span>
        {stage?.probability !== undefined && (
          <span className="font-mono text-xs text-fg-muted">
            {stage.probability}%
          </span>
        )}
      </header>
      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {column.deals.length === 0 ? (
          <div className="text-sm text-fg-muted text-center py-4">
            Drop deals here
          </div>
        ) : (
          column.deals.map((deal) => (
            <DealCardTile key={deal.objectId} deal={deal} basePath={basePath} />
          ))
        )}
      </div>
    </section>
  )
}
