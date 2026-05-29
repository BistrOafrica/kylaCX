import { services, unary } from "@/lib/rpc"
import {
  ReadQueueRequest,
  ReadQueuesRequest,
  type CallQueue,
  CallersInQueueRequest,
  RingStrategy,
  QueueType,
  PickupStrategy,
} from "@/pb/call_queue"

export async function listQueues(args: {
  pageNumber?: number
  pageSize?: number
}): Promise<{ items: CallQueue[]; total: number }> {
  const res = await unary(
    services.callQueue.readQueues(
      ReadQueuesRequest.create({
        pageNumber: args.pageNumber ?? 1,
        pageSize: args.pageSize ?? 50,
      }) as ReadQueuesRequest,
    ),
  )
  return {
    items: (res as { items?: CallQueue[] }).items ?? [],
    total: Number((res as { total?: bigint | number }).total ?? 0),
  }
}

export async function getQueue(id: string): Promise<CallQueue> {
  return unary(
    services.callQueue.readQueue(
      ReadQueueRequest.create({ id }) as ReadQueueRequest,
    ),
  )
}

export async function getCallersInQueue(queueId: string) {
  return unary(
    services.callQueue.getCallersInQueue(
      CallersInQueueRequest.create({ queueId }) as CallersInQueueRequest,
    ),
  )
}

export { RingStrategy, QueueType, PickupStrategy }
export type { CallQueue }
