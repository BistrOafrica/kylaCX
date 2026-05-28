import { useEffect, useState } from "react"
import { cn } from "@/lib/utils"
import { formatSlaWindow } from "../utils/format"

/**
 * SlaBadge — Linear-dense chip showing time-to-SLA with breach styling.
 *
 *   <SlaBadge deadline={conv.slaDeadline} />
 *
 * Ticks every 30s so the countdown stays fresh without burning CPU.
 */
export function SlaBadge({
  deadline,
  size = "md",
}: {
  deadline?: string | null
  size?: "sm" | "md"
}) {
  const [tick, setTick] = useState(0)

  useEffect(() => {
    if (!deadline) return
    const id = setInterval(() => setTick((t) => t + 1), 30_000)
    return () => clearInterval(id)
  }, [deadline])

  // Reference `tick` so React knows the value is read and re-renders.
  void tick

  const sla = formatSlaWindow(deadline)
  if (!sla) return null

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-xs font-mono font-medium",
        size === "sm" ? "h-4 px-1 text-[10px]" : "h-5 px-1.5 text-[11px]",
        sla.breached
          ? "bg-danger-subtle text-danger"
          : "bg-warn-subtle text-warn",
      )}
      title={`SLA · ${sla.text}`}
    >
      SLA · {sla.text}
    </span>
  )
}
