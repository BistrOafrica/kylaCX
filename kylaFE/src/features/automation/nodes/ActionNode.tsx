import { memo } from "react"
import { Handle, Position, type NodeProps } from "@xyflow/react"
import { cn } from "@/lib/utils"
import { getActionSpec, type ActionType } from "../utils/actions"

export interface ActionNodeData extends Record<string, unknown> {
  actionType: ActionType
  params: Record<string, unknown>
}

function ActionNodeImpl({ data, selected }: NodeProps) {
  const d = data as ActionNodeData
  const spec = getActionSpec(d.actionType)
  const Icon = spec?.icon

  return (
    <div
      className={cn(
        "min-w-[220px] rounded-md border-2 bg-surface shadow-elev-1",
        selected ? "border-accent" : "border-border",
      )}
      data-testid="action-node"
      data-action-type={d.actionType}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!size-2 !bg-accent !border-2 !border-surface"
      />
      <header className="flex items-center gap-2 px-3 h-8 border-b border-border-subtle bg-subtle rounded-t-[4px]">
        <span
          className={cn(
            "inline-flex items-center justify-center size-5 rounded-xs",
            "bg-accent-subtle",
          )}
          style={{ color: spec?.color }}
          aria-hidden
        >
          {Icon ? <Icon className="size-3.5" /> : null}
        </span>
        <span className="text-md font-medium text-fg flex-1 truncate">
          {spec?.label ?? d.actionType}
        </span>
      </header>
      <div className="px-3 py-2 space-y-0.5">
        {spec ? (
          <ParamPreview params={d.params} keys={spec.params.slice(0, 2).map((p) => p.key)} />
        ) : (
          <div className="text-sm text-fg-muted">Unknown action</div>
        )}
      </div>
      <Handle
        type="source"
        position={Position.Bottom}
        className="!size-2 !bg-accent !border-2 !border-surface"
      />
    </div>
  )
}

export const ActionNode = memo(ActionNodeImpl)

function ParamPreview({
  params,
  keys,
}: {
  params: Record<string, unknown>
  keys: string[]
}) {
  const entries = keys
    .map((k) => [k, params[k]] as const)
    .filter(([, v]) => v !== undefined && v !== "" && v !== null)

  if (entries.length === 0) {
    return <div className="text-sm text-fg-muted">Click to configure…</div>
  }

  return (
    <ul className="space-y-0.5">
      {entries.map(([k, v]) => (
        <li key={k} className="text-sm font-mono text-fg-secondary truncate">
          <span className="text-fg-muted">{k}:</span>{" "}
          <span className="text-fg">{previewValue(v)}</span>
        </li>
      ))}
    </ul>
  )
}

function previewValue(v: unknown): string {
  if (typeof v === "string") return v.length > 32 ? v.slice(0, 32) + "…" : v
  if (typeof v === "object") {
    const s = JSON.stringify(v)
    return s.length > 32 ? s.slice(0, 32) + "…" : s
  }
  return String(v)
}
