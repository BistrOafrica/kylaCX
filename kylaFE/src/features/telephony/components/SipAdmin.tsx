import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  IconPlus,
  IconTrash,
  IconLoader2,
  IconRefresh,
  IconDeviceFloppy,
  IconUsers,
  IconNetwork,
  IconWorld,
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
import { useWorkspaceStore } from "@/lib/workspace"
import type { SipExtension, SipTrunk, SipDomain } from "@/pb/telephony"
import {
  listSipExtensions,
  createSipExtension,
  deleteSipExtension,
  listSipTrunks,
  createSipTrunk,
  updateSipTrunk,
  deleteSipTrunk,
  listSipDomains,
  createSipDomain,
  deleteSipDomain,
} from "../api/sipAdmin"

/**
 * SipAdmin — three-tab control surface for the self-hosted SIP stack:
 * Extensions (one per agent), Trunks (outbound PSTN), Domains (SIP realms).
 *
 * All three lists poll on a slow cadence (10s) so registration state stays
 * roughly current without overwhelming the API.
 */

type Tab = "extensions" | "trunks" | "domains"

export function SipAdmin() {
  const [tab, setTab] = useState<Tab>("extensions")
  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="SIP infrastructure"
        description="Extensions, trunks, and domains for the self-hosted PBX."
      />
      <div className="px-3 pt-2 flex gap-1 border-b border-default">
        <TabButton current={tab} value="extensions" onClick={setTab} icon={<IconUsers className="size-4" />}>
          Extensions
        </TabButton>
        <TabButton current={tab} value="trunks" onClick={setTab} icon={<IconNetwork className="size-4" />}>
          Trunks
        </TabButton>
        <TabButton current={tab} value="domains" onClick={setTab} icon={<IconWorld className="size-4" />}>
          Domains
        </TabButton>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto p-3">
        {tab === "extensions" && <ExtensionsPanel />}
        {tab === "trunks" && <TrunksPanel />}
        {tab === "domains" && <DomainsPanel />}
      </div>
    </div>
  )
}

function TabButton({
  current,
  value,
  onClick,
  icon,
  children,
}: {
  current: Tab
  value: Tab
  onClick: (v: Tab) => void
  icon: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <button
      onClick={() => onClick(value)}
      className={cn(
        "inline-flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-t-sm border-b-2",
        current === value
          ? "border-accent text-fg font-medium"
          : "border-transparent text-fg-muted hover:text-fg",
      )}
    >
      {icon}
      {children}
    </button>
  )
}

// ── Extensions ──────────────────────────────────────────────────────────────

