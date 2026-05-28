import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * EmptyState — used everywhere a list, view, or surface has no data.
 *
 *   <EmptyState
 *     icon={<IconInbox />}
 *     title="Inbox zero"
 *     description="Nothing waiting for you. New conversations will appear here."
 *     action={<Button>Refresh</Button>}
 *   />
 */
export interface EmptyStateProps {
  icon?: React.ReactNode
  title: React.ReactNode
  description?: React.ReactNode
  action?: React.ReactNode
  className?: string
  size?: "sm" | "md" | "lg"
}

export function EmptyState({
  icon,
  title,
  description,
  action,
  size = "md",
  className,
}: EmptyStateProps) {
  return (
    <div
      role="status"
      className={cn(
        "flex flex-col items-center justify-center text-center",
        size === "sm" && "py-8 gap-2",
        size === "md" && "py-12 gap-3",
        size === "lg" && "py-20 gap-4",
        className,
      )}
    >
      {icon && (
        <div
          className={cn(
            "flex items-center justify-center rounded-md text-fg-muted",
            "bg-subtle border border-border",
            size === "sm" && "size-9",
            size === "md" && "size-11",
            size === "lg" && "size-14",
          )}
          aria-hidden
        >
          {icon}
        </div>
      )}
      <div className="space-y-1">
        <div
          className={cn(
            "font-medium text-fg",
            size === "sm" && "text-md",
            size === "md" && "text-lg",
            size === "lg" && "text-xl",
          )}
        >
          {title}
        </div>
        {description && (
          <div className="text-fg-muted max-w-md text-base">
            {description}
          </div>
        )}
      </div>
      {action && <div className="pt-1">{action}</div>}
    </div>
  )
}
