import * as React from "react"
import { IconAlertTriangle } from "@tabler/icons-react"
import { cn } from "@/lib/utils"

/**
 * ErrorState — failure surface for lists, panels, and full pages.
 *
 *   <ErrorState
 *     title="Couldn't load conversations"
 *     description={error.message}
 *     onRetry={() => refetch()}
 *   />
 */
export interface ErrorStateProps {
  title?: React.ReactNode
  description?: React.ReactNode
  onRetry?: () => void
  retryLabel?: React.ReactNode
  size?: "sm" | "md" | "lg"
  className?: string
}

export function ErrorState({
  title = "Something went wrong",
  description,
  onRetry,
  retryLabel = "Try again",
  size = "md",
  className,
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className={cn(
        "flex flex-col items-center justify-center text-center",
        size === "sm" && "py-8 gap-2",
        size === "md" && "py-12 gap-3",
        size === "lg" && "py-20 gap-4",
        className,
      )}
    >
      <div
        className={cn(
          "flex items-center justify-center rounded-md",
          "bg-danger-subtle text-danger",
          size === "sm" && "size-9",
          size === "md" && "size-11",
          size === "lg" && "size-14",
        )}
        aria-hidden
      >
        <IconAlertTriangle className="size-5" />
      </div>
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
          <div className="text-fg-muted max-w-md text-base break-words">
            {description}
          </div>
        )}
      </div>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-surface border border-border hover:bg-subtle transition-colors",
          )}
        >
          {retryLabel}
        </button>
      )}
    </div>
  )
}
