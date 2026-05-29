import { useMemo, useState } from "react"
import {
  IconCircleCheck,
  IconCircleX,
  IconLoader2,
  IconCircleDashed,
  IconRefresh,
} from "@tabler/icons-react"
import { EmptyState, ErrorState, ListRowSkeleton } from "@/design-system"
import { cn } from "@/lib/utils"
import { useWorkflowRuns } from "../hooks/queries"
import type { WorkflowRunRow } from "../utils/types"
import { relativeShort } from "@/features/crm/utils/relative"

type StatusFilter = "all" | "success" | "running" | "failed"

/**
 * RunHistory — projection of Temporal workflow executions for the
 * active workflow. Polls every 3s while there are in-flight runs.
 */
export function RunHistory({ workflowId }: { workflowId: string }) {
  const runs = useWorkflowRuns(workflowId)
  const [filter, setFilter] = useState<StatusFilter>("all")

  const filtered = useMemo(() => {
    const all = runs.data?.runs ?? []
    if (filter === "all") return all
    if (filter === "running")
      return all.filter((r) => r.status === "running" || r.status === "pending")
    return all.filter((r) => r.status === filter)
  }, [runs.data, filter])

  return (
    <div className="flex flex-col h-full bg-canvas">
      <header className="flex items-center gap-2 px-3 h-9 border-b border-border bg-canvas">
        <span className="text-md font-medium text-fg">Run history</span>
        <span className="font-mono text-xs text-fg-muted">
          {runs.data?.total ?? 0}
        </span>
        <div className="ms-auto flex items-center gap-1">
          {(["all", "success", "running", "failed"] as const).map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setFilter(f)}
              aria-pressed={filter === f}
              className={cn(
                "inline-flex items-center h-6 px-2 rounded-xs text-sm",
                filter === f
                  ? "bg-accent-subtle text-fg"
                  : "text-fg-muted hover:text-fg hover:bg-subtle",
              )}
            >
              {f}
            </button>
          ))}
          <button
            type="button"
            onClick={() => void runs.refetch()}
            aria-label="Refresh"
            className="inline-flex items-center justify-center size-6 rounded-xs text-fg-muted hover:text-fg hover:bg-subtle"
          >
            <IconRefresh
              className={cn("size-3.5", runs.isFetching && "animate-spin")}
            />
          </button>
        </div>
      </header>

      {runs.isPending ? (
        <div className="flex-1 p-2">
          <ListRowSkeleton count={6} />
        </div>
      ) : runs.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load runs"
            description={(runs.error as Error).message}
            onRetry={() => void runs.refetch()}
          />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconCircleDashed className="size-5" />}
            title="No runs yet"
            description="Trigger this workflow or use the Test run button to see executions here."
            size="sm"
          />
        </div>
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {filtered.map((run) => (
            <RunRow key={run.id} run={run} />
          ))}
        </ul>
      )}
    </div>
  )
}

function RunRow({ run }: { run: WorkflowRunRow }) {
  return (
    <li
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <StatusIcon status={run.status} />
      <div className="min-w-0 flex-1">
        <div className="font-mono text-sm text-fg truncate">
          {run.temporalRunId || run.id.slice(0, 12)}
        </div>
        <div className="text-sm text-fg-muted truncate">
          Trigger event {run.triggerEventId.slice(0, 12)}
          {run.error ? ` · ${run.error}` : ""}
        </div>
      </div>
      <span className="font-mono text-xs text-fg-muted">
        {run.startedAt ? relativeShort(run.startedAt) : ""}
      </span>
    </li>
  )
}

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case "success":
      return (
        <span className="inline-flex items-center justify-center size-6 rounded-xs bg-success-subtle text-success">
          <IconCircleCheck className="size-3.5" />
        </span>
      )
    case "failed":
      return (
        <span className="inline-flex items-center justify-center size-6 rounded-xs bg-danger-subtle text-danger">
          <IconCircleX className="size-3.5" />
        </span>
      )
    case "running":
    case "pending":
      return (
        <span className="inline-flex items-center justify-center size-6 rounded-xs bg-info-subtle text-info">
          <IconLoader2 className="size-3.5 animate-spin" />
        </span>
      )
    default:
      return (
        <span className="inline-flex items-center justify-center size-6 rounded-xs bg-subtle text-fg-muted">
          <IconCircleDashed className="size-3.5" />
        </span>
      )
  }
}
