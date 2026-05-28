import { type ReactNode } from "react"
import { Navigate, useLocation } from "react-router-dom"
import { useAuthStore } from "./store"

/**
 * RequireAuth — route wrapper that bounces unauthenticated users to /login.
 * Preserves the original destination in `state.from` so post-login can
 * return the user to where they were going.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const status = useAuthStore((s) => s.status)
  const location = useLocation()

  if (status === "idle") {
    // Still hydrating from localStorage on first mount — render nothing.
    return null
  }

  if (status === "unauthenticated") {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <>{children}</>
}

/**
 * RequireGuest — inverse: bounce authenticated users off /login etc.
 */
export function RequireGuest({ children }: { children: ReactNode }) {
  const status = useAuthStore((s) => s.status)
  if (status === "idle") return null
  if (status === "authenticated" || status === "refreshing") {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}
