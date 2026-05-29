import { useMemo, useState } from "react"
import {
  IconBuilding,
  IconBuildingCommunity,
  IconUsers,
  IconPlus,
  IconTrash,
  IconChevronRight,
} from "@tabler/icons-react"
import {
  PageHeader,
  CardSkeleton,
  ErrorState,
  EmptyState,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useBranches,
  useDepartments,
  useTeams,
  useCreateBranch,
  useCreateDepartment,
  useCreateTeam,
  useDeleteBranch,
  useDeleteDepartment,
  useDeleteTeam,
} from "../hooks/queries"
import { OwnerType } from "../utils/scope"

/**
 * OrganisationTree — three-column drill-down editor for the
 * org → branches → departments → teams hierarchy.
 *
 * Selecting a branch narrows the departments column to that branch;
 * selecting a department narrows teams. Each column has its own
 * inline "+" creator.
 */
export function OrganisationTree() {
  const branches = useBranches()
  const departments = useDepartments()
  const teams = useTeams()

  const [activeBranch, setActiveBranch] = useState<string | null>(null)
  const [activeDept, setActiveDept] = useState<string | null>(null)

  const departmentsInBranch = useMemo(() => {
    if (!activeBranch) return departments.data ?? []
    return (departments.data ?? []).filter(
      (d) =>
        d.ownerType === OwnerType.BRANCHES && d.ownerId === activeBranch,
    )
  }, [departments.data, activeBranch])

  const teamsForScope = useMemo(() => {
    const ws = teams.data ?? []
    if (activeDept) {
      return ws.filter(
        (t) =>
          t.ownerType === OwnerType.DEPARTMENTS && t.ownerId === activeDept,
      )
    }
    if (activeBranch) {
      return ws.filter(
        (t) =>
          (t.ownerType === OwnerType.BRANCHES && t.ownerId === activeBranch) ||
          departmentsInBranch.some((d) => d.id === t.ownerId),
      )
    }
    return ws
  }, [teams.data, activeDept, activeBranch, departmentsInBranch])

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Organisation"
        description="Branches, departments and teams"
      />
      {branches.isError ? (
        <ErrorState
          title="Couldn't load organisation"
          description={(branches.error as Error).message}
          onRetry={() => void branches.refetch()}
        />
      ) : (
        <div className="flex-1 min-h-0 grid grid-cols-3 gap-0 border-t border-border">
          <Column
            icon={<IconBuilding className="size-3.5" />}
            title="Branches"
            count={branches.data?.length ?? 0}
            isPending={branches.isPending}
            isEmpty={(branches.data ?? []).length === 0}
            emptyLabel="No branches yet."
            create={<CreateBranchInput />}
          >
            {(branches.data ?? []).map((b) => (
              <TreeRow
                key={b.id}
                label={b.name}
                meta={b.description}
                selected={activeBranch === b.id}
                onSelect={() => {
                  setActiveBranch(b.id === activeBranch ? null : b.id)
                  setActiveDept(null)
                }}
                onDelete={() => void deleteBranch(b.id, branches.refetch)}
              />
            ))}
          </Column>

          <Column
            icon={<IconBuildingCommunity className="size-3.5" />}
            title="Departments"
            count={departmentsInBranch.length}
            scopeLabel={activeBranch ? "in this branch" : "across org"}
            isPending={departments.isPending}
            isEmpty={departmentsInBranch.length === 0}
            emptyLabel="No departments here."
            create={
              <CreateDepartmentInput
                branchId={activeBranch}
              />
            }
          >
            {departmentsInBranch.map((d) => (
              <TreeRow
                key={d.id}
                label={d.departmentName}
                meta={d.departmentBio}
                selected={activeDept === d.id}
                onSelect={() =>
                  setActiveDept(d.id === activeDept ? null : d.id)
                }
                onDelete={() => void deleteDept(d.id, departments.refetch)}
              />
            ))}
          </Column>

          <Column
            icon={<IconUsers className="size-3.5" />}
            title="Teams"
            count={teamsForScope.length}
            scopeLabel={
              activeDept
                ? "in this department"
                : activeBranch
                  ? "in this branch"
                  : "across org"
            }
            isPending={teams.isPending}
            isEmpty={teamsForScope.length === 0}
            emptyLabel="No teams here."
            create={
              <CreateTeamInput
                ownerType={
                  activeDept
                    ? OwnerType.DEPARTMENTS
                    : activeBranch
                      ? OwnerType.BRANCHES
                      : OwnerType.ORGANISATIONS
                }
                ownerId={activeDept ?? activeBranch ?? null}
              />
            }
          >
            {teamsForScope.map((t) => (
              <TreeRow
                key={t.id}
                label={t.name}
                meta={t.description}
                onDelete={() => void deleteTeam(t.id, teams.refetch)}
              />
            ))}
          </Column>
        </div>
      )}
    </div>
  )
}

function deleteBranch(id: string, refetch: () => unknown) {
  if (!confirm("Delete this branch and its departments?")) return
  void (async () => {
    const { deleteBranch } = await import("../api/organisation")
    await deleteBranch(id)
    refetch()
  })()
}

function deleteDept(id: string, refetch: () => unknown) {
  if (!confirm("Delete this department?")) return
  void (async () => {
    const { deleteDepartment } = await import("../api/organisation")
    await deleteDepartment(id)
    refetch()
  })()
}

function deleteTeam(id: string, refetch: () => unknown) {
  if (!confirm("Delete this team?")) return
  void (async () => {
    const { deleteTeam } = await import("../api/organisation")
    await deleteTeam(id)
    refetch()
  })()
}

