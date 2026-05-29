import { memo } from "react"
import { Handle, Position, type NodeProps } from "@xyflow/react"
import { IconFilter } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { type Condition } from "../utils/actions"

export interface ConditionNodeData extends Record<string, unknown> {
  conditions: Condition[]
}

function ConditionNodeImpl({ data, selected }: NodeProps) {
  const d = data as ConditionNodeData
  const conditions = d.conditions ?? []

  return (
    <div
      className={cn(
        "min-w-[220px] rounded-md border-2 bg-surface shadow-elev-1",
        selected ? "border-accent" : "border-border",
      )}
      data-testid="condition-node"
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
            "bg-info-subtle text-info",
          )}
          aria-hidden
        >
          <IconFilter className="size-3.5" />
        </span>
        <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          If
        </span>
      </header>
      <div className="px-3 py-2 space-y-1">
        {conditions.length === 0 ? (
          <div className="text-sm text-fg-muted">No conditions — always proceeds.</div>
        ) : (
          conditions.map((c, i) => (
            <div key={i} className="font-mono text-sm text-fg-secondary truncate">
              <span className="text-fg">{c.field || "?"}</span>{" "}
              <span className="text-fg-muted">{c.op}</span>{" "}
              <span className="text-fg">{stringifyValue(c.value)}</span>
            </div>
          ))
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

export const ConditionNode = memo(ConditionNodeImpl)

function stringifyValue(v: unknown): string {
  if (v === undefined || v === null) return ""
  if (typeof v === "string") return `"${v}"`
  if (Array.isArray(v)) return JSON.stringify(v)
  return String(v)
}
