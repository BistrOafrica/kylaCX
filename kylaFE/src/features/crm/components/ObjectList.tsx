import { useEffect, useMemo, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { useVirtualizer } from "@tanstack/react-virtual"
import {
  IconSearch,
  IconRefresh,
  IconPlus,
  IconUsers,
} from "@tabler/icons-react"
import { ListRowSkeleton, EmptyState, ErrorState, PageHeader } from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { defaultColumns } from "../utils/fields"
import type { ObjectRecord } from "../utils/types"
import { FieldValue } from "./FieldValue"
import { useObjects, useObjectType } from "../hooks/queries"

const ROW_HEIGHT = 36

/**
 * Schema-aware object list — the workhorse for Contacts, Companies,
 * Leads, Deals, and any future custom type.
 *
 *   <ObjectList typeSlug="contact" title="Contacts" />
 *
 * Columns are derived from the ObjectType schema; the first ~5 fields
 * are surfaced by default. Saved Views (F2.x) will let users pick
 * columns per view.
 */
export interface ObjectListProps {
  typeSlug: string
  title: string
  description?: string
  basePath: string                 // e.g. "/crm/contacts"
}

export function ObjectList({ typeSlug, title, description, basePath }: ObjectListProps) {
  const [search, setSearch] = useState("")
  const objectType = useObjectType(typeSlug)
  const query = useObjects({ typeSlug })

  const fields = useMemo(
    () => objectType.data?.schema?.fields ?? [],
    [objectType.data],
  )
  const columns = useMemo(() => defaultColumns(fields), [fields])

  const records: ObjectRecord[] = useMemo(
    () => query.data?.pages.flatMap((p) => p.records) ?? [],
    [query.data],
  )

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return records
    return records.filter((r) =>
      JSON.stringify(r.data).toLowerCase().includes(q),
    )
  }, [records, search])

  const total = query.data?.pages[query.data.pages.length - 1]?.total

  const parentRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  })

  useEffect(() => {
    const items = virtualizer.getVirtualItems()
    const last = items[items.length - 1]
    if (!last) return
    if (
      last.index >= filtered.length - 4 &&
      query.hasNextPage &&
      !query.isFetchingNextPage
    ) {
      void query.fetchNextPage()
    }
  }, [virtualizer, filtered.length, query])

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title={title}
        description={description ?? objectType.data?.pluralName ?? ""}
        actions={
          <button
            type="button"
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover"
          >
            <IconPlus className="size-3.5" />
            New
          </button>
        }
      />

      <div className="flex items-center gap-2 px-3 h-9 border-b border-border bg-canvas">
        <div className="relative flex-1 max-w-md">
          <IconSearch
            className="absolute start-2 top-1/2 -translate-y-1/2 size-3.5 text-fg-muted"
            aria-hidden
          />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search…"
            className="h-7 ps-7 text-base"
            aria-label="Search records"
          />
        </div>
        {(total !== undefined || filtered.length > 0) && (
          <span className="ms-auto font-mono text-xs text-fg-muted">
            {filtered.length}{total ? ` / ${total}` : ""}
          </span>
        )}
      </div>

      {objectType.isPending || query.isPending ? (
        <div className="flex-1 overflow-hidden p-2">
          <ListRowSkeleton count={12} />
        </div>
      ) : objectType.isError || query.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load records"
            description={
              (objectType.error as Error | undefined)?.message ??
              (query.error as Error | undefined)?.message
            }
            onRetry={() => {
              void objectType.refetch()
              void query.refetch()
            }}
          />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconUsers className="size-5" />}
            title={search ? "No matches" : `No ${title.toLowerCase()} yet`}
            description={
              search
                ? "Try a different search term or clear filters."
                : "Records you create or import will appear here."
            }
            action={
              <button
                type="button"
                onClick={() => void query.refetch()}
                className="inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md border border-border hover:bg-subtle"
              >
                <IconRefresh className="size-3.5" />
                Refresh
              </button>
            }
          />
        </div>
      ) : (
        <>
          <ColumnHeader columns={columns} />
          <div
            ref={parentRef}
            className="flex-1 overflow-y-auto outline-none"
            tabIndex={-1}
            role="listbox"
            aria-label={title}
          >
            <div
              style={{
                height: virtualizer.getTotalSize(),
                position: "relative",
              }}
            >
              {virtualizer.getVirtualItems().map((item) => {
                const record = filtered[item.index]
                if (!record) return null
                return (
                  <div
                    key={record.id}
                    style={{
                      position: "absolute",
                      top: 0,
                      insetInlineStart: 0,
                      insetInlineEnd: 0,
                      transform: `translateY(${item.start}px)`,
                      height: item.size,
                    }}
                  >
                    <Row
                      record={record}
                      columns={columns}
                      basePath={basePath}
                    />
                  </div>
                )
              })}
            </div>
            {query.isFetchingNextPage && (
              <div className="p-2">
                <ListRowSkeleton count={3} />
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}

function ColumnHeader({
  columns,
}: {
  columns: ReturnType<typeof defaultColumns>
}) {
  return (
    <div
      role="row"
      className={cn(
        "flex items-center h-7 px-3 gap-3 border-b border-border bg-subtle",
        "text-xs font-mono uppercase tracking-wider text-fg-muted",
      )}
    >
      <span className="w-48 truncate">Name</span>
      {columns.map((c) => (
        <span key={c.key} className="flex-1 min-w-0 truncate">
          {c.label}
        </span>
      ))}
      <span className="w-20 text-end">Updated</span>
    </div>
  )
}

function Row({
  record,
  columns,
  basePath,
}: {
  record: ObjectRecord
  columns: ReturnType<typeof defaultColumns>
  basePath: string
}) {
  return (
    <Link
      to={`${basePath}/${record.id}`}
      className={cn(
        "flex items-center gap-3 px-3 h-9 border-b border-border-subtle",
        "hover:bg-subtle transition-colors text-base text-fg",
      )}
    >
      <span className="w-48 truncate font-medium">
        {recordName(record, columns)}
      </span>
      {columns.map((c) => (
        <span key={c.key} className="flex-1 min-w-0 truncate">
          <FieldValue field={c} value={record.data[c.key]} />
        </span>
      ))}
      <span className="w-20 text-end font-mono text-xs text-fg-muted">
        {relativeShort(record.updatedAt)}
      </span>
    </Link>
  )
}

function recordName(
  record: ObjectRecord,
  columns: ReturnType<typeof defaultColumns>,
): string {
  // Prefer a `name` field if the schema has one; else fall back to
  // first non-empty value; else show the ID prefix.
  const named = record.data.name ?? record.data.full_name ?? record.data.title
  if (typeof named === "string" && named) return named

  for (const c of columns) {
    const v = record.data[c.key]
    if (typeof v === "string" && v) return v
  }
  return record.id.slice(0, 8) + "…"
}

function relativeShort(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  const diffMs = Date.now() - d.getTime()
  const min = Math.floor(diffMs / 60_000)
  if (min < 1) return "now"
  if (min < 60) return `${min}m`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h}h`
  const days = Math.floor(h / 24)
  if (days < 30) return `${days}d`
  return `${Math.floor(days / 30)}mo`
}