function ExtensionsPanel() {
  const workspaceId = useWorkspaceStore((s) => s.workspace?.id ?? "")
  const qc = useQueryClient()
  const list = useQuery({
    queryKey: ["sip-extensions", workspaceId],
    queryFn: () => listSipExtensions(workspaceId),
    refetchInterval: 10_000,
    enabled: !!workspaceId,
  })
  const create = useMutation({
    mutationFn: (e: { extension: string; displayName: string; userId: string }) =>
      createSipExtension({
        ...e,
        workspaceId,
      }),
    onSuccess: () => {
      toast.success("Extension created")
      qc.invalidateQueries({ queryKey: ["sip-extensions", workspaceId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })
  const remove = useMutation({
    mutationFn: deleteSipExtension,
    onSuccess: () => {
      toast.success("Extension deleted")
      qc.invalidateQueries({ queryKey: ["sip-extensions", workspaceId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })

  const [ext, setExt] = useState("")
  const [name, setName] = useState("")
  const [userId, setUserId] = useState("")

  return (
    <div className="space-y-3">
      <Surface level={1} radius="md" className="p-3 space-y-2">
        <div className="text-sm font-medium">New extension</div>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-2">
          <Input placeholder="Extension (e.g. 1001)" value={ext} onChange={(e) => setExt(e.target.value)} />
          <Input placeholder="Display name" value={name} onChange={(e) => setName(e.target.value)} />
          <Input placeholder="User ID" value={userId} onChange={(e) => setUserId(e.target.value)} />
          <button
            disabled={!ext.trim() || !userId.trim() || create.isPending}
            onClick={() => {
              create.mutate({ extension: ext.trim(), displayName: name.trim(), userId: userId.trim() })
              setExt("")
              setName("")
              setUserId("")
            }}
            className={cn(
              "inline-flex items-center justify-center gap-1.5 h-9 px-3 rounded-sm",
              "bg-accent text-accent-fg disabled:opacity-50",
            )}
          >
            {create.isPending ? <IconLoader2 className="size-4 animate-spin" /> : <IconPlus className="size-4" />}
            Create
          </button>
        </div>
      </Surface>

      <ListContainer
        loading={list.isPending}
        error={list.error as Error | undefined}
        empty={(list.data ?? []).length === 0}
        onRefresh={() => list.refetch()}
        emptyTitle="No SIP extensions yet"
        emptyDescription="Create an extension to provision an agent's softphone."
      >
        {(list.data ?? []).map((e) => (
          <ExtensionRow key={e.id} extension={e} onDelete={() => remove.mutate(e.id)} />
        ))}
      </ListContainer>
    </div>
  )
}

function ExtensionRow({ extension, onDelete }: { extension: SipExtension; onDelete: () => void }) {
  return (
    <div className="flex items-center gap-3 py-2 px-3 border-b border-default last:border-b-0">
      <div className="flex-1">
        <div className="text-sm font-medium font-mono">{extension.extension}</div>
        <div className="text-xs text-fg-muted">{extension.displayName || extension.userId}</div>
      </div>
      <StatusBadge value={extension.status} okValue="registered" />
      <button
        onClick={onDelete}
        className="text-fg-muted hover:text-danger"
        aria-label="Delete extension"
      >
        <IconTrash className="size-4" />
      </button>
    </div>
  )
}

// ── Trunks ──────────────────────────────────────────────────────────────────

function TrunksPanel() {
  const orgId = useWorkspaceStore((s) => s.organisation?.id ?? "")
  const qc = useQueryClient()
  const list = useQuery({
    queryKey: ["sip-trunks", orgId],
    queryFn: () => listSipTrunks(orgId),
    refetchInterval: 30_000,
    enabled: !!orgId,
  })
  const create = useMutation({
    mutationFn: (t: Partial<SipTrunk>) => createSipTrunk(t),
    onSuccess: () => {
      toast.success("Trunk created")
      qc.invalidateQueries({ queryKey: ["sip-trunks", orgId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })
  const update = useMutation({
    mutationFn: (t: SipTrunk) => updateSipTrunk(t),
    onSuccess: () => {
      toast.success("Trunk updated")
      qc.invalidateQueries({ queryKey: ["sip-trunks", orgId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })
  const remove = useMutation({
    mutationFn: deleteSipTrunk,
    onSuccess: () => {
      toast.success("Trunk deleted")
      qc.invalidateQueries({ queryKey: ["sip-trunks", orgId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })

  const [draft, setDraft] = useState({
    name: "",
    gatewayName: "",
    provider: "custom",
    sipServer: "",
    username: "",
    password: "",
    fromUri: "",
  })

  return (
    <div className="space-y-3">
      <Surface level={1} radius="md" className="p-3 space-y-2">
        <div className="text-sm font-medium">New trunk</div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
          <Input placeholder="Name (e.g. Twilio Elastic)" value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
          <Input placeholder="Gateway name (FS profile)" value={draft.gatewayName} onChange={(e) => setDraft({ ...draft, gatewayName: e.target.value })} />
          <Input placeholder="SIP server" value={draft.sipServer} onChange={(e) => setDraft({ ...draft, sipServer: e.target.value })} />
          <Input placeholder="From URI" value={draft.fromUri} onChange={(e) => setDraft({ ...draft, fromUri: e.target.value })} />
          <Input placeholder="Username" value={draft.username} onChange={(e) => setDraft({ ...draft, username: e.target.value })} />
          <Input placeholder="Password" type="password" value={draft.password} onChange={(e) => setDraft({ ...draft, password: e.target.value })} />
        </div>
        <button
          disabled={!draft.name.trim() || !draft.gatewayName.trim() || create.isPending}
          onClick={() => {
            create.mutate(draft)
            setDraft({ name: "", gatewayName: "", provider: "custom", sipServer: "", username: "", password: "", fromUri: "" })
          }}
          className={cn(
            "inline-flex items-center gap-1.5 h-9 px-3 rounded-sm",
            "bg-accent text-accent-fg disabled:opacity-50",
          )}
        >
          {create.isPending ? <IconLoader2 className="size-4 animate-spin" /> : <IconPlus className="size-4" />}
          Create
        </button>
      </Surface>

      <ListContainer
        loading={list.isPending}
        error={list.error as Error | undefined}
        empty={(list.data ?? []).length === 0}
        onRefresh={() => list.refetch()}
        emptyTitle="No SIP trunks yet"
        emptyDescription="Add a trunk to make outbound PSTN calls."
      >
        {(list.data ?? []).map((t) => (
          <TrunkRow
            key={t.id}
            trunk={t}
            onToggleActive={() => update.mutate({ ...t, isActive: !t.isActive })}
            onDelete={() => remove.mutate(t.id)}
          />
        ))}
      </ListContainer>
    </div>
  )
}

function TrunkRow({
  trunk,
  onToggleActive,
  onDelete,
}: {
  trunk: SipTrunk
  onToggleActive: () => void
  onDelete: () => void
}) {
  return (
    <div className="flex items-center gap-3 py-2 px-3 border-b border-default last:border-b-0">
      <div className="flex-1">
        <div className="text-sm font-medium">{trunk.name}</div>
        <div className="text-xs text-fg-muted font-mono">{trunk.gatewayName} · {trunk.sipServer || "—"}</div>
      </div>
      <button
        onClick={onToggleActive}
        className={cn(
          "text-xs px-2 py-0.5 rounded-xs",
          trunk.isActive ? "bg-success-subtle text-success" : "bg-subtle text-fg-muted",
        )}
      >
        {trunk.isActive ? "active" : "inactive"}
      </button>
      <button onClick={onDelete} className="text-fg-muted hover:text-danger" aria-label="Delete trunk">
        <IconTrash className="size-4" />
      </button>
    </div>
  )
}

// ── Domains ─────────────────────────────────────────────────────────────────

function DomainsPanel() {
  const orgId = useWorkspaceStore((s) => s.organisation?.id ?? "")
  const qc = useQueryClient()
  const list = useQuery({
    queryKey: ["sip-domains", orgId],
    queryFn: () => listSipDomains(orgId),
    refetchInterval: 60_000,
    enabled: !!orgId,
  })
  const create = useMutation({
    mutationFn: (d: { domain: string; isDefault: boolean }) => createSipDomain(d),
    onSuccess: () => {
      toast.success("Domain created")
      qc.invalidateQueries({ queryKey: ["sip-domains", orgId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })
  const remove = useMutation({
    mutationFn: deleteSipDomain,
    onSuccess: () => {
      toast.success("Domain deleted")
      qc.invalidateQueries({ queryKey: ["sip-domains", orgId] })
    },
    onError: (e: Error) => toast.error(e.message),
  })

  const [domain, setDomain] = useState("")
  const [isDefault, setIsDefault] = useState(false)

  return (
    <div className="space-y-3">
      <Surface level={1} radius="md" className="p-3 space-y-2">
        <div className="text-sm font-medium">New domain</div>
        <div className="flex items-center gap-2">
          <Input placeholder="kyla.example.com" value={domain} onChange={(e) => setDomain(e.target.value)} className="flex-1" />
          <label className="inline-flex items-center gap-1 text-xs text-fg-muted">
            <input type="checkbox" checked={isDefault} onChange={(e) => setIsDefault(e.target.checked)} />
            Default
          </label>
          <button
            disabled={!domain.trim() || create.isPending}
            onClick={() => {
              create.mutate({ domain: domain.trim(), isDefault })
              setDomain("")
              setIsDefault(false)
            }}
            className={cn(
              "inline-flex items-center gap-1.5 h-9 px-3 rounded-sm",
              "bg-accent text-accent-fg disabled:opacity-50",
            )}
          >
            {create.isPending ? <IconLoader2 className="size-4 animate-spin" /> : <IconPlus className="size-4" />}
            Create
          </button>
        </div>
      </Surface>

      <ListContainer
        loading={list.isPending}
        error={list.error as Error | undefined}
        empty={(list.data ?? []).length === 0}
        onRefresh={() => list.refetch()}
        emptyTitle="No SIP domains configured"
        emptyDescription="At least one domain is needed before extensions can register."
      >
        {(list.data ?? []).map((d) => (
          <DomainRow key={d.id} domain={d} onDelete={() => remove.mutate(d.id)} />
        ))}
      </ListContainer>
    </div>
  )
}

function DomainRow({ domain, onDelete }: { domain: SipDomain; onDelete: () => void }) {
  return (
    <div className="flex items-center gap-3 py-2 px-3 border-b border-default last:border-b-0">
      <div className="flex-1 text-sm font-mono">{domain.domain}</div>
      {domain.isDefault && (
        <span className="text-xs px-2 py-0.5 rounded-xs bg-accent-subtle text-accent">default</span>
      )}
      <button onClick={onDelete} className="text-fg-muted hover:text-danger" aria-label="Delete domain">
        <IconTrash className="size-4" />
      </button>
    </div>
  )
}

// ── Shared list shell ───────────────────────────────────────────────────────

function ListContainer({
  loading,
  error,
  empty,
  onRefresh,
  emptyTitle,
  emptyDescription,
  children,
}: {
  loading: boolean
  error?: Error
  empty: boolean
  onRefresh: () => void
  emptyTitle: string
  emptyDescription: string
  children: React.ReactNode
}) {
  return (
    <Surface level={1} radius="md" className="overflow-hidden">
      <div className="flex items-center justify-between px-3 py-2 border-b border-default">
        <div className="text-xs text-fg-muted">Showing live state</div>
        <button onClick={onRefresh} className="text-fg-muted hover:text-fg" aria-label="Refresh">
          <IconRefresh className="size-4" />
        </button>
      </div>
      {loading ? (
        <div className="p-3 space-y-1">
          <ListRowSkeleton />
          <ListRowSkeleton />
          <ListRowSkeleton />
        </div>
      ) : error ? (
        <ErrorState title="Couldn't load" description={error.message} onRetry={onRefresh} />
      ) : empty ? (
        <EmptyState title={emptyTitle} description={emptyDescription} />
      ) : (
        <div>{children}</div>
      )}
    </Surface>
  )
}

function StatusBadge({ value, okValue }: { value: string; okValue: string }) {
  const ok = value === okValue
  return (
    <span
      className={cn(
        "text-xs px-2 py-0.5 rounded-xs",
        ok ? "bg-success-subtle text-success" : "bg-subtle text-fg-muted",
      )}
    >
      {value || "—"}
    </span>
  )
}

// Re-exporting IconDeviceFloppy keeps it available for follow-up edit modals
// without forcing a re-import. Intentionally unused at the file level today.
void IconDeviceFloppy
