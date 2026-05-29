import { describe, it, expect } from "vitest"
import {
  ACTIONS,
  ACTION_BY_TYPE,
  getActionSpec,
  TRIGGERS,
  getTriggerSpec,
  CONDITION_OPS,
} from "@/features/automation/utils/actions"

describe("action registry", () => {
  it("exposes all 11 canonical action types", () => {
    expect(ACTIONS).toHaveLength(11)
    const types = ACTIONS.map((a) => a.type).sort()
    expect(types).toEqual(
      [
        "assign_user",
        "create_object",
        "create_task",
        "delay",
        "invoke_webhook",
        "run_ai_skill",
        "send_message",
        "send_notification",
        "set_sla",
        "start_workflow",
        "update_object",
      ].sort(),
    )
  })

  it("each action has at least one required param or no required params at all", () => {
    for (const spec of ACTIONS) {
      expect(spec.label).toBeTruthy()
      expect(spec.category).toBeTruthy()
      expect(spec.params).toBeInstanceOf(Array)
    }
  })

  it("getActionSpec resolves valid types and returns null for unknown", () => {
    expect(getActionSpec("delay")?.label).toBe("Delay")
    expect(getActionSpec("send_message")?.category).toBe("messaging")
    expect(getActionSpec("not_a_real_action")).toBeNull()
  })

  it("ACTION_BY_TYPE map is consistent with ACTIONS array", () => {
    for (const spec of ACTIONS) {
      expect(ACTION_BY_TYPE[spec.type]).toBe(spec)
    }
  })
})

describe("trigger registry", () => {
  it("exposes the documented trigger set", () => {
    const types = TRIGGERS.map((t) => t.type).sort()
    expect(types).toContain("object.created")
    expect(types).toContain("conversation.message_received")
    expect(types).toContain("form.submitted")
    expect(types).toContain("schedule")
    expect(types).toContain("webhook")
  })

  it("getTriggerSpec resolves known triggers", () => {
    expect(getTriggerSpec("object.created")?.label).toBe("Object created")
    expect(getTriggerSpec("not_a_trigger")).toBeNull()
  })
})

describe("condition operators", () => {
  it("covers all the documented operators", () => {
    const ops = CONDITION_OPS.map((o) => o.value)
    expect(ops).toContain("eq")
    expect(ops).toContain("neq")
    expect(ops).toContain("contains")
    expect(ops).toContain("empty")
    expect(ops).toContain("in")
  })
})
