import { Link } from "react-router-dom"
import { useMemo, useState } from "react"
import {
  IconArrowLeft,
  IconCopy,
  IconFileText,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  CardSkeleton,
  ErrorState,
  Surface,
  EmptyState,
  ListRowSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"
import { useForm, useFormSubmissions } from "../hooks/queries"
import {
  parseFields,
  type FormFieldSpec,
} from "../api/forms"
import { relativeShort } from "@/features/crm/utils/relative"
import { env } from "@/lib/rpc"

/**
 * FormDetail — form schema preview + submissions inbox.
 *
 * F3 ships read + submissions; the builder UI (drag-drop fields,
 * conditional logic) lands in F3.x. For now, schema view is read-only;
 * editing falls back to the JSON.
 */
export function FormDetail({ formId }: { formId: string }) {
  const form = useForm(formId)
  const submissions = useFormSubmissions(formId)
  const [tab, setTab] = useState<"schema" | "submissions" | "embed">("schema")

  if (form.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
      </div>
    )
  }

  if (form.isError || !form.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load form"
          description={(form.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void form.refetch()}
        />
      </div>
    )
  }

  const fields = parseFields(form.data.fields)

  return (
    <div className="flex flex-col h-full bg-canvas">
      <header className="flex items-start gap-3 px-6 py-3 border-b border-border">
        <Link
          to="/forms"
          className={cn(
            "inline-flex items-center justify-center size-7 rounded-sm",
            "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
          )}
          aria-label="Back to forms"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className={cn(
            "size-10 rounded-md flex items-center justify-center shrink-0",
            "bg-accent-subtle text-fg-secondary",
          )}
          aria-hidden
        >
          <IconFileText className="size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
            {form.data.name}
          </h1>
          <div className="text-base text-fg-muted">{form.data.description}</div>
        </div>
      </header>

      <nav
        role="tablist"
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
      >
        {(
          [
            { id: "schema", label: "Schema" },
            { id: "submissions", label: "Submissions" },
            { id: "embed", label: "Embed" },
          ] as const
        ).map((t) => (
          <button
            key={t.id}
            role="tab"
            aria-selected={tab === t.id}
            onClick={() => setTab(t.id)}
            className={cn(
              "inline-flex items-center h-7 px-2.5 rounded-sm text-md",
              "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
              tab === t.id && "bg-accent-subtle text-fg font-medium",
            )}
          >
            {t.label}
          </button>
        ))}
      </nav>

      <div className="flex-1 min-h-0 overflow-y-auto p-6">
        {tab === "schema" && <SchemaView fields={fields} />}
        {tab === "submissions" && (
          <SubmissionsView
            isPending={submissions.isPending}
            error={submissions.error as Error | undefined}
            submissions={submissions.data?.submissions ?? []}
            total={submissions.data?.total ?? 0}
            fields={fields}
            refetch={submissions.refetch}
          />
        )}
        {tab === "embed" && <EmbedView formId={formId} />}
      </div>
    </div>
  )
}

function SchemaView({ fields }: { fields: FormFieldSpec[] }) {
  if (fields.length === 0) {
    return (
      <EmptyState
        icon={<IconFileText className="size-5" />}
        title="No fields yet"
        description="Define the schema for this form. F3.x adds a drag-drop builder."
      />
    )
  }
  return (
    <Surface level={1} radius="md" className="p-4 max-w-2xl">
      <ol className="space-y-2">
        {fields.map((f, i) => (
          <li
            key={f.key}
            className="flex items-center gap-3 px-2 py-2 rounded-sm border border-border-subtle"
          >
            <span className="font-mono text-xs text-fg-muted w-6 text-end">
              {i + 1}
            </span>
            <div className="min-w-0 flex-1">
              <div className="text-md font-medium text-fg">
                {f.label}
                {f.required && (
                  <span className="text-danger ms-1" aria-label="Required">*</span>
                )}
              </div>
              <div className="text-sm text-fg-muted font-mono">
                {f.key} · {f.type}
              </div>
              {f.helpText && (
                <div className="text-sm text-fg-muted">{f.helpText}</div>
              )}
            </div>
          </li>
        ))}
      </ol>
    </Surface>
  )
}

function SubmissionsView({
  isPending,
  error,
  submissions,
  total,
  fields,
  refetch,
}: {
  isPending: boolean
  error?: Error
  submissions: { id: string; formId: string; data: string; createdAt: string }[]
  total: number
  fields: FormFieldSpec[]
  refetch: () => void
}) {
  if (isPending) return <ListRowSkeleton count={6} />
  if (error) {
    return (
      <ErrorState
        title="Couldn't load submissions"
        description={error.message}
        onRetry={() => void refetch()}
      />
    )
  }
  if (submissions.length === 0) {
    return (
      <EmptyState
        icon={<IconFileText className="size-5" />}
        title="No submissions yet"
        description="Once respondents start submitting, their answers land here."
      />
    )
  }
  return (
    <div className="space-y-2 max-w-3xl">
      <div className="text-sm text-fg-muted">
        {total} submission{total === 1 ? "" : "s"}
      </div>
      <ol className="space-y-2">
        {submissions.map((sub) => {
          let parsed: Record<string, unknown> = {}
          try {
            parsed = JSON.parse(sub.data) as Record<string, unknown>
          } catch {
            /* ignore */
          }
          return (
            <li
              key={sub.id}
              className="rounded-md border border-border bg-surface p-3"
            >
              <div className="text-xs font-mono text-fg-muted">
                {sub.id.slice(0, 12)} · {relativeShort(sub.createdAt)}
              </div>
              <dl className="mt-2 grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1.5">
                {fields.map((f) => (
                  <div key={f.key} className="flex flex-col">
                    <dt className="text-xs font-mono uppercase tracking-wider text-fg-muted">
                      {f.label}
                    </dt>
                    <dd className="text-base text-fg break-words">
                      {String(parsed[f.key] ?? "—")}
                    </dd>
                  </div>
                ))}
              </dl>
            </li>
          )
        })}
      </ol>
    </div>
  )
}

function EmbedView({ formId }: { formId: string }) {
  const embedUrl = useMemo(
    () => `${env.apiUrl.replace(/\/$/, "")}/forms/${formId}/embed.js`,
    [formId],
  )
  const snippet = useMemo(
    () =>
      `<!-- Kyla form embed -->
<div id="kyla-form-${formId}"></div>
<script async src="${embedUrl}"></script>`,
    [formId, embedUrl],
  )

  const onCopy = async () => {
    await navigator.clipboard.writeText(snippet)
    toast.success("Embed snippet copied")
  }

  return (
    <Surface level={1} radius="md" className="p-4 max-w-2xl space-y-3">
      <div>
        <h2 className="text-md font-semibold text-fg">Embed snippet</h2>
        <p className="text-base text-fg-muted">
          Drop this anywhere on your site to render the live form.
        </p>
      </div>
      <label className="block">
        <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          Public URL
        </span>
        <Input
          readOnly
          value={embedUrl}
          className="h-8 mt-1 font-mono text-sm"
        />
      </label>
      <div>
        <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          HTML snippet
        </span>
        <pre className="mt-1 p-3 rounded-sm bg-subtle border border-border text-sm font-mono whitespace-pre-wrap text-fg overflow-x-auto">
          {snippet}
        </pre>
      </div>
      <button
        type="button"
        onClick={() => void onCopy()}
        className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md border border-border hover:bg-subtle"
      >
        <IconCopy className="size-3.5" />
        Copy snippet
      </button>
    </Surface>
  )
}
