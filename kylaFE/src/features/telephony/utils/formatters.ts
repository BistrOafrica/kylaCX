import { CallDirection, CallStatus } from "@/pb/call_session"

export function formatDuration(seconds: bigint | number | string): string {
  const s = typeof seconds === "bigint" ? Number(seconds) : Number(seconds)
  if (!Number.isFinite(s) || s <= 0) return "0s"
  if (s < 60) return `${Math.round(s)}s`
  const m = Math.floor(s / 60)
  const rem = Math.floor(s % 60)
  if (m < 60) return rem === 0 ? `${m}m` : `${m}m ${rem}s`
  const h = Math.floor(m / 60)
  return `${h}h ${m % 60}m`
}

export const DIRECTION_LABEL: Record<CallDirection, string> = {
  [CallDirection.INBOUND]:  "Inbound",
  [CallDirection.OUTBOUND]: "Outbound",
  [CallDirection.LOCAL]:    "Local",
}

export const STATUS_LABEL: Record<CallStatus, string> = {
  [CallStatus.ONGOING]:   "Ongoing",
  [CallStatus.REJECTED]:  "Rejected",
  [CallStatus.DROPPED]:   "Dropped",
  [CallStatus.COMPLETED]: "Completed",
  [CallStatus.ON_HOLD]:   "On hold",
}

export function statusTone(status: CallStatus): "success" | "warn" | "danger" | "info" | "muted" {
  switch (status) {
    case CallStatus.COMPLETED: return "success"
    case CallStatus.ONGOING:   return "info"
    case CallStatus.ON_HOLD:   return "warn"
    case CallStatus.REJECTED:
    case CallStatus.DROPPED:   return "danger"
    default:                   return "muted"
  }
}
