import { create } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"
import { registerMetaProvider } from "@/lib/rpc"

/**
 * Workspace store.
 *
 * Holds the current organisation + workspace IDs. Every workspace-scoped
 * gRPC call attaches these as `x-org-id` / `x-workspace-id` metadata via
 * the registered provider.
 *
 * Selection persists across reloads. The shell hydrates from this store
 * on boot; if no workspace is set after auth completes, the user is
 * routed to a workspace picker (F1+).
 */

export interface WorkspaceRef {
  id: string
  name: string
  slug?: string
}

export interface OrganisationRef {
  id: string
  name: string
}

interface WorkspaceState {
  organisation: OrganisationRef | null
  workspace: WorkspaceRef | null

  setOrganisation: (org: OrganisationRef | null) => void
  setWorkspace: (ws: WorkspaceRef | null) => void
  reset: () => void
}

export const useWorkspaceStore = create<WorkspaceState>()(
  persist(
    (set) => ({
      organisation: null,
      workspace: null,

      setOrganisation: (organisation) => set({ organisation }),
      setWorkspace: (workspace) => set({ workspace }),
      reset: () => set({ organisation: null, workspace: null }),
    }),
    {
      name: "kyla.workspace.v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
)

/**
 * Register the workspace metadata provider so every RPC carries the
 * current org + workspace IDs as gRPC metadata. Backend interceptors
 * read these for scope resolution.
 */
export function initWorkspaceRpcMetadata(): () => void {
  return registerMetaProvider(() => {
    const { organisation, workspace } = useWorkspaceStore.getState()
    const meta: Record<string, string> = {}
    if (organisation) meta["x-org-id"] = organisation.id
    if (workspace) meta["x-workspace-id"] = workspace.id
    return meta
  })
}
