import { useState } from "react"

/**
 * Sidebar collapsed/expanded state.
 *
 * Per-tab UI preference — not persisted globally because window
 * widths vary. Toggle via ⌘\ in AppShell.
 */
export function useSidebarCollapse() {
  return useState(false)
}
