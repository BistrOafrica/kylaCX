import { useMemo, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import {
  IconSpeakerphone,
  IconPlus,
  IconBrandWhatsapp,
  IconPhoneCall,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useWhatsappCampaigns,
  useAutodialerCampaigns,
  useCreateWhatsappCampaign,
  useWhatsappTemplates,
} from "../hooks/queries"
import {
  bucketCampaignStatus,
  STATUS_LABEL,
  STATUS_TONE,
  parseContacts,
} from "../utils/status"
import { relativeShort } from "@/features/crm/utils/relative"
import type { WhatsappCampaign } from "../api/whatsapp"
import type { AutoDialerCampaign } from "../api/autodialer"

const TONE_CLS = {
  muted:   "bg-subtle text-fg-muted",
  info:    "bg-info-subtle text-info",
  success: "bg-success-subtle text-success",
  warn:    "bg-warn-subtle text-warn",
  danger:  "bg-danger-subtle text-danger",
}

type Channel = "whatsapp" | "voice"

/**
 * CampaignsList — unified list across WhatsApp + Auto-dialer campaigns
 * with a channel filter.
 */
export function CampaignsList() {
  const wa = useWhatsappCampaigns()
  const ad = useAutodialerCampaigns()
  const templates = useWhatsappTemplates()
  const navigate = useNavigate()
  const create = useCreateWhatsappCampaign()
  const [channel, setChannel] = useState<Channel | "all">("all")
  const [query, setQuery] = useState("")

  const isPending = wa.isPending || ad.isPending
  const isError = wa.isError || ad.isError
  const error =
    (wa.error as Error | undefined)?.message ??
    (ad.error as Error | undefined)?.message

  const rows = useMemo(() => {
    const xs: Array<{
      kind: Channel
      id: string
      name: string
      status: string
      contacts: string
      startDate: string
      templateId?: string
    }> = []
    if (channel === "all" || channel === "whatsapp") {
      for (const c of wa.data ?? []) {
        xs.push({
          kind: "whatsapp",
          id: c.id,
          name: c.name || "(untitled campaign)",
          status: c.status,
          contacts: c.contacts,
          startDate: c.startDate,
          templateId: c.templateId,
        })
      }
    }
    if (channel === "all" || channel === "voice") {
      for (const c of ad.data ?? []) {
        xs.push({
          kind: "voice",
          id: c.id,
          name: c.name || "(untitled campaign)",
          status: "running",     // autodialer proto has no status field today
          contacts: c.contacts,
          startDate: c.startDate,
        })
      }
    }
    const q = query.trim().toLowerCase()
    return q ? xs.filter((x) => x.name.toLowerCase().includes(q)) : xs
  }, [wa.data, ad.data, channel, query])

  const onNewWhatsapp = async () => {
    const tpl = templates.data?.[0]
    if (!tpl) {
      toast.info("Create at least one approved template before launching a WhatsApp campaign.")
      navigate("/campaigns/templates")
      return
    }
    try {
      const c = await create.mutateAsync({
        name: "Untitled campaign",
        templateId: tpl.id,
        contacts: "",
      })
      if (c?.id) navigate(`/campaigns/whatsapp/${c.id}`)
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Campaigns"
        description="WhatsApp broadcasts and auto-dialer voice campaigns"
        actions={
          <button
            type="button"
            onClick={() => void onNewWhatsapp()}
            className={cn(
              "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
              "bg-accent text-accent-fg hover:bg-accent-hover",
            )}
          >
            <IconPlus className="size-3.5" />
            New campaign
          </button>
        }
      />

      <div className="flex items-center gap-2 px-3 h-9 border-b border-border bg-canvas">
        <Pill active={channel === "all"} onClick={() => setChannel("all")}>
          All ({((wa.data?.length ?? 0) + (ad.data?.length ?? 0)).toLocaleString()})
        </Pill>
        <Pill active={channel === "whatsapp"} onClick={() => setChannel("whatsapp")}>
          <IconBrandWhatsapp className="size-3.5" />
          WhatsApp ({wa.data?.length ?? 0})
        </Pill>
        <Pill active={channel === "voice"} onClick={() => setChannel("voice")}>
          <IconPhoneCall className="size-3.5" />
          Voice ({ad.data?.length ?? 0})
        </Pill>
        <Link
          to="/campaigns/templates"
          className="ms-2 text-sm text-fg-muted hover:text-fg hover:underline"
        >
          Templates
        </Link>
        <div className="ms-auto w-64">
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search campaigns…"
            className="h-7 text-base"
          />
        </div>
      </div>

      {isPending ? (
        <div className="p-2">
          <ListRowSkeleton count={6} />
        </div>
      ) : isError ? (
        <ErrorState
          title="Couldn't load campaigns"
          description={error}
          onRetry={() => {
            void wa.refetch()
            void ad.refetch()
          }}
        />
      ) : rows.length === 0 ? (
        <EmptyState
          icon={<IconSpeakerphone className="size-5" />}
          title={query ? "No matches" : "No campaigns yet"}
          description={
            query
              ? "Try a different search term or clear the filter."
              : "Launch a WhatsApp broadcast or auto-dialer voice campaign."
          }
        />
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {rows.map((row) => (
            <CampaignRow key={`${row.kind}-${row.id}`} row={row} />
          ))}
        </ul>
      )}
    </div>
  )
}

function CampaignRow({
  row,
}: {
  row: {
    kind: Channel
    id: string
    name: string
    status: string
    contacts: string
    startDate: string
  }
}) {
  const bucket = bucketCampaignStatus(row.status)
  const Icon = row.kind === "whatsapp" ? IconBrandWhatsapp : IconPhoneCall
  const href =
    row.kind === "whatsapp"
      ? `/campaigns/whatsapp/${row.id}`
      : `/campaigns/voice/${row.id}`
  const contactCount = parseContacts(row.contacts).length

  return (
    <li>
      <Link
        to={href}
        className="flex items-center gap-3 px-3 py-2 rounded-sm hover:bg-subtle transition-colors"
      >
        <div
          className={cn(
            "size-7 rounded-sm flex items-center justify-center shrink-0",
            row.kind === "whatsapp"
              ? "bg-channel-whatsapp/15 text-channel-whatsapp"
              : "bg-channel-voice/20 text-channel-voice",
          )}
          aria-hidden
        >
          <Icon className="size-3.5" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-md font-medium text-fg truncate">{row.name}</div>
          <div className="text-sm text-fg-muted truncate font-mono">
            {row.kind === "whatsapp" ? "WhatsApp" : "Voice"} ·{" "}
            {contactCount.toLocaleString()} recipient{contactCount === 1 ? "" : "s"}
          </div>
        </div>
        <span
          className={cn(
            "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
            TONE_CLS[STATUS_TONE[bucket]],
          )}
        >
          {STATUS_LABEL[bucket]}
        </span>
        <span className="font-mono text-xs text-fg-muted w-20 text-end">
          {relativeShort(row.startDate)}
        </span>
      </Link>
    </li>
  )
}

function Pill({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "inline-flex items-center gap-1.5 h-6 px-2 rounded-xs text-sm",
        active
          ? "bg-accent-subtle text-fg font-medium"
          : "text-fg-secondary hover:text-fg hover:bg-subtle",
      )}
    >
      {children}
    </button>
  )
}

// Keep the campaign type re-exports used elsewhere.
export type { WhatsappCampaign, AutoDialerCampaign }
