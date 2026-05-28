import { NavLink } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { IconChevronsLeft } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { PRIMARY_NAV, SECONDARY_NAV, type NavItem } from "./nav-items"

/**
 * Sidebar — primary navigation.
 *
 * Two density states:
 *   expanded   208px, label + icon + badge
 *   collapsed  48px,  icon only with tooltip on hover
 *
 * Collapsing the sidebar is a per-tab UI state (not persisted globally —
 * it's a workspace preference for that window). Toggle via ⌘\.
 */

interface SidebarProps {
  collapsed: boolean
  onToggleCollapsed: () => void
}

export function Sidebar({ collapsed, onToggleCollapsed }: SidebarProps) {
  return (
    <aside
      data-collapsed={collapsed}
      className={cn(
        "shrink-0 flex flex-col bg-surface border-e border-border",
        "transition-[width] duration-200 ease-out",
        collapsed ? "w-12" : "w-52",
      )}
      aria-label="Primary"
    >
      <div className="flex items-center px-3 h-10 border-b border-border">
        <Brand collapsed={collapsed} />
      </div>

      <nav className="flex-1 overflow-y-auto p-1.5 space-y-px">
        {PRIMARY_NAV.map((item) => (
          <SidebarItem key={item.id} item={item} collapsed={collapsed} />
        ))}
      </nav>

      <div className="border-t border-border p-1.5 space-y-px">
        {SECONDARY_NAV.map((item) => (
          <SidebarItem key={item.id} item={item} collapsed={collapsed} />
        ))}
        <CollapseToggle collapsed={collapsed} onClick={onToggleCollapsed} />
      </div>
    </aside>
  )
}

function Brand({ collapsed }: { collapsed: boolean }) {
  return (
    <div className="flex items-center gap-2 min-w-0">
      <span
        className={cn(
          "inline-flex items-center justify-center size-6 rounded-sm",
          "bg-accent text-accent-fg font-semibold text-md",
        )}
        aria-hidden
      >
        K
      </span>
      {!collapsed && (
        <span className="text-md font-semibold text-fg tracking-tight">
          Kyla
        </span>
      )}
    </div>
  )
}

function SidebarItem({ item, collapsed }: { item: NavItem; collapsed: boolean }) {
  const { t } = useTranslation()
  const Icon = item.icon
  const label = t(item.i18nKey)

  const content = (
    <>
      <Icon className="size-4 shrink-0" aria-hidden />
      {!collapsed && (
        <>
          <span className="truncate text-md">{label}</span>
          {item.disabled && (
            <span
              className={cn(
                "ms-auto text-[10px] font-mono uppercase tracking-wider",
                "text-fg-muted",
              )}
              aria-label="Coming soon"
            >
              {item.phase}
            </span>
          )}
        </>
      )}
    </>
  )

  if (item.disabled) {
    return (
      <div
        role="link"
        aria-disabled
        title={collapsed ? `${label} · ${item.phase}` : undefined}
        className={cn(
          "flex items-center gap-2.5 px-2 h-7 rounded-sm",
          "text-fg-muted cursor-not-allowed",
          collapsed && "justify-center px-0",
        )}
      >
        {content}
      </div>
    )
  }

  return (
    <NavLink
      to={item.href}
      title={collapsed ? label : undefined}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-2.5 px-2 h-7 rounded-sm transition-colors",
          "text-fg-secondary hover:text-fg hover:bg-subtle",
          isActive && "bg-accent-subtle text-fg font-medium",
          collapsed && "justify-center px-0",
        )
      }
    >
      {content}
    </NavLink>
  )
}

function CollapseToggle({
  collapsed,
  onClick,
}: {
  collapsed: boolean
  onClick: () => void
}) {
  const { t } = useTranslation()
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={t("shell.toggleSidebar")}
      title={t("shell.toggleSidebar")}
      className={cn(
        "flex items-center gap-2.5 px-2 h-7 w-full rounded-sm",
        "text-fg-muted hover:text-fg hover:bg-subtle transition-colors",
        collapsed && "justify-center px-0",
      )}
    >
      <IconChevronsLeft
        className={cn(
          "size-4 transition-transform",
          collapsed && "rotate-180",
        )}
        aria-hidden
      />
      {!collapsed && <span className="truncate text-sm">Collapse</span>}
    </button>
  )
}
