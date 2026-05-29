import { NavLink, Outlet, useParams } from "react-router-dom"
import {
  IconPhone,
  IconActivity,
  IconTree,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { CallsList } from "../components/CallsList"
import { CallDetail } from "../components/CallDetail"
import { Wallboard } from "../components/Wallboard"
import { IvrList } from "../components/IvrList"
import { IvrFlowBuilder } from "../components/IvrFlowBuilder"

const TABS = [
  { to: "/calls/history",   icon: IconPhone,    label: "History" },
  { to: "/calls/wallboard", icon: IconActivity, label: "Wallboard" },
  { to: "/calls/ivr",       icon: IconTree,     label: "IVR flows" },
]

export function CallsRoute() {
  return (
    <div className="flex flex-col h-full bg-canvas">
      <nav
        role="tablist"
        aria-label="Calls sections"
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

export function CallsHistoryRoute() {
  return <CallsList />
}

export function CallDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return <CallDetail callId={id} />
}

export function WallboardRoute() {
  return <Wallboard />
}

export function IvrListRoute() {
  return <IvrList />
}

export function IvrFlowRoute() {
  const { id } = useParams()
  if (!id) return null
  return <IvrFlowBuilder id={id} />
}
