import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import { useWorkspaceStore } from "@/lib/workspace"
import * as workflowApi from "../api/workflows"

export function useWorkflows() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  return useQuery({
    queryKey: ["workflows", workspaceId],
    enabled: Boolean(workspaceId),
    queryFn: () => workflowApi.listWorkflows(),
  })
}

export function useWorkflow(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["workflow", id] : ["workflow", "none"],
    enabled: Boolean(id),
    queryFn: () => workflowApi.getWorkflow(id!),
  })
}

export function useWorkflowRuns(id: string | null | undefined) {
  return useQuery({
    queryKey: id ? ["workflow", id, "runs"] : ["workflow", "none", "runs"],
    enabled: Boolean(id),
    queryFn: () => workflowApi.getRunHistory(id!),
    refetchInterval: (q) => {
      const data = q.state.data as workflowApi.RunHistoryPage | undefined
      const hasInFlight = data?.runs.some((r) => r.status === "running" || r.status === "pending")
      return hasInFlight ? 3_000 : false
    },
  })
}

export function useCreateWorkflow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: workflowApi.createWorkflow,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["workflows"] }),
  })
}

export function useUpdateWorkflow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: workflowApi.updateWorkflow,
    onSuccess: (w) => {
      qc.setQueryData(["workflow", w.id], w)
      void qc.invalidateQueries({ queryKey: ["workflows"] })
    },
  })
}

export function useDeleteWorkflow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: workflowApi.deleteWorkflow,
    onSuccess: (_v, id) => {
      qc.removeQueries({ queryKey: ["workflow", id] })
      void qc.invalidateQueries({ queryKey: ["workflows"] })
    },
  })
}

export function useTestRunWorkflow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { workflowId: string; sampleEvent?: Record<string, unknown> }) =>
      workflowApi.testRunWorkflow(input.workflowId, input.sampleEvent ?? {}),
    onSuccess: (_r, vars) => {
      void qc.invalidateQueries({ queryKey: ["workflow", vars.workflowId, "runs"] })
    },
  })
}
