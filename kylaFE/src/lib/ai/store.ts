import { create } from "zustand"

/**
 * AI rail store.
 *
 * F0 ships the rail UI + state; the actual LLM calls land in F1
 * alongside the inbox copilot. Until then any consumer can push a
 * mock suggestion or set a context to exercise the UI.
 */

export type AIContextKind =
  | "none"
  | "conversation"
  | "ticket"
  | "deal"
  | "contact"
  | "workflow"

export interface AIContext {
  kind: AIContextKind
  resourceId?: string
  title?: string
}

export type SuggestionStatus = "streaming" | "complete" | "error"

export interface AISuggestion {
  id: string
  kind: "summary" | "reply" | "translate" | "classify" | "custom"
  title: string
  body: string                  // accumulating as tokens stream in
  status: SuggestionStatus
  citations?: { label: string; href?: string }[]
  createdAt: number
}

interface AIState {
  isOpen: boolean
  context: AIContext
  suggestions: AISuggestion[]

  open: () => void
  close: () => void
  toggle: () => void

  setContext: (ctx: AIContext) => void
  clearContext: () => void

  pushSuggestion: (suggestion: AISuggestion) => void
  appendToSuggestion: (id: string, chunk: string) => void
  finalizeSuggestion: (id: string, status?: SuggestionStatus) => void
  removeSuggestion: (id: string) => void
  clearSuggestions: () => void
}

export const useAIStore = create<AIState>((set) => ({
  isOpen: false,
  context: { kind: "none" },
  suggestions: [],

  open:   () => set({ isOpen: true }),
  close:  () => set({ isOpen: false }),
  toggle: () => set((s) => ({ isOpen: !s.isOpen })),

  setContext: (context) => set({ context, suggestions: [] }),
  clearContext: () => set({ context: { kind: "none" }, suggestions: [] }),

  pushSuggestion: (suggestion) =>
    set((s) => ({ suggestions: [suggestion, ...s.suggestions] })),

  appendToSuggestion: (id, chunk) =>
    set((s) => ({
      suggestions: s.suggestions.map((sug) =>
        sug.id === id ? { ...sug, body: sug.body + chunk } : sug,
      ),
    })),

  finalizeSuggestion: (id, status = "complete") =>
    set((s) => ({
      suggestions: s.suggestions.map((sug) =>
        sug.id === id ? { ...sug, status } : sug,
      ),
    })),

  removeSuggestion: (id) =>
    set((s) => ({ suggestions: s.suggestions.filter((sug) => sug.id !== id) })),

  clearSuggestions: () => set({ suggestions: [] }),
}))
