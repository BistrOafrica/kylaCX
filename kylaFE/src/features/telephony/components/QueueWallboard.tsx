import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  IconUsers,
  IconClockHour4,
  IconActivity,
  IconHeadphones,
} from "@tabler/icons-react"
import {
  PageHeader,
  Surface,
  EmptyState,
} from "@/design-system"
import { cn } from "@/lib/utils"
import {
  listQueuesV2,
  listQueueEntriesV2,
  listQueueMembersV2,
} from "../api/queuesV2"
import type { Queue, QueueEntry, QueueMembership } from "@/pb/queues"
import { useWorkspaceStore } from "@/lib/workspace"

/**
 * QueueWallboard — live view of every active queue in the current workspace.
 *
 * Polls the new Phase 5d QueueService every 2 seconds. Each queue card
 * shows waiting + ringing + connected counts, the longest current wait,
 * and how many agents are currently active in the queue.
 *
 * Streaming-over-gRPC is a follow-up; polling is sufficient for the
 * sub-minute granularity contact-centre operators need today.
 */
export function QueueWallboard() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")

  const queuesQuery = useQuery({
    queryKey: ["queues-v2", workspaceId],
    queryFn: () => listQueuesV2(workspaceId, true),
    refetchInterval: 5_000,
    enabled: !!workspaceId,
  })

  return (
    <div className="p-6 space-y-6">
      <PageHeader
        title="Wallboard"
        description="Live queue state across the workspace. Updates every 2 seconds."
        icon={<IconActivity className="size-5" />}
      />

      {queuesQuery.isPending ? (
        <Surface level={1} className="p-6 text-fg-muted">
          Loading queues…
        </Surface>
      ) : queuesQuery.isError ? (
        <EmptyState
          title="Couldn't load queues"
          description={(queuesQuery.error as Error).message}
        />
      ) : queuesQuery.data?.length === 0 ? (
        <EmptyState
          title="No active queues"
          description="Create a queue and add agents to see live activity here."
        />
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          {queuesQuery.data?.map((q) => <QueueCard key={q.id} queue={q} />)}
        </div>
      )}
    </div>
  )
}

interface QueueCardProps {
  queue: Queue
}

function QueueCard({ queue }: QueueCardProps) {
  // Live entry stream — refetches at the same cadence the operator expects.
  const entries = useQuery({
    queryKey: ["queue-entries-v2", queue.id],
    queryFn: () => listQueueEntriesV2(queue.id, ""),
    refetchInterval: 2_000,
  })
  const members = useQuery({
    queryKey: ["queue-members-v2", queue.id],
    queryFn: () => listQueueMembersV2(queue.id),
    refetchInterval: 10_000,
  })

  const stats = useMemo(() => summariseEntries(entries.data ?? []), [entries.data])
  const activeMembers = useMemo(
    () => (members.data ?? []).filter((m) => m.isActive).length,
    [members.data],
  )
  const totalMembers = members.data?.length ?? 0

  return (
    <Surface level={1} className="p-4 space-y-3">
      <header className="flex items-start justify-between">
        <div>
          <div className="text-sm font-medium">{queue.name}</div>
          <div className="text-xs text-fg-muted">{queue.strategy}</div>
        </div>
        <Badge
          tone={stats.waiting > 0 ? "warn" : "default"}
          icon={<IconUsers className="size-3.5" />}
        >
          {stats.waiting} waiting
        </Badge>
      </header>

      <div className="grid grid-cols-3 gap-2">
        <Tile label="Ringing" value={stats.ringing} />
        <Tile label="Connected" value={stats.connected} />
        <Tile
          label="Longest wait"
          value={stats.waiting > 0 ? formatSeconds(stats.longestWaitSeconds) : "—"}
        />
      </div>

      <div className="text-xs text-fg-muted flex items-center gap-1.5">
        <IconHeadphones className="size-3.5" />
        {activeMembers}/{totalMembers} agents active
      </div>

      {/* Waiting callers, oldest-first. */}
      <div className="space-y-1">
        {(entries.data ?? [])
          .filter((e) => e.status === "waiting")
          .slice(0, 5)
          .map((e) => (
            <WaitingEntryRow key={e.id} entry={e} />
          ))}
      </div>
    </Surface>
  )
}

function summariseEntries(entries: QueueEntry[]) {
  let waiting = 0
  let ringing = 0
  let connected = 0
  let oldestWaiting = Number.POSITIVE_INFINITY
  for (const e of entries) {
    switch (e.status) {
      case "waiting":
        waiting++
        if (e.enteredAt) {
          const enteredMs = Number(e.enteredAt.seconds) * 1000
          if (enteredMs < oldestWaiting) oldestWaiting = enteredMs
        }
        break
      case "ringing":
        ringing++
        break
      case "connected":
        connected++
        break
    }
  }
  const longestWaitSeconds =
    oldestWaiting === Number.POSITIVE_INFINITY
      ? 0
      : Math.max(0, Math.floor((Date.now() - oldestWaiting) / 1000))
  return { waiting, ringing, connected, longestWaitSeconds }
}

function WaitingEntryRow({ entry }: { entry: QueueEntry }) {
  const enteredMs = entry.enteredAt ? Number(entry.enteredAt.seconds) * 1000 : Date.now()
  const waitedSec = Math.max(0, Math.floor((Date.now() - enteredMs) / 1000))
  return (
    <div className="flex items-center justify-between text-xs border-t border-default pt-1">
      <span className="font-mono">{entry.callId.slice(0, 8)}…</span>
      <span className="text-fg-muted inline-flex items-center gap-1">
        <IconClockHour4 className="size-3" />
        {formatSeconds(waitedSec)}
      </span>
    </div>
  )
}

// ── tiny presentational helpers ───────────────────────────────────────────

function Tile({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-sm bg-subtle p-2 text-center">
      <div className="text-xs uppercase text-fg-muted">{label}</div>
      <div className="text-base font-medium tabular-nums">{value}</div>
    </div>
  )
}

function Badge({
  children,
  tone,
  icon,
}: {
  children: React.ReactNode
  tone: "default" | "warn"
  icon?: React.ReactNode
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 px-1.5 py-0.5 text-xs rounded-xs",
        tone === "warn" ? "bg-warn-subtle text-warn" : "bg-subtle text-fg-muted",
      )}
    >
      {icon}
      {children}
    </span>
  )
}

function formatSeconds(s: number): string {
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  const r = s % 60
  return `${m}m${r.toString().padStart(2, "0")}s`
}
