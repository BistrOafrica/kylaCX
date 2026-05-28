import { refreshTokens } from "./api"
import { useAuthStore } from "./store"
import { RpcError } from "@/lib/rpc"

/**
 * Background token refresh.
 *
 * Schedules a single refresh at ~80% of the access token's TTL so
 * the user never sees an expired token. Runs while authenticated;
 * tears down on sign-out.
 */

let pendingTimer: ReturnType<typeof setTimeout> | null = null

const REFRESH_WINDOW_RATIO = 0.8
const MIN_REFRESH_DELAY_MS = 30_000
const MAX_REFRESH_DELAY_MS = 6 * 60 * 60 * 1000   // 6h cap to retry against drift

export function scheduleNextRefresh() {
  if (pendingTimer) {
    clearTimeout(pendingTimer)
    pendingTimer = null
  }

  const { status, expiresAt } = useAuthStore.getState()
  if (status !== "authenticated" || !expiresAt) return

  const now = Date.now()
  const ttl = expiresAt - now
  if (ttl <= 0) {
    void doRefresh()
    return
  }

  const delay = Math.max(
    MIN_REFRESH_DELAY_MS,
    Math.min(MAX_REFRESH_DELAY_MS, ttl * REFRESH_WINDOW_RATIO),
  )

  pendingTimer = setTimeout(() => { void doRefresh() }, delay)
}

async function doRefresh() {
  try {
    await refreshTokens()
    scheduleNextRefresh()
  } catch (err) {
    if (err instanceof RpcError && err.isUnauthenticated) {
      useAuthStore.getState().signOut()
      return
    }
    // Transient — retry in 30s.
    pendingTimer = setTimeout(() => { void doRefresh() }, MIN_REFRESH_DELAY_MS)
  }
}

/**
 * Wire the auto-refresh lifecycle. Subscribe to auth state changes and
 * (re)schedule the timer whenever tokens rotate or sign-out occurs.
 */
export function startAutoRefresh(): () => void {
  scheduleNextRefresh()
  const unsub = useAuthStore.subscribe(scheduleNextRefresh)
  return () => {
    unsub()
    if (pendingTimer) {
      clearTimeout(pendingTimer)
      pendingTimer = null
    }
  }
}

/**
 * 401-on-call recovery helper. Wraps a call so it transparently
 * refreshes and retries once on UNAUTHENTICATED. Callers that don't
 * need this (e.g. the refresh call itself) should call the underlying
 * RPC directly.
 */
export async function callWithRefresh<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    if (err instanceof RpcError && err.isUnauthenticated) {
      try {
        await refreshTokens()
        return await fn()
      } catch (refreshErr) {
        if (refreshErr instanceof RpcError && refreshErr.isUnauthenticated) {
          useAuthStore.getState().signOut()
        }
        throw refreshErr
      }
    }
    throw err
  }
}
