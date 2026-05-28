import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * PageHeader — the standard heading strip on every workbench page.
 *
 *   <PageHeader
 *     title="Inbox"
 *     description="All conversations across channels"
 *     actions={<Button>New</Button>}
 *   />
 *
 * Renders inline with the surface (no extra padding) — the consuming
 * page chooses its own container padding so headers can sit flush
 * with side panels or extend full bleed.
 */
export interface PageHeaderProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, "title"> {
  title: React.ReactNode
  description?: React.ReactNode
  breadcrumb?: React.ReactNode
  actions?: React.ReactNode
  sticky?: boolean
}

export function PageHeader({
  title,
  description,
  breadcrumb,
  actions,
  sticky = false,
  className,
  ...rest
}: PageHeaderProps) {
  return (
    <header
      className={cn(
        "flex items-start justify-between gap-3 px-4 py-3 border-b border-border",
        "bg-canvas",
        sticky && "sticky top-0 z-10",
        className,
      )}
      {...rest}
    >
      <div className="min-w-0 flex-1 space-y-1">
        {breadcrumb && (
          <div className="text-sm text-fg-muted">{breadcrumb}</div>
        )}
        <h1 className="text-lg font-semibold text-fg truncate">{title}</h1>
        {description && (
          <p className="text-base text-fg-secondary">{description}</p>
        )}
      </div>
      {actions && (
        <div className="flex items-center gap-2 shrink-0">{actions}</div>
      )}
    </header>
  )
}
