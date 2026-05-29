import { useState } from "react"
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts"
import { PageHeader } from "@/design-system"
import { KpiTile } from "./KpiTile"
import { ChartCard } from "./ChartCard"
import { TimeRangePicker } from "./TimeRangePicker"
import type { TimeRangePreset } from "../utils/time-range"
import {
  useTicketKPIs,
  useTicketVolume,
  useChannelDistribution,
  useStatusDistribution,
  usePriorityDistribution,
  useAgentPerformance,
  useSLACompliance,
} from "../hooks/queries"

/**
 * OverviewDashboard — the cross-domain analytics homepage.
 *
 *   ┌──────────────────────────────────────────────────────┐
 *   │ Page header + time range picker                       │
 *   ├──────────────────────────────────────────────────────┤
 *   │ KPI row (5 tiles)                                     │
 *   ├────────────────────────────────┬─────────────────────┤
 *   │ Ticket volume (area)           │ Channel mix (pie)   │
 *   ├────────────────────────────────┼─────────────────────┤
 *   │ Status (bar)                   │ Priority (bar)      │
 *   ├────────────────────────────────┴─────────────────────┤
 *   │ Top agents (bar)                                     │
 *   └──────────────────────────────────────────────────────┘
 */
export function OverviewDashboard() {
  const [preset, setPreset] = useState<TimeRangePreset>("last_30d")

  const kpi = useTicketKPIs(preset)
  const volume = useTicketVolume(preset)
  const channels = useChannelDistribution(preset)
  const statuses = useStatusDistribution(preset)
  const priorities = usePriorityDistribution(preset)
  const agents = useAgentPerformance(preset)
  const sla = useSLACompliance(preset)

  const ax = AXIS_STYLE
  const grid = GRID_STYLE
  const palette = PALETTE

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <PageHeader
        title="Analytics"
        description="Real-time view across the workbench"
        actions={<TimeRangePicker value={preset} onChange={setPreset} />}
      />

      <div className="p-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        <KpiTile
          label="Total tickets"
          value={kpi.data?.totalTickets?.toLocaleString() ?? "—"}
          pending={kpi.isPending}
        />
        <KpiTile
          label="First-contact resolution"
          value={
            kpi.data
              ? `${(kpi.data.firstContactResolutionPercent ?? 0).toFixed(1)}%`
              : "—"
          }
          pending={kpi.isPending}
        />
        <KpiTile
          label="SLA compliance"
          value={
            kpi.data
              ? `${(kpi.data.slaCompliancePercent ?? 0).toFixed(1)}%`
              : "—"
          }
          pending={kpi.isPending}
        />
        <KpiTile
          label="Avg resolution"
          value={
            kpi.data
              ? formatMinutes(kpi.data.averageResolutionTimeMinutes)
              : "—"
          }
          pending={kpi.isPending}
        />
        <KpiTile
          label="Avg sentiment"
          value={
            kpi.data
              ? (kpi.data.averageSentimentScore ?? 0).toFixed(2)
              : "—"
          }
          hint="−1 → 1 (positive)"
          pending={kpi.isPending}
        />
      </div>

      <div className="px-4 pb-4 grid grid-cols-1 lg:grid-cols-3 gap-3">
        <ChartCard
          title="Ticket volume"
          subtitle="New tickets over time"
          isPending={volume.isPending}
          isError={volume.isError}
          errorMessage={(volume.error as Error | undefined)?.message}
          onRetry={() => void volume.refetch()}
          isEmpty={(volume.data ?? []).length === 0}
          className="lg:col-span-2"
        >
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={volume.data ?? []}>
              <CartesianGrid {...grid} />
              <XAxis dataKey="date" {...ax} />
              <YAxis {...ax} />
              <Tooltip {...TOOLTIP_STYLE} />
              <Area
                type="monotone"
                dataKey="newTickets"
                stroke="var(--accent-solid)"
                fill="var(--accent-solid)"
                fillOpacity={0.18}
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard
          title="Channel mix"
          subtitle={`SLA: ${sla.data ? (sla.data.overallCompliance ?? 0).toFixed(1) + "% compliant" : "—"}`}
          isPending={channels.isPending}
          isError={channels.isError}
          errorMessage={(channels.error as Error | undefined)?.message}
          onRetry={() => void channels.refetch()}
          isEmpty={(channels.data ?? []).length === 0}
        >
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie
                data={(channels.data ?? []).map((c) => ({
                  name: c.channel,
                  value: c.ticketCount,
                }))}
                dataKey="value"
                nameKey="name"
                innerRadius={50}
                outerRadius={85}
                paddingAngle={2}
              >
                {(channels.data ?? []).map((_, i) => (
                  <Cell key={i} fill={palette[i % palette.length]!} />
                ))}
              </Pie>
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Tooltip {...TOOLTIP_STYLE} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      <div className="px-4 pb-4 grid grid-cols-1 lg:grid-cols-2 gap-3">
        <ChartCard
          title="Status distribution"
          isPending={statuses.isPending}
          isError={statuses.isError}
          isEmpty={(statuses.data ?? []).length === 0}
        >
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={statuses.data ?? []}>
              <CartesianGrid {...grid} />
              <XAxis dataKey="status" {...ax} />
              <YAxis {...ax} />
              <Tooltip {...TOOLTIP_STYLE} />
              <Bar dataKey="count" fill="var(--accent-solid)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard
          title="Priority distribution"
          isPending={priorities.isPending}
          isError={priorities.isError}
          isEmpty={(priorities.data ?? []).length === 0}
        >
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={priorities.data ?? []}>
              <CartesianGrid {...grid} />
              <XAxis dataKey="priority" {...ax} />
              <YAxis {...ax} />
              <Tooltip {...TOOLTIP_STYLE} />
              <Bar dataKey="count" fill="var(--status-warn-solid)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      <div className="px-4 pb-6">
        <ChartCard
          title="Top agents"
          subtitle="By tickets resolved"
          isPending={agents.isPending}
          isError={agents.isError}
          isEmpty={(agents.data ?? []).length === 0}
        >
          <ResponsiveContainer width="100%" height={280}>
            <BarChart
              data={(agents.data ?? []).slice(0, 10)}
              layout="vertical"
              margin={{ left: 60 }}
            >
              <CartesianGrid {...grid} />
              <XAxis type="number" {...ax} />
              <YAxis
                type="category"
                dataKey="agentName"
                {...ax}
                width={120}
              />
              <Tooltip {...TOOLTIP_STYLE} />
              <Bar
                dataKey="ticketsHandled"
                fill="var(--accent-solid)"
                radius={[0, 4, 4, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>
    </div>
  )
}

function formatMinutes(min: number): string {
  if (!min) return "—"
  if (min < 60) return `${Math.round(min)}m`
  const h = Math.floor(min / 60)
  const m = Math.round(min % 60)
  return m === 0 ? `${h}h` : `${h}h ${m}m`
}

// Shared Recharts styling so every chart on every page reads the
// active design-system theme. Changing the accent token cascades here.
const AXIS_STYLE = {
  stroke: "var(--text-muted)",
  fontSize: 11,
  tick: { fill: "var(--text-muted)" },
  tickLine: false,
  axisLine: false,
}
const GRID_STYLE = {
  stroke: "var(--border-default)",
  strokeDasharray: "3 3",
  vertical: false,
}
const TOOLTIP_STYLE = {
  contentStyle: {
    background: "var(--bg-elevated)",
    border: "1px solid var(--border-default)",
    borderRadius: 6,
    fontSize: 12,
  },
  cursor: { fill: "var(--bg-subtle)" },
}
const PALETTE = [
  "var(--channel-whatsapp)",
  "var(--channel-email)",
  "var(--channel-sms)",
  "var(--channel-voice)",
  "var(--channel-webchat)",
  "var(--channel-instagram)",
  "var(--channel-messenger)",
]
