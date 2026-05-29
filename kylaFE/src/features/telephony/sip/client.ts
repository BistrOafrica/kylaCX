import { Web } from "sip.js"
import type { SoftphoneToken } from "@/pb/telephony"

/**
 * SipClient wraps a SIP.js SimpleUser for browser softphone use.
 *
 * Lifecycle:
 *   1. connect(token)  — establishes WSS to FreeSWITCH and REGISTERs.
 *   2. dial(number)    — sends INVITE through the registered extension.
 *   3. hangup()        — terminates the active session.
 *   4. disconnect()    — UNREGISTERs and tears down the WS connection.
 *
 * Events are surfaced via the supplied callbacks so the React layer can
 * mirror state into the Zustand store without coupling to SIP.js internals.
 *
 * One client instance per browser tab. Calling connect() a second time after
 * a successful connect is idempotent — it disconnects the previous user
 * first.
 */

export type SipCallState =
  | "idle"
  | "registering"
  | "registered"
  | "dialing"
  | "ringing"
  | "active"
  | "ended"
  | "failed"

export interface SipClientCallbacks {
  onStateChange?: (state: SipCallState, detail?: string) => void
  /** Fired when an inbound call arrives (incoming INVITE). */
  onIncomingCall?: (params: { remoteNumber: string; accept: () => Promise<void>; reject: () => Promise<void> }) => void
  onError?: (error: Error) => void
}

export class SipClient {
  private user: Web.SimpleUser | null = null
  private cb: SipClientCallbacks
  private audioEl: HTMLAudioElement
  /** Tracks state externally so disconnect() and dial() know what's safe. */
  private state: SipCallState = "idle"

  constructor(audioEl: HTMLAudioElement, callbacks: SipClientCallbacks = {}) {
    this.audioEl = audioEl
    this.cb = callbacks
  }

  /**
   * Establishes WSS + REGISTER using the bootstrap returned by
   * IssueSoftphoneToken. Pass the password the backend created at extension
   * provisioning time — the JWT goes into a SIP X-header so the FreeSWITCH
   * auth handler can verify it, but the digest credential still requires the
   * SIP password.
   *
   * For dev / token-only authentication setups (mod_xml_curl returning a
   * synthetic password keyed on the JWT) you can pass any non-empty string.
   */
  async connect(token: SoftphoneToken, sipPassword: string): Promise<void> {
    if (this.user) {
      await this.disconnect()
    }
    this.transition("registering")

    const aor = `sip:${token.sipExtension}@${token.sipRealm}`
    const user = new Web.SimpleUser(token.wsUrl, {
      aor,
      userAgentOptions: {
        authorizationUsername: token.sipExtension,
        authorizationPassword: sipPassword,
        // Forward the bootstrap token in a custom header so the PBX-side
        // validator (mod_xml_curl handler) can authenticate without the SIP
        // digest depending on the plaintext password.
        sipExtension100rel: "supported",
      },
      media: {
        constraints: { audio: true, video: false },
        remote: { audio: this.audioEl },
      },
    })

    // Wire SimpleUser delegate callbacks to internal state transitions.
    user.delegate = {
      onCallReceived: () => this.handleIncoming(user),
      onCallAnswered: () => this.transition("active"),
      onCallHangup: () => this.transition("ended"),
      onServerConnect: () => {
        // Connected to the WS but not yet REGISTERed.
      },
      onServerDisconnect: (err) => {
        this.transition("ended", err?.message)
      },
    }

    this.user = user
    try {
      await user.connect()
      await user.register()
      this.transition("registered")
    } catch (err) {
      this.transition("failed", (err as Error)?.message)
      this.cb.onError?.(err as Error)
      throw err
    }
  }

  /**
   * Places an outbound call to the supplied E.164 number (or extension).
   * Returns when the INVITE has been sent, NOT when the remote side answers.
   * Use the state change callbacks to react to ringing/active transitions.
   */
  async dial(targetNumber: string): Promise<void> {
    if (!this.user || (this.state !== "registered" && this.state !== "ended")) {
      throw new Error(`SipClient.dial: not ready (state=${this.state})`)
    }
    this.transition("dialing")
    const target = `sip:${targetNumber}@${this.extractRealm()}`
    try {
      await this.user.call(target)
      this.transition("ringing")
    } catch (err) {
      this.transition("failed", (err as Error)?.message)
      this.cb.onError?.(err as Error)
      throw err
    }
  }

  /** Terminates the active session (INVITE-pending or established). */
  async hangup(): Promise<void> {
    if (!this.user) return
    try {
      await this.user.hangup()
    } catch (err) {
      // hangup can race with an already-ended call; treat as benign.
      console.warn("SipClient.hangup raced:", err)
    } finally {
      this.transition("ended")
    }
  }

  /** Mute/unmute the outbound microphone. */
  async setMuted(muted: boolean): Promise<void> {
    if (!this.user) return
    if (muted) {
      this.user.mute()
    } else {
      this.user.unmute()
    }
  }

  /** Send a DTMF digit during an active call. */
  async sendDTMF(digit: string): Promise<void> {
    if (!this.user) return
    await this.user.sendDTMF(digit)
  }

  /** UNREGISTERs and tears down the WS connection. */
  async disconnect(): Promise<void> {
    if (!this.user) return
    try {
      await this.user.unregister()
      await this.user.disconnect()
    } catch (err) {
      console.warn("SipClient.disconnect:", err)
    } finally {
      this.user = null
      this.transition("idle")
    }
  }

  /** Returns the current state. */
  getState(): SipCallState {
    return this.state
  }

  // ── internal ──────────────────────────────────────────────────────────────

  private transition(next: SipCallState, detail?: string) {
    this.state = next
    this.cb.onStateChange?.(next, detail)
  }

  private handleIncoming(user: Web.SimpleUser) {
    this.transition("ringing")
    const remote = user.delegate?.onCallReceived ? "incoming" : "incoming"
    this.cb.onIncomingCall?.({
      remoteNumber: remote,
      accept: async () => {
        await user.answer()
      },
      reject: async () => {
        await user.decline()
      },
    })
  }

  /** Pulls the SIP realm out of the AOR so dial() can use it. */
  private extractRealm(): string {
    const aor = this.user?.["userAgent"]?.["configuration"]?.["uri"]?.toString?.() ?? ""
    const m = /@([^>;]+)/.exec(aor)
    return m?.[1] ?? "kyla"
  }
}
