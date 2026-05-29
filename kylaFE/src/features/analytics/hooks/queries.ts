import { useQuery } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/lib/workspace"
import * as ticketsApi from "../api/tickets"
import * as callsApi from "../api/calls"
import * as billingApi from "../api/billing"
import type { TimeRangePreset } from "../utils/time-range"

const wsId = () => useWorkspaceStore.getState().workspace?.id ?? ""

// ── Ticket analytics ─────────────────────────────────────────────────────────

export function useTicketKPIs(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "kpi", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getKPIs(preset),
  })
}

export function useTicketVolume(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "volume", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getTicketVolume(preset),
  })
}

export function useChannelDistribution(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "channels", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getChannelDistribution(preset),
  })
}

export function useAgentPerformance(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "agents", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getAgentPerformance(preset),
  })
}

export function useSLACompliance(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "sla", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getSLACompliance(preset),
  })
}

export function useStatusDistribution(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "status", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getStatusDistribution(preset),
  })
}

export function usePriorityDistribution(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "priority", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getPriorityDistribution(preset),
  })
}

export function useSentiment(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "tickets", "sentiment", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => ticketsApi.getSentiment(preset),
  })
}

// ── Call analytics ───────────────────────────────────────────────────────────

export function useCallOverview(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "calls", "overview", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => callsApi.getCallOverview(preset),
  })
}

export function useCallTraffic(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "calls", "traffic", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => callsApi.getCallTraffic(preset),
  })
}

export function useCallHandling(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "calls", "handling", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => callsApi.getCallHandling(preset),
  })
}

export function useCustomerExperience(preset: TimeRangePreset) {
  return useQuery({
    queryKey: ["analytics", "calls", "cx", wsId(), preset],
    enabled: Boolean(wsId()),
    queryFn: () => callsApi.getCustomerExperience(preset),
  })
}

// ── Billing ──────────────────────────────────────────────────────────────────

export function useWallets() {
  return useQuery({
    queryKey: ["billing", "wallets", wsId()],
    enabled: Boolean(wsId()),
    queryFn: billingApi.listWallets,
  })
}

export function useWalletTransactions(walletId: string | null) {
  return useQuery({
    queryKey: walletId
      ? ["billing", "wallet", walletId, "transactions"]
      : ["billing", "wallet", "none", "transactions"],
    enabled: Boolean(walletId),
    queryFn: () => billingApi.listWalletTransactions(walletId!),
  })
}
