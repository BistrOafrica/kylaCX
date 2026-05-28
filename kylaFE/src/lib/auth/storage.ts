/**
 * Hardened localStorage wrapper for session tokens.
 *
 * Backend investigation (kylaRM/FRONTEND_OVERHAUL.md §3) determined:
 *   - Backend issues access + refresh JWTs (10-day TTL each)
 *   - Backend does NOT set cookies
 *   - Envoy CORS does NOT permit credentialed cross-origin requests
 *
 * Therefore F0 stores tokens in localStorage under a namespaced key.
 * Mitigations against XSS:
 *   - Strong CSP delivered via the Vite-served HTML
 *   - Tokens accessed only via this module; components read via the
 *     `useAuthStore()` selector, never localStorage directly
 *   - Versioned storage key so a breaking format change forces re-login
 *   - Quick wipe (`clearSession()`) on any decode failure
 *
 * Future hardening (post-F0):
 *   - Add HttpOnly refresh cookie path (requires backend + Envoy work)
 *   - Shorten access-token TTL (backend currently 10 days)
 */

const STORAGE_KEY = "kyla.auth.v1"

export interface PersistedSession {
  accessToken: string
  refreshToken: string
  userId: string
  expiresAt?: number       // ms epoch — if known from JWT claims
}

export function readSession(): PersistedSession | null {
  if (typeof localStorage === "undefined") return null
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as PersistedSession
    if (!parsed.accessToken || !parsed.refreshToken || !parsed.userId) {
      return null
    }
    return parsed
  } catch {
    // Corrupted — wipe and force re-login.
    try { localStorage.removeItem(STORAGE_KEY) } catch { /* ignore */ }
    return null
  }
}

export function writeSession(session: PersistedSession): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  } catch {
    // Quota or private mode — fail silent. App still works for the
    // current tab via the in-memory Zustand store; refresh will lose
    // the session.
  }
}

export function clearSession(): void {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    /* ignore */
  }
}

/**
 * Pull the `exp` claim from a JWT without verifying the signature
 * (verification is the backend's job — we only need the timestamp to
 * schedule a refresh).
 */
export function readJwtExpiry(token: string): number | undefined {
  try {
    const [, payloadB64] = token.split(".")
    if (!payloadB64) return undefined
    const json = atob(payloadB64.replace(/-/g, "+").replace(/_/g, "/"))
    const claims = JSON.parse(json) as { exp?: number }
    return typeof claims.exp === "number" ? claims.exp * 1000 : undefined
  } catch {
    return undefined
  }
}
