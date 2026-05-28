import { services, unary, RpcError } from "@/lib/rpc"
import { useAuthStore, type AuthIdentity } from "./store"
import {
  LoginRequest,
  LoginWithMFARequest,
  LoginWithPasskeyRequest,
  RefreshTokenRequest,
  LogoutRequest,
  MFAVerifyRequest,
  MFASetupRequest,
} from "@/pb/auth"
import type { User } from "@/pb/user"

/**
 * Auth API — thin wrappers around AuthService RPCs.
 *
 * These functions update the Zustand auth store as their side effect
 * so callers (login form, refresh interceptor) only deal with the
 * happy path. Failures bubble as RpcError.
 */

function identityFromUser(user: User | undefined, fallbackEmail?: string): AuthIdentity {
  if (!user) {
    return { userId: "", email: fallbackEmail }
  }
  const displayName = [user.firstName, user.lastName].filter(Boolean).join(" ").trim()
  return {
    userId: user.id,
    email: user.email || fallbackEmail,
    displayName: displayName || undefined,
  }
}

/** Standard email + password login. */
export async function login(email: string, password: string) {
  const res = await unary(
    services.auth.login(
      LoginRequest.create({ email, password }) as LoginRequest,
    ),
  )

  useAuthStore.getState().setSession({
    accessToken: res.accessToken,
    refreshToken: res.refreshToken,
    identity: identityFromUser(res.user, email),
  })
  return res
}

/** Two-step MFA login: send code + user_id from the MFA challenge. */
export async function loginWithMfa(userId: string, code: string) {
  const res = await unary(
    services.auth.loginWithMFA(
      LoginWithMFARequest.create({ userId, code }) as LoginWithMFARequest,
    ),
  )
  useAuthStore.getState().setSession({
    accessToken: res.accessToken,
    refreshToken: res.refreshToken,
    identity: identityFromUser(res.user),
  })
  return res
}

/** Login via WebAuthn assertion. The credential bytes come from the browser. */
export async function loginWithPasskey(payload: LoginWithPasskeyRequest) {
  const res = await unary(services.auth.loginWithPasskey(payload))
  useAuthStore.getState().setSession({
    accessToken: res.accessToken,
    refreshToken: res.refreshToken,
    identity: identityFromUser(res.user),
  })
  return res
}

/** Verify a TOTP code without rotating tokens (used by MFA setup confirmation). */
export async function verifyMfa(userId: string, code: string) {
  return unary(
    services.auth.verifyMFA(
      MFAVerifyRequest.create({ userId, code }) as MFAVerifyRequest,
    ),
  )
}

/** Generate a TOTP secret + QR code for first-time MFA enrollment. */
export async function setupMfa(userId: string) {
  return unary(
    services.auth.setupMFA(
      MFASetupRequest.create({ userId }) as MFASetupRequest,
    ),
  )
}

/**
 * Refresh tokens. Throws if the refresh token itself is no longer
 * valid — callers should sign out on that branch.
 */
export async function refreshTokens(): Promise<void> {
  const state = useAuthStore.getState()
  if (!state.refreshToken) {
    throw new RpcError("No refresh token", "UNAUTHENTICATED")
  }

  state.beginRefresh()

  const res = await unary(
    services.auth.refreshToken(
      RefreshTokenRequest.create({
        refreshToken: state.refreshToken,
      }) as RefreshTokenRequest,
    ),
  )

  useAuthStore.getState().setSession({
    accessToken: res.accessToken,
    refreshToken: res.refreshToken,
    identity: state.identity ?? { userId: "" },
  })
}

/** Server-side logout + local wipe. Best-effort: local wipe always happens. */
export async function logout() {
  const { accessToken } = useAuthStore.getState()
  try {
    if (accessToken) {
      await unary(
        services.auth.logout(
          LogoutRequest.create({ accessToken }) as LogoutRequest,
        ),
      )
    }
  } catch {
    // Best-effort. Local wipe is the source of truth.
  } finally {
    useAuthStore.getState().signOut()
  }
}
