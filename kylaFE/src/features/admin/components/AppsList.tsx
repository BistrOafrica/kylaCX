import { useState } from "react"
import {
  IconApps,
  IconPlus,
  IconRefresh,
  IconCopy,
  IconTrash,
  IconLoader2,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  Surface,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useApps,
  useCreateApp,
  useRegenerateAppSecret,
  useDeleteApp,
} from "../hooks/queries"
import type { App } from "../api/apps"
import { relativeShort } from "@/features/crm/utils/relative"

export function AppsList() {
  const apps = useApps()
  const create = useCreateApp()
  const [name, setName] = useState("")

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Apps"
        description="API tokens and webhooks"
      />

      <Surface level={1} radius="md" className="m-3 p-3 flex items-center gap-2">
        <IconApps className="size-4 text-fg-muted shrink-0" aria-hidden />
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={async (e) => {
            if (e.key === "Enter" && name.trim()) {
              const app = await create.mutateAsync({ name: name.trim() })
              if (app?.secret) {
                toast.success("App created — copy the secret now, it won't be shown again.")
              }
              setName("")
            }
          }}
          placeholder="New app name (e.g. CRM Integration)"
          disabled={create.isPending}
          className="h-8 flex-1"
        />
        <button
          type="button"
          disabled={create.isPending || !name.trim()}
          onClick={async () => {
            const app = await create.mutateAsync({ name: name.trim() })
            if (app?.secret) {
              toast.success("App created — copy the secret now, it won't be shown again.")
            }
            setName("")
          }}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          <IconPlus className="size-3.5" />
          Create
        </button>
      </Surface>

      {apps.isPending ? (
        <div className="p-3">
          <ListRowSkeleton count={4} />
        </div>
      ) : apps.isError ? (
        <ErrorState
          title="Couldn't load apps"
          description={(apps.error as Error).message}
          onRetry={() => void apps.refetch()}
        />
      ) : (apps.data ?? []).length === 0 ? (
        <EmptyState
          icon={<IconApps className="size-5" />}
          title="No apps yet"
          description="Apps own API tokens and webhook URLs for integrations."
        />
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {apps.data!.map((app) => (
            <AppRow key={app.id} app={app} />
          ))}
        </ul>
      )}
    </div>
  )
}

function AppRow({ app }: { app: App }) {
  const regen = useRegenerateAppSecret()
  const remove = useDeleteApp()

  return (
    <li className={cn("p-3 rounded-sm border border-border-subtle hover:bg-subtle transition-colors space-y-1.5")}>
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "size-9 rounded-sm bg-accent-subtle text-fg-secondary flex items-center justify-center shrink-0",
          )}
          aria-hidden
        >
          <IconApps className="size-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-md font-medium text-fg truncate">{app.name}</div>
          <div className="text-sm text-fg-muted truncate">
            {app.description || "No description"}
          </div>
        </div>
        <span
          className={cn(
            "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
            app.status === "ACTIVE" || app.status === "APPROVED"
              ? "bg-success-subtle text-success"
              : app.status === "PENDING"
                ? "bg-warn-subtle text-warn"
                : "bg-subtle text-fg-muted",
          )}
        >
          {app.status || "—"}
        </span>
        <button
          type="button"
          onClick={async () => {
            const next = await regen.mutateAsync(app)
            if (next?.secret) {
              await navigator.clipboard.writeText(next.secret)
              toast.success("New secret generated and copied to clipboard.")
            }
          }}
          disabled={regen.isPending}
          aria-label="Regenerate secret"
          className="inline-flex items-center justify-center size-7 rounded-xs text-fg-muted hover:text-fg hover:bg-canvas"
        >
          {regen.isPending ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconRefresh className="size-3.5" />
          )}
        </button>
        <button
          type="button"
          onClick={() => {
            if (confirm(`Delete app "${app.name}"?`)) remove.mutate(app.id)
          }}
          aria-label="Delete"
          className="inline-flex items-center justify-center size-7 rounded-xs text-fg-muted hover:text-danger hover:bg-canvas"
        >
          <IconTrash className="size-3.5" />
        </button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 ms-12">
        <CredField label="Token" value={app.token} />
        {app.secret && <CredField label="Secret" value={app.secret} />}
      </div>
      <div className="ms-12 text-xs font-mono text-fg-muted">
        Created {relativeShort(app.createdAt)}
      </div>
    </li>
  )
}

function CredField({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center gap-1.5">
      <span className="text-xs font-mono uppercase tracking-wider text-fg-muted w-12">
        {label}
      </span>
      <code className="flex-1 px-1.5 py-0.5 rounded-xs bg-canvas border border-border text-xs font-mono text-fg-secondary truncate">
        {value || "—"}
      </code>
      {value && (
        <button
          type="button"
          onClick={async () => {
            await navigator.clipboard.writeText(value)
            toast.success(`${label} copied`)
          }}
          aria-label={`Copy ${label}`}
          className="inline-flex items-center justify-center size-6 rounded-xs text-fg-muted hover:text-fg hover:bg-canvas"
        >
          <IconCopy className="size-3" />
        </button>
      )}
    </div>
  )
}
