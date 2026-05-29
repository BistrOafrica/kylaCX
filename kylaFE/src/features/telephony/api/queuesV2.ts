import { services } from "@/lib/rpc/services"
import { stream, type Subscription } from "@/lib/rpc"
import type {
  Queue,
  QueueEntry,
  QueueMembership,
  WatchQueueEntriesUpdate,
} from "@/pb/queues"

/**
 * Helpers for the Phase 5d QueueService gRPC. The legacy CallQueueService
 * helpers in api/queues.ts continue to work; new components should use these.
 *
 * Naming: V2 keeps this file's symbols from colliding with the legacy
 * exports — both are simultaneously importable for a transitional period.
 */

export async function listQueuesV2(workspaceId: string, activeOnly = true): Promise<Queue[]> {
  const res = await services.queue.listQueues({ workspaceId, activeOnly })
  return res.response.queues
}

export async function getQueueV2(id: string): Promise<Queue> {
  const res = await services.queue.getQueue({ id })
  return res.response
}

export async function listQueueMembersV2(queueId: string): Promise<QueueMembership[]> {
  const res = await services.queue.listQueueMembers({ queueId })
  return res.response.members
}

/**
 * listQueueEntriesV2 fetches live entries for a queue. status filter is
 * optional — empty fetches all non-ended entries. Used by the wallboard
 * polling loop.
 */
export async function listQueueEntriesV2(queueId: string, status = ""): Promise<QueueEntry[]> {
  const res = await services.queue.listQueueEntries({ queueId, status })
  return res.response.entries
}

export async function setMemberActiveV2(
  queueId: string,
  userId: string,
  isActive: boolean,
): Promise<QueueMembership> {
  const res = await services.queue.setQueueMemberActive({ queueId, userId, isActive })
  return res.response.member!
}

/**
 * subscribeQueueEntriesV2 opens a server-side stream that pushes a snapshot
 * whenever the queue's entries change. The server emits an initial snapshot
 * and then only on transitions (plus a heartbeat every 15s).
 *
 * intervalMs hints the server's polling cadence; it's clamped to [500, 10000].
 * Returns the same Subscription handle used elsewhere in the codebase so
 * components can `cancel()` on unmount.
 */
export function subscribeQueueEntriesV2(
  queueId: string,
  onUpdate: (msg: WatchQueueEntriesUpdate) => void,
  intervalMs = 1500,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.queue.watchQueueEntries(
        { queueId, intervalMs },
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}
