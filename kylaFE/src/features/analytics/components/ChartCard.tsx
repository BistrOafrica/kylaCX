import { Surface, CardSkeleton, ErrorState, EmptyState } from "@/design-system"
import { cn } from "@/lib/utils"

/**
 * ChartCard — wraps a chart with title + loading / error / empty states.
 *
 *   <ChartCard title="Ticket volume" subtitle="Last 30 days" isPending={…}>
 *     <ResponsiveContainer> … </ResponsiveContainer>
 *   </ChartCard>
 */
export interface ChartCardProps {
  title: React.ReactNode
  subtitle?: React.ReactNode
  actions?: React.ReactNode
  isPending?: boolean
  isError?: boolean
  errorMessage?: string
  onRetry?: () => void
  isEmpty?: boolean
  emptyMessage?: string
  height?: number
  className?: string
  children?: React.ReactNode
}

export function ChartCard({
  title,
  subtitle,
  actions,
  isPending,
  isError,
  errorMessage,
  onRetry,
  isEmpty,
  emptyMessage = "No data for this period",
  height = 240,
  className,
  children,
}: ChartCardProps) {
  return (
    <Surface level={1} radius="md" className={cn("flex flex-col", className)}>
      <header className="flex items-start gap-2 px-4 py-3 border-b border-border">
        <div className="min-w-0 flex-1">
          <div className="text-md font-medium text-fg truncate">{title}</div>
          {subtitle && (
            <div className="text-sm text-fg-muted truncate">{subtitle}</div>
          )}
        </div>
        {actions && <div className="flex items-center gap-1.5 shrink-0">{actions}</div>}
      </header>
      <div className="p-3" style={{ minHeight: height }}>
        {isPending ? (
          <CardSkeleton lines={4} />
        ) : isError ? (
          <ErrorState
            title="Couldn't load chart"
            description={errorMessage}
            onRetry={onRetry}
            size="sm"
          />
        ) : isEmpty ? (
          <EmptyState title={emptyMessage} size="sm" />
        ) : (
          children
        )}
      </div>
    </Surface>
  )
}
