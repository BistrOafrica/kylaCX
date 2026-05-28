import { create } from "zustand"
import { readSession, writeSession, clearSession, readJwtExpiry } from "./storage"
import { registerMetaProvider } from "@/lib/rpc"

export type AuthStatus =
  | "idle"             // not initialized yet
  | "authenticated"
  | "unauthenticated"
  | "mfa_required"     // login succeeded but MFA challenge pending
  | "refreshing"       // token refresh in flight

export interface AuthIdentity {
  userId: string
  email?: string
  displayName?: string
  avatarUrl?: string
}

export interface AuthState {
  status: AuthStatus
  identity: AuthIdentity | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null     // ms epoch
  pendingMfaUserId: string | null

  /** Replace tokens after a successful login or refresh. */
  setSession: (input: {
    accessToken: string
    refreshToken: string
    identity: AuthIdentity
  }) => void

  /** Begin MFA challenge — credentials accepted but MFA required. */
  beginMfaChallenge: (userId: string) => void

  /** Mark current refresh in-flight. */
  beginRefresh: () => void

  /** Wipe local session and force the user back to /login. */
  signOut: () => void

  /** Hydrate from localStorage on app boot. */
  hydrate: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  status: "idle",
  identity: null,
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
  pendingMfaUserId: null,

  setSession: ({ accessToken, refreshToken, identity }) => {
    const expiresAt = readJwtExpiry(accessToken) ?? null
    writeSession({
      accessToken,
      refreshToken,
      userId: identity.userId,
      expiresAt: expiresAt ?? undefined,
    })
    set({
      status: "authenticated",
      identity,
      accessToken,
      refreshToken,
      expiresAt,
      pendingMfaUserId: null,
    })
  },

  beginMfaChallenge: (userId) =>
    set({ status: "mfa_required", pendingMfaUserId: userId }),

  beginRefresh: () => {
    if (get().status === "authenticated") set({ status: "refreshing" })
  },

  signOut: () => {
    clearSession()
    set({
      status: "unauthenticated",
      identity: null,
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      pendingMfaUserId: null,
    })
  },

  hydrate: () => {
    const persisted = readSession()
    if (persisted) {
      set({
        status: "authenticated",
        accessToken: persisted.accessToken,
        refreshToken: persisted.refreshToken,
        expiresAt: persisted.expiresAt ?? null,
        identity: { userId: persisted.userId },
      })
    } else {
      set({ status: "unauthenticated" })
    }
  },
}))

/**
 * Register the auth metadata provider so every gRPC call carries the
 * current access token as `authorization`. Called once at app boot.
 */
export function initAuthRpcMetadata(): () => void {
  return registerMetaProvider(() => {
    const token = useAuthStore.getState().accessToken
    const meta: Record<string, string> = {}
    if (token) meta.authorization = token
    return meta
  })
}

/** Convenience selectors used widely in components. */
export const selectIsAuthenticated = (s: AuthState) =>
  s.status === "authenticated" || s.status === "refreshing"
export const selectIdentity = (s: AuthState) => s.identity
