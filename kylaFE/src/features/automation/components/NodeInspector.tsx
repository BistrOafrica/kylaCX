import { IconTrash, IconX } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"
import {
  ACTION_BY_TYPE,
  TRIGGERS,
  CONDITION_OPS,
  type ActionParam,
  type ActionType,
  type TriggerType,
  type Condition,
} from "../utils/actions"
import type { Node } from "@xyflow/react"

/**
 * NodeInspector — right-side editor for whichever node is currently
 * selected on the canvas. Reads/writes node data via the
 * `onPatch` / `onDelete` callbacks the parent canvas owns.
 */
export function NodeInspector({
  node,
  onPatch,
  onDelete,
  onClose,
}: {
  node: Node | null
  onPatch: (id: string, patch: Record<string, unknown>) => void
  onDelete: (id: string) => void
  onClose: () => void
}) {
  if (!node) return null

  return (
    <aside
      aria-label="Node inspector"
      className="w-80 shrink-0 overflow-y-auto bg-surface border-s border-border"
    >
      <header className="flex items-center gap-2 px-3 h-9 border-b border-border">
        <span className="text-md font-medium text-fg flex-1 truncate">
          {nodeTitle(node)}
        </span>
        <button
          type="button"
          onClick={() => onDelete(node.id)}
          aria-label="Delete node"
          className={cn(
            "inline-flex items-center justify-center size-6 rounded-xs",
            "text-fg-muted hover:text-danger hover:bg-subtle",
          )}
        >
          <IconTrash className="size-3.5" />
        </button>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close inspector"
          className={cn(
            "inline-flex items-center justify-center size-6 rounded-xs",
            "text-fg-muted hover:text-fg hover:bg-subtle",
          )}
        >
          <IconX className="size-3.5" />
        </button>
      </header>

      <div className="p-3">
        {node.type === "trigger" && (
          <TriggerEditor
            value={(node.data as { triggerType: TriggerType | null }).triggerType}
            config={(node.data as { config?: Record<string, unknown> }).config ?? {}}
            onChange={(triggerType, config) => onPatch(node.id, { triggerType, config })}
          />
        )}
        {node.type === "condition" && (
          <ConditionEditor
            conditions={(node.data as { conditions: Condition[] }).conditions ?? []}
            onChange={(conditions) => onPatch(node.id, { conditions })}
          />
        )}
        {node.type === "action" && (
          <ActionEditor
            actionType={(node.data as { actionType: ActionType }).actionType}
            params={(node.data as { params: Record<string, unknown> }).params ?? {}}
            onPatchParam={(k, v) =>
              onPatch(node.id, {
                params: {
                  ...((node.data as { params: Record<string, unknown> }).params ?? {}),
                  [k]: v,
                },
              })
            }
          />
        )}
      </div>
    </aside>
  )
}

function nodeTitle(node: Node): string {
  if (node.type === "trigger") return "Trigger"
  if (node.type === "condition") return "If condition"
  if (node.type === "action") {
    const t = (node.data as { actionType: ActionType }).actionType
    return ACTION_BY_TYPE[t]?.label ?? "Action"
  }
  return "Node"
}

// ── Trigger editor ───────────────────────────────────────────────────────────

function TriggerEditor({
  value,
  config,
  onChange,
}: {
  value: TriggerType | null
  config: Record<string, unknown>
  onChange: (type: TriggerType | null, config: Record<string, unknown>) => void
}) {
  return (
    <div className="space-y-3">
      <Field label="Trigger event">
        <select
          value={value ?? ""}
          onChange={(e) =>
            onChange((e.target.value || null) as TriggerType | null, config)
          }
          className="w-full h-8 rounded-sm border border-border bg-surface px-2 text-base"
        >
          <option value="">Pick a trigger…</option>
          {TRIGGERS.map((t) => (
            <option key={t.type} value={t.type}>
              {t.label}
            </option>
          ))}
        </select>
      </Field>
      <Field label="Filter config (JSON)" help="Pattern matched against the event payload.">
        <JsonField
          value={config}
          onChange={(v) => onChange(value, v)}
        />
      </Field>
    </div>
  )
}

// ── Condition editor ─────────────────────────────────────────────────────────

