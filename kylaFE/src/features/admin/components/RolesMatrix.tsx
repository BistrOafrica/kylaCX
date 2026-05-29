import { useMemo, useState } from "react"
import {
  IconShieldLock,
  IconPlus,
  IconTrash,
} from "@tabler/icons-react"
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
  useRoles,
  usePermissions,
  useCreateRole,
  useDeleteRole,
  useTogglePermission,
} from "../hooks/queries"
import type { Role, Permission } from "../api/rbac"

/**
 * RolesMatrix — left column lists roles; right grid is a permission
 * matrix keyed by service. Toggling a cell calls
 * Add/RemovePermissionFromRole.
 */
export function RolesMatrix() {
  const roles = useRoles()
  const perms = usePermissions()
  const create = useCreateRole()
  const remove = useDeleteRole()
  const toggle = useTogglePermission()
  const [activeId, setActiveId] = useState<string | null>(null)
  const [newRoleName, setNewRoleName] = useState("")

  const activeRole = useMemo<Role | null>(
    () => roles.data?.find((r) => r.id === activeId) ?? roles.data?.[0] ?? null,
    [roles.data, activeId],
  )

  const permsByService = useMemo(() => {
    const map = new Map<string, Permission[]>()
    for (const p of perms.data ?? []) {
      const arr = map.get(p.service || "general") ?? []
      arr.push(p)
      map.set(p.service || "general", arr)
    }
    return [...map.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [perms.data])

  const granted = useMemo(
    () => new Set(activeRole?.permissions ?? []),
    [activeRole],
  )

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader title="Roles" description="Permissions per role" />

      {roles.isError ? (
        <ErrorState
          title="Couldn't load roles"
          description={(roles.error as Error).message}
          onRetry={() => void roles.refetch()}
        />
      ) : (
        <div className="flex-1 min-h-0 grid grid-cols-12 gap-0 border-t border-border">
          <aside className="col-span-3 border-e border-border flex flex-col">
            <div className="p-2 border-b border-border-subtle space-y-1.5">
              <Input
                value={newRoleName}
                onChange={(e) => setNewRoleName(e.target.value)}
                placeholder="New role name"
                className="h-7 text-base"
              />
              <button
                type="button"
                disabled={create.isPending || !newRoleName.trim()}
                onClick={async () => {
                  const r = await create.mutateAsync({ name: newRoleName.trim() })
                  if (r) setActiveId(r.id)
                  setNewRoleName("")
                }}
                className={cn(
                  "w-full inline-flex items-center justify-center gap-1.5 h-7 px-2 rounded-sm text-md font-medium",
                  "bg-accent text-accent-fg hover:bg-accent-hover disabled:opacity-40",
                )}
              >
                <IconPlus className="size-3.5" />
                Create role
              </button>
            </div>
            {roles.isPending ? (
              <div className="p-2">
                <ListRowSkeleton count={4} />
              </div>
            ) : (
              <ul className="flex-1 overflow-y-auto p-1 space-y-px" role="list">
                {roles.data!.map((r) => (
                  <li key={r.id}>
                    <button
                      type="button"
                      onClick={() => setActiveId(r.id)}
                      className={cn(
                        "w-full group flex items-center gap-2 px-2 h-8 rounded-sm text-start",
                        "hover:bg-subtle transition-colors text-md text-fg",
                        activeRole?.id === r.id && "bg-accent-subtle font-medium",
                      )}
                    >
                      <IconShieldLock className="size-3.5 text-fg-muted" />
                      <span className="truncate flex-1">{r.name}</span>
                      {!r.isDefault && (
                        <span
                          role="button"
                          tabIndex={0}
                          aria-label="Delete"
                          onClick={(e) => {
                            e.stopPropagation()
                            if (confirm(`Delete role "${r.name}"?`))
                              remove.mutate(r.id)
                          }}
                          className="opacity-0 group-hover:opacity-100 text-fg-muted hover:text-danger"
                        >
                          <IconTrash className="size-3" />
                        </span>
                      )}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </aside>

          <main className="col-span-9 overflow-y-auto">
            {!activeRole ? (
              <EmptyState
                icon={<IconShieldLock className="size-5" />}
                title="Pick a role"
                description="Select a role on the left to edit its permissions."
              />
            ) : perms.isPending ? (
              <div className="p-4">
                <ListRowSkeleton count={8} />
              </div>
            ) : permsByService.length === 0 ? (
              <EmptyState
                title="No permissions"
                description="The backend hasn't published any permission codenames yet."
              />
            ) : (
              <div className="p-4 space-y-4">
                <Surface level={0} radius="md" className="p-3">
                  <div className="text-md font-medium text-fg">
                    {activeRole.name}
                  </div>
                  <div className="text-sm text-fg-muted">
                    {activeRole.description || "No description"}
                  </div>
                </Surface>
                {permsByService.map(([service, list]) => (
                  <Surface key={service} level={1} radius="md" className="p-3">
                    <h3 className="text-xs font-mono uppercase tracking-wider text-fg-muted mb-2">
                      {service}
                    </h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
                      {list.map((p) => {
                        const on = granted.has(p.codeName)
                        return (
                          <label
                            key={p.codeName}
                            className={cn(
                              "flex items-start gap-2 px-2 py-1.5 rounded-sm cursor-pointer",
                              "hover:bg-subtle transition-colors",
                            )}
                          >
                            <input
                              type="checkbox"
                              checked={on}
                              disabled={toggle.isPending}
                              onChange={() =>
                                toggle.mutate({
                                  roleId: activeRole.id,
                                  codename: p.codeName,
                                  on: !on,
                                })
                              }
                              className="accent-emerald-500 mt-0.5"
                            />
                            <span className="min-w-0">
                              <span className="block text-md text-fg truncate">
                                {p.name || p.codeName}
                              </span>
                              {p.description && (
                                <span className="block text-sm text-fg-muted truncate">
                                  {p.description}
                                </span>
                              )}
                              <span className="block text-xs font-mono text-fg-disabled truncate">
                                {p.codeName}
                              </span>
                            </span>
                          </label>
                        )
                      })}
                    </div>
                  </Surface>
                ))}
              </div>
            )}
          </main>
        </div>
      )}
    </div>
  )
}
