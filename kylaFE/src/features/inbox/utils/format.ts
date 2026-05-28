import { formatDistanceToNowStrict, parseISO } from "date-fns"

/**
 * Compact "2m" / "4h" / "3d" formatter used in dense rows.
 * Returns "now" for <60s ago. Empty input → empty string.
 */
export function relativeTime(iso: string | undefined | null): string {
  if (!iso) return ""
  let date: Date
  try {
    date = parseISO(iso)
  } catch {
    return ""
  }
  if (Number.isNaN(date.getTime())) return ""

  const diffMs = Date.now() - date.getTime()
  if (diffMs < 60_000) return "now"
  return formatDistanceToNowStrict(date)
    .replace(" seconds", "s").replace(" second", "s")
    .replace(" minutes", "m").replace(" minute", "m")
    .replace(" hours",   "h").replace(" hour",   "h")
    .replace(" days",    "d").replace(" day",    "d")
    .replace(" weeks",   "w").replace(" week",   "w")
    .replace(" months",  "mo").replace(" month", "mo")
    .replace(" years",   "y").replace(" year",   "y")
}

/**
 * Parse the message content JSON ({"text": "..."} for plain text) and
 * fall back to the raw string when the payload isn't JSON.
 */
export function readContentText(content: string): string {
  try {
    const parsed = JSON.parse(content) as { text?: string }
    if (typeof parsed.text === "string") return parsed.text
  } catch {
    /* not JSON */
  }
  return content
}

/**
 * Compute time-to-deadline in seconds. Negative when breached.
 */
export function secondsUntil(iso: string | undefined | null): number | null {
  if (!iso) return null
  try {
    const target = parseISO(iso).getTime()
    if (Number.isNaN(target)) return null
    return Math.round((target - Date.now()) / 1000)
  } catch {
    return null
  }
}

/**
 * "28m" / "1h 12m" / "breached" — readable SLA chip body.
 */
export function formatSlaWindow(deadline: string | undefined | null): {
  text: string
  breached: boolean
} | null {
  const secs = secondsUntil(deadline)
  if (secs === null) return null
  if (secs <= 0) return { text: "Breached", breached: true }
  const totalMin = Math.floor(secs / 60)
  if (totalMin < 60) return { text: `${totalMin}m`, breached: false }
  const h = Math.floor(totalMin / 60)
  const m = totalMin % 60
  return { text: m === 0 ? `${h}h` : `${h}h ${m}m`, breached: false }
}
