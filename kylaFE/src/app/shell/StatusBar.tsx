import { useTranslation } from "react-i18next"
import { StatusDot } from "@/design-system"
import { cn } from "@/lib/utils"

/**
 * StatusBar — 28px strip at the bottom of the shell.
 *
 *   [● Available ▾]   SLA risk: 2   Active: 5   Queue: 14            v1.0.0
 *
 * Counts in F0 are placeholders sourced from the workspace store
 * (always zero until F1 wires real data). Agent status menu opens the
 * AgentOps presence picker — also F1.
 */

const APP_VERSION = "0.1.0-f0"

export function StatusBar() {
  const { t } = useTranslation()

  return (
    <footer
      role="contentinfo"
      className={cn(
        "h-7 shrink-0 flex items-center gap-4 px-3",
        "bg-canvas border-t border-border text-sm",
      )}
    >
      <button
        type="button"
        className={cn(
          "inline-flex items-center gap-1.5 h-5 px-1.5 rounded-xs",
          "hover:bg-subtle text-fg-secondary transition-colors",
        )}
        aria-label={t("shell.agentStatus")}
      >
        <StatusDot tone="online" size={6} />
        <span className="text-sm">{t("agentStatus.online")}</span>
      </button>

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
