/**
 * Campaign status helpers.
 *
 * The proto stores status as a free-form string (backend writes
 * "draft" / "scheduled" / "running" / "completed" / "paused" / etc.).
 * The UI normalizes the value into one of five buckets so badges stay
 * consistent across surfaces.
 */

export type CampaignStatusBucket =
  | "draft"
  | "scheduled"
  | "running"
  | "completed"
  | "paused"
  | "failed"

export function bucketCampaignStatus(raw: string): CampaignStatusBucket {
  const s = raw.toLowerCase()
  if (s.includes("draft")) return "draft"
  if (s.includes("schedul")) return "scheduled"
  if (s.includes("run") || s === "active") return "running"
  if (s.includes("complete") || s.includes("done")) return "completed"
  if (s.includes("paus")) return "paused"
  if (s.includes("fail") || s.includes("error")) return "failed"
  return "draft"
}

export const STATUS_LABEL: Record<CampaignStatusBucket, string> = {
  draft:     "Draft",
  scheduled: "Scheduled",
  running:   "Running",
  completed: "Completed",
  paused:    "Paused",
  failed:    "Failed",
}

export const STATUS_TONE: Record<CampaignStatusBucket, "muted" | "info" | "success" | "warn" | "danger"> = {
  draft:     "muted",
  scheduled: "info",
  running:   "info",
  completed: "success",
  paused:    "warn",
  failed:    "danger",
}

/**
 * Parse a contact list — protos store them as a single string
 * (sometimes JSON, sometimes CSV). Accept both and return a flat
 * list of phone numbers / contact-ids.
 */
export function parseContacts(raw: string): string[] {
  if (!raw) return []
  const trimmed = raw.trim()
  // JSON array.
  if (trimmed.startsWith("[")) {
    try {
      const arr = JSON.parse(trimmed) as unknown[]
      return arr.map((x) => String(x).trim()).filter(Boolean)
    } catch {
      /* fall through */
    }
  }
  return trimmed
    .split(/[\s,;]+/)
    .map((s) => s.trim())
    .filter(Boolean)
}

/**
 * Compute delivery rate (delivered / sent) as a percentage 0–100.
 * Returns 0 when sent is zero (avoids NaN).
 */
export function deliveryRate(sent: bigint | number, delivered: bigint | number): number {
  const s = Number(sent)
  const d = Number(delivered)
  if (!s) return 0
  return Math.round((d / s) * 1000) / 10
}

/**
 * Read rate (read / sent) as a percentage 0–100.
 */
export function readRate(sent: bigint | number, read: bigint | number): number {
  const s = Number(sent)
  const r = Number(read)
  if (!s) return 0
  return Math.round((r / s) * 1000) / 10
}
