import { Outlet } from "react-router-dom"
import { useHotkeys } from "react-hotkeys-hook"
import { useTheme } from "@/components/theme-provider"
import { useCommandStore } from "@/lib/command"
import { useAIStore } from "@/lib/ai"
import { Sidebar } from "./Sidebar"
import { useSidebarCollapse } from "./useSidebarCollapse"
import { TopBar } from "./TopBar"
import { AIRail } from "./AIRail"
import { StatusBar } from "./StatusBar"
import { CommandPalette } from "./CommandPalette"
import { Softphone } from "@/features/telephony/components/Softphone"
import { SoftphoneButton } from "@/features/telephony/components/SoftphoneButton"

/**
 * AppShell — the authenticated 4-pane workbench layout.
 *
 *   ┌────────────────────────────────────────────────────────┐
 *   │ TopBar                                                 │
 *   ├──────────┬─────────────────────────────┬───────────────┤
 *   │ Sidebar  │ Outlet (primary surface)    │ AIRail        │
 *   │          │                             │ (toggle ⌘J)   │
 *   ├──────────┴─────────────────────────────┴───────────────┤
 *   │ StatusBar                                              │
 *   └────────────────────────────────────────────────────────┘
 *
 * Routes registered under <RequireAuth><AppShell/></RequireAuth> in
 * src/app/routes/config.tsx render into <Outlet/>.
 */
export function AppShell() {
  const [collapsed, setCollapsed] = useSidebarCollapse()
  const togglePalette = useCommandStore((s) => s.toggle)
  const toggleAI = useAIStore((s) => s.toggle)
  const { theme, setTheme } = useTheme()

  // Global keyboard shortcuts
  useHotkeys("mod+k", (e) => {
    e.preventDefault()
    togglePalette()
  }, { enableOnFormTags: true })

  useHotkeys("mod+j", (e) => {
    e.preventDefault()
    toggleAI()
  }, { enableOnFormTags: false })

  useHotkeys("mod+\\", (e) => {
    e.preventDefault()
    setCollapsed((c) => !c)
  }, { enableOnFormTags: false })

  useHotkeys("mod+/", (e) => {
    e.preventDefault()
    setTheme(theme === "dark" ? "light" : "dark")
  }, { enableOnFormTags: false })

  return (
    <div className="h-dvh w-dvw flex flex-col overflow-hidden bg-canvas">
      <TopBar />
      <div className="flex-1 flex min-h-0">
        <Sidebar collapsed={collapsed} onToggleCollapsed={() => setCollapsed((c) => !c)} />
        <main className="flex-1 min-w-0 overflow-hidden flex">
          <div className="flex-1 min-w-0 overflow-y-auto">
            <Outlet />
          </div>
          <AIRail />
        </main>
      </div>
      <StatusBar />
      <CommandPalette />
      <Softphone />
      <SoftphoneButton />
    </div>
  )
}
