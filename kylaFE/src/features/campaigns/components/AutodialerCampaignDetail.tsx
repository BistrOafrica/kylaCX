import { Link } from "react-router-dom"
import {
  IconArrowLeft,
  IconPhoneCall,
  IconUsers,
  IconCalendar,
} from "@tabler/icons-react"
import {
  CardSkeleton,
  ErrorState,
  Surface,
} from "@/design-system"
import { useAutodialerCampaign } from "../hooks/queries"
import { parseContacts } from "../utils/status"

export function AutodialerCampaignDetail({ id }: { id: string }) {
  const campaign = useAutodialerCampaign(id)

  if (campaign.isPending) {
    return (
      <div className="p-6">
        <CardSkeleton lines={4} />
      </div>
    )
  }

  if (campaign.isError || !campaign.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load campaign"
          description={(campaign.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void campaign.refetch()}
        />
      </div>
    )
  }

  const c = campaign.data
  const recipients = parseContacts(c.contacts)

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <header className="flex items-start gap-3 px-6 py-3 border-b border-border">
        <Link
          to="/campaigns"
          aria-label="Back"
          className="inline-flex items-center justify-center size-7 rounded-sm text-fg-secondary hover:text-fg hover:bg-subtle mt-0.5"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className="size-10 rounded-md bg-channel-voice/20 text-channel-voice flex items-center justify-center shrink-0"
          aria-hidden
        >
          <IconPhoneCall className="size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
            {c.name || "(untitled campaign)"}
          </h1>
          <div className="text-base text-fg-muted">
            Voice · {recipients.length.toLocaleString()} recipients
          </div>
        </div>
      </header>

      <div className="p-6 max-w-3xl space-y-4">
        <Surface level={1} radius="md" className="p-4 space-y-2.5">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted flex items-center gap-1.5">
            <IconCalendar className="size-3" />
            Schedule
          </h2>
          <KV label="Start" value={c.startDate || "—"} mono />
          <KV label="End" value={c.endDate || "—"} mono />
          <KV label="Frequency" value={c.frequency || "once"} />
          <KV label="Target type" value={c.targetType || "—"} />
          <KV
            label="Target"
            value={c.target ? c.target.slice(0, 16) + "…" : "—"}
            mono
          />
        </Surface>

        <Surface level={1} radius="md" className="p-4 space-y-2">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted flex items-center gap-1.5">
            <IconUsers className="size-3" />
            Recipients
          </h2>
          <pre className="text-xs font-mono whitespace-pre-wrap break-words rounded-sm bg-canvas border border-border px-3 py-2 max-h-64 overflow-y-auto text-fg-secondary">
            {recipients.slice(0, 100).join("\n") || "—"}
          </pre>
          {recipients.length > 100 && (
            <div className="text-sm text-fg-muted">
              … and {(recipients.length - 100).toLocaleString()} more
            </div>
          )}
        </Surface>

        <Surface level={1} radius="md" className="p-4">
          <p className="text-base text-fg-muted">
            The autodialer proto exposes CRUD but no live progress or
            analytics RPCs yet. Per-call performance shows up on{" "}
            <Link to="/analytics/calls" className="text-fg-link hover:underline">
              /analytics/calls
            </Link>
            ; campaign-level rollups land once a backend ListProgress RPC ships.
          </p>
        </Surface>
      </div>
    </div>
  )
}

function KV({
  label,
  value,
  mono,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className="flex items-center gap-2">
      <dt className="w-24 text-sm text-fg-muted">{label}</dt>
      <dd className={mono ? "font-mono text-sm text-fg-secondary" : "text-base text-fg"}>
        {value}
      </dd>
    </div>
  )
}
