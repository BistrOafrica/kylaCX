import { services, unary } from "@/lib/rpc"
import { useWorkspaceStore } from "@/lib/workspace"
import {
  CreateFormRequest,
  GetFormRequest,
  ListFormsRequest,
  UpdateFormRequest,
  DeleteFormRequest,
  SubmitFormRequest,
  ListSubmissionsRequest,
  FormStatus,
  type FormDefinition,
  type FormSubmission,
} from "@/pb/forms"

function scope() {
  const { organisation, workspace } = useWorkspaceStore.getState()
  return {
    orgId: organisation?.id ?? "",
    workspaceId: workspace?.id ?? "",
  }
}

/**
 * Form-field schema used by the builder UI. Serialised into
 * `FormDefinition.fields` as JSON. Backend treats `fields` as opaque;
 * we own this format end-to-end (only the SubmitForm payload echoes
 * back the same keys).
 */
export interface FormFieldSpec {
  key: string
  label: string
  type:
    | "text" | "textarea" | "email" | "phone" | "number"
    | "select" | "checkbox" | "date"
  required?: boolean
  placeholder?: string
  options?: { value: string; label: string }[]
  helpText?: string
}

export function parseFields(json: string): FormFieldSpec[] {
  if (!json) return []
  try {
    const arr = JSON.parse(json) as FormFieldSpec[]
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

export function serializeFields(fields: FormFieldSpec[]): string {
  return JSON.stringify(fields)
}

// ── Forms ────────────────────────────────────────────────────────────────────

export async function listForms(status?: FormStatus): Promise<FormDefinition[]> {
  const { orgId, workspaceId } = scope()
  const res = await unary(
    services.forms.listForms(
      ListFormsRequest.create({
        orgId,
        workspaceId,
        status: status ?? FormStatus.UNSPECIFIED,
      }) as ListFormsRequest,
    ),
  )
  return res.forms
}

export async function getForm(id: string): Promise<FormDefinition> {
  return unary(
    services.forms.getForm(
      GetFormRequest.create({
        id,
        orgId: scope().orgId,
      }) as GetFormRequest,
    ),
  )
}

export async function createForm(input: {
  name: string
  description?: string
  fields: FormFieldSpec[]
  submitRedirect?: string
}): Promise<FormDefinition> {
  const { orgId, workspaceId } = scope()
  return unary(
    services.forms.createForm(
      CreateFormRequest.create({
        orgId,
        workspaceId,
        name: input.name,
        description: input.description ?? "",
        fields: serializeFields(input.fields),
        submitRedirect: input.submitRedirect ?? "",
      }) as CreateFormRequest,
    ),
  )
}

export async function updateForm(input: {
  id: string
  name: string
  description?: string
  fields: FormFieldSpec[]
  status: FormStatus
  submitRedirect?: string
}): Promise<FormDefinition> {
  return unary(
    services.forms.updateForm(
      UpdateFormRequest.create({
        id: input.id,
        orgId: scope().orgId,
        name: input.name,
        description: input.description ?? "",
        fields: serializeFields(input.fields),
        status: input.status,
        submitRedirect: input.submitRedirect ?? "",
      }) as UpdateFormRequest,
    ),
  )
}

export async function deleteForm(id: string): Promise<void> {
  await unary(
    services.forms.deleteForm(
      DeleteFormRequest.create({
        id,
        orgId: scope().orgId,
      }) as DeleteFormRequest,
    ),
  )
}

// ── Submissions ──────────────────────────────────────────────────────────────

export async function submitForm(input: {
  formId: string
  data: Record<string, unknown>
}): Promise<FormSubmission> {
  return unary(
    services.forms.submitForm(
      SubmitFormRequest.create({
        orgId: scope().orgId,
        formId: input.formId,
        data: JSON.stringify(input.data),
      }) as SubmitFormRequest,
    ),
  )
}

export interface SubmissionsPage {
  submissions: FormSubmission[]
  nextPageToken: string
  total: number
}

export async function listSubmissions(
  formId: string,
  pageToken = "",
  pageSize = 50,
): Promise<SubmissionsPage> {
  const res = await unary(
    services.forms.listSubmissions(
      ListSubmissionsRequest.create({
        orgId: scope().orgId,
        formId,
        pageToken,
        pageSize,
      }) as ListSubmissionsRequest,
    ),
  )
  return {
    submissions: res.submissions,
    nextPageToken: res.nextPageToken,
    total: Number(res.total),
  }
}

export { FormStatus }
export type { FormDefinition, FormSubmission }
