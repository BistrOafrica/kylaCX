import { Link } from "react-router-dom"
import {
  IconPlus,
  IconFileText,
  IconCircleCheck,
  IconCircle,
} from "@tabler/icons-react"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useForms } from "../hooks/queries"
import { FormStatus, type FormDefinition } from "../api/forms"
import { relativeShort } from "@/features/crm/utils/relative"

export function FormsList() {
  const forms = useForms()

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Forms"
        description="Public data-collection forms"
        actions={
          <button
            type="button"
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover"
          >
            <IconPlus className="size-3.5" />
            New form
          </button>
        }
      />

      {forms.isPending ? (
        <div className="flex-1 p-2">
          <ListRowSkeleton count={6} />
        </div>
      ) : forms.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load forms"
            description={(forms.error as Error).message}
            onRetry={() => void forms.refetch()}
          />
        </div>
      ) : (forms.data?.length ?? 0) === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconFileText className="size-5" />}
            title="No forms yet"
            description="Build forms to collect leads, support requests, or surveys."
          />
        </div>
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {forms.data!.map((form) => (
            <li key={form.id}>
              <FormRow form={form} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function FormRow({ form }: { form: FormDefinition }) {
  return (
    <Link
      to={`/forms/${form.id}`}
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <div
        className={cn(
          "size-7 rounded-sm flex items-center justify-center shrink-0",
          "bg-accent-subtle text-fg-secondary",
        )}
        aria-hidden
      >
        <IconFileText className="size-3.5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-md font-medium text-fg truncate">{form.name}</div>
        {form.description && (
          <div className="text-sm text-fg-muted truncate">{form.description}</div>
        )}
      </div>
      <StatusBadge status={form.status} />
      <span className="font-mono text-xs text-fg-muted w-20 text-end">
        {form.submissionCount} resp
      </span>
      <span className="font-mono text-xs text-fg-muted w-16 text-end">
        {relativeShort(form.updatedAt)}
      </span>
    </Link>
  )
}

function StatusBadge({ status }: { status: FormStatus }) {
  switch (status) {
    case FormStatus.ACTIVE:
      return (
        <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-success-subtle text-success">
          <IconCircleCheck className="size-3" />
          Active
        </span>
      )
    case FormStatus.DRAFT:
      return (
        <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-subtle text-fg-muted">
          <IconCircle className="size-3" />
          Draft
        </span>
      )
    case FormStatus.CLOSED:
      return (
        <span className="inline-flex items-center gap-1 h-5 px-1.5 rounded-xs text-xs font-medium bg-warn-subtle text-warn">
          Closed
        </span>
      )
    default:
      return null
  }
}
