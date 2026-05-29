import { useState } from "react"
import { Link } from "react-router-dom"
import {
  IconArrowLeft,
  IconDeviceFloppy,
  IconPlayerPlay,
  IconLoader2,
  IconBolt,
} from "@tabler/icons-react"
import { toast } from "sonner"
import { ErrorState, CardSkeleton } from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useWorkflow,
  useUpdateWorkflow,
  useTestRunWorkflow,
} from "../hooks/queries"
import { WorkflowCanvas } from "./WorkflowCanvas"
import { RunHistory } from "./RunHistory"
import type { WorkflowSpec, ActionNode } from "../utils/types"
import type { Condition, TriggerType } from "../utils/actions"

/**
 * WorkflowDetail — the full edit surface for one workflow.
 *
 *   ┌────────────────────────────────────────────┐
 *   │ Header (name · status · save · test run)   │
 *   ├──────────────────────┬─────────────────────┤
 *   │ Canvas (React Flow)  │ Runs history        │
 *   └──────────────────────┴─────────────────────┘
 */
export function WorkflowDetail({ workflowId }: { workflowId: string }) {
  const wf = useWorkflow(workflowId)

  if (wf.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
      </div>
    )
  }

  if (wf.isError || !wf.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load workflow"
          description={(wf.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void wf.refetch()}
        />
      </div>
    )
  }

  return <Editor key={wf.data.id} initial={wf.data} />
}

function Editor({ initial }: { initial: WorkflowSpec }) {
  const update = useUpdateWorkflow()
  const testRun = useTestRunWorkflow()
  const [spec, setSpec] = useState<WorkflowSpec>(initial)
  const [graph, setGraph] = useState<{
    trigger: { type: TriggerType; config?: Record<string, unknown> } | null
    conditions: Condition[]
    actions: ActionNode[]
  }>({
    trigger: initial.trigger,
    conditions: initial.conditions,
    actions: initial.actions,
  })
  const [dirty, setDirty] = useState(false)

  const onSave = async () => {
    const next: WorkflowSpec = {
      ...spec,
      trigger: graph.trigger,
      conditions: graph.conditions,
      actions: graph.actions,
    }
    try {
      await update.mutateAsync(next)
      setDirty(false)
      toast.success("Workflow saved")
    } catch {
      /* mutationCache toasts */
    }
  }

  const onTestRun = async () => {
    try {
      const res = await testRun.mutateAsync({
        workflowId: spec.id,
        sampleEvent: { type: graph.trigger?.type ?? "test", id: "evt-test" },
      })
      toast.success(`Test run started · ${res.temporalRunId.slice(0, 12)}…`)
    } catch {
      /* mutationCache toasts */
    }
  }

  const onPatch = (patch: Partial<WorkflowSpec>) => {
    setSpec((s) => ({ ...s, ...patch }))
    setDirty(true)
  }

  return (
    <div className="h-full flex flex-col">
      <header className="flex items-center gap-3 px-4 py-2 border-b border-border bg-canvas">
        <Link
          to="/automation"
          className={cn(
            "inline-flex items-center justify-center size-7 rounded-sm",
            "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
          )}
          aria-label="Back to workflows"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className={cn(
            "size-8 rounded-md flex items-center justify-center shrink-0",
            "bg-accent-subtle text-fg-secondary",
          )}
          aria-hidden
        >
          <IconBolt className="size-4" />
        </div>
        <div className="min-w-0 flex-1">
          <Input
            value={spec.name}
            onChange={(e) => onPatch({ name: e.target.value })}
            placeholder="Workflow name"
            className="h-8 text-lg font-semibold border-0 bg-transparent shadow-none px-0"
          />
        </div>
        <select
          value={spec.status}
          onChange={(e) =>
            onPatch({ status: e.target.value as WorkflowSpec["status"] })
          }
          className="h-8 rounded-sm border border-border bg-surface px-2 text-md"
        >
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="inactive">Paused</option>
        </select>
        <button
          type="button"
          onClick={() => void onTestRun()}
          disabled={testRun.isPending || !graph.trigger}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md",
            "border border-border hover:bg-subtle",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          {testRun.isPending ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconPlayerPlay className="size-3.5" />
          )}
          Test run
        </button>
        <button
          type="button"
          onClick={() => void onSave()}
          disabled={!dirty || update.isPending}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          {update.isPending ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconDeviceFloppy className="size-3.5" />
          )}
          {dirty ? "Save" : "Saved"}
        </button>
      </header>

      <div className="flex-1 min-h-0 flex">
        <div className="flex-1 min-w-0">
          <WorkflowCanvas
            initial={spec}
            onChange={(next) => {
              setGraph(next)
              setDirty(true)
            }}
          />
        </div>
        <div className="w-80 shrink-0 border-s border-border">
          <RunHistory workflowId={spec.id} />
        </div>
      </div>
    </div>
  )
}
