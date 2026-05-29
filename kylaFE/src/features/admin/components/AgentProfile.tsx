import { useAuthStore } from "@/lib/auth"
import {
  PageHeader,
  Surface,
  EmptyState,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useUser } from "../hooks/queries"
import { IconUser, IconShieldCheck, IconKey } from "@tabler/icons-react"

/**
 * AgentProfile — the agent's own settings page.
 *
 * Read-only in F5; F5.x adds the avatar uploader, email-signature
 * editor, MFA enrolment dialog, and passkey manager once the backend
 * exposes the corresponding endpoints in a more agent-friendly shape.
 */
export function AgentProfile() {
  const identity = useAuthStore((s) => s.identity)
  const user = useUser(identity?.userId ?? null)

  if (!identity) {
    return (
      <EmptyState
        title="Sign in to view your profile"
        description="You need to be authenticated."
      />
    )
  }

  const u = user.data
  const name = u
    ? [u.firstName, u.lastName].filter(Boolean).join(" ") || u.email
    : identity.email ?? identity.userId

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <PageHeader title="Profile" description="Your account and security" />

      <div className="p-6 max-w-3xl space-y-4">
        <Surface level={1} radius="md" className="p-4">
          <div className="flex items-start gap-4">
            <div
              className={cn(
                "size-14 rounded-md bg-accent-subtle text-fg flex items-center justify-center",
                "font-medium text-xl",
              )}
              aria-hidden
            >
              {(name?.[0] ?? "?").toUpperCase()}
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-lg font-semibold text-fg truncate">{name}</div>
              <div className="text-base text-fg-muted truncate font-mono">
                {u?.email ?? identity.email ?? ""}
              </div>
              {u?.phone && (
                <div className="text-sm text-fg-muted font-mono">{u.phone}</div>
              )}
            </div>
            <span
              className={cn(
                "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
                u?.status === "active"
                  ? "bg-success-subtle text-success"
                  : "bg-subtle text-fg-muted",
              )}
            >
              {u?.status ?? "—"}
            </span>
          </div>
        </Surface>

        <Surface level={1} radius="md" className="p-4 space-y-3">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            Security
          </h2>
          <SecurityRow
            icon={<IconShieldCheck className="size-4" />}
            label="Multi-factor authentication"
            value={u?.mfaEnabled ? "Enabled" : "Disabled"}
            tone={u?.mfaEnabled ? "success" : "muted"}
          />
          <SecurityRow
            icon={<IconKey className="size-4" />}
            label="Passkeys"
            value={
              u?.passkeys?.length
                ? `${u.passkeys.length} registered`
                : "None"
            }
            tone={u?.passkeys?.length ? "success" : "muted"}
          />
          <SecurityRow
            icon={<IconUser className="size-4" />}
            label="User ID"
            value={identity.userId.slice(0, 24) + "…"}
            mono
          />
        </Surface>
      </div>
    </div>
  )
}

function SecurityRow({
  icon,
  label,
  value,
  tone = "muted",
  mono,
}: {
  icon: React.ReactNode
  label: string
  value: string
  tone?: "success" | "muted"
  mono?: boolean
}) {
  return (
    <div className="flex items-center gap-3">
      <span className="text-fg-muted" aria-hidden>{icon}</span>
      <span className="text-md text-fg flex-1">{label}</span>
      <span
        className={cn(
          "text-sm",
          mono && "font-mono",
          tone === "success" ? "text-success" : "text-fg-muted",
        )}
      >
        {value}
      </span>
    </div>
  )
}
