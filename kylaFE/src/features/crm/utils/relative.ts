/**
 * Compact "now / Nm / Nh / Nd / Nmo / Ny" relative-time formatter.
 *
 * Mirrors the inbox's `relativeTime` but uses month/year cutoffs more
 * suitable for CRM timelines (events span years; conversations rarely).
 */
export function relativeShort(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  const diffMs = Date.now() - d.getTime()
  const sec = Math.floor(diffMs / 1000)
  if (sec < 60) return "now"
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h}h`
  const days = Math.floor(h / 24)
  if (days < 30) return `${days}d`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo`
  return `${Math.floor(months / 12)}y`
}
