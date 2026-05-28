import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreateSavedViewRequest,
  GetSavedViewRequest,
  ListSavedViewsRequest,
  UpdateSavedViewRequest,
  DeleteSavedViewRequest,
  SavedView,
} from "@/pb/object_core"

/**
 * Saved-view persistence — the same primitive used by inbox filters in
 * the future, but F2 ships it first because CRM lists drive the
 * heaviest demand for shared filters.
 *
 * The proto wraps Create/Update with a full SavedView message rather
 * than scalar fields, so callers pass a partial body and we build the
 * envelope here.
 */

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

export interface SavedViewBody {
  name: string
  typeSlug: string
  filters?: Record<string, unknown>
  sort?: { by: string; desc: boolean }
  columns?: string[]
  isShared?: boolean
}

export async function listSavedViews(typeSlug: string): Promise<SavedView[]> {
  const { workspaceId } = scope()
  const res = await unary(
    services.views.listSavedViews(
      ListSavedViewsRequest.create({
        workspaceId,
        typeSlug,
      }) as ListSavedViewsRequest,
    ),
  )
  return res.views
}

export async function getSavedView(id: string): Promise<SavedView> {
  const { workspaceId } = scope()
  return unary(
    services.views.getSavedView(
      GetSavedViewRequest.create({
        id,
        workspaceId,
      }) as GetSavedViewRequest,
    ),
  )
}

export async function createSavedView(input: SavedViewBody): Promise<SavedView> {
  const { orgId, workspaceId } = scope()
  const view = SavedView.create({
    id: "",
    orgId,
    workspaceId,
    name: input.name,
    typeSlug: input.typeSlug,
    filters: input.filters ? JSON.stringify(input.filters) : "",
    sort: input.sort ? JSON.stringify(input.sort) : "",
    columns: input.columns ? JSON.stringify(input.columns) : "",
    isShared: input.isShared ?? false,
    createdBy: "",
    createdAt: "",
    updatedAt: "",
  }) as SavedView

  return unary(
    services.views.createSavedView(
      CreateSavedViewRequest.create({ view }) as CreateSavedViewRequest,
    ),
  )
}

export async function updateSavedView(
  id: string,
  patch: Partial<SavedViewBody>,
): Promise<SavedView> {
  const current = await getSavedView(id)
  const view = SavedView.create({
    ...current,
    name: patch.name ?? current.name,
    filters: patch.filters ? JSON.stringify(patch.filters) : current.filters,
    sort: patch.sort ? JSON.stringify(patch.sort) : current.sort,
    columns: patch.columns ? JSON.stringify(patch.columns) : current.columns,
    isShared: patch.isShared ?? current.isShared,
  }) as SavedView
  return unary(
    services.views.updateSavedView(
      UpdateSavedViewRequest.create({ view }) as UpdateSavedViewRequest,
    ),
  )
}

export async function deleteSavedView(id: string): Promise<void> {
  const { workspaceId } = scope()
  await unary(
    services.views.deleteSavedView(
      DeleteSavedViewRequest.create({
        id,
        workspaceId,
      }) as DeleteSavedViewRequest,
    ),
  )
}

export type { SavedView }
