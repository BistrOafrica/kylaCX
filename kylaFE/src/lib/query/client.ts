import { QueryClient, MutationCache, QueryCache } from "@tanstack/react-query"
import { toast } from "sonner"
import { RpcError } from "@/lib/rpc"
import { useAuthStore } from "@/lib/auth/store"

/**
 * TanStack Query client.
 *
 * Conventions:
 *   - Queries are typed via the `qk` factory in keys.ts.
 *   - Errors normalize to RpcError before being inspected here.
 *   - UNAUTHENTICATED clears the session and lets the route guard
 *     redirect the user. Other user-facing errors surface as a toast.
 *   - 30s stale time by default — most domain lists are not refetched
 *     on every focus; explicit invalidation drives updates instead.
 */

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        gcTime: 5 * 60_000,
        retry: (failureCount, err) => {
          if (err instanceof RpcError) {
            // No point retrying user errors or auth failures.
            if (err.isUserFacing || err.isUnauthenticated) return false
          }
          return failureCount < 1
        },
        refetchOnWindowFocus: false,
      },
      mutations: {
        retry: false,
      },
    },
    queryCache: new QueryCache({
      onError(err, query) {
        const normalized = err instanceof RpcError ? err : RpcError.from(err)
        if (normalized.isUnauthenticated) {
          useAuthStore.getState().signOut()
          return
        }
        // Background refetch failures shouldn't toast unless the query
        // actually has an observer mounted. TanStack Query already
        // surfaces error state via `query.state.error`; the toast is
        // a courtesy for visible-but-silent failures.
        if (query.getObserversCount() > 0 && normalized.isUserFacing) {
          toast.error(normalized.message)
        }
      },
    }),
    mutationCache: new MutationCache({
      onError(err) {
        const normalized = err instanceof RpcError ? err : RpcError.from(err)
        if (normalized.isUnauthenticated) {
          useAuthStore.getState().signOut()
          return
        }
        if (normalized.isUserFacing) {
          toast.error(normalized.message)
        } else {
          toast.error("Something went wrong. Please try again.")
        }
      },
    }),
  })
}
