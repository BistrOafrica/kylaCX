import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * Surface — neutral container with hairline border.
 *
 * The fundamental building block for cards, panels, sidebars and
 * everything else that needs a "surface raised once" against the
 * canvas. Three elevations:
 *
 *   level 0   flat on canvas (no border, just a fill)
 *   level 1   raised once     (border, no shadow)        — default
 *   level 2   raised twice    (border + shadow-1)
 *   level 3   floating        (border + shadow-2)        — popovers
 *
 * Used by Card, AIRail, Sidebar, dialog body, etc.
 */
export interface SurfaceProps extends React.HTMLAttributes<HTMLDivElement> {
  as?: keyof React.JSX.IntrinsicElements
  level?: 0 | 1 | 2 | 3
  radius?: "none" | "sm" | "md" | "lg"
  inset?: boolean
}

export function Surface({
  as = "div",
  level = 1,
  radius = "md",
  inset = false,
  className,
  children,
  ...rest
}: SurfaceProps) {
  const Tag = as as React.ElementType
  return (
    <Tag
      className={cn(
        "bg-surface text-fg",
        level === 0 && "bg-subtle",
        level === 1 && "border border-border",
        level === 2 && "border border-border shadow-elev-1",
        level === 3 && "border border-border shadow-elev-2",
        radius === "sm" && "rounded-sm",
        radius === "md" && "rounded-md",
        radius === "lg" && "rounded-lg",
        inset && "p-3",
        className,
      )}
      {...rest}
    >
      {children}
    </Tag>
  )
}
