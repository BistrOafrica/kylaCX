import { cn } from "@/lib/utils"
import { AgentStatusButton } from "@/features/admin/components/AgentStatusButton"

/**
 * StatusBar — 28px strip at the bottom of the shell.
 *
 *   [● Available ▾]   SLA risk: 2   Active: 5   Queue: 14            v1.0.0
 *
 * Counts are placeholders sourced from the workspace store (always
 * zero until streaming-driven counters land in F5.x). Agent status
 * comes from AgentStatusService via the AgentStatusButton.
 */

const APP_VERSION = "0.5.0-f5"

export function StatusBar() {
  return (
    <footer
      role="contentinfo"
      className={cn(
        "h-7 shrink-0 flex items-center gap-4 px-3",
        "bg-canvas border-t border-border text-sm",
      )}
    >
      <AgentStatusButton />

      <Stat label="SLA risk" value={0} tone="danger" />
      <Stat label="Active" value={0} />
      <Stat label="Queue" value={0} />

      <span className="ms-auto font-mono text-xs text-fg-muted">
        v{APP_VERSION}
      </span>
    </footer>
  )
}

function Stat({
  label,
  value,
  tone,
}: {
  label: string
  value: number
  tone?: "danger" | "warn"
}) {
  const isZero = value === 0
  return (
    <div className="inline-flex items-center gap-1.5 text-fg-muted">
      <span className="text-sm">{label}:</span>
      <span
        className={cn(
          "font-mono text-sm font-medium",
          tone === "danger" && !isZero && "text-danger",
          tone === "warn" && !isZero && "text-warn",
          (!tone || isZero) && "text-fg-secondary",
        )}
      >
        {value}
      </span>
    </div>
  )
}
