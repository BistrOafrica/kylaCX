import { useState } from "react"
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
} from "recharts"
import {
  IconCircleCheck,
  IconCircleX,
  IconEye,
  IconSend,
} from "@tabler/icons-react"
import {
  Surface,
  EmptyState,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { KpiTile } from "@/features/analytics/components/KpiTile"
import { ChartCard } from "@/features/analytics/components/ChartCard"
import { useCampaignAnalytics } from "../hooks/queries"
import { deliveryRate, readRate } from "../utils/status"
import { relativeShort } from "@/features/crm/utils/relative"

const palette = [
  "var(--accent-solid)",
  "var(--status-info-solid)",
  "var(--status-warn-solid)",
  "var(--status-danger-solid)",
]

type StatusFilter = "" | "sent" | "delivered" | "read" | "failed"

/**
 * CampaignAnalytics — per-campaign delivery funnel.
 *
 * Polls every 30s while the campaign is open.
 */
export function CampaignAnalytics({ campaignId }: { campaignId: string }) {
  const [filter, setFilter] = useState<StatusFilter>("")
  const q = useCampaignAnalytics(campaignId, filter)

  const counts = q.data?.count
  const rows = q.data?.rows ?? []

  const sent = Number(counts?.sent ?? 0)
  const delivered = Number(counts?.delivered ?? 0)
  const read = Number(counts?.read ?? 0)
  const failed = Number(counts?.failed ?? 0)

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <KpiTile
          label="Sent"
          value={sent.toLocaleString()}
          pending={q.isPending}
        />
        <KpiTile
          label="Delivered"
          value={delivered.toLocaleString()}
          hint={`${deliveryRate(sent, delivered)}% delivery rate`}
          pending={q.isPending}
        />
        <KpiTile
          label="Read"
          value={read.toLocaleString()}
          hint={`${readRate(sent, read)}% read rate`}
          pending={q.isPending}
        />
        <KpiTile
          label="Failed"
          value={failed.toLocaleString()}
          pending={q.isPending}
          trendInvert
        />
      </div>

      <ChartCard
        title="Delivery funnel"
        subtitle="Sent → delivered → read · failed plotted separately"
        isPending={q.isPending}
        isError={q.isError}
        errorMessage={(q.error as Error | undefined)?.message}
        isEmpty={!counts || sent + delivered + read + failed === 0}
      >
        <ResponsiveContainer width="100%" height={260}>
          <PieChart>
            <Pie
              data={[
                { name: "Sent",      value: sent },
                { name: "Delivered", value: delivered },
                { name: "Read",      value: read },
                { name: "Failed",    value: failed },
              ].filter((d) => d.value > 0)}
              dataKey="value"
              nameKey="name"
              innerRadius={50}
              outerRadius={90}
              paddingAngle={2}
            >
              {palette.map((color, i) => (
                <Cell key={i} fill={color} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                background: "var(--bg-elevated)",
                border: "1px solid var(--border-default)",
                borderRadius: 6,
                fontSize: 12,
              }}
            />
            <Legend wrapperStyle={{ fontSize: 12 }} />
          </PieChart>
        </ResponsiveContainer>
      </ChartCard>

      <Surface level={1} radius="md" className="flex flex-col">
        <header className="flex items-center gap-2 px-4 py-3 border-b border-border">
          <span className="text-md font-medium text-fg">Recipients</span>
          <div className="ms-auto flex items-center gap-1">
            {(
              [
                { id: "",          label: "All" },
                { id: "sent",      label: "Sent" },
                { id: "delivered", label: "Delivered" },
                { id: "read",      label: "Read" },
                { id: "failed",    label: "Failed" },
              ] as const
            ).map((f) => (
              <button
                key={f.id}
                type="button"
                onClick={() => setFilter(f.id as StatusFilter)}
                aria-pressed={filter === f.id}
                className={cn(
                  "inline-flex items-center h-6 px-2 rounded-xs text-sm",
                  filter === f.id
                    ? "bg-accent-subtle text-fg"
                    : "text-fg-muted hover:text-fg hover:bg-subtle",
                )}
              >
                {f.label}
              </button>
            ))}
          </div>
        </header>
        {rows.length === 0 ? (
          <EmptyState title="No recipients yet" size="sm" />
        ) : (
          <ul role="list" className="divide-y divide-border-subtle">
            {rows.slice(0, 50).map((r) => (
              <li
                key={r.id}
                className="flex items-center gap-3 px-4 py-2"
              >
                <RecipientIcon row={r} />
                <div className="min-w-0 flex-1">
                  <div className="font-mono text-md text-fg truncate">
                    {r.contact || "—"}
                  </div>
                  <div className="text-sm text-fg-muted truncate">
                    {r.status || (r.failed ? "Failed" : r.read ? "Read" : r.delivered ? "Delivered" : "Sent")}
                    {r.reason ? ` · ${r.reason}` : ""}
                  </div>
                </div>
                <span className="font-mono text-xs text-fg-muted w-20 text-end">
                  {relativeShort(r.readAt || r.deliveredAt || r.sentAt)}
                </span>
              </li>
            ))}
          </ul>
        )}
      </Surface>
    </div>
  )
}

function RecipientIcon({
  row,
}: {
  row: {
    sent: boolean
    delivered: boolean
    read: boolean
    failed: boolean
  }
}) {
  if (row.failed) {
    return (
      <span className="size-6 rounded-sm bg-danger-subtle text-danger flex items-center justify-center" aria-hidden>
        <IconCircleX className="size-3.5" />
      </span>
    )
  }
  if (row.read) {
    return (
      <span className="size-6 rounded-sm bg-info-subtle text-info flex items-center justify-center" aria-hidden>
        <IconEye className="size-3.5" />
      </span>
    )
  }
  if (row.delivered) {
    return (
      <span className="size-6 rounded-sm bg-success-subtle text-success flex items-center justify-center" aria-hidden>
        <IconCircleCheck className="size-3.5" />
      </span>
    )
  }
  return (
    <span className="size-6 rounded-sm bg-subtle text-fg-muted flex items-center justify-center" aria-hidden>
      <IconSend className="size-3.5" />
    </span>
  )
}
