import { Link } from "react-router-dom"
import {
  IconArrowLeft,
  IconPhone,
  IconUser,
  IconClock,
  IconHash,
  IconPhoneCall,
} from "@tabler/icons-react"
import {
  CardSkeleton,
  ErrorState,
  Surface,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useCall } from "../hooks/queries"
import {
  DIRECTION_LABEL,
  STATUS_LABEL,
  formatDuration,
  statusTone,
} from "../utils/formatters"
import { CallDirection } from "@/pb/call_session"
import { useSoftphoneStore } from "../store/softphone"

export function CallDetail({ callId }: { callId: string }) {
  const call = useCall(callId)
  const requestDial = useSoftphoneStore((s) => s.requestDial)

  if (call.isPending) {
    return (
      <div className="p-6">
        <CardSkeleton lines={4} />
      </div>
    )
  }

  if (call.isError || !call.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load call"
          description={(call.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void call.refetch()}
        />
      </div>
    )
  }

  const c = call.data
  const otherParty =
    c.direction === CallDirection.OUTBOUND
      ? c.destinationNumber || c.callerIdNumber
      : c.callerIdNumber || c.destinationNumber
  const tone = statusTone(c.status)

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <header className="flex items-start gap-3 px-6 py-3 border-b border-border">
        <Link
          to="/calls"
          aria-label="Back"
          className="inline-flex items-center justify-center size-7 rounded-sm text-fg-secondary hover:text-fg hover:bg-subtle"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className={cn(
            "size-10 rounded-md flex items-center justify-center shrink-0",
            "bg-accent-subtle text-fg-secondary",
          )}
          aria-hidden
        >
          <IconPhone className="size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
            {c.callerIdName || otherParty || "(unknown caller)"}
          </h1>
          <div className="text-base text-fg-muted font-mono">{c.id.slice(0, 16)}…</div>
        </div>
        <button
          type="button"
          disabled={!otherParty}
          onClick={() => otherParty && requestDial(otherParty, c.callerIdName)}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-success text-success-fg hover:opacity-90 disabled:opacity-50",
          )}
        >
          <IconPhoneCall className="size-3.5" />
          Call back
        </button>
      </header>

      <div className="p-6 space-y-4 max-w-3xl">
        <Surface level={1} radius="md" className="p-4">
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2.5">
            <KV
              icon={<IconUser className="size-3.5" />}
              label="From"
              value={`${c.callerIdName || "—"} (${c.callerIdNumber || "—"})`}
            />
            <KV
              icon={<IconUser className="size-3.5" />}
              label="To"
              value={c.destinationNumber || "—"}
            />
            <KV
              icon={<IconHash className="size-3.5" />}
              label="Direction"
              value={DIRECTION_LABEL[c.direction] ?? "—"}
            />
            <KV
              icon={<IconHash className="size-3.5" />}
              label="Status"
              value={
                <span
                  className={cn(
                    "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
                    tone === "success" && "bg-success-subtle text-success",
                    tone === "warn"    && "bg-warn-subtle text-warn",
                    tone === "danger"  && "bg-danger-subtle text-danger",
                    tone === "info"    && "bg-info-subtle text-info",
                    tone === "muted"   && "bg-subtle text-fg-muted",
                  )}
                >
                  {STATUS_LABEL[c.status] ?? "—"}
                </span>
              }
            />
            <KV
              icon={<IconClock className="size-3.5" />}
              label="Duration"
              value={formatDuration(c.duration)}
            />
            <KV
              icon={<IconHash className="size-3.5" />}
              label="Queue"
              value={c.queueId ? c.queueId.slice(0, 12) + "…" : "—"}
              mono
            />
            <KV
              icon={<IconUser className="size-3.5" />}
              label="Agent"
              value={c.agentId ? c.agentId.slice(0, 12) + "…" : "—"}
              mono
            />
            <KV
              icon={<IconHash className="size-3.5" />}
              label="Call ID"
              value={c.callId ? c.callId.slice(0, 12) + "…" : "—"}
              mono
            />
          </dl>
        </Surface>

        <Surface level={1} radius="md" className="p-4 space-y-2">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            Recording & transcript
          </h2>
          <p className="text-base text-fg-muted">
            Recording / transcript playback ships when the audio + transcript
            services land in the proto layer. The CDR UUID is{" "}
            <code className="font-mono text-sm text-fg-secondary">
              {c.cdrUuid || "—"}
            </code>
            .
          </p>
        </Surface>
      </div>
    </div>
  )
}

function KV({
  icon,
  label,
  value,
  mono,
}: {
  icon: React.ReactNode
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="flex items-center gap-1 text-xs font-mono uppercase tracking-wider text-fg-muted">
        <span className="text-fg-muted" aria-hidden>{icon}</span>
        {label}
      </dt>
      <dd className={cn("text-base text-fg", mono && "font-mono")}>{value}</dd>
    </div>
  )
}
