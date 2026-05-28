import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreateObjectRequest,
  GetObjectRequest,
  ListObjectsRequest,
  SearchObjectsRequest,
  UpdateObjectRequest,
  DeleteObjectRequest,
  GetObjectTimelineRequest,
  GetObjectRelationsRequest,
  LinkObjectsRequest,
  UnlinkObjectsRequest,
  GetObjectTypeRequest,
  ListObjectTypesRequest,
  UpdateObjectSchemaRequest,
  type ObjectSchema,
  type Object as CoreObjectMessage,
} from "@/pb/object_core"
import { toRecord, type ObjectRecord, type ObjectType } from "../utils/types"

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

// ── Object types (schema) ────────────────────────────────────────────────────

export async function listObjectTypes(): Promise<ObjectType[]> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.objectCore.listObjectTypes(
      ListObjectTypesRequest.create({ orgId, workspaceId }) as ListObjectTypesRequest,
    ),
  )
  return res.objectTypes
}

export async function getObjectType(slug: string): Promise<ObjectType> {
  return unary(
    services.objectCore.getObjectType(
      GetObjectTypeRequest.create({
        orgId: scope().orgId,
        slug,
      }) as GetObjectTypeRequest,
    ),
  )
}

export async function updateObjectSchema(
  slug: string,
  schema: ObjectSchema,
): Promise<ObjectType> {
  return unary(
    services.objectCore.updateObjectSchema(
      UpdateObjectSchemaRequest.create({
        orgId: scope().orgId,
        slug,
        schema,
      }) as UpdateObjectSchemaRequest,
    ),
  )
}

// ── Objects (records) ────────────────────────────────────────────────────────

export interface ListObjectsArgs {
  typeSlug: string
  pageToken?: string
  pageSize?: number
  filter?: string         // JSON
  sortBy?: string
  sortDesc?: boolean
}

export interface ObjectsPage {
  records: ObjectRecord[]
  nextPageToken: string
  total: number
}

export async function listObjects(args: ListObjectsArgs): Promise<ObjectsPage> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.objectCore.listObjects(
      ListObjectsRequest.create({
        orgId,
        workspaceId,
        typeSlug: args.typeSlug,
        pageSize: args.pageSize ?? 50,
        pageToken: args.pageToken ?? "",
        filter: args.filter ?? "",
        sortBy: args.sortBy ?? "",
        sortDesc: args.sortDesc ?? false,
      }) as ListObjectsRequest,
    ),
  )
  return {
    records: res.objects.map(toRecord),
    nextPageToken: res.nextPageToken,
    total: Number(res.total),
  }
}

export async function searchObjects(
  query: string,
  typeSlug = "",
  pageSize = 25,
): Promise<ObjectRecord[]> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.objectCore.searchObjects(
      SearchObjectsRequest.create({
        orgId,
        workspaceId,
        typeSlug,
        query,
        pageSize,
      }) as SearchObjectsRequest,
    ),
  )
  return res.objects.map(toRecord)
}

export async function getObject(id: string): Promise<ObjectRecord> {
  const obj = await unary(
    services.objectCore.getObject(
      GetObjectRequest.create({
        id,
        orgId: scope().orgId,
      }) as GetObjectRequest,
    ),
  )
  return toRecord(obj as CoreObjectMessage)
}

export interface CreateObjectInput {
  typeSlug: string
  data: Record<string, unknown>
}

export async function createObject(input: CreateObjectInput): Promise<ObjectRecord> {
  const { orgId, workspaceId } = scope()
  const obj = await unary(
    services.objectCore.createObject(
      CreateObjectRequest.create({
        orgId,
        workspaceId,
        typeSlug: input.typeSlug,
        data: JSON.stringify(input.data),
      }) as CreateObjectRequest,
    ),
  )
  return toRecord(obj as CoreObjectMessage)
}

export interface UpdateObjectInput {
  id: string
  data: Record<string, unknown>
}

export async function updateObject(input: UpdateObjectInput): Promise<ObjectRecord> {
  const obj = await unary(
    services.objectCore.updateObject(
      UpdateObjectRequest.create({
        id: input.id,
        orgId: scope().orgId,
        data: JSON.stringify(input.data),
      }) as UpdateObjectRequest,
    ),
  )
  return toRecord(obj as CoreObjectMessage)
}

export async function deleteObject(id: string): Promise<void> {
  await unary(
    services.objectCore.deleteObject(
      DeleteObjectRequest.create({
        id,
        orgId: scope().orgId,
      }) as DeleteObjectRequest,
    ),
  )
}

// ── Timeline + relations ─────────────────────────────────────────────────────

export interface TimelineEvent {
  id: string
  objectId: string
  actorId: string
  actorType: string
  eventType: string
  payload: string
  createdAt: string
}

export async function getObjectTimeline(objectId: string): Promise<TimelineEvent[]> {
  const res = await unary(
    services.objectCore.getObjectTimeline(
      GetObjectTimelineRequest.create({
        objectId,
        orgId: scope().orgId,
      }) as GetObjectTimelineRequest,
    ),
  )
  return res.events.map((e) => ({
    id: e.id,
    objectId: e.objectId,
    actorId: e.actorId,
    actorType: e.actorType,
    eventType: e.eventType,
    payload: e.payload,
    createdAt: e.createdAt,
  }))
}

export async function getObjectRelations(objectId: string) {
  const res = await unary(
    services.objectCore.getObjectRelations(
      GetObjectRelationsRequest.create({
        objectId,
        orgId: scope().orgId,
      }) as GetObjectRelationsRequest,
    ),
  )
  return res.relations
}

export async function linkObjects(
  fromId: string,
  toId: string,
  relation: string,
) {
  return unary(
    services.objectCore.linkObjects(
      LinkObjectsRequest.create({
        fromId,
        toId,
        relation,
        orgId: scope().orgId,
      }) as LinkObjectsRequest,
    ),
  )
}

export async function unlinkObjects(
  fromId: string,
  toId: string,
  relation: string,
) {
  return unary(
    services.objectCore.unlinkObjects(
      UnlinkObjectsRequest.create({
        fromId,
        toId,
        relation,
      }) as UnlinkObjectsRequest,
    ),
  )
}
