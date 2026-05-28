import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * Skeleton — shimmer placeholder used while data is loading.
 *
 *   <Skeleton className="h-8 w-32" />
 *
 * For composite loading layouts use the named skeletons below
 * (ListRowSkeleton, CardSkeleton). Always render at the same height
 * as the real content to avoid layout shift on swap.
 */
export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  shape?: "rect" | "circle" | "pill"
}

export function Skeleton({
  shape = "rect",
  className,
  ...rest
}: SkeletonProps) {
  return (
    <div
      aria-hidden
      className={cn(
        "animate-pulse bg-subtle",
        shape === "rect" && "rounded-sm",
        shape === "pill" && "rounded-full",
        shape === "circle" && "rounded-full aspect-square",
        className,
      )}
      {...rest}
    />
  )
}

/**
 * ListRowSkeleton — dense row placeholder matching inbox / list density.
 * Renders count rows; gap inferred from row-default.
 */
export function ListRowSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div role="status" aria-label="Loading" className="space-y-px">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="row-default flex items-center gap-3 px-3 border-b border-border-subtle"
        >
          <Skeleton shape="circle" className="size-1.5" />
          <Skeleton className="h-3.5 w-32" />
          <Skeleton className="h-3 w-12 ml-auto" />
          <Skeleton className="h-3 w-10" />
        </div>
      ))}
    </div>
  )
}

/**
 * CardSkeleton — composite card placeholder.
 */
export function CardSkeleton({ lines = 3 }: { lines?: number }) {
  return (
    <div className="surface rounded-md p-4 space-y-3" aria-label="Loading">
      <div className="flex items-center gap-3">
        <Skeleton shape="circle" className="size-8" />
        <div className="flex-1 space-y-1.5">
          <Skeleton className="h-3.5 w-32" />
          <Skeleton className="h-3 w-20" />
        </div>
      </div>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton key={i} className="h-3 w-full" />
      ))}
    </div>
  )
}
