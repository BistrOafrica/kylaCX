import { NavLink, Outlet } from "react-router-dom"
import { useTranslation } from "react-i18next"
import {
  IconUser,
  IconBuilding,
  IconUserStar,
  IconLayoutKanban,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"

/**
 * CrmRoute — owns the CRM tab strip + outlet.
 *
 *   ┌──────────────────────────────────────┐
 *   │ Contacts · Companies · Leads · Deals │  ← horizontal tabs
 *   ├──────────────────────────────────────┤
 *   │  Outlet                              │
 *   └──────────────────────────────────────┘
 *
 * Default redirects to /crm/contacts when the user lands on /crm.
 */
const TABS: { to: string; icon: typeof IconUser; key: string; label: string }[] = [
  { to: "/crm/contacts",  icon: IconUser,         key: "contacts",  label: "Contacts" },
  { to: "/crm/companies", icon: IconBuilding,     key: "companies", label: "Companies" },
  { to: "/crm/leads",     icon: IconUserStar,     key: "leads",     label: "Leads" },
  { to: "/crm/deals",     icon: IconLayoutKanban, key: "deals",     label: "Deals" },
]

export function CrmRoute() {
  const { t } = useTranslation()
  void t // i18n keys for tabs land when locale work catches up

  return (
    <div className="flex flex-col h-full bg-canvas">
      <nav
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
        role="tablist"
        aria-label="CRM sections"
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
      <div className="flex-1 min-h-0">
        <Outlet />
      </div>
    </div>
  )
}
