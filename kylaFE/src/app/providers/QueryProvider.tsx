import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { useState, type ReactNode } from "react"
import { createQueryClient } from "@/lib/query"
import { env } from "@/lib/rpc"

/**
 * Owns the QueryClient lifecycle. One client per app instance, kept in
 * state so React Refresh doesn't blow away the cache mid-session.
 */
export function QueryProvider({ children }: { children: ReactNode }) {
  const [client] = useState<QueryClient>(() => createQueryClient())
  return (
    <QueryClientProvider client={client}>
      {children}
      {env.isDev && <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-left" />}
    </QueryClientProvider>
  )
}
