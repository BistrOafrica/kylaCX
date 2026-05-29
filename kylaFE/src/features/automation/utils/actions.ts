import {
  IconClock,
  IconEdit,
  IconPlus,
  IconMessage,
  IconUserCheck,
  IconChecklist,
  IconWebhook,
  IconSparkles,
  IconBell,
  IconAlarm,
  IconArrowsShuffle2,
  type Icon,
} from "@tabler/icons-react"

/**
 * Action registry — the canonical definition of every action type
 * the workflow builder supports. Each entry tells the canvas how to
 * label/colour/icon the node, which params to expose in the
 * inspector, and how to serialize to the proto's JSON shape.
 *
 * The proto stores actions as `google.protobuf.Struct[]` with the
 * runtime contract `{ "type": "<action>", "params": { … } }`. This
 * registry is the single source of truth for valid types + params.
 */

export type ActionType =
  | "delay"
  | "update_object"
  | "create_object"
  | "send_message"
  | "assign_user"
  | "create_task"
  | "invoke_webhook"
  | "run_ai_skill"
  | "send_notification"
  | "set_sla"
  | "start_workflow"

export type ParamType = "text" | "textarea" | "number" | "boolean" | "select" | "json"

export interface ActionParam {
  key: string
  label: string
  type: ParamType
  required?: boolean
  placeholder?: string
  helpText?: string
  options?: { value: string; label: string }[]
  default?: string | number | boolean | Record<string, unknown>
}

export interface ActionSpec {
  type: ActionType
  label: string
  category: "control" | "data" | "messaging" | "ai" | "integrations"
  icon: Icon
  color: string                // text colour for the node accent
  description: string
  params: ActionParam[]
}

export const ACTIONS: ActionSpec[] = [
  {
    type: "delay",
    label: "Delay",
    category: "control",
    icon: IconClock,
    color: "var(--info-solid, #0EA5E9)",
    description: "Pause execution before the next action.",
    params: [
      { key: "duration", label: "Duration (seconds)", type: "number", required: true, default: 60 },
    ],
  },
  {
    type: "update_object",
    label: "Update object",
    category: "data",
    icon: IconEdit,
    color: "var(--accent-solid)",
    description: "Update fields on an Object Core record.",
    params: [
      { key: "object_id", label: "Object ID (or trigger.id)", type: "text", required: true, placeholder: "{{ event.object_id }}" },
      { key: "data", label: "Patch (JSON)", type: "json", required: true, placeholder: '{"status":"closed"}' },
    ],
  },
  {
    type: "create_object",
    label: "Create object",
    category: "data",
    icon: IconPlus,
    color: "var(--accent-solid)",
    description: "Create a new Object Core record.",
    params: [
      { key: "type_slug", label: "Object type", type: "text", required: true, placeholder: "task" },
      { key: "data", label: "Field values (JSON)", type: "json", required: true, default: {} },
    ],
  },
  {
    type: "send_message",
    label: "Send message",
    category: "messaging",
    icon: IconMessage,
    color: "var(--channel-whatsapp, #10B981)",
    description: "Dispatch a message via the channel adapter (WhatsApp / Email / SMS).",
    params: [
      { key: "channel", label: "Channel", type: "select", required: true, options: [
        { value: "whatsapp", label: "WhatsApp" },
        { value: "email",    label: "Email" },
        { value: "sms",      label: "SMS" },
      ]},
      { key: "to", label: "Recipient", type: "text", required: true, placeholder: "{{ contact.phone }}" },
      { key: "body", label: "Message body", type: "textarea", required: true, placeholder: "Hi {{ contact.first_name }} …" },
    ],
  },
  {
    type: "assign_user",
    label: "Assign user",
    category: "data",
    icon: IconUserCheck,
    color: "var(--accent-solid)",
    description: "Assign the triggering object to an agent or team.",
    params: [
      { key: "object_id", label: "Object ID", type: "text", required: true, placeholder: "{{ event.object_id }}" },
      { key: "user_id", label: "Agent user ID", type: "text", placeholder: "(blank to unassign)" },
      { key: "team_id", label: "Team ID", type: "text" },
    ],
  },
  {
    type: "create_task",
    label: "Create task",
    category: "data",
    icon: IconChecklist,
    color: "var(--accent-solid)",
    description: "Create a follow-up task object.",
    params: [
      { key: "title", label: "Title", type: "text", required: true, placeholder: "Follow up with customer" },
      { key: "due_at", label: "Due (ISO timestamp)", type: "text", placeholder: "2026-06-01T12:00:00Z" },
      { key: "assignee_id", label: "Assignee user ID", type: "text" },
    ],
  },
  {
    type: "invoke_webhook",
    label: "Invoke webhook",
    category: "integrations",
    icon: IconWebhook,
    color: "var(--status-info-solid, #0EA5E9)",
    description: "POST a JSON payload to an external URL with retry.",
    params: [
      { key: "url", label: "URL", type: "text", required: true, placeholder: "https://example.com/hook" },
      { key: "payload", label: "Payload (JSON)", type: "json", required: true, default: {} },
    ],
  },
  {
    type: "run_ai_skill",
    label: "Run AI skill",
    category: "ai",
    icon: IconSparkles,
    color: "var(--accent-solid)",
    description: "Classify, summarize, or generate a reply with the LLM.",
    params: [
      { key: "skill", label: "Skill", type: "select", required: true, options: [
        { value: "classify",  label: "Classify text" },
        { value: "summarize", label: "Summarize text" },
        { value: "generate",  label: "Generate reply" },
      ]},
      { key: "text", label: "Input text", type: "textarea", required: true, placeholder: "{{ event.message }}" },
      { key: "labels", label: "Labels (comma-separated, classify only)", type: "text" },
    ],
  },
  {
    type: "send_notification",
    label: "Send notification",
    category: "messaging",
    icon: IconBell,
    color: "var(--status-warn-solid, #F59E0B)",
    description: "Push an in-app or device notification.",
    params: [
      { key: "user_id", label: "Recipient user ID", type: "text", required: true },
      { key: "title", label: "Title", type: "text", required: true, placeholder: "New ticket assigned" },
      { key: "body", label: "Body", type: "textarea" },
    ],
  },
  {
    type: "set_sla",
    label: "Set SLA",
    category: "data",
    icon: IconAlarm,
    color: "var(--status-warn-solid, #F59E0B)",
    description: "Set or reset the SLA deadline on the triggering object.",
    params: [
      { key: "policy_id", label: "SLA policy ID", type: "text", required: true },
      { key: "object_id", label: "Object ID", type: "text", required: true, placeholder: "{{ event.object_id }}" },
    ],
  },
  {
    type: "start_workflow",
    label: "Start workflow",
    category: "control",
    icon: IconArrowsShuffle2,
    color: "var(--info-solid, #0EA5E9)",
    description: "Chain into another workflow as a child execution.",
    params: [
      { key: "workflow_id", label: "Workflow ID", type: "text", required: true },
      { key: "payload", label: "Payload (JSON)", type: "json", default: {} },
    ],
  },
]

