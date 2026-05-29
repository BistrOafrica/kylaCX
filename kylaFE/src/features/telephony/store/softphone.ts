import { create } from "zustand"

/**
 * Softphone store — shell-wide call state.
 *
 * The widget itself reads from this store; click-to-call buttons
 * elsewhere in the app (contact panel, conversation header) push
 * numbers in via `requestDial`. The store doesn't talk to the backend
 * directly — the widget owns the StartCallSession mutation so the
 * Query cache stays the single source of truth for call state.
 */

export type CallState =
  | "idle"
  | "dialing"
  | "ringing"
  | "active"
  | "on_hold"
  | "ended"

interface SoftphoneState {
  isOpen: boolean
  state: CallState
  sessionId: string | null
  dialNumber: string
  contactName: string | null
  startedAt: number | null
  /** Seconds elapsed since `startedAt`. Updated externally via tick. */
  elapsedSec: number

  setOpen: (open: boolean) => void
  open: () => void
  close: () => void
  toggle: () => void

  setDialNumber: (n: string) => void
  requestDial: (n: string, contactName?: string) => void
  setSession: (sessionId: string) => void
  setState: (s: CallState) => void
  setContactName: (name: string | null) => void

  tick: () => void
  reset: () => void
}

export const useSoftphoneStore = create<SoftphoneState>((set, get) => ({
  isOpen: false,
  state: "idle",
  sessionId: null,
  dialNumber: "",
  contactName: null,
  startedAt: null,
  elapsedSec: 0,

  setOpen: (isOpen) => set({ isOpen }),
  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
  toggle: () => set((s) => ({ isOpen: !s.isOpen })),

  setDialNumber: (dialNumber) => set({ dialNumber }),
  requestDial: (n, contactName) =>
    set({
      isOpen: true,
      dialNumber: n,
      contactName: contactName ?? null,
    }),
  setSession: (sessionId) =>
    set({
      sessionId,
      state: "dialing",
      startedAt: Date.now(),
      elapsedSec: 0,
    }),
  setState: (state) => set({ state }),
  setContactName: (contactName) => set({ contactName }),

  tick: () => {
    const { startedAt, state } = get()
    if (!startedAt) return
    if (state === "ended" || state === "idle") return
    set({ elapsedSec: Math.floor((Date.now() - startedAt) / 1000) })
  },
  reset: () =>
    set({
      state: "idle",
      sessionId: null,
      startedAt: null,
      elapsedSec: 0,
    }),
}))
