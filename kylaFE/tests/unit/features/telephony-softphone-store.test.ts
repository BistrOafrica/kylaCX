import { describe, it, expect, beforeEach } from "vitest"
import { useSoftphoneStore } from "@/features/telephony/store/softphone"

beforeEach(() => {
  useSoftphoneStore.setState({
    isOpen: false,
    state: "idle",
    sessionId: null,
    dialNumber: "",
    contactName: null,
    startedAt: null,
    elapsedSec: 0,
  })
})

describe("softphone store", () => {
  it("requestDial opens the widget and fills the dial number", () => {
    useSoftphoneStore.getState().requestDial("+15550101234", "Maria K.")
    const s = useSoftphoneStore.getState()
    expect(s.isOpen).toBe(true)
    expect(s.dialNumber).toBe("+15550101234")
    expect(s.contactName).toBe("Maria K.")
  })

  it("setSession transitions to dialing and records start time", () => {
    useSoftphoneStore.getState().setSession("session-123")
    const s = useSoftphoneStore.getState()
    expect(s.state).toBe("dialing")
    expect(s.sessionId).toBe("session-123")
    expect(s.startedAt).toBeTypeOf("number")
  })

  it("setState moves through call lifecycle states", () => {
    useSoftphoneStore.getState().setSession("session-1")
    useSoftphoneStore.getState().setState("active")
    expect(useSoftphoneStore.getState().state).toBe("active")
    useSoftphoneStore.getState().setState("on_hold")
    expect(useSoftphoneStore.getState().state).toBe("on_hold")
    useSoftphoneStore.getState().setState("ended")
    expect(useSoftphoneStore.getState().state).toBe("ended")
  })

  it("reset returns the store to idle", () => {
    useSoftphoneStore.getState().setSession("session-1")
    useSoftphoneStore.getState().setState("active")
    useSoftphoneStore.getState().reset()
    const s = useSoftphoneStore.getState()
    expect(s.state).toBe("idle")
    expect(s.sessionId).toBeNull()
    expect(s.startedAt).toBeNull()
    expect(s.elapsedSec).toBe(0)
  })

  it("toggle flips open ↔ closed", () => {
    useSoftphoneStore.getState().toggle()
    expect(useSoftphoneStore.getState().isOpen).toBe(true)
    useSoftphoneStore.getState().toggle()
    expect(useSoftphoneStore.getState().isOpen).toBe(false)
  })
})
