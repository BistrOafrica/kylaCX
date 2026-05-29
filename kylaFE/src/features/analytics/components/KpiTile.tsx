import { Surface } from "@/design-system"
import { cn } from "@/lib/utils"
import { IconArrowUpRight, IconArrowDownRight, IconMinus } from "@tabler/icons-react"

/**
 * KpiTile — single metric card.
 *
 *   <KpiTile label="Total tickets" value={1284} hint="Last 30 days" />
 *   <KpiTile label="SLA compliance" value="94%" trend={2.4} />
 *
 * Trend is optional and rendered as a coloured delta indicator
 * (green when positive, rose when negative, muted when zero).
 */
export interface KpiTileProps {
  label: string
  value: React.ReactNode
  hint?: string
  trend?: number              // percent change vs prev period
  trendInvert?: boolean       // true when "down" is good (e.g. resolution time)
  pending?: boolean
}

export function KpiTile({
  label,
  value,
  hint,
  trend,
  trendInvert,
  pending,
}: KpiTileProps) {
  return (
    <Surface level={1} radius="md" className="p-3 space-y-1">
      <div className="text-xs font-mono uppercase tracking-wider text-fg-muted truncate">
        {label}
      </div>
      <div className="flex items-baseline gap-2">
        <div
          className={cn(
            "text-2xl font-semibold tabular-nums text-fg truncate",
            pending && "opacity-50",
          )}
        >
          {pending ? "—" : value}
        </div>
        {trend !== undefined && !pending && <TrendBadge value={trend} invert={trendInvert} />}
      </div>
      {hint && (
        <div className="text-sm text-fg-muted truncate">{hint}</div>
      )}
    </Surface>
  )
}

function TrendBadge({ value, invert }: { value: number; invert?: boolean }) {
  const dir = value > 0 ? "up" : value < 0 ? "down" : "flat"
  const goodWhenUp = !invert
  const good =
    dir === "flat"
      ? null
      : goodWhenUp
        ? dir === "up"
        : dir === "down"

  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 h-4 px-1 rounded-xs text-xs font-mono",
        dir === "flat" && "bg-subtle text-fg-muted",
        good === true && "bg-success-subtle text-success",
        good === false && "bg-danger-subtle text-danger",
      )}
    >
      {dir === "up" && <IconArrowUpRight className="size-3" />}
      {dir === "down" && <IconArrowDownRight className="size-3" />}
      {dir === "flat" && <IconMinus className="size-3" />}
      {Math.abs(value).toFixed(1)}%
    </span>
  )
}
