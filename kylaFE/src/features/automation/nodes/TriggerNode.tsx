import { memo } from "react"
import { Handle, Position, type NodeProps } from "@xyflow/react"
import { IconBolt } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { getTriggerSpec, type TriggerType } from "../utils/actions"

export interface TriggerNodeData extends Record<string, unknown> {
  triggerType: TriggerType | null
  config?: Record<string, unknown>
  selected?: boolean
}

function TriggerNodeImpl({ data, selected }: NodeProps) {
  const d = data as TriggerNodeData
  const spec = d.triggerType ? getTriggerSpec(d.triggerType) : null

  return (
    <div
      className={cn(
        "min-w-[200px] rounded-md border-2 bg-surface shadow-elev-1",
        selected ? "border-accent" : "border-border",
      )}
      data-testid="trigger-node"
    >
      <header className="flex items-center gap-2 px-3 h-8 border-b border-border-subtle bg-subtle rounded-t-[4px]">
        <span
          className={cn(
            "inline-flex items-center justify-center size-5 rounded-xs",
            "bg-warn-subtle text-warn",
          )}
          aria-hidden
        >
          <IconBolt className="size-3.5" />
        </span>
        <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          Trigger
        </span>
      </header>
      <div className="px-3 py-2">
        <div className="text-md font-medium text-fg">
          {spec ? spec.label : "Pick a trigger"}
        </div>
        {spec && (
          <div className="text-sm text-fg-muted">{spec.description}</div>
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

export const TriggerNode = memo(TriggerNodeImpl)
