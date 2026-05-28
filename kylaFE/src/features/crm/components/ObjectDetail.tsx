import { Link } from "react-router-dom"
import { IconArrowLeft, IconUsers } from "@tabler/icons-react"
import { Surface, ErrorState, CardSkeleton } from "@/design-system"
import { cn } from "@/lib/utils"
import { useObject, useObjectType } from "../hooks/queries"
import type { ObjectRecord } from "../utils/types"
import { FieldValue } from "./FieldValue"
import { ObjectTimeline } from "./ObjectTimeline"
import { ObjectRelations } from "./ObjectRelations"
import { relativeShort } from "../utils/relative"

/**
 * ObjectDetail — schema-aware detail view used by Contacts / Companies /
 * Leads / Deals / any object type.
 *
 * Layout:
 *   ┌──────────────────────────────┬──────────────┐
 *   │ Header + identity            │ Sidebar      │
 *   │ Custom fields grid           │  - Timeline  │
 *   │ Linked objects               │  - Relations │
 *   └──────────────────────────────┴──────────────┘
 */
export interface ObjectDetailProps {
  objectId: string
  backHref: string
  backLabel: string
}

export function ObjectDetail({ objectId, backHref, backLabel }: ObjectDetailProps) {
  const obj = useObject(objectId)
  const type = useObjectType(obj.data?.typeSlug)

  if (obj.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
        <CardSkeleton lines={3} />
      </div>
    )
  }

  if (obj.isError || !obj.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load record"
          description={(obj.error as Error | undefined)?.message}
          onRetry={() => void obj.refetch()}
        />
      </div>
    )
  }

  const record = obj.data
  const fields = type.data?.schema?.fields ?? []
  const title = primaryName(record, fields.map((f) => f.key))

  return (
    <div className="h-full overflow-hidden flex">
      <div className="flex-1 min-w-0 overflow-y-auto">
        <header className="flex items-start gap-3 px-6 py-4 border-b border-border">
          <Link
            to={backHref}
            className={cn(
              "inline-flex items-center justify-center size-7 rounded-sm",
              "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
            )}
            aria-label={`Back to ${backLabel}`}
          >
            <IconArrowLeft className="size-4" />
          </Link>
          <div
            className={cn(
              "size-10 rounded-md flex items-center justify-center",
              "bg-accent-subtle text-fg-secondary",
            )}
            aria-hidden
          >
            <IconUsers className="size-5" />
          </div>
          <div className="min-w-0 flex-1">
            <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
              {title}
            </h1>
            <div className="text-base text-fg-muted">
              {type.data?.name ?? record.typeSlug} · created{" "}
              {relativeShort(record.createdAt)}
            </div>
          </div>
        </header>

        <section className="p-6 space-y-6">
          <Surface level={1} radius="md" className="p-4 space-y-4">
            <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
              Fields
            </h2>
            <FieldGrid record={record} fields={fields} />
          </Surface>

          <Surface level={1} radius="md" className="p-4 space-y-3">
            <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
              Related
            </h2>
            <ObjectRelations objectId={objectId} />
          </Surface>
        </section>
      </div>

      <aside className="w-80 shrink-0 overflow-y-auto border-s border-border bg-surface">
        <div className="p-4 space-y-4">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            Activity
          </h2>
          <ObjectTimeline objectId={objectId} />
        </div>
      </aside>
    </div>
  )
}

function FieldGrid({
  record,
  fields,
}: {
  record: ObjectRecord
  fields: { key: string; label: string; type: number }[]
}) {
  if (fields.length === 0) {
    return (
      <div className="text-base text-fg-muted">
        This object type has no schema fields. Define some in the schema editor.
      </div>
    )
  }
  return (
    <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2.5">
      {fields.map((f) => (
        <div key={f.key} className="flex flex-col">
          <dt className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            {f.label}
          </dt>
          <dd className="text-base text-fg">
            <FieldValue
              field={f as Parameters<typeof FieldValue>[0]["field"]}
              value={record.data[f.key]}
            />
          </dd>
        </div>
      ))}
    </dl>
  )
}

function primaryName(
  record: ObjectRecord,
  fieldKeys: string[],
): string {
  for (const key of ["name", "full_name", "title", "subject", "deal_name"]) {
    const v = record.data[key]
    if (typeof v === "string" && v) return v
  }
  for (const key of fieldKeys) {
    const v = record.data[key]
    if (typeof v === "string" && v) return v
  }
  return record.id.slice(0, 12)
}
