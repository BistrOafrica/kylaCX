import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query"
import { qk } from "@/lib/query"
import { useWorkspaceStore } from "@/lib/workspace"
import * as objApi from "../api/objects"
import * as crmApi from "../api/crm"
import * as viewsApi from "../api/views"

// ── Object types (schema) ────────────────────────────────────────────────────

export function useObjectTypes() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["object-types", workspaceId],
    enabled: Boolean(workspaceId),
    queryFn: objApi.listObjectTypes,
    staleTime: 5 * 60_000,
  })
}

export function useObjectType(slug: string | null | undefined) {
  return useQuery({
    queryKey: slug ? ["object-type", slug] : ["object-type", "none"],
    enabled: Boolean(slug),
    queryFn: () => objApi.getObjectType(slug!),
    staleTime: 5 * 60_000,
  })
}

// ── Objects (records) ────────────────────────────────────────────────────────

export interface ObjectsQueryArgs {
  typeSlug: string
  filter?: string
  sortBy?: string
  sortDesc?: boolean
}

export function useObjects(args: ObjectsQueryArgs) {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useInfiniteQuery({
    queryKey: ["objects", workspaceId, args.typeSlug, args.filter ?? "", args.sortBy ?? "", args.sortDesc ?? false],
    enabled: Boolean(workspaceId && args.typeSlug),
    initialPageParam: "",
    queryFn: ({ pageParam }) =>
      objApi.listObjects({
        typeSlug: args.typeSlug,
        pageToken: pageParam as string,
        filter: args.filter,
        sortBy: args.sortBy,
        sortDesc: args.sortDesc,
      }),
    getNextPageParam: (last) =>
      last.nextPageToken && last.records.length > 0
        ? last.nextPageToken
        : undefined,
    staleTime: 15_000,
  })
}

export function useObject(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["object", id] : ["object", "none"],
    enabled: Boolean(id),
    queryFn: () => objApi.getObject(id!),
  })
}

export function useObjectTimeline(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["object", id, "timeline"] : ["object", "none", "timeline"],
    enabled: Boolean(id),
    queryFn: () => objApi.getObjectTimeline(id!),
  })
}

export function useObjectRelations(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["object", id, "relations"] : ["object", "none", "relations"],
    enabled: Boolean(id),
    queryFn: () => objApi.getObjectRelations(id!),
  })
}

export function useCreateObject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: objApi.CreateObjectInput) => objApi.createObject(input),
    onSuccess: (record) => {
      void qc.invalidateQueries({ queryKey: ["objects"] })
      qc.setQueryData(["object", record.id], record)
    },
  })
}

export function useUpdateObject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: objApi.UpdateObjectInput) => objApi.updateObject(input),
    onSuccess: (record) => {
      qc.setQueryData(["object", record.id], record)
      void qc.invalidateQueries({ queryKey: ["objects"] })
      void qc.invalidateQueries({ queryKey: ["object", record.id, "timeline"] })
    },
  })
}

export function useDeleteObject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => objApi.deleteObject(id),
    onSuccess: (_v, id) => {
      qc.removeQueries({ queryKey: ["object", id] })
      void qc.invalidateQueries({ queryKey: ["objects"] })
    },
  })
}

// ── CRM (pipelines + stages + kanban) ────────────────────────────────────────

export function usePipelines() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["pipelines", workspaceId],
    enabled: Boolean(workspaceId),
    queryFn: crmApi.listPipelines,
  })
}

export function usePipeline(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["pipeline", id] : ["pipeline", "none"],
    enabled: Boolean(id),
    queryFn: () => crmApi.getPipeline(id!),
  })
}

export function useStages(pipelineId: string | null | undefined) {
  return useQuery({
    queryKey: pipelineId ? ["pipeline", pipelineId, "stages"] : ["pipeline", "none", "stages"],
    enabled: Boolean(pipelineId),
    queryFn: () => crmApi.listStages(pipelineId!),
  })
}

export function usePipelineBoard(pipelineId: string | null | undefined) {
  return useQuery({
    queryKey: pipelineId ? ["pipeline", pipelineId, "board"] : ["pipeline", "none", "board"],
    enabled: Boolean(pipelineId),
    queryFn: () => crmApi.getPipelineBoard(pipelineId!),
    staleTime: 10_000,
  })
}

export function useMoveDeal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { dealObjectId: string; stageId: string }) =>
      crmApi.moveDeal(input.dealObjectId, input.stageId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["pipeline"] })
      void qc.invalidateQueries({ queryKey: ["objects"] })
    },
  })
}

// ── Saved views ──────────────────────────────────────────────────────────────

export function useSavedViews(typeSlug: string | null | undefined) {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["saved-views", workspaceId, typeSlug],
    enabled: Boolean(workspaceId && typeSlug),
    queryFn: () => viewsApi.listSavedViews(typeSlug!),
  })
}

export function useCreateSavedView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: viewsApi.SavedViewBody) => viewsApi.createSavedView(body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["saved-views"] }),
  })
}

// Augment the qk barrel with crm-friendly keys for inter-module use.
void qk
