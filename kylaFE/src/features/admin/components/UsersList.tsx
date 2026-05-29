import { useMemo, useState } from "react"
import {
  IconUsers,
  IconUserOff,
  IconUserCheck,
  IconMail,
  IconSearch,
  IconSend,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
  Surface,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useUsers,
  useInvitations,
  useCreateInvitation,
  useCancelInvitation,
  useActivateUser,
  useDeactivateUser,
} from "../hooks/queries"
import { InvitationStatus } from "../api/invitations"
import { relativeShort } from "@/features/crm/utils/relative"

/**
 * UsersList — combined users + invitations admin panel.
 *
 *   ┌───────────────────────────────────────────────┐
 *   │ Tabs: Active users · Pending invites          │
 *   ├───────────────────────────────────────────────┤
 *   │ Invite form (always visible on top)            │
 *   ├───────────────────────────────────────────────┤
 *   │ List                                           │
 *   └───────────────────────────────────────────────┘
 */
export function UsersList() {
  const [tab, setTab] = useState<"users" | "invites">("users")
  const users = useUsers()
  const invites = useInvitations(
    tab === "invites" ? InvitationStatus.PENDING : undefined,
  )

  const [search, setSearch] = useState("")
  const filteredUsers = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return users.data ?? []
    return (users.data ?? []).filter(
      (u) =>
        u.email.toLowerCase().includes(q) ||
        u.firstName.toLowerCase().includes(q) ||
        u.lastName.toLowerCase().includes(q),
    )
  }, [users.data, search])

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Users & invitations"
        description="People in your organisation"
      />

      <InviteForm />

      <nav
        role="tablist"
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
      >
        <TabBtn
          active={tab === "users"}
          onClick={() => setTab("users")}
          label="Active users"
          count={users.data?.length ?? 0}
        />
        <TabBtn
          active={tab === "invites"}
          onClick={() => setTab("invites")}
          label="Pending invites"
          count={invites.data?.length ?? 0}
        />
        {tab === "users" && (
          <div className="ms-auto relative w-64">
            <IconSearch className="absolute start-2 top-1/2 -translate-y-1/2 size-3.5 text-fg-muted" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search users…"
              className="h-7 ps-7 text-base"
            />
          </div>
        )}
      </nav>

      {tab === "users" ? (
        users.isPending ? (
          <ListRowSkeleton count={6} />
        ) : users.isError ? (
          <ErrorState
            title="Couldn't load users"
            description={(users.error as Error).message}
            onRetry={() => void users.refetch()}
          />
        ) : filteredUsers.length === 0 ? (
          <EmptyState
            icon={<IconUsers className="size-5" />}
            title={search ? "No matches" : "No users yet"}
            description="Invite teammates above to populate this list."
          />
        ) : (
          <ul className="flex-1 overflow-y-auto p-1" role="list">
            {filteredUsers.map((u) => (
              <UserRow key={u.id} user={u} />
            ))}
          </ul>
        )
      ) : invites.isPending ? (
        <ListRowSkeleton count={6} />
      ) : invites.isError ? (
        <ErrorState
          title="Couldn't load invitations"
          description={(invites.error as Error).message}
          onRetry={() => void invites.refetch()}
        />
      ) : (invites.data ?? []).length === 0 ? (
        <EmptyState
          icon={<IconMail className="size-5" />}
          title="No pending invites"
          description="Use the form above to send one."
        />
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {invites.data!.map((inv) => (
            <InviteRow key={inv.id} invite={inv} />
          ))}
        </ul>
      )}
    </div>
  )
}

function TabBtn({
  active,
  onClick,
  label,
  count,
}: {
  active: boolean
  onClick: () => void
  label: string
  count: number
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md",
        "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
        active && "bg-accent-subtle text-fg font-medium",
      )}
    >
      {label}
      <span className="font-mono text-xs text-fg-muted">{count}</span>
    </button>
  )
}

function InviteForm() {
  const [email, setEmail] = useState("")
  const create = useCreateInvitation()
  const submit = async () => {
    const value = email.trim()
    if (!value) return
    try {
      await create.mutateAsync({ email: value })
      toast.success(`Invitation sent to ${value}`)
      setEmail("")
    } catch {
      /* mutationCache toasts */
    }
  }
  return (
    <Surface level={1} radius="md" className="m-3 p-3">
      <div className="flex items-center gap-2">
        <IconMail className="size-4 text-fg-muted shrink-0" aria-hidden />
        <Input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void submit()
          }}
          placeholder="teammate@company.com"
          disabled={create.isPending}
          className="h-8 flex-1"
        />
        <button
          type="button"
          onClick={() => void submit()}
          disabled={create.isPending || !email.trim()}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          <IconSend className="size-3.5" />
          Send invite
        </button>
      </div>
    </Surface>
  )
}

function UserRow({
  user,
}: {
  user: {
    id: string
    firstName: string
    lastName: string
    email: string
    status: string
    lastLogin?: string
  }
}) {
  const activate = useActivateUser()
  const deactivate = useDeactivateUser()
  const isActive = user.status === "active" || user.status === "ACTIVE"
  const name = [user.firstName, user.lastName].filter(Boolean).join(" ") || user.email

  return (
    <li
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <div
        className={cn(
          "size-8 rounded-full flex items-center justify-center font-medium text-sm shrink-0",
          isActive ? "bg-accent-subtle text-fg" : "bg-subtle text-fg-muted",
        )}
        aria-hidden
      >
        {(name[0] ?? "?").toUpperCase()}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-md text-fg truncate">{name}</div>
        <div className="text-sm text-fg-muted truncate font-mono">
          {user.email}
        </div>
      </div>
      <span
        className={cn(
          "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
          isActive ? "bg-success-subtle text-success" : "bg-subtle text-fg-muted",
        )}
      >
        {user.status || "inactive"}
      </span>
      {user.lastLogin && (
        <span className="font-mono text-xs text-fg-muted w-16 text-end">
          {relativeShort(user.lastLogin)}
        </span>
      )}
      <button
        type="button"
        onClick={() =>
          isActive ? deactivate.mutate(user.id) : activate.mutate(user.id)
        }
        disabled={activate.isPending || deactivate.isPending}
        className={cn(
          "inline-flex items-center justify-center size-7 rounded-xs",
          "text-fg-muted hover:bg-canvas",
          isActive ? "hover:text-warn" : "hover:text-success",
        )}
        aria-label={isActive ? "Deactivate" : "Activate"}
      >
        {isActive ? (
          <IconUserOff className="size-3.5" />
        ) : (
          <IconUserCheck className="size-3.5" />
        )}
      </button>
    </li>
  )
}

function InviteRow({
  invite,
}: {
  invite: {
    id: string
    email: string
    status: InvitationStatus
    createdAt?: { seconds: bigint | string }
  }
}) {
  const cancel = useCancelInvitation()
  return (
    <li
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <IconMail className="size-4 text-fg-muted" aria-hidden />
      <span className="font-mono text-md text-fg flex-1 truncate">
        {invite.email}
      </span>
      <span className="inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium bg-warn-subtle text-warn">
        Pending
      </span>
      <button
        type="button"
        onClick={() => void cancel.mutateAsync(invite.id)}
        disabled={cancel.isPending}
        className="text-sm text-danger hover:underline disabled:opacity-50"
      >
        Cancel
      </button>
    </li>
  )
}
