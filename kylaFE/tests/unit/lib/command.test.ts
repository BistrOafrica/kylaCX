import { describe, it, expect, beforeEach, vi } from "vitest"
import { useCommandStore } from "@/lib/command/registry"

beforeEach(() => {
  // Reset store between tests.
  useCommandStore.setState({
    isOpen: false,
    recent: [],
    commands: new Map(),
  })
})

describe("command registry", () => {
  it("open/close/toggle drive the isOpen flag", () => {
    useCommandStore.getState().open()
    expect(useCommandStore.getState().isOpen).toBe(true)
    useCommandStore.getState().close()
    expect(useCommandStore.getState().isOpen).toBe(false)
    useCommandStore.getState().toggle()
    expect(useCommandStore.getState().isOpen).toBe(true)
  })

  it("register adds a command and returns an unregister", () => {
    const run = vi.fn()
    const unreg = useCommandStore.getState().register({
      id: "test.echo",
      section: "actions",
      label: "Echo",
      run,
    })
    expect(useCommandStore.getState().commands.has("test.echo")).toBe(true)
    unreg()
    expect(useCommandStore.getState().commands.has("test.echo")).toBe(false)
  })

  it("recordRun moves the id to the front of recent, capped at 8", () => {
    const ids = ["a", "b", "c", "d", "e", "f", "g", "h", "i"]
    for (const id of ids) useCommandStore.getState().recordRun(id)
    const { recent } = useCommandStore.getState()
    expect(recent.length).toBe(8)
    expect(recent[0]).toBe("i")
  })

  it("listForRender surfaces recents first", () => {
    useCommandStore.getState().registerMany([
      { id: "x", section: "actions", label: "x", run: () => {} },
      { id: "y", section: "actions", label: "y", run: () => {} },
      { id: "z", section: "actions", label: "z", run: () => {} },
    ])
    useCommandStore.getState().recordRun("y")
    const list = useCommandStore.getState().listForRender()
    expect(list[0]?.id).toBe("y")
  })

  it("hidden commands are filtered out of listForRender", () => {
    useCommandStore.getState().registerMany([
      { id: "visible", section: "actions", label: "v", run: () => {} },
      { id: "ghost", section: "actions", label: "g", run: () => {}, hidden: true },
    ])
    const list = useCommandStore.getState().listForRender()
    expect(list.map((c) => c.id)).toEqual(["visible"])
  })
})
