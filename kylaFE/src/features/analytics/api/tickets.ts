import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  AnalyticsFilters,
  TimeRange,
  GetKPIMetricsRequest,
  GetTicketVolumeRequest,
  GetChannelDistributionRequest,
  GetAgentPerformanceRequest,
  GetSLAComplianceRequest,
  GetStatusDistributionRequest,
  GetPriorityDistributionRequest,
  GetSentimentAnalyticsRequest,
} from "@/pb/ticket_analytics"
import { OwnerType } from "@/pb/owner_type"
import { resolveRange, granularityFor, type TimeRangePreset } from "../utils/time-range"

/**
 * Ticket analytics — every request takes a shared AnalyticsFilters.
 *
 * Filters come from a preset; consumers pass the preset id and these
 * wrappers expand it into the proto's TimeRange + owner scope.
 */
function buildFilters(preset: TimeRangePreset): AnalyticsFilters {
  const range = resolveRange(preset)
  const orgId = useWorkspaceStore.getState().organisation?.id ?? ""
  return AnalyticsFilters.create({
    timeRange: TimeRange.create({
      startDate: range.startDate,
      endDate: range.endDate,
    }) as TimeRange,
    teams: [],
    agents: [],
    categories: [],
    priorities: [],
    statuses: [],
    ownerId: orgId,
    ownerType: OwnerType.ORGANISATIONS,
    includeSpam: false,
  }) as AnalyticsFilters
}

export async function getKPIs(preset: TimeRangePreset) {
  return unary(
    services.ticketAnalytics.getKPIMetrics(
      GetKPIMetricsRequest.create({
        filters: buildFilters(preset),
      }) as GetKPIMetricsRequest,
    ),
  )
}

export async function getTicketVolume(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getTicketVolume(
      GetTicketVolumeRequest.create({
        filters: buildFilters(preset),
        granularity: granularityFor(preset),
      }) as GetTicketVolumeRequest,
    ),
  )
  return res.dataPoints
}

export async function getChannelDistribution(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getChannelDistribution(
      GetChannelDistributionRequest.create({
        filters: buildFilters(preset),
      }) as GetChannelDistributionRequest,
    ),
  )
  return res.channels
}

export async function getAgentPerformance(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getAgentPerformance(
      GetAgentPerformanceRequest.create({
        filters: buildFilters(preset),
        page: 1,
        perPage: 10,
      }) as GetAgentPerformanceRequest,
    ),
  )
  return res.agents
}

export async function getSLACompliance(preset: TimeRangePreset) {
  return unary(
    services.ticketAnalytics.getSLACompliance(
      GetSLAComplianceRequest.create({
        filters: buildFilters(preset),
      }) as GetSLAComplianceRequest,
    ),
  )
}

export async function getStatusDistribution(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getStatusDistribution(
      GetStatusDistributionRequest.create({
        filters: buildFilters(preset),
      }) as GetStatusDistributionRequest,
    ),
  )
  return res.statuses
}

export async function getPriorityDistribution(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getPriorityDistribution(
      GetPriorityDistributionRequest.create({
        filters: buildFilters(preset),
      }) as GetPriorityDistributionRequest,
    ),
  )
  return res.priorities
}

export async function getSentiment(preset: TimeRangePreset) {
  const res = await unary(
    services.ticketAnalytics.getSentimentAnalytics(
      GetSentimentAnalyticsRequest.create({
        filters: buildFilters(preset),
      }) as GetSentimentAnalyticsRequest,
    ),
  )
  return res
}
