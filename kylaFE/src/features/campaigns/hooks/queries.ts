import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import * as whatsappApi from "../api/whatsapp"
import * as autodialerApi from "../api/autodialer"

// ── WhatsApp ─────────────────────────────────────────────────────────────────

export function useWhatsappCampaigns() {
  return useQuery({
    queryKey: ["campaigns", "whatsapp"],
    queryFn: whatsappApi.listWhatsappCampaigns,
  })
}

export function useWhatsappCampaign(id: string | null) {
  return useQuery({
    queryKey: id ? ["campaign", "whatsapp", id] : ["campaign", "whatsapp", "none"],
    enabled: Boolean(id),
    queryFn: () => whatsappApi.getWhatsappCampaign(id!),
  })
}

export function useWhatsappTemplates() {
  return useQuery({
    queryKey: ["campaigns", "whatsapp", "templates"],
    queryFn: whatsappApi.listWhatsappTemplates,
    staleTime: 5 * 60_000,
  })
}

export function useCreateWhatsappCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: whatsappApi.createWhatsappCampaign,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["campaigns", "whatsapp"] }),
  })
}

export function useCreateWhatsappTemplate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: whatsappApi.createWhatsappTemplate,
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["campaigns", "whatsapp", "templates"] }),
  })
}

export function useDeleteWhatsappTemplate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: whatsappApi.deleteWhatsappTemplate,
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["campaigns", "whatsapp", "templates"] }),
  })
}

export function useCampaignAnalytics(campaignId: string | null, status = "") {
  return useQuery({
    queryKey: campaignId
      ? ["campaign", "whatsapp", campaignId, "analytics", status]
      : ["campaign", "whatsapp", "none", "analytics"],
    enabled: Boolean(campaignId),
    queryFn: () => whatsappApi.getCampaignAnalytics(campaignId!, status),
    refetchInterval: 30_000,
  })
}

// ── Auto-dialer ──────────────────────────────────────────────────────────────

export function useAutodialerCampaigns() {
  return useQuery({
    queryKey: ["campaigns", "autodialer"],
    queryFn: autodialerApi.listAutoDialerCampaigns,
  })
}

export function useAutodialerCampaign(id: string | null) {
  return useQuery({
    queryKey: id
      ? ["campaign", "autodialer", id]
      : ["campaign", "autodialer", "none"],
    enabled: Boolean(id),
    queryFn: () => autodialerApi.getAutoDialerCampaign(id!),
  })
}

export function useCreateAutodialerCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: autodialerApi.createAutoDialerCampaign,
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["campaigns", "autodialer"] }),
  })
}

export function useDeleteAutodialerCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: autodialerApi.deleteAutoDialerCampaign,
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["campaigns", "autodialer"] }),
  })
}
