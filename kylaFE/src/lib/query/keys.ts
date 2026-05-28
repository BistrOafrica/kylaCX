/**
 * Query-key factory.
 *
 * Every cached query in the app references its key through this
 * factory, so refactors and invalidations are typed end-to-end.
 *
 *   useQuery({ queryKey: qk.conversations.list(wsId, filters), ... })
 *
 * Convention:
 *   qk.<domain>.<view>(<scope>, <args?>)  →  readonly tuple
 */

export const qk = {
  auth: {
    me:          () => ["auth", "me"] as const,
    permissions: (userId: string) => ["auth", "permissions", userId] as const,
  },
  workspaces: {
    list:    ()           => ["workspaces", "list"] as const,
    detail:  (id: string) => ["workspaces", "detail", id] as const,
    members: (id: string) => ["workspaces", "members", id] as const,
  },
  organisations: {
    me:     ()           => ["organisations", "me"] as const,
    list:   ()           => ["organisations", "list"] as const,
    detail: (id: string) => ["organisations", "detail", id] as const,
  },
  users: {
    list:   (orgId: string) => ["users", "list", orgId] as const,
    detail: (id: string)    => ["users", "detail", id] as const,
  },
  conversations: {
    all:     ()                                => ["conversations"] as const,
    list:    (wsId: string, filters?: unknown) =>
      ["conversations", "list", wsId, filters] as const,
    detail:   (id: string)                     => ["conversations", "detail", id] as const,
    messages: (id: string)                     => ["conversations", "messages", id] as const,
    timeline: (id: string)                     => ["conversations", "timeline", id] as const,
    summary:  (id: string)                     => ["conversations", "summary", id] as const,
    sentiment: (id: string)                    => ["conversations", "sentiment", id] as const,
  },
  contacts: {
    detail:  (id: string) => ["contacts", "detail", id] as const,
  },
} as const
