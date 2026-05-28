import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import { useWorkspaceStore } from "@/lib/workspace"
import * as ticketingApi from "../api/ticketing"
import * as knowledgeApi from "../api/knowledge"
import * as formsApi from "../api/forms"

// ── Ticketing ────────────────────────────────────────────────────────────────

export function useTicketRooms(ticketId: string | null | undefined) {
  return useQuery({
    queryKey: ticketId ? ["ticket", ticketId, "rooms"] : ["ticket", "none", "rooms"],
    enabled: Boolean(ticketId),
    queryFn: () => ticketingApi.listTicketRooms(ticketId!),
  })
}

export function useRoomMessages(roomId: string | null | undefined) {
  return useQuery({
    queryKey: roomId ? ["room", roomId, "messages"] : ["room", "none", "messages"],
    enabled: Boolean(roomId),
    queryFn: () => ticketingApi.listRoomMessages(roomId!).then((r) => r.messages),
    staleTime: 5_000,
  })
}

export function useAddRoomMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { roomId: string; content: string; isPrivate?: boolean }) =>
      ticketingApi.addRoomMessage(input),
    onSuccess: (_msg, vars) => {
      void qc.invalidateQueries({ queryKey: ["room", vars.roomId, "messages"] })
    },
  })
}

export function useCreateTicketRoom() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { ticketId: string; name: string; type?: ticketingApi.RoomType }) =>
      ticketingApi.createTicketRoom(input),
    onSuccess: (_room, vars) => {
      void qc.invalidateQueries({ queryKey: ["ticket", vars.ticketId, "rooms"] })
    },
  })
}

export function useMacros() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["macros", workspaceId],
    enabled: Boolean(workspaceId),
    queryFn: ticketingApi.listMacros,
  })
}

export function useApplyMacro() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { macroId: string; ticketId: string; roomId?: string }) =>
      ticketingApi.applyMacro(input),
    onSuccess: (_v, vars) => {
      void qc.invalidateQueries({ queryKey: ["objects"] })
      if (vars.roomId) {
        void qc.invalidateQueries({ queryKey: ["room", vars.roomId, "messages"] })
      }
    },
  })
}

export function useCreateMacro() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ticketingApi.createMacro,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["macros"] }),
  })
}

export function useUpdateMacro() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ticketingApi.updateMacro,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["macros"] }),
  })
}

export function useDeleteMacro() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ticketingApi.deleteMacro,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["macros"] }),
  })
}

// ── Knowledge ────────────────────────────────────────────────────────────────

export function useKnowledgeBases() {
  return useQuery({
    queryKey: ["knowledge", "list"],
    queryFn: () => knowledgeApi.listKnowledgeBases(100),
  })
}

export function useKnowledgeBase(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["knowledge", id] : ["knowledge", "none"],
    enabled: Boolean(id),
    queryFn: () => knowledgeApi.getKnowledgeBase(id!),
  })
}

export function useKnowledgeSearch(query: string) {
  return useQuery({
    queryKey: ["knowledge", "search", query],
    enabled: query.trim().length >= 2,
    queryFn: () => knowledgeApi.searchKnowledge(query),
    staleTime: 30_000,
  })
}

export function useCreateKnowledgeBase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: knowledgeApi.createKnowledgeBase,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["knowledge"] }),
  })
}

export function useUpdateKnowledgeBase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: knowledgeApi.updateKnowledgeBase,
    onSuccess: (kb) => {
      if (kb) qc.setQueryData(["knowledge", kb.id], kb)
      void qc.invalidateQueries({ queryKey: ["knowledge", "list"] })
    },
  })
}

export function useDeleteKnowledgeBase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: knowledgeApi.deleteKnowledgeBase,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["knowledge"] }),
  })
}

// ── Forms ────────────────────────────────────────────────────────────────────

export function useForms() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["forms", workspaceId],
    enabled: Boolean(workspaceId),
    queryFn: () => formsApi.listForms(),
  })
}

export function useForm(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["form", id] : ["form", "none"],
    enabled: Boolean(id),
    queryFn: () => formsApi.getForm(id!),
  })
}

export function useFormSubmissions(formId: string | null | undefined) {
  return useQuery({
    queryKey: formId ? ["form", formId, "submissions"] : ["form", "none", "submissions"],
    enabled: Boolean(formId),
    queryFn: () => formsApi.listSubmissions(formId!),
  })
}

export function useCreateForm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: formsApi.createForm,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["forms"] }),
  })
}

export function useUpdateForm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: formsApi.updateForm,
    onSuccess: (form) => {
      qc.setQueryData(["form", form.id], form)
      void qc.invalidateQueries({ queryKey: ["forms"] })
    },
  })
}

export function useDeleteForm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: formsApi.deleteForm,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["forms"] }),
  })
}
