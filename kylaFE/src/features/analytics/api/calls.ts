import { services, unary } from "@/lib/rpc"
import { AnalyticsRequest } from "@/pb/call_analytics"
import { Timestamp } from "@/pb/google/protobuf/timestamp"
import { resolveRange, type TimeRangePreset } from "../utils/time-range"

/**
 * Call analytics — every method takes the same `AnalyticsRequest`,
 * which uses Timestamp fields rather than ISO strings (different
 * convention from ticket analytics).
 */
function buildRequest(preset: TimeRangePreset): AnalyticsRequest {
  const range = resolveRange(preset)
  return AnalyticsRequest.create({
    membershipId: "",
    startDate: Timestamp.fromDate(new Date(range.startDate)) as Timestamp,
    endDate: Timestamp.fromDate(new Date(range.endDate)) as Timestamp,
    dateRange: preset,
    queueIds: [],
    agentIds: [],
    campaignIds: [],
  }) as AnalyticsRequest
}

export async function getCallOverview(preset: TimeRangePreset) {
  const res = await unary(
    services.callAnalytics.getAnalyticsOverview(buildRequest(preset)),
  )
  return res.data ?? null
}

export async function getCallTraffic(preset: TimeRangePreset) {
  const res = await unary(
    services.callAnalytics.getCallTraffic(buildRequest(preset)),
  )
  return res.data ?? null
}

export async function getCallHandling(preset: TimeRangePreset) {
  const res = await unary(
    services.callAnalytics.getCallHandling(buildRequest(preset)),
  )
  return res.data ?? null
}

export async function getQueueIVR(preset: TimeRangePreset) {
  const res = await unary(
    services.callAnalytics.getQueueIVR(buildRequest(preset)),
  )
  return res.data ?? null
}

export async function getCustomerExperience(preset: TimeRangePreset) {
  const res = await unary(
    services.callAnalytics.getCustomerExperience(buildRequest(preset)),
  )
  return res.data ?? null
}