export const ACTION_BY_TYPE: Record<ActionType, ActionSpec> = ACTIONS.reduce(
  (acc, spec) => {
    acc[spec.type] = spec
    return acc
  },
  {} as Record<ActionType, ActionSpec>,
)

export function getActionSpec(type: string): ActionSpec | null {
  return (ACTION_BY_TYPE as Record<string, ActionSpec | undefined>)[type] ?? null
}

// ── Trigger registry ─────────────────────────────────────────────────────────

export type TriggerType =
  | "object.created"
  | "object.updated"
  | "conversation.message_received"
  | "conversation.resolved"
  | "deal.stage_changed"
  | "form.submitted"
  | "schedule"
  | "webhook"

export interface TriggerSpec {
  type: TriggerType
  label: string
  description: string
}

export const TRIGGERS: TriggerSpec[] = [
  { type: "object.created",                  label: "Object created",        description: "Fires when any Object Core record is created." },
  { type: "object.updated",                  label: "Object updated",        description: "Fires when an object's fields change." },
  { type: "conversation.message_received",   label: "Message received",      description: "Fires on every inbound message in the inbox." },
  { type: "conversation.resolved",           label: "Conversation resolved", description: "Fires when a conversation is closed." },
  { type: "deal.stage_changed",              label: "Deal stage changed",    description: "Fires when a deal moves between pipeline stages." },
  { type: "form.submitted",                  label: "Form submitted",        description: "Fires on every form submission." },
  { type: "schedule",                        label: "On a schedule",         description: "Cron-style timed trigger." },
  { type: "webhook",                         label: "External webhook",      description: "POST to a workflow-specific URL fires this trigger." },
]

export function getTriggerSpec(type: string): TriggerSpec | null {
  return TRIGGERS.find((t) => t.type === type) ?? null
}

// ── Condition operators ──────────────────────────────────────────────────────

export type ConditionOp =
  | "eq" | "neq"
  | "gt" | "gte" | "lt" | "lte"
  | "contains" | "starts_with" | "ends_with"
  | "in" | "not_in"
  | "empty" | "not_empty"

export const CONDITION_OPS: { value: ConditionOp; label: string }[] = [
  { value: "eq",          label: "equals" },
  { value: "neq",         label: "does not equal" },
  { value: "gt",          label: ">" },
  { value: "gte",         label: "≥" },
  { value: "lt",          label: "<" },
  { value: "lte",         label: "≤" },
  { value: "contains",    label: "contains" },
  { value: "starts_with", label: "starts with" },
  { value: "ends_with",   label: "ends with" },
  { value: "in",          label: "is one of" },
  { value: "not_in",      label: "is not one of" },
  { value: "empty",       label: "is empty" },
  { value: "not_empty",   label: "is not empty" },
]

export interface Condition {
  field: string
  op: ConditionOp
  value?: unknown
}
