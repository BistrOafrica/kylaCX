import { NavLink, Outlet } from "react-router-dom"
import {
  IconBuildingSkyscraper,
  IconUsersGroup,
  IconShieldLock,
  IconApps,
  IconUserCircle,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"

const TABS = [
  { to: "/admin/organisation", icon: IconBuildingSkyscraper, label: "Organisation" },
  { to: "/admin/users",        icon: IconUsersGroup,         label: "Users & invites" },
  { to: "/admin/rbac",         icon: IconShieldLock,         label: "Roles" },
  { to: "/admin/apps",         icon: IconApps,               label: "Apps" },
  { to: "/admin/profile",      icon: IconUserCircle,         label: "Profile" },
]

export function AdminRoute() {
  return (
    <div className="flex flex-col h-full bg-canvas">
      <nav
        role="tablist"
        aria-label="Admin sections"
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
      >
        {TABS.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            role="tab"
            className={({ isActive }) =>
              cn(
                "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md transition-colors",
                "text-fg-secondary hover:text-fg hover:bg-subtle",
                isActive && "bg-accent-subtle text-fg font-medium",
              )
            }
          >
            <tab.icon className="size-3.5" aria-hidden />
            {tab.label}
          </NavLink>
        ))}
      </nav>
      <div className="flex-1 min-h-0 overflow-hidden">
        <Outlet />
      </div>
    </div>
  )
}
