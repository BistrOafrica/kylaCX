import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import * as callsApi from "../api/calls"
import * as sessionsApi from "../api/sessions"
import * as queuesApi from "../api/queues"
import * as extensionsApi from "../api/extensions"
import * as ivrApi from "../api/ivr"
import { CallDirection, CallStatus } from "@/pb/call_session"

export function useCalls(args: {
  pageNumber?: number
  agentId?: string
  direction?: CallDirection
  status?: CallStatus
} = {}) {
  return useQuery({
    queryKey: ["calls", args],
    queryFn: () => callsApi.listCalls(args),
    staleTime: 15_000,
  })
}

export function useCall(id: string | null) {
  return useQuery({
    queryKey: id ? ["call", id] : ["call", "none"],
    enabled: Boolean(id),
    queryFn: () => callsApi.getCall(id!),
  })
}

export function useQueues() {
  return useQuery({
    queryKey: ["queues"],
    queryFn: () => queuesApi.listQueues({}),
  })
}

export function useQueue(id: string | null) {
  return useQuery({
    queryKey: id ? ["queue", id] : ["queue", "none"],
    enabled: Boolean(id),
    queryFn: () => queuesApi.getQueue(id!),
  })
}

export function useExtensions() {
  return useQuery({
    queryKey: ["extensions"],
    queryFn: () => extensionsApi.listExtensions({}),
  })
}

export function useExtensionByAgent(agentId: string | null) {
  return useQuery({
    queryKey: agentId ? ["extension", "agent", agentId] : ["extension", "none"],
    enabled: Boolean(agentId),
    queryFn: () => extensionsApi.getExtensionByAgent(agentId!),
  })
}

export function useIvrFlows() {
  return useQuery({
    queryKey: ["ivr-flows"],
    queryFn: () => ivrApi.listIvrFlows({}),
  })
}

export function useIvrFlow(id: string | null) {
  return useQuery({
    queryKey: id ? ["ivr-flow", id] : ["ivr-flow", "none"],
    enabled: Boolean(id),
    queryFn: () => ivrApi.getIvrFlow(id!),
  })
}

// ── Session mutations (softphone controls) ───────────────────────────────────

export function useStartCallSession() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: sessionsApi.startSession,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["calls"] }),
  })
}

export function useEndCallSession() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: sessionsApi.endSession,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["calls"] }),
  })
}

export function usePlaceOnHold() {
  return useMutation({ mutationFn: sessionsApi.placeOnHold })
}

export function useRemoveFromHold() {
  return useMutation({ mutationFn: sessionsApi.removeFromHold })
}

export function useAddCallNote() {
  return useMutation({ mutationFn: sessionsApi.addNote })
}

export function useHangupCall() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: callsApi.hangupCall,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["calls"] }),
  })
}
