import { create } from "zustand"

/**
 * Command palette registry.
 *
 * Routes and features register their commands at mount time; ⌘K
 * surfaces them and invokes the chosen handler. The store is the
 * source of truth, so the palette UI renders from it and persists
 * recently-used items.
 *
 * A command is just a typed handle:
 *
 *   {
 *     id: "nav.inbox",
 *     section: "navigation",
 *     label: t("nav.inbox"),
 *     keywords: ["inbox", "messages"],
 *     shortcut: ["g", "i"],
 *     run: () => navigate("/inbox"),
 *   }
 */

export type CommandSection =
  | "navigation"
  | "actions"
  | "theme"
  | "language"
  | "workspaces"
  | "account"
  | "help"

export interface Command {
  id: string
  section: CommandSection
  label: string
  description?: string
  keywords?: string[]
  shortcut?: string[]
  icon?: React.ReactNode
  run: () => void | Promise<void>
  /** If true the command is filtered out — useful for context-aware items. */
  hidden?: boolean
}

interface PaletteState {
  isOpen: boolean
  recent: string[]            // command ids, most-recent-first, capped at 8
  commands: Map<string, Command>

  open: () => void
  close: () => void
  toggle: () => void

  register: (command: Command) => () => void
  registerMany: (commands: Command[]) => () => void
  unregister: (id: string) => void
  recordRun: (id: string) => void

  listForRender: () => Command[]
}

const RECENT_CAP = 8

export const useCommandStore = create<PaletteState>((set, get) => ({
  isOpen: false,
  recent: [],
  commands: new Map(),

  open:   () => set({ isOpen: true }),
  close:  () => set({ isOpen: false }),
  toggle: () => set((s) => ({ isOpen: !s.isOpen })),

  register: (command) => {
    set((s) => {
      const next = new Map(s.commands)
      next.set(command.id, command)
      return { commands: next }
    })
    return () => get().unregister(command.id)
  },

  registerMany: (commands) => {
    set((s) => {
      const next = new Map(s.commands)
      for (const c of commands) next.set(c.id, c)
      return { commands: next }
    })
    return () => {
      set((s) => {
        const next = new Map(s.commands)
        for (const c of commands) next.delete(c.id)
        return { commands: next }
      })
    }
  },

  unregister: (id) => {
    set((s) => {
      if (!s.commands.has(id)) return s
      const next = new Map(s.commands)
      next.delete(id)
      return { commands: next }
    })
  },

  recordRun: (id) => {
    set((s) => {
      const filtered = s.recent.filter((r) => r !== id)
      return { recent: [id, ...filtered].slice(0, RECENT_CAP) }
    })
  },

  listForRender: () => {
    const { commands, recent } = get()
    const recentCommands = recent
      .map((id) => commands.get(id))
      .filter((c): c is Command => Boolean(c) && !c?.hidden)
    const others = [...commands.values()].filter(
      (c) => !c.hidden && !recent.includes(c.id),
    )
    return [...recentCommands, ...others]
  },
}))