// ── Reusable column ──────────────────────────────────────────────────────────

function Column({
  icon,
  title,
  count,
  scopeLabel,
  isPending,
  isEmpty,
  emptyLabel,
  create,
  children,
}: {
  icon: React.ReactNode
  title: string
  count: number
  scopeLabel?: string
  isPending: boolean
  isEmpty: boolean
  emptyLabel: string
  create: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="flex flex-col min-h-0 border-e border-border last:border-e-0">
      <header className="flex items-center gap-2 px-3 h-9 border-b border-border">
        <span className="text-fg-muted" aria-hidden>{icon}</span>
        <span className="text-md font-medium text-fg flex-1">{title}</span>
        <span className="font-mono text-xs text-fg-muted">
          {count}
          {scopeLabel && <span className="ms-1 text-fg-disabled">· {scopeLabel}</span>}
        </span>
      </header>
      <div className="border-b border-border-subtle p-2">{create}</div>
      {isPending ? (
        <div className="p-2">
          <CardSkeleton lines={1} />
        </div>
      ) : isEmpty ? (
        <EmptyState
          title={emptyLabel}
          description="Create one with the input above."
          size="sm"
        />
      ) : (
        <div className="flex-1 overflow-y-auto p-1 space-y-px" role="list">
          {children}
        </div>
      )}
    </section>
  )
}

function TreeRow({
  label,
  meta,
  selected,
  onSelect,
  onDelete,
}: {
  label: string
  meta?: string
  selected?: boolean
  onSelect?: () => void
  onDelete?: () => void
}) {
  return (
    <div
      role="listitem"
      className={cn(
        "group flex items-center gap-2 px-2 h-9 rounded-sm",
        "hover:bg-subtle transition-colors",
        selected && "bg-accent-subtle",
      )}
    >
      <button
        type="button"
        onClick={onSelect}
        className="flex items-center gap-2 min-w-0 flex-1 text-start"
      >
        {onSelect && (
          <IconChevronRight
            className={cn(
              "size-3 shrink-0 text-fg-muted transition-transform",
              selected && "rotate-90 text-accent",
            )}
          />
        )}
        <div className="min-w-0">
          <div className="text-md text-fg truncate">{label || "(unnamed)"}</div>
          {meta && <div className="text-sm text-fg-muted truncate">{meta}</div>}
        </div>
      </button>
      {onDelete && (
        <button
          type="button"
          onClick={onDelete}
          aria-label="Delete"
          className={cn(
            "opacity-0 group-hover:opacity-100",
            "inline-flex items-center justify-center size-6 rounded-xs",
            "text-fg-muted hover:text-danger hover:bg-canvas",
          )}
        >
          <IconTrash className="size-3.5" />
        </button>
      )}
    </div>
  )
}

// ── Creators ─────────────────────────────────────────────────────────────────

function CreateBranchInput() {
  const [name, setName] = useState("")
  const create = useCreateBranch()
  const submit = async () => {
    const value = name.trim()
    if (!value) return
    await create.mutateAsync({ name: value })
    setName("")
  }
  return (
    <CreatorRow
      placeholder="New branch name"
      value={name}
      onChange={setName}
      onSubmit={submit}
      pending={create.isPending}
    />
  )
}

function CreateDepartmentInput({ branchId }: { branchId: string | null }) {
  const [name, setName] = useState("")
  const create = useCreateDepartment()
  const submit = async () => {
    const value = name.trim()
    if (!value) return
    await create.mutateAsync({
      departmentName: value,
      ownerType: branchId ? OwnerType.BRANCHES : OwnerType.ORGANISATIONS,
      ownerId: branchId ?? undefined,
    })
    setName("")
  }
  return (
    <CreatorRow
      placeholder={branchId ? "New dept in this branch" : "New department"}
      value={name}
      onChange={setName}
      onSubmit={submit}
      pending={create.isPending}
    />
  )
}

function CreateTeamInput({
  ownerType,
  ownerId,
}: {
  ownerType: OwnerType
  ownerId: string | null
}) {
  const [name, setName] = useState("")
  const create = useCreateTeam()
  const submit = async () => {
    const value = name.trim()
    if (!value) return
    await create.mutateAsync({
      name: value,
      ownerType,
      ownerId: ownerId ?? undefined,
    })
    setName("")
  }
  return (
    <CreatorRow
      placeholder="New team"
      value={name}
      onChange={setName}
      onSubmit={submit}
      pending={create.isPending}
    />
  )
}

function CreatorRow({
  placeholder,
  value,
  onChange,
  onSubmit,
  pending,
}: {
  placeholder: string
  value: string
  onChange: (v: string) => void
  onSubmit: () => void | Promise<void>
  pending: boolean
}) {
  return (
    <div className="flex items-center gap-1">
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") void onSubmit()
        }}
        placeholder={placeholder}
        disabled={pending}
        className="h-7 text-base"
      />
      <button
        type="button"
        onClick={() => void onSubmit()}
        disabled={pending || !value.trim()}
        className={cn(
          "inline-flex items-center justify-center size-7 rounded-sm",
          "bg-accent text-accent-fg hover:bg-accent-hover",
          "disabled:opacity-40 disabled:pointer-events-none",
        )}
        aria-label="Create"
      >
        <IconPlus className="size-3.5" />
      </button>
    </div>
  )
}

// Re-export shape used to delete via async-imported helper above.
void useDeleteBranch
void useDeleteDepartment
void useDeleteTeam
