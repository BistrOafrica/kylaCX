/**
 * Keyboard shortcut definitions — single source of truth for what each
 * key combo does. The shell binds these via react-hotkeys-hook; the
 * shortcuts overlay (? key) renders the full list from here.
 *
 * Scope rules:
 *   - global   : always active
 *   - shell    : active while authenticated and in the app shell
 *   - inbox    : active inside the inbox surface
 *   - palette  : active while the command palette is open
 */

export type ShortcutScope =
  | "global"
  | "shell"
  | "inbox"
  | "crm"
  | "palette"

export interface Shortcut {
  id: string
  scope: ShortcutScope
  keys: string             // hotkeys-hook syntax: "mod+k", "g i", "esc"
  description: string
}

export const SHORTCUTS: Shortcut[] = [
  { id: "command.open",      scope: "global", keys: "mod+k",  description: "Open command palette" },
  { id: "command.close",     scope: "palette", keys: "esc",   description: "Close command palette" },
  { id: "shell.ai-rail",     scope: "shell",  keys: "mod+j",  description: "Toggle Kyla copilot" },
  { id: "shell.sidebar",     scope: "shell",  keys: "mod+\\", description: "Toggle sidebar" },
  { id: "shell.theme",       scope: "shell",  keys: "mod+/",  description: "Toggle light/dark theme" },
  { id: "shell.settings",    scope: "shell",  keys: "mod+,",  description: "Open settings" },
  { id: "shell.back",        scope: "shell",  keys: "mod+b",  description: "Back / close detail" },

  { id: "nav.inbox",         scope: "shell",  keys: "g>i",    description: "Go to Inbox" },
  { id: "nav.crm",           scope: "shell",  keys: "g>c",    description: "Go to CRM" },
  { id: "nav.tickets",       scope: "shell",  keys: "g>t",    description: "Go to Tickets" },
  { id: "nav.knowledge",     scope: "shell",  keys: "g>k",    description: "Go to Knowledge" },
  { id: "nav.automation",    scope: "shell",  keys: "g>a",    description: "Go to Automation" },
  { id: "nav.admin",         scope: "shell",  keys: "g>s",    description: "Go to Settings" },

  { id: "help.shortcuts",    scope: "global", keys: "shift+?", description: "Show keyboard shortcuts" },
]

export function shortcutsForScope(scope: ShortcutScope): Shortcut[] {
  return SHORTCUTS.filter((s) => s.scope === scope)
}
