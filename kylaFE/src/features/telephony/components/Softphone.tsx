import { useEffect, useState } from "react"
import {
  IconPhone,
  IconPhoneOff,
  IconBackspace,
  IconPlayerPause,
  IconPlayerPlay,
  IconNote,
  IconX,
  IconLoader2,
  IconMicrophone,
  IconMicrophoneOff,
} from "@tabler/icons-react"
import { toast } from "sonner"
import { Surface } from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useSoftphoneStore, type CallState } from "../store/softphone"
import {
  useEndCallSession,
  usePlaceOnHold,
  useRemoveFromHold,
  useAddCallNote,
} from "../hooks/queries"
import { useSipClient } from "../sip/useSipClient"

const DIGITS = [
  ["1", "2", "3"],
  ["4", "5", "6"],
  ["7", "8", "9"],
  ["*", "0", "#"],
]

/**
 * Softphone — floating dialer.
 *
 * Mounts once at the AppShell level. Hidden by default; opens via the
 * floating phone button or any click-to-call elsewhere in the app
 * (which calls `useSoftphoneStore.getState().requestDial(number)`).
 */
export function Softphone() {
  const isOpen = useSoftphoneStore((s) => s.isOpen)
  const callState = useSoftphoneStore((s) => s.state)
  const sessionId = useSoftphoneStore((s) => s.sessionId)
  const dial = useSoftphoneStore((s) => s.dialNumber)
  const contactName = useSoftphoneStore((s) => s.contactName)
  const elapsed = useSoftphoneStore((s) => s.elapsedSec)
  const setDialNumber = useSoftphoneStore((s) => s.setDialNumber)
  const setSession = useSoftphoneStore((s) => s.setSession)
  const setState = useSoftphoneStore((s) => s.setState)
  const close = useSoftphoneStore((s) => s.close)
  const reset = useSoftphoneStore((s) => s.reset)
  const tick = useSoftphoneStore((s) => s.tick)

  const end = useEndCallSession()
  const hold = usePlaceOnHold()
  const unhold = useRemoveFromHold()
  const addNote = useAddCallNote()
  // Phase 5: real SIP/WebRTC media path. The hook owns SIP.js lifecycle —
  // dial/hangup go straight to the PBX; the legacy mutations stay for hold
  // and note operations until those are ported off CallSessionService.
  const sip = useSipClient()

  const [muted, setMuted] = useState(false)
  const [noteMode, setNoteMode] = useState(false)
  const [noteText, setNoteText] = useState("")

  // Tick the elapsed counter every second while the call is live.
  useEffect(() => {
    if (callState === "idle" || callState === "ended") return
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [callState, tick])

  // Auto-close + reset 2s after a call ends so the next dial is fresh.
  useEffect(() => {
    if (callState !== "ended") return
    const t = setTimeout(reset, 2_000)
    return () => clearTimeout(t)
  }, [callState, reset])

  if (!isOpen) return null

  const onDigit = (d: string) => {
    if (callState === "active" || callState === "ringing") {
      // Send DTMF via the active SIP session (RFC 2833 INFO).
      void sip.sendDTMF(d).catch(() => {/* benign — call may have ended */})
      return
    }
    setDialNumber(dial + d)
  }

  const onBackspace = () => setDialNumber(dial.slice(0, -1))

  const onCall = async () => {
    if (!dial.trim()) return
    try {
      // Real SIP dial via the registered UserAgent. State transitions arrive
      // through the SipClient callbacks (mirrored into the store), so we
      // don't optimistically advance here — the UI reflects the actual
      // INVITE/ringing/answered sequence.
      await sip.dial(dial.trim())
      // The backend creates a Call row when the CHANNEL_CREATE event fires;
      // the session_id for hold/notes is fetched out of the call history
      // list — for now we leave it empty and skip those ops mid-call.
      setSession(`pending-${Date.now()}`)
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  const onHangup = async () => {
    try {
      await sip.hangup()
    } catch {
      /* benign — race with remote BYE */
    }
    if (sessionId) {
      try {
        await end.mutateAsync(sessionId)
      } catch {
        /* mutation cache toasts */
      }
    }
    setState("ended")
  }

  const onToggleHold = async () => {
    if (!sessionId) return
    if (callState === "on_hold") {
      await unhold.mutateAsync(sessionId)
      setState("active")
    } else if (callState === "active") {
      await hold.mutateAsync(sessionId)
      setState("on_hold")
    }
  }

  const onSaveNote = async () => {
    if (!sessionId || !noteText.trim()) {
      setNoteMode(false)
      return
    }
    try {
      await addNote.mutateAsync({
        sessionId,
        name: "Call note",
        content: noteText,
      })
      toast.success("Note saved")
      setNoteText("")
      setNoteMode(false)
    } catch {
      /* mutation cache toasts */
    }
  }

  const isActive = callState === "active" || callState === "on_hold"

  return (
    <Surface
      level={3}
      radius="lg"
      className={cn(
        "fixed bottom-9 end-6 z-40 w-64 overflow-hidden",
        "shadow-elev-3",
      )}
    >
      <header
        className={cn(
          "flex items-center gap-2 px-3 h-9 border-b border-border",
          "bg-surface",
        )}
      >
        <StateDot state={callState} />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium text-fg truncate">
            {contactName ?? labelForState(callState)}
          </div>
          {isActive && (
            <div className="text-xs font-mono text-fg-muted">
              {formatElapsed(elapsed)}
            </div>
          )}
        </div>
        <button
          type="button"
          onClick={close}
          aria-label="Close"
          className="inline-flex items-center justify-center size-5 rounded-xs text-fg-muted hover:text-fg hover:bg-subtle"
        >
          <IconX className="size-3.5" />
        </button>
      </header>

      <div className="p-3 space-y-3">
        {!noteMode && (
          <>
            <Input
              value={dial}
              onChange={(e) =>
                setDialNumber(e.target.value.replace(/[^\d+*#]/g, ""))
              }
              placeholder="+1 555 010 1234"
              className="h-9 text-lg font-mono tabular-nums text-center tracking-wide"
              readOnly={isActive}
            />

            <div className="grid grid-cols-3 gap-1.5">
              {DIGITS.flat().map((d) => (
                <button
                  key={d}
                  type="button"
                  onClick={() => onDigit(d)}
                  className={cn(
                    "h-10 rounded-sm font-mono text-lg",
                    "bg-subtle hover:bg-muted active:bg-accent-subtle",
                    "text-fg transition-colors",
                  )}
                >
                  {d}
                </button>
              ))}
            </div>

            <div className="flex items-center justify-between gap-1.5">
              <button
                type="button"
                onClick={onBackspace}
                aria-label="Backspace"
                disabled={!dial}
                className={cn(
                  "inline-flex items-center justify-center size-8 rounded-sm",
                  "text-fg-muted hover:text-fg hover:bg-subtle",
                  "disabled:opacity-40 disabled:pointer-events-none",
                )}
              >
                <IconBackspace className="size-4" />
              </button>

              {isActive ? (
                <>
                  <button
                    type="button"
                    onClick={() => {
                      const next = !muted
                      setMuted(next)
                      void sip.setMuted(next)
                    }}
                    aria-label={muted ? "Unmute" : "Mute"}
                    className={cn(
                      "inline-flex items-center justify-center size-8 rounded-sm",
                      muted ? "bg-warn-subtle text-warn" : "text-fg-muted hover:bg-subtle",
                    )}
                  >
                    {muted ? (
                      <IconMicrophoneOff className="size-4" />
                    ) : (
                      <IconMicrophone className="size-4" />
                    )}
                  </button>
                  <button
                    type="button"
                    onClick={() => void onToggleHold()}
                    aria-label={callState === "on_hold" ? "Resume" : "Hold"}
                    className={cn(
                      "inline-flex items-center justify-center size-8 rounded-sm",
                      callState === "on_hold"
                        ? "bg-warn-subtle text-warn"
                        : "text-fg-muted hover:bg-subtle",
                    )}
                  >
                    {callState === "on_hold" ? (
                      <IconPlayerPlay className="size-4" />
                    ) : (
                      <IconPlayerPause className="size-4" />
                    )}
                  </button>
                  <button
                    type="button"
                    onClick={() => setNoteMode(true)}
                    aria-label="Note"
                    className="inline-flex items-center justify-center size-8 rounded-sm text-fg-muted hover:bg-subtle"
                  >
                    <IconNote className="size-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => void onHangup()}
                    disabled={end.isPending}
                    aria-label="Hang up"
                    className={cn(
                      "inline-flex items-center justify-center size-9 rounded-full",
                      "bg-danger text-danger-fg hover:opacity-90 disabled:opacity-50",
                    )}
                  >
                    <IconPhoneOff className="size-4" />
                  </button>
                </>
              ) : (
                <button
                  type="button"
                  onClick={() => void onCall()}
                  disabled={sip.state === "dialing" || sip.state === "registering" || !dial.trim()}
                  aria-label="Call"
                  className={cn(
                    "ms-auto inline-flex items-center gap-1.5 h-8 px-3 rounded-sm",
                    "bg-success text-success-fg hover:opacity-90",
                    "disabled:opacity-50 disabled:pointer-events-none",
                  )}
                >
                  {sip.state === "dialing" || sip.state === "registering" ? (
                    <IconLoader2 className="size-4 animate-spin" />
                  ) : (
                    <IconPhone className="size-4" />
                  )}
                  Call
                </button>
              )}
            </div>
          </>
        )}

        {noteMode && (
          <div className="space-y-2">
            <textarea
              value={noteText}
              onChange={(e) => setNoteText(e.target.value)}
              autoFocus
              rows={4}
              placeholder="Note about this call…"
              className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent"
            />
            <div className="flex items-center justify-end gap-2">
              <button
                type="button"
                onClick={() => {
                  setNoteText("")
                  setNoteMode(false)
                }}
                className="text-sm text-fg-muted hover:text-fg"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={() => void onSaveNote()}
                disabled={addNote.isPending || !noteText.trim()}
                className={cn(
                  "inline-flex items-center gap-1 h-7 px-2.5 rounded-sm text-md font-medium",
                  "bg-accent text-accent-fg disabled:opacity-50",
                )}
              >
                Save
              </button>
            </div>
          </div>
        )}
      </div>
    </Surface>
  )
}

function StateDot({ state }: { state: CallState }) {
  const cls =
    state === "active"
      ? "bg-success animate-pulse"
      : state === "on_hold"
        ? "bg-warn"
        : state === "dialing" || state === "ringing"
          ? "bg-info animate-pulse"
          : state === "ended"
            ? "bg-danger"
            : "bg-border-strong"
  return <span aria-hidden className={cn("size-1.5 rounded-full shrink-0", cls)} />
}

function labelForState(s: CallState): string {
  switch (s) {
    case "dialing":  return "Dialing…"
    case "ringing":  return "Ringing…"
    case "active":   return "In call"
    case "on_hold":  return "On hold"
    case "ended":    return "Call ended"
    case "idle":
    default:         return "Softphone"
  }
}

function formatElapsed(sec: number): string {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`
}
