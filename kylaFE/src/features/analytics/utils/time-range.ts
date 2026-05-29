/**
 * Time-range presets shared across every analytics surface.
 *
 * Every backend analytics request takes `AnalyticsFilters.timeRange`
 * with ISO 8601 start/end strings. Components consume this preset
 * list; the resolver below converts a preset id into the concrete
 * `{ startDate, endDate }` payload.
 */

export type TimeRangePreset =
  | "today"
  | "yesterday"
  | "last_7d"
  | "last_30d"
  | "last_90d"
  | "this_quarter"
  | "ytd"

export const PRESETS: { id: TimeRangePreset; label: string }[] = [
  { id: "today",        label: "Today" },
  { id: "yesterday",    label: "Yesterday" },
  { id: "last_7d",      label: "Last 7 days" },
  { id: "last_30d",     label: "Last 30 days" },
  { id: "last_90d",     label: "Last 90 days" },
  { id: "this_quarter", label: "This quarter" },
  { id: "ytd",          label: "Year to date" },
]

export function resolveRange(preset: TimeRangePreset): {
  startDate: string
  endDate: string
} {
  const now = new Date()
  const end = endOfDay(now).toISOString()

  switch (preset) {
    case "today":
      return { startDate: startOfDay(now).toISOString(), endDate: end }

    case "yesterday": {
      const y = new Date(now)
      y.setDate(y.getDate() - 1)
      return {
        startDate: startOfDay(y).toISOString(),
        endDate: endOfDay(y).toISOString(),
      }
    }

    case "last_7d":
      return { startDate: daysAgo(7).toISOString(), endDate: end }

    case "last_30d":
      return { startDate: daysAgo(30).toISOString(), endDate: end }

    case "last_90d":
      return { startDate: daysAgo(90).toISOString(), endDate: end }

    case "this_quarter": {
      const q = Math.floor(now.getMonth() / 3)
      const start = new Date(now.getFullYear(), q * 3, 1)
      return { startDate: start.toISOString(), endDate: end }
    }

    case "ytd": {
      const start = new Date(now.getFullYear(), 0, 1)
      return { startDate: start.toISOString(), endDate: end }
    }
  }
}

function startOfDay(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}
function endOfDay(d: Date) {
  const x = new Date(d)
  x.setHours(23, 59, 59, 999)
  return x
}
function daysAgo(n: number) {
  const x = new Date()
  x.setDate(x.getDate() - n)
  x.setHours(0, 0, 0, 0)
  return x
}

/**
 * Pick a friendly granularity ("DAILY"/"WEEKLY"/"MONTHLY") for a
 * chosen preset — the backend's volume / trend RPCs take this as a
 * raw string field.
 */
export function granularityFor(preset: TimeRangePreset): "DAILY" | "WEEKLY" | "MONTHLY" {
  switch (preset) {
    case "today":
    case "yesterday":
    case "last_7d":
      return "DAILY"
    case "last_30d":
    case "last_90d":
    case "this_quarter":
      return "WEEKLY"
    case "ytd":
      return "MONTHLY"
  }
}
