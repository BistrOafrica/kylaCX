import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"
import { fetchSoftphoneToken } from "../api/softphone"
import { useSoftphoneStore, type CallState } from "../store/softphone"
import { SipClient, type SipCallState } from "./client"

/**
 * useSipClient owns the SipClient instance for the current browser session.
 *
 * Responsibilities:
 *   1. Mounts an off-screen <audio> element that receives the remote stream.
 *   2. Fetches a softphone token + connects the SIP.js UserAgent on demand
 *      (lazy — registration only happens when the operator opens the
 *      softphone widget or dials a number).
 *   3. Mirrors SIP state into the softphone Zustand store so the existing
 *      Softphone.tsx UI reflects the call lifecycle without re-renders.
 *   4. Exposes dial/answer/hangup/mute as stable callbacks.
 *
 * The hook is safe to mount in multiple places — internally it uses a
 * module-level singleton so REGISTER only happens once per tab.
 */

let singleton: SipClient | null = null
let singletonAudio: HTMLAudioElement | null = null

export interface UseSipClient {
  ready: boolean
  state: SipCallState
  error: string | null
  /**
   * Optional SIP password — operator can store this in browser storage at
   * first registration if mod_xml_curl auth isn't wired yet. Empty string
   * means "let the backend's auth mode handle it" (typically token-only).
   */
  connect: (sipPassword?: string) => Promise<void>
  disconnect: () => Promise<void>
  dial: (number: string) => Promise<void>
  hangup: () => Promise<void>
  setMuted: (muted: boolean) => Promise<void>
  sendDTMF: (digit: string) => Promise<void>
}

export function useSipClient(): UseSipClient {
  const audioRef = useRef<HTMLAudioElement | null>(singletonAudio)
  const [state, setState] = useState<SipCallState>(singleton?.getState() ?? "idle")
  const [error, setError] = useState<string | null>(null)
  const [ready, setReady] = useState(singleton?.getState() === "registered")

  // Reflect SIP transitions into the existing softphone store so the
  // floating dialer UI keeps showing the right state.
  const setStoreState = useSoftphoneStore((s) => s.setState)
  const reset = useSoftphoneStore((s) => s.reset)

  // Lazily create the audio element + SipClient on first mount.
  useEffect(() => {
    if (!singletonAudio) {
      const el = document.createElement("audio")
      el.autoplay = true
      el.style.display = "none"
      document.body.appendChild(el)
      singletonAudio = el
    }
    audioRef.current = singletonAudio

    if (!singleton) {
      singleton = new SipClient(singletonAudio, {
        onStateChange: (next, detail) => {
          setState(next)
          setReady(next === "registered" || next === "active" || next === "ringing")
          setStoreState(mapStoreState(next))
          if (detail && (next === "failed" || next === "ended")) {
            setError(detail)
          } else if (next === "registered") {
            setError(null)
          }
        },
        onIncomingCall: ({ remoteNumber, accept, reject }) => {
          // Best-effort inbound UX. The richer accept/reject UI lives in
          // Softphone.tsx; this just opens the widget so the operator sees
          // the incoming state.
          useSoftphoneStore.getState().requestDial(remoteNumber)
          toast.info(`Incoming call from ${remoteNumber}`, {
            action: { label: "Answer", onClick: () => void accept() },
            cancel: { label: "Decline", onClick: () => void reject() },
          })
        },
        onError: (e) => {
          setError(e.message)
          toast.error(`Softphone error: ${e.message}`)
        },
      })
    }
    // Intentionally NOT disconnecting on unmount — the singleton lives for
    // the lifetime of the tab. Disconnect requires an explicit call.
  }, [setStoreState])

  const connect = useCallback(
    async (sipPassword = "") => {
      if (!singleton) throw new Error("SipClient not initialised yet")
      try {
        const token = await fetchSoftphoneToken()
        await singleton.connect(token, sipPassword || "kyla-token-auth")
      } catch (e) {
        setError((e as Error).message)
        throw e
      }
    },
    [],
  )

  const disconnect = useCallback(async () => {
    if (!singleton) return
    await singleton.disconnect()
    reset()
  }, [reset])

  const dial = useCallback(
    async (number: string) => {
      if (!singleton) throw new Error("SipClient not initialised yet")
      // Lazy register the first time the operator dials.
      if (singleton.getState() === "idle") {
        await connect()
      }
      await singleton.dial(number)
    },
    [connect],
  )

  const hangup = useCallback(async () => {
    if (!singleton) return
    await singleton.hangup()
  }, [])

  const setMuted = useCallback(async (muted: boolean) => {
    if (!singleton) return
    await singleton.setMuted(muted)
  }, [])

  const sendDTMF = useCallback(async (digit: string) => {
    if (!singleton) return
    await singleton.sendDTMF(digit)
  }, [])

  return { ready, state, error, connect, disconnect, dial, hangup, setMuted, sendDTMF }
}

// mapStoreState collapses the richer SIP state into the legacy 5-state
// model the existing softphone store and UI expect.
function mapStoreState(s: SipCallState): CallState {
  switch (s) {
    case "dialing":
      return "dialing"
    case "ringing":
      return "ringing"
    case "active":
      return "active"
    case "ended":
    case "failed":
      return "ended"
    case "registering":
    case "registered":
    case "idle":
    default:
      return "idle"
  }
}
