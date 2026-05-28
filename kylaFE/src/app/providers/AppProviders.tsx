import { useEffect, type ReactNode } from "react"
import { I18nextProvider } from "react-i18next"
import { ThemeProvider } from "@/components/theme-provider"
import { Toaster } from "@/components/ui/sonner"
import { QueryProvider } from "./QueryProvider"
import { i18n } from "@/lib/i18n"
import {
  initAuthRpcMetadata,
  startAutoRefresh,
  useAuthStore,
} from "@/lib/auth"
import { initWorkspaceRpcMetadata } from "@/lib/workspace"

/**
 * App-wide providers. Order matters:
 *
 *   ThemeProvider     dark/light + storage key
 *   I18nextProvider   translation context
 *   QueryProvider     TanStack Query client
 *
 * Auth + workspace metadata providers are wired as side-effects at boot;
 * the auth store hydrates from localStorage on first render.
 */

export function AppProviders({ children }: { children: ReactNode }) {
  const hydrate = useAuthStore((s) => s.hydrate)
  const authStatus = useAuthStore((s) => s.status)

  useEffect(() => {
    // Hydrate the auth store from localStorage exactly once on mount.
    hydrate()

    // Register gRPC metadata providers — these read the current auth +
    // workspace state on every outbound call.
    const unregAuth = initAuthRpcMetadata()
    const unregWs = initWorkspaceRpcMetadata()

    return () => {
      unregAuth()
      unregWs()
    }
  }, [hydrate])

  // Auto-refresh runs only while authenticated; subscribe/teardown based
  // on status. The refresh module itself watches the store, so we only
  // need to start it once.
  useEffect(() => {
    if (authStatus !== "authenticated") return
    const stop = startAutoRefresh()
    return stop
  }, [authStatus])

  return (
    <ThemeProvider defaultTheme="system" storageKey="kyla.theme">
      <I18nextProvider i18n={i18n}>
        <QueryProvider>
          {children}
          <Toaster />
        </QueryProvider>
      </I18nextProvider>
    </ThemeProvider>
  )
}