function ConditionEditor({
  conditions,
  onChange,
}: {
  conditions: Condition[]
  onChange: (next: Condition[]) => void
}) {
  const update = (i: number, patch: Partial<Condition>) => {
    const next = [...conditions]
    next[i] = { ...next[i]!, ...patch }
    onChange(next)
  }

  return (
    <div className="space-y-3">
      {conditions.map((c, i) => (
        <div key={i} className="space-y-1.5 p-2 rounded-sm border border-border bg-canvas">
          <Field label="Field">
            <Input
              value={c.field}
              onChange={(e) => update(i, { field: e.target.value })}
              placeholder="event.priority"
              className="h-7"
            />
          </Field>
          <Field label="Operator">
            <select
              value={c.op}
              onChange={(e) => update(i, { op: e.target.value as Condition["op"] })}
              className="w-full h-7 rounded-sm border border-border bg-surface px-2 text-base"
            >
              {CONDITION_OPS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </Field>
          {c.op !== "empty" && c.op !== "not_empty" && (
            <Field label="Value">
              <Input
                value={String(c.value ?? "")}
                onChange={(e) => update(i, { value: e.target.value })}
                className="h-7"
              />
            </Field>
          )}
          <button
            type="button"
            onClick={() => onChange(conditions.filter((_, idx) => idx !== i))}
            className="text-sm text-danger hover:underline"
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() =>
          onChange([...conditions, { field: "", op: "eq", value: "" }])
        }
        className="w-full text-md h-8 rounded-sm border border-border hover:bg-subtle"
      >
        + Add condition
      </button>
    </div>
  )
}

// ── Action editor ────────────────────────────────────────────────────────────

function ActionEditor({
  actionType,
  params,
  onPatchParam,
}: {
  actionType: ActionType
  params: Record<string, unknown>
  onPatchParam: (key: string, value: unknown) => void
}) {
  const spec = ACTION_BY_TYPE[actionType]
  if (!spec) return <div className="text-sm text-fg-muted">Unknown action type.</div>

  return (
    <div className="space-y-3">
      <div className="text-sm text-fg-muted">{spec.description}</div>
      {spec.params.map((p) => (
        <Field key={p.key} label={`${p.label}${p.required ? " *" : ""}`} help={p.helpText}>
          <ParamInput
            param={p}
            value={params[p.key]}
            onChange={(v) => onPatchParam(p.key, v)}
          />
        </Field>
      ))}
    </div>
  )
}

function ParamInput({
  param,
  value,
  onChange,
}: {
  param: ActionParam
  value: unknown
  onChange: (v: unknown) => void
}) {
  switch (param.type) {
    case "textarea":
      return (
        <textarea
          value={String(value ?? "")}
          onChange={(e) => onChange(e.target.value)}
          placeholder={param.placeholder}
          rows={3}
          className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent placeholder:text-fg-muted"
        />
      )
    case "number":
      return (
        <Input
          type="number"
          value={value === undefined || value === null ? "" : String(value)}
          onChange={(e) =>
            onChange(e.target.value === "" ? "" : Number(e.target.value))
          }
          className="h-7"
        />
      )
    case "boolean":
      return (
        <label className="flex items-center gap-2 text-base">
          <input
            type="checkbox"
            checked={Boolean(value)}
            onChange={(e) => onChange(e.target.checked)}
            className="accent-emerald-500"
          />
          {param.label}
        </label>
      )
    case "select":
      return (
        <select
          value={String(value ?? "")}
          onChange={(e) => onChange(e.target.value)}
          className="w-full h-7 rounded-sm border border-border bg-surface px-2 text-base"
        >
          <option value="">Select…</option>
          {param.options?.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      )
    case "json":
      return (
        <JsonField
          value={(value as Record<string, unknown>) ?? {}}
          onChange={onChange}
        />
      )
    case "text":
    default:
      return (
        <Input
          value={String(value ?? "")}
          onChange={(e) => onChange(e.target.value)}
          placeholder={param.placeholder}
          className="h-7"
        />
      )
  }
}

function JsonField({
  value,
  onChange,
}: {
  value: Record<string, unknown>
  onChange: (v: Record<string, unknown>) => void
}) {
  const text = JSON.stringify(value, null, 2)
  return (
    <textarea
      value={text}
      onChange={(e) => {
        try {
          const parsed = JSON.parse(e.target.value) as Record<string, unknown>
          onChange(parsed)
        } catch {
          // ignore until JSON parses cleanly
        }
      }}
      rows={6}
      className="w-full font-mono text-sm rounded-sm border border-border bg-canvas px-2 py-1.5 outline-none focus:border-accent"
    />
  )
}

function Field({
  label,
  help,
  children,
}: {
  label: string
  help?: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1">
      <label className="block text-xs font-mono uppercase tracking-wider text-fg-muted">
        {label}
      </label>
      {children}
      {help && <p className="text-sm text-fg-muted">{help}</p>}
    </div>
  )
}
