import { Link } from "react-router-dom"
import { IconPlus, IconBolt, IconCircleCheck, IconCircle } from "@tabler/icons-react"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useWorkflows, useCreateWorkflow } from "../hooks/queries"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import type { WorkflowSpec } from "../utils/types"
import { getTriggerSpec } from "../utils/actions"

export function WorkflowsList() {
  const list = useWorkflows()
  const create = useCreateWorkflow()
  const navigate = useNavigate()

  const onNew = async () => {
    try {
      const wf = await create.mutateAsync({
        name: "Untitled workflow",
        description: "",
        status: "draft",
        trigger: null,
        conditions: [],
        actions: [],
      })
      navigate(`/automation/${wf.id}`)
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Automation"
        description="Temporal-backed workflows that fire on events"
        actions={
          <button
            type="button"
            onClick={() => void onNew()}
            disabled={create.isPending}
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover disabled:opacity-50"
          >
            <IconPlus className="size-3.5" />
            New workflow
          </button>
        }
      />

      {list.isPending ? (
        <div className="flex-1 p-2">
          <ListRowSkeleton count={8} />
        </div>
      ) : list.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load workflows"
            description={(list.error as Error).message}
            onRetry={() => void list.refetch()}
          />
        </div>
      ) : (list.data?.items.length ?? 0) === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconBolt className="size-5" />}
            title="No workflows yet"
            description="Automate routing, replies, notifications, and integrations. Start with a template or build from scratch."
            action={
              <button
                type="button"
                onClick={() => void onNew()}
                className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover"
              >
                <IconPlus className="size-3.5" />
                New workflow
              </button>
            }
          />
        </div>
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {list.data!.items.map((w) => (
            <li key={w.id}>
              <WorkflowRow w={w} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function WorkflowRow({ w }: { w: WorkflowSpec }) {
  const trigger = w.trigger ? getTriggerSpec(w.trigger.type) : null

  return (
    <Link
      to={`/automation/${w.id}`}
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <div
        className={cn(
          "size-7 rounded-sm flex items-center justify-center shrink-0",
          "bg-accent-subtle text-fg-secondary",
        )}
        aria-hidden
      >
        <IconBolt className="size-3.5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-md font-medium text-fg truncate">{w.name}</div>
        <div className="text-sm text-fg-muted truncate">
          {trigger ? trigger.label : "No trigger"} ·{" "}
          {w.actions.length} action{w.actions.length === 1 ? "" : "s"}
        </div>
      </div>
      <StatusBadge status={w.status} />
    </Link>
  )
}

function StatusBadge({ status }: { status: string }) {
  if (status === "active") {
    return (
      <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-success-subtle text-success">
        <IconCircleCheck className="size-3" />
        Active
      </span>
    )
  }
  if (status === "inactive") {
    return (
      <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-warn-subtle text-warn">
        Paused
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-subtle text-fg-muted">
      <IconCircle className="size-3" />
      Draft
    </span>
  )
}
