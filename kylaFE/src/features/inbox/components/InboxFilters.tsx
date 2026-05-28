import { useTranslation } from "react-i18next"
import { IconFilter, IconX, IconSearch } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"
import {
  CHANNEL_META,
  STATUS_META,
  PRIORITY_META,
  type ConversationStatus,
  type ConversationPriority,
  type Channel,
} from "../utils/enums"
import type { InboxFilters } from "../api/conversations"

/**
 * Compact filter strip above the inbox list.
 *
 * Channel / status / priority pill rows toggle on click; the search
 * box drives the `query` filter. Active filter count visible in the
 * left badge so agents know what's narrowing their view.
 */
export interface InboxFiltersProps {
  value: InboxFilters
  onChange: (next: InboxFilters) => void
  total?: number
  loaded?: number
}

export function InboxFiltersBar({
  value,
  onChange,
  total,
  loaded,
}: InboxFiltersProps) {
  const { t } = useTranslation()
  const activeCount = countActiveFilters(value)

  return (
    <div className="flex flex-col bg-canvas border-b border-border">
      <div className="flex items-center gap-2 px-3 h-9">
        <div className="relative flex-1 max-w-md">
          <IconSearch className="absolute start-2 top-1/2 -translate-y-1/2 size-3.5 text-fg-muted" aria-hidden />
          <Input
            value={value.query ?? ""}
            onChange={(e) => onChange({ ...value, query: e.target.value })}
            placeholder={t("common.search")}
            className="h-7 ps-7 text-base"
            aria-label={t("common.search")}
          />
        </div>
        <button
          type="button"
          onClick={() => onChange({ ...value, activeOnly: !value.activeOnly })}
          className={cn(
            "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm",
            "border border-border hover:bg-subtle",
            value.activeOnly && "bg-accent-subtle text-fg border-accent/30",
          )}
          aria-pressed={value.activeOnly ?? false}
        >
          <IconFilter className="size-3" />
          Active only
        </button>

        {activeCount > 0 && (
          <button
            type="button"
            onClick={() => onChange({ activeOnly: value.activeOnly })}
            className="inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm text-fg-muted hover:text-fg hover:bg-subtle"
          >
            <IconX className="size-3" />
            Clear ({activeCount})
          </button>
        )}

        {(total !== undefined || loaded !== undefined) && (
          <span className="ms-auto font-mono text-xs text-fg-muted">
            {loaded ?? 0}{total ? ` / ${total}` : ""}
          </span>
        )}
      </div>

      <PillRow
        label="Channels"
        items={CHANNEL_META.map((m) => ({
          key: m.protoEnum,
          label: m.label,
          active: value.channels?.includes(m.protoEnum) ?? false,
        }))}
        onToggle={(key) => {
          const k = key as Channel
          const set = new Set(value.channels ?? [])
          if (set.has(k)) set.delete(k)
          else set.add(k)
          onChange({ ...value, channels: set.size ? [...set] : undefined })
        }}
      />
      <PillRow
        label="Status"
        items={STATUS_META.map((m) => ({
          key: m.protoEnum,
          label: m.label,
          active: value.statuses?.includes(m.protoEnum) ?? false,
        }))}
        onToggle={(key) => {
          const k = key as ConversationStatus
          const set = new Set(value.statuses ?? [])
          if (set.has(k)) set.delete(k)
          else set.add(k)
          onChange({ ...value, statuses: set.size ? [...set] : undefined })
        }}
      />
      <PillRow
        label="Priority"
        items={PRIORITY_META.map((m) => ({
          key: m.protoEnum,
          label: m.label,
          active: value.priorities?.includes(m.protoEnum) ?? false,
        }))}
        onToggle={(key) => {
          const k = key as ConversationPriority
          const set = new Set(value.priorities ?? [])
          if (set.has(k)) set.delete(k)
          else set.add(k)
          onChange({ ...value, priorities: set.size ? [...set] : undefined })
        }}
      />
    </div>
  )
}

interface PillRowItem {
  key: number
  label: string
  active: boolean
}

function PillRow({
  label,
  items,
  onToggle,
}: {
  label: string
  items: PillRowItem[]
  onToggle: (key: number) => void
}) {
  return (
    <div className="flex items-center gap-1.5 px-3 py-1 border-t border-border-subtle">
      <span className="text-xs font-mono uppercase tracking-wider text-fg-muted w-16 shrink-0">
        {label}
      </span>
      <div className="flex flex-wrap gap-1">
        {items.map((item) => (
          <button
            key={item.key}
            type="button"
            onClick={() => onToggle(item.key)}
            aria-pressed={item.active}
            className={cn(
              "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium transition-colors",
              item.active
                ? "bg-accent-subtle text-fg border border-accent/30"
                : "border border-border text-fg-secondary hover:bg-subtle",
            )}
          >
            {item.label}
          </button>
        ))}
      </div>
    </div>
  )
}

function countActiveFilters(f: InboxFilters): number {
  let n = 0
  if (f.channels?.length) n++
  if (f.statuses?.length) n++
  if (f.priorities?.length) n++
  if (f.assignedTo) n++
  if (f.unassignedOnly) n++
  if (f.query?.trim()) n++
  return n
}
