import { useEffect, useState } from "react"
import {
  IconActivity,
  IconPhoneIncoming,
  IconPhoneOutgoing,
  IconClock,
} from "@tabler/icons-react"
import {
  PageHeader,
  Surface,
  EmptyState,
} from "@/design-system"
import { cn } from "@/lib/utils"
import {
  subscribeOnlineAgents,
  subscribeQueueData,
  subscribeCallMetrics,
  subscribeRecentCalls,
} from "../api/calls"
import { KpiTile } from "@/features/analytics/components/KpiTile"
import {
  DIRECTION_LABEL,
  STATUS_LABEL,
  statusTone,
} from "../utils/formatters"
import {
  type CallDirection,
  type CallStatus,
} from "@/pb/call_session"

interface OnlineAgent {
  agentName: string
  extension: string
  queueName: string
  status: number
}

interface QueueRow {
  queueId: string
  queueName: string
  agents: string
  ongoingCalls: string
  type: number
}

interface RecentCall {
  contact: string
  agent: string
  queue: string
  queueId: string
  callId: string
  time: string
  direction: number
  status: number
}

const TONE_CLS = {
  success: "bg-success-subtle text-success",
  warn:    "bg-warn-subtle text-warn",
  danger:  "bg-danger-subtle text-danger",
  info:    "bg-info-subtle text-info",
  muted:   "bg-subtle text-fg-muted",
}

/**
 * Wallboard — supervisor live view.
 *
 * Subscribes to four server-streaming RPCs from CallMonitoringService:
 *   - ReadOnlineAgents     who's logged in + status
 *   - ReadQueues           live queue depth and member count
 *   - ReadCallMetrics      aggregate inbound / outbound / missed counters
 *   - ReadRecentCalls      most-recent activity feed
 *
 * Each subscription mounts on first render and tears down on unmount.
 * Stream payloads are typed loosely because the proto uses string-typed
 * counters; numeric coercion lives in formatters.
 */
