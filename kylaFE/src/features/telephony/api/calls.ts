import { services, unary, stream, type Subscription } from "@/lib/rpc"
import {
  ReadCallLogsRequest,
  ReadCallLogRequest,
  CallLogFilter,
  type CallLog,
} from "@/pb/call_log"
import { CallDirection, CallStatus } from "@/pb/call_session"

/**
 * Call log API — read-side wrappers for call history.
 */

export interface CallsPage {
  items: CallLog[]
  total: number
}

export async function listCalls(args: {
  pageNumber?: number
  pageSize?: number
  agentId?: string
  queueId?: string
  direction?: CallDirection
  status?: CallStatus
}): Promise<CallsPage> {
  const res = await unary(
    services.callLog.readLogs(
      ReadCallLogsRequest.create({
        pageNumber: args.pageNumber ?? 1,
        pageSize: args.pageSize ?? 50,
        filters: CallLogFilter.create({
          agentId: args.agentId ?? "",
          queueId: args.queueId ?? "",
          direction: args.direction ?? CallDirection.INBOUND,
          status: args.status ?? CallStatus.COMPLETED,
        }) as CallLogFilter,
      }) as ReadCallLogsRequest,
    ),
  )
  return { items: res.items, total: Number(res.total) }
}

export async function getCall(id: string): Promise<CallLog> {
  return unary(
    services.callLog.readLog(
      ReadCallLogRequest.create({ id }) as ReadCallLogRequest,
    ),
  )
}

export async function getCallByCallId(callId: string): Promise<CallLog> {
  return unary(
    services.callLog.readLogByCallId(
      ReadCallLogRequest.create({ id: callId }) as ReadCallLogRequest,
    ),
  )
}

// ── Monitoring streams (server-streaming) ────────────────────────────────────

import {
  ReadOnlineAgentsResponse,
  ReadRecentCallsRequest,
  type ReadRecentCallsResponse,
  ReadAgentStatusesRequest,
  type ReadAgentStatusesResponse,
  ReadQueueDataRequest,
  type ReadQueueDataResponse,
  ReadCallMetricsRequest,
  type ReadCallMetricsResponse,
  HangupCallRequest,
} from "@/pb/call_monitoring"

export function subscribeOnlineAgents(
  onUpdate: (msg: ReadOnlineAgentsResponse) => void,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.callMonitoring.readOnlineAgents(
        ReadOnlineAgentsResponse.create({}) as ReadOnlineAgentsResponse,
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}

export function subscribeRecentCalls(
  onUpdate: (msg: ReadRecentCallsResponse) => void,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.callMonitoring.readRecentCalls(
        ReadRecentCallsRequest.create({}) as ReadRecentCallsRequest,
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}

export function subscribeAgentStatuses(
  onUpdate: (msg: ReadAgentStatusesResponse) => void,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.callMonitoring.readAgentStatuses(
        ReadAgentStatusesRequest.create({}) as ReadAgentStatusesRequest,
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}

export function subscribeQueueData(
  onUpdate: (msg: ReadQueueDataResponse) => void,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.callMonitoring.readQueues(
        ReadQueueDataRequest.create({}) as ReadQueueDataRequest,
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}

export function subscribeCallMetrics(
  onUpdate: (msg: ReadCallMetricsResponse) => void,
  onError?: (err: Error) => void,
): Subscription {
  return stream(
    (opts) =>
      services.callMonitoring.readCallMetrics(
        ReadCallMetricsRequest.create({}) as ReadCallMetricsRequest,
        opts,
      ),
    {
      onMessage: onUpdate,
      onError,
    },
  )
}

// ── Supervisor actions ───────────────────────────────────────────────────────

export async function hangupCall(callId: string): Promise<void> {
  await unary(
    services.callMonitoring.hangupCall(
      HangupCallRequest.create({ callId }) as HangupCallRequest,
    ),
  )
}

export { CallDirection, CallStatus }
export type { CallLog }
