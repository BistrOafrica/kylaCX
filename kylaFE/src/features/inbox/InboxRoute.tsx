import { Outlet } from "react-router-dom"
import { InboxList } from "./components/InboxList"
import { useInboxNavigation } from "./hooks/useInboxNavigation"

/**
 * InboxRoute — owns the two-pane inbox layout that AppShell renders
 * into. Left: virtualized conversation list. Right: <Outlet/> which
 * resolves to either the empty state or a specific conversation.
 *
 * Inbox-scoped keyboard nav (j/k) lives at this layout level so it's
 * always active when the inbox is on screen.
 */
export function InboxRoute() {
  useInboxNavigation({ activeOnly: true })

  return (
    <div className="flex h-full bg-canvas">
      <div className="w-[var(--shell-list)] shrink-0 border-e border-border">
        <InboxList />
      </div>
      <div className="flex-1 min-w-0 flex">
        <Outlet />
      </div>
    </div>
  )
}