export function Wallboard() {
  const [agents, setAgents] = useState<OnlineAgent[]>([])
  const [queues, setQueues] = useState<QueueRow[]>([])
  const [metrics, setMetrics] = useState({
    inboundCalls: "0",
    outboundCalls: "0",
    missedCalls: "0",
    answeredCalls: "0",
    averageCallDuration: "0s",
    averageWaitTime: "0s",
  })
  const [recent, setRecent] = useState<RecentCall[]>([])

  useEffect(() => {
    const a = subscribeOnlineAgents(
      (msg) => setAgents(((msg as { agents?: OnlineAgent[] }).agents ?? [])),
      () => { /* keep last good state */ },
    )
    const q = subscribeQueueData(
      (msg) => setQueues(((msg as { queues?: QueueRow[] }).queues ?? [])),
      () => { /* keep last good state */ },
    )
    const m = subscribeCallMetrics(
      (msg) => {
        const x = msg as Partial<typeof metrics>
        setMetrics((cur) => ({ ...cur, ...x }))
      },
      () => { /* keep last good state */ },
    )
    const r = subscribeRecentCalls(
      (msg) => setRecent(((msg as { calls?: RecentCall[] }).calls ?? [])),
      () => { /* keep last good state */ },
    )
    return () => {
      a.cancel()
      q.cancel()
      m.cancel()
      r.cancel()
    }
  }, [])

  const totalCalls =
    Number(metrics.inboundCalls) + Number(metrics.outboundCalls)
  const onlineCount = agents.length
  const queueCallsWaiting = queues.reduce(
    (acc, q) => acc + Number(q.ongoingCalls || 0),
    0,
  )

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <PageHeader
        title="Wallboard"
        description="Live telephony view — auto-updates from monitoring streams"
        actions={
          <div className="flex items-center gap-1.5 text-xs text-fg-muted">
            <span className="size-1.5 rounded-full bg-success animate-pulse" aria-hidden />
            Live
          </div>
        }
      />

      <div className="p-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        <KpiTile label="Online agents" value={onlineCount} />
        <KpiTile label="Calls waiting" value={queueCallsWaiting} />
        <KpiTile label="Inbound today" value={metrics.inboundCalls} />
        <KpiTile label="Outbound today" value={metrics.outboundCalls} />
        <KpiTile
          label="Avg wait"
          value={metrics.averageWaitTime || "—"}
          hint={`Avg call ${metrics.averageCallDuration || "—"}`}
        />
      </div>

      <div className="px-4 pb-4 grid grid-cols-1 lg:grid-cols-2 gap-3">
        <Surface level={1} radius="md" className="flex flex-col">
          <header className="flex items-center gap-2 px-4 py-3 border-b border-border">
            <IconActivity className="size-3.5 text-fg-muted" />
            <span className="text-md font-medium text-fg flex-1">Queues</span>
            <span className="font-mono text-xs text-fg-muted">{queues.length}</span>
          </header>
          {queues.length === 0 ? (
            <EmptyState title="No queues live yet" size="sm" />
          ) : (
            <ul role="list" className="divide-y divide-border-subtle">
              {queues.map((q) => (
                <li key={q.queueId} className="flex items-center gap-3 px-4 py-2.5">
                  <div className="min-w-0 flex-1">
                    <div className="text-md font-medium text-fg truncate">
                      {q.queueName || "(unnamed queue)"}
                    </div>
                    <div className="text-sm text-fg-muted">
                      {q.agents} agent{q.agents === "1" ? "" : "s"} online
                    </div>
                  </div>
                  <span className="font-mono text-md tabular-nums">
                    {q.ongoingCalls}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Surface>

        <Surface level={1} radius="md" className="flex flex-col">
          <header className="flex items-center gap-2 px-4 py-3 border-b border-border">
            <IconClock className="size-3.5 text-fg-muted" />
            <span className="text-md font-medium text-fg flex-1">Online agents</span>
            <span className="font-mono text-xs text-fg-muted">{agents.length}</span>
          </header>
          {agents.length === 0 ? (
            <EmptyState title="No agents online" size="sm" />
          ) : (
            <ul role="list" className="divide-y divide-border-subtle">
              {agents.map((a, i) => (
                <li key={`${a.agentName}-${i}`} className="flex items-center gap-3 px-4 py-2.5">
                  <div
                    className={cn(
                      "size-7 rounded-full bg-accent-subtle text-fg flex items-center justify-center font-medium text-sm",
                    )}
                    aria-hidden
                  >
                    {(a.agentName?.[0] ?? "?").toUpperCase()}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="text-md text-fg truncate">
                      {a.agentName || "(unknown)"}
                    </div>
                    <div className="text-sm text-fg-muted truncate">
                      ext {a.extension || "—"} · {a.queueName || "—"}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </Surface>
      </div>

      <div className="px-4 pb-6">
        <Surface level={1} radius="md" className="flex flex-col">
          <header className="flex items-center gap-2 px-4 py-3 border-b border-border">
            <IconActivity className="size-3.5 text-fg-muted" />
            <span className="text-md font-medium text-fg flex-1">
              Recent activity
            </span>
            <span className="font-mono text-xs text-fg-muted">
              {totalCalls} today
            </span>
          </header>
          {recent.length === 0 ? (
            <EmptyState title="No recent calls" size="sm" />
          ) : (
            <ul role="list" className="divide-y divide-border-subtle">
              {recent.slice(0, 25).map((r) => {
                const Icon =
                  r.direction === 0 ? IconPhoneIncoming : IconPhoneOutgoing
                const tone = statusTone(r.status as CallStatus)
                return (
                  <li
                    key={`${r.callId}-${r.time}`}
                    className="flex items-center gap-3 px-4 py-2"
                  >
                    <div
                      className={cn(
                        "size-7 rounded-sm flex items-center justify-center shrink-0",
                        TONE_CLS[tone],
                      )}
                      aria-hidden
                    >
                      <Icon className="size-3.5" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="text-md text-fg truncate">
                        {r.contact || "—"}
                      </div>
                      <div className="text-sm text-fg-muted truncate">
                        {r.agent || "—"} · {r.queue || "—"}
                      </div>
                    </div>
                    <span
                      className={cn(
                        "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
                        TONE_CLS[tone],
                      )}
                    >
                      {STATUS_LABEL[r.status as CallStatus] ??
                        DIRECTION_LABEL[r.direction as CallDirection] ??
                        "—"}
                    </span>
                    <span className="font-mono text-xs text-fg-muted w-20 text-end truncate">
                      {r.time || ""}
                    </span>
                  </li>
                )
              })}
            </ul>
          )}
        </Surface>
      </div>
    </div>
  )
}
