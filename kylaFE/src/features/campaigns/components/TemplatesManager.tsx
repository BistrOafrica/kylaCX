import { useState } from "react"
import {
  IconTemplate,
  IconTrash,
  IconPlus,
  IconLoader2,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  Surface,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useWhatsappTemplates,
  useCreateWhatsappTemplate,
  useDeleteWhatsappTemplate,
} from "../hooks/queries"
import type { WhatsappTemplate } from "../api/whatsapp"

const CATEGORIES = ["MARKETING", "UTILITY", "AUTHENTICATION"]
const LANGUAGES = ["en", "ar", "fr", "sw", "es", "pt"]

/**
 * TemplatesManager — full-page editor for WhatsApp message templates.
 *
 * Two-column layout: list on the left, editor on the right. Saving
 * creates a new template (status = PENDING approval); delete removes
 * from Meta + our DB.
 */
export function TemplatesManager() {
  const templates = useWhatsappTemplates()
  const [activeId, setActiveId] = useState<string | null>(null)

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="WhatsApp templates"
        description="Approved message templates for outbound campaigns"
      />

      {templates.isError ? (
        <ErrorState
          title="Couldn't load templates"
          description={(templates.error as Error).message}
          onRetry={() => void templates.refetch()}
        />
      ) : (
        <div className="flex-1 min-h-0 grid grid-cols-12 gap-0 border-t border-border">
          <aside className="col-span-4 border-e border-border flex flex-col">
            <header className="flex items-center gap-2 px-3 h-9 border-b border-border">
              <span className="text-md font-medium text-fg flex-1">Templates</span>
              <button
                type="button"
                onClick={() => setActiveId("new")}
                aria-label="New template"
                className="inline-flex items-center justify-center size-6 rounded-xs text-fg-muted hover:text-fg hover:bg-subtle"
              >
                <IconPlus className="size-3.5" />
              </button>
            </header>
            {templates.isPending ? (
              <div className="p-2">
                <ListRowSkeleton count={4} />
              </div>
            ) : (templates.data?.length ?? 0) === 0 && activeId !== "new" ? (
              <EmptyState
                icon={<IconTemplate className="size-5" />}
                title="No templates yet"
                description="Create one to launch a campaign."
                size="sm"
              />
            ) : (
              <ul className="flex-1 overflow-y-auto p-1 space-y-px" role="list">
                {activeId === "new" && (
                  <li>
                    <button
                      type="button"
                      onClick={() => setActiveId("new")}
                      className="w-full flex items-center gap-2 px-2 h-8 rounded-sm bg-accent-subtle text-fg text-md"
                    >
                      <IconTemplate className="size-3.5 text-fg-muted" />
                      <span className="truncate">(new template)</span>
                    </button>
                  </li>
                )}
                {(templates.data ?? []).map((t) => (
                  <li key={t.id}>
                    <button
                      type="button"
                      onClick={() => setActiveId(t.id)}
                      className={cn(
                        "w-full flex items-center gap-2 px-2 h-8 rounded-sm text-md text-start",
                        "hover:bg-subtle transition-colors",
                        activeId === t.id && "bg-accent-subtle font-medium",
                      )}
                    >
                      <IconTemplate className="size-3.5 text-fg-muted" />
                      <span className="truncate flex-1">{t.name}</span>
                      <StatusPill status={t.status} />
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </aside>
          <main className="col-span-8 overflow-y-auto p-4">
            {activeId === "new" ? (
              <Editor template={null} onSaved={(id) => setActiveId(id)} />
            ) : activeId ? (
              <Editor
                template={
                  templates.data?.find((t) => t.id === activeId) ?? null
                }
                onSaved={(id) => setActiveId(id)}
              />
            ) : (
              <EmptyState
                icon={<IconTemplate className="size-5" />}
                title="Pick a template"
                description="Or create a new one to start editing."
              />
            )}
          </main>
        </div>
      )}
    </div>
  )
}

function StatusPill({ status }: { status: string }) {
  const s = status.toUpperCase()
  const cls =
    s === "APPROVED"
      ? "bg-success-subtle text-success"
      : s === "REJECTED"
        ? "bg-danger-subtle text-danger"
        : "bg-warn-subtle text-warn"
  return (
    <span
      className={cn(
        "inline-flex items-center h-4 px-1 rounded-xs text-[10px] font-mono uppercase",
        cls,
      )}
    >
      {s || "PENDING"}
    </span>
  )
}

function Editor({
  template,
  onSaved,
}: {
  template: WhatsappTemplate | null
  onSaved: (id: string | null) => void
}) {
  const create = useCreateWhatsappTemplate()
  const remove = useDeleteWhatsappTemplate()

  const [name, setName] = useState(template?.name ?? "")
  const [language, setLanguage] = useState(template?.language || "en")
  const [category, setCategory] = useState(template?.category || "MARKETING")
  const [header, setHeader] = useState(template?.header ?? "")
  const [body, setBody] = useState(template?.body ?? "")
  const [footer, setFooter] = useState(template?.footer ?? "")
  const isNew = template === null

  const onSubmit = async () => {
    if (!name.trim() || !body.trim()) {
      toast.error("Name and body are required")
      return
    }
    if (isNew) {
      const created = await create.mutateAsync({
        name,
        language,
        category,
        header,
        body,
        footer,
      })
      if (created?.id) {
        toast.success("Template submitted for approval")
        onSaved(created.id)
      }
    } else {
      toast.info("Update RPC ships when proto exposes UpdateWhatsappTemplate")
    }
  }

  const onDelete = async () => {
    if (!template?.id) return
    if (!confirm(`Delete template "${template.name}"?`)) return
    await remove.mutateAsync(template.id)
    onSaved(null)
  }

  return (
    <Surface level={1} radius="md" className="p-4 space-y-4 max-w-2xl">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <Field label="Name *">
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="welcome_offer_v2"
            className="h-8 font-mono"
          />
        </Field>
        <Field label="Language">
          <select
            value={language}
            onChange={(e) => setLanguage(e.target.value)}
            className="w-full h-8 rounded-sm border border-border bg-surface px-2 text-base"
          >
            {LANGUAGES.map((l) => (
              <option key={l} value={l}>{l}</option>
            ))}
          </select>
        </Field>
        <Field label="Category">
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full h-8 rounded-sm border border-border bg-surface px-2 text-base"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </Field>
      </div>

      <Field label="Header (optional)">
        <Input
          value={header}
          onChange={(e) => setHeader(e.target.value)}
          placeholder="Welcome to Kyla!"
          className="h-8"
        />
      </Field>
      <Field label="Body *">
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={6}
          placeholder="Hi {{1}}, thanks for joining…"
          className="w-full rounded-sm border border-border bg-canvas px-3 py-2 text-base outline-none focus:border-accent"
        />
      </Field>
      <Field label="Footer (optional)">
        <Input
          value={footer}
          onChange={(e) => setFooter(e.target.value)}
          placeholder="Reply STOP to unsubscribe"
          className="h-8"
        />
      </Field>

      <div className="flex items-center justify-between pt-2 border-t border-border">
        {template?.id ? (
          <button
            type="button"
            onClick={() => void onDelete()}
            disabled={remove.isPending}
            className="inline-flex items-center gap-1 h-8 px-2.5 rounded-sm text-md text-danger hover:bg-danger-subtle disabled:opacity-50"
          >
            <IconTrash className="size-3.5" />
            Delete
          </button>
        ) : <span />}
        <button
          type="button"
          onClick={() => void onSubmit()}
          disabled={create.isPending || !name.trim() || !body.trim()}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          {create.isPending && <IconLoader2 className="size-3.5 animate-spin" />}
          {isNew ? "Submit for approval" : "Save"}
        </button>
      </div>
    </Surface>
  )
}

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1">
      <label className="block text-xs font-mono uppercase tracking-wider text-fg-muted">
        {label}
      </label>
      {children}
    </div>
  )
}
