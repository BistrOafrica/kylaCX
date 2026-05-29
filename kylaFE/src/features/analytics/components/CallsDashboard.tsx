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
} from "recharts"
import { PageHeader } from "@/design-system"
import { KpiTile } from "./KpiTile"
import { ChartCard } from "./ChartCard"
import { TimeRangePicker } from "./TimeRangePicker"
import {
  useCallOverview,
  useCallTraffic,
  useCallHandling,
  useCustomerExperience,
} from "../hooks/queries"
import type { TimeRangePreset } from "../utils/time-range"

const ax = {
  stroke: "var(--text-muted)",
  fontSize: 11,
  tick: { fill: "var(--text-muted)" },
  tickLine: false,
  axisLine: false,
}
const grid = {
  stroke: "var(--border-default)",
  strokeDasharray: "3 3",
  vertical: false,
}
const tooltip = {
  contentStyle: {
    background: "var(--bg-elevated)",
    border: "1px solid var(--border-default)",
    borderRadius: 6,
    fontSize: 12,
  },
  cursor: { fill: "var(--bg-subtle)" },
}

/**
 * CallsDashboard — call-centre operations view.
 *
 * The CallAnalytics proto stores most KPI values as preformatted
 * strings (the backend has done the maths and units), so KPI tiles
 * render the strings verbatim instead of re-formatting on the
 * frontend.
 */
export function CallsDashboard() {
  const [preset, setPreset] = useState<TimeRangePreset>("last_30d")
  const overview = useCallOverview(preset)
  const traffic = useCallTraffic(preset)
  const handling = useCallHandling(preset)
  const cx = useCustomerExperience(preset)

  const kpi = overview.data?.kpiMetrics
  const trafficSeries = traffic.data?.hourlyVolume ?? []
  const handlingMetrics = handling.data?.metrics
  const cxData = cx.data

  // Combine inbound + outbound + internal into a single "total" for the
  // hourly volume chart; the proto exposes the three columns separately.
  const trafficForChart = trafficSeries.map((p) => ({
    hour: p.hour,
    inbound: p.inbound,
    outbound: p.outbound,
    internal: p.internal,
    total: p.inbound + p.outbound + p.internal,
  }))

  // Handling metrics come back as strings ("4m 12s") — render the chart
  // by parsing them into seconds-equivalents only if they're numeric.
  const handlingChart = handlingMetrics
    ? [
        { stage: "Talk",  seconds: secondsOrZero(handlingMetrics.averageTalkTime) },
        { stage: "Hold",  seconds: secondsOrZero(handlingMetrics.averageHoldTime) },
        { stage: "Wrap",  seconds: secondsOrZero(handlingMetrics.averageWrapUpTime) },
      ]
    : []

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <PageHeader
        title="Calls"
        description="Telephony performance and customer experience"
        actions={<TimeRangePicker value={preset} onChange={setPreset} />}
      />

      <div className="p-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        <KpiTile
          label="Total calls"
          value={kpi?.totalCalls ?? "—"}
          pending={overview.isPending}
        />
        <KpiTile
          label="Avg handle time"
          value={kpi?.averageHandleTime ?? "—"}
          pending={overview.isPending}
          trendInvert
        />
        <KpiTile
          label="Avg speed of answer"
          value={kpi?.averageSpeedOfAnswer ?? "—"}
          pending={overview.isPending}
          trendInvert
        />
        <KpiTile
          label="Service level"
          value={kpi?.serviceLevel ?? "—"}
          pending={overview.isPending}
        />
        <KpiTile
          label="CSAT"
          value={cxData?.metrics ? cxData.metrics.csatScore.toFixed(2) : "—"}
          pending={cx.isPending}
        />
      </div>

      <div className="px-4 pb-4 grid grid-cols-1 lg:grid-cols-3 gap-3">
        <ChartCard
          title="Hourly call volume"
          subtitle="Inbound + outbound + internal"
          isPending={traffic.isPending}
          isError={traffic.isError}
          isEmpty={trafficForChart.length === 0}
          className="lg:col-span-2"
        >
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={trafficForChart}>
              <CartesianGrid {...grid} />
              <XAxis dataKey="hour" {...ax} />
              <YAxis {...ax} />
              <Tooltip {...tooltip} />
              <Area
                type="monotone"
                dataKey="total"
                stroke="var(--accent-solid)"
                fill="var(--accent-solid)"
                fillOpacity={0.18}
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard
          title="Handle-time breakdown"
          subtitle="Talk + hold + wrap (when reported numerically)"
          isPending={handling.isPending}
          isError={handling.isError}
          isEmpty={handlingChart.every((p) => p.seconds === 0)}
        >
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={handlingChart}>
              <CartesianGrid {...grid} />
              <XAxis dataKey="stage" {...ax} />
              <YAxis {...ax} />
              <Tooltip {...tooltip} />
              <Bar dataKey="seconds" fill="var(--status-info-solid)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>
    </div>
  )
}

/** Best-effort string → seconds parser for the backend's formatted strings. */
function secondsOrZero(raw: string | undefined): number {
  if (!raw) return 0
  // "4m 12s" / "12s" / "00:04:12" / plain numeric.
  const n = Number(raw)
  if (!Number.isNaN(n)) return n
  const m = raw.match(/(\d+)m\s*(\d+)?s?/i)
  if (m) return Number(m[1]) * 60 + (Number(m[2]) || 0)
  return 0
}
