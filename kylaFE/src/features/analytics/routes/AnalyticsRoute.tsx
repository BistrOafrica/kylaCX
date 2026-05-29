import { NavLink, Outlet } from "react-router-dom"
import {
  IconChartBar,
  IconPhone,
  IconCreditCard,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"

const TABS = [
  { to: "/analytics/overview", icon: IconChartBar,   label: "Overview" },
  { to: "/analytics/calls",    icon: IconPhone,      label: "Calls" },
  { to: "/analytics/billing",  icon: IconCreditCard, label: "Billing" },
]

export function AnalyticsRoute() {
  return (
    <div className="flex flex-col h-full bg-canvas">
      <nav
        role="tablist"
        aria-label="Analytics sections"
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
