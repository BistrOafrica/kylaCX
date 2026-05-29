import { useState } from "react"
import { Link } from "react-router-dom"
import {
  IconArrowLeft,
  IconBrandWhatsapp,
  IconUsers,
  IconCalendar,
  IconClock,
  IconDeviceFloppy,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  CardSkeleton,
  ErrorState,
  Surface,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu"
import {
  useWhatsappCampaign,
  useWhatsappTemplates,
} from "../hooks/queries"
import {
  bucketCampaignStatus,
  STATUS_LABEL,
  STATUS_TONE,
  parseContacts,
} from "../utils/status"
import { CampaignAnalytics } from "./CampaignAnalytics"

const TONE_CLS = {
  muted:   "bg-subtle text-fg-muted",
  info:    "bg-info-subtle text-info",
  success: "bg-success-subtle text-success",
  warn:    "bg-warn-subtle text-warn",
  danger:  "bg-danger-subtle text-danger",
}

/**
 * WhatsappCampaignDetail — composer + analytics in two tabs.
 *
 * The composer side reads + edits campaign fields locally; saving
 * fires a backend Update (TODO when proto adds Update RPC — today the
 * proto only exposes Create/View/List for campaigns, so the
 * inline-editor surfaces a "Read-only after creation" notice for now).
 */
export function WhatsappCampaignDetail({ id }: { id: string }) {
  const campaign = useWhatsappCampaign(id)
  const templates = useWhatsappTemplates()
  const [tab, setTab] = useState<"composer" | "analytics">("composer")

  if (campaign.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
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
  const bucket = bucketCampaignStatus(c.status)
  const template = templates.data?.find((t) => t.id === c.templateId)
  const recipients = parseContacts(c.contacts).length

  return (
    <div className="flex flex-col h-full bg-canvas">
      <header className="flex items-start gap-3 px-6 py-3 border-b border-border">
        <Link
          to="/campaigns"
          aria-label="Back"
          className="inline-flex items-center justify-center size-7 rounded-sm text-fg-secondary hover:text-fg hover:bg-subtle mt-0.5"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div
          className="size-10 rounded-md bg-channel-whatsapp/15 text-channel-whatsapp flex items-center justify-center shrink-0"
          aria-hidden
        >
          <IconBrandWhatsapp className="size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold tracking-tight text-fg truncate">
            {c.name || "(untitled campaign)"}
          </h1>
          <div className="text-base text-fg-muted">
            {recipients.toLocaleString()} recipients · WhatsApp
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
      </header>

      <nav
        role="tablist"
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
      >
        <TabBtn active={tab === "composer"} onClick={() => setTab("composer")} label="Composer" />
        <TabBtn active={tab === "analytics"} onClick={() => setTab("analytics")} label="Analytics" />
      </nav>

      <div className="flex-1 min-h-0 overflow-y-auto p-6 max-w-3xl">
        {tab === "composer" && (
          <Composer
            name={c.name}
            description={c.description}
            templateId={c.templateId}
            contacts={c.contacts}
            startDate={c.startDate}
            endDate={c.endDate}
            frequency={c.frequency}
            templates={templates.data ?? []}
            templateName={template?.name}
            templateBody={template?.body}
          />
        )}
        {tab === "analytics" && <CampaignAnalytics campaignId={id} />}
      </div>
    </div>
  )
}

function TabBtn({
  active,
  onClick,
  label,
}: {
  active: boolean
  onClick: () => void
  label: string
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "inline-flex items-center h-7 px-2.5 rounded-sm text-md transition-colors",
        "text-fg-secondary hover:text-fg hover:bg-subtle",
        active && "bg-accent-subtle text-fg font-medium",
      )}
    >
      {label}
    </button>
  )
}

function Composer({
  name,
  description,
  templateId,
  contacts,
  startDate,
  endDate,
  frequency,
  templates,
  templateName,
  templateBody,
}: {
  name: string
  description: string
  templateId: string
  contacts: string
  startDate: string
  endDate: string
  frequency: string
  templates: { id: string; name: string; body: string }[]
  templateName?: string
  templateBody?: string
}) {
  const [nameValue, setName] = useState(name)
  const [descriptionValue, setDescription] = useState(description)
  const [templateIdValue, setTemplateId] = useState(templateId)
  const [contactsValue, setContacts] = useState(contacts)
  const [startDateValue, setStartDate] = useState(startDate || "")
  const [endDateValue, setEndDate] = useState(endDate || "")
  const [frequencyValue, setFrequency] = useState(frequency || "once")

  const recipientCount = parseContacts(contactsValue).length
  const activeTemplate =
    templates.find((t) => t.id === templateIdValue) ??
    (templateName ? { id: templateId, name: templateName, body: templateBody ?? "" } : null)

  return (
    <div className="space-y-4">
      <Surface level={1} radius="md" className="p-4 space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          Campaign
        </h2>
        <Field label="Name">
          <Input
            value={nameValue}
            onChange={(e) => setName(e.target.value)}
            className="h-8"
          />
        </Field>
        <Field label="Description">
          <Input
            value={descriptionValue}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional summary"
            className="h-8"
          />
        </Field>
      </Surface>

      <Surface level={1} radius="md" className="p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            Template
          </h2>
          <Link
            to="/campaigns/templates"
            className="text-sm text-fg-muted hover:text-fg hover:underline"
          >
            Manage templates →
          </Link>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger
            className={cn(
              "w-full inline-flex items-center justify-between h-9 px-3 rounded-sm",
              "border border-border bg-surface hover:bg-subtle text-base text-fg",
            )}
          >
            {activeTemplate ? activeTemplate.name : "Pick a template…"}
            <span className="text-fg-muted">▾</span>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-72">
            {templates.length === 0 ? (
              <DropdownMenuItem disabled>No templates yet</DropdownMenuItem>
            ) : (
              templates.map((t) => (
                <DropdownMenuItem
                  key={t.id}
                  onClick={() => setTemplateId(t.id)}
                  className="flex flex-col items-start gap-0.5"
                >
                  <span className="text-md text-fg">{t.name}</span>
                  <span className="text-sm text-fg-muted line-clamp-1">
                    {t.body}
                  </span>
                </DropdownMenuItem>
              ))
            )}
          </DropdownMenuContent>
        </DropdownMenu>
        {activeTemplate?.body && (
          <pre className="text-base whitespace-pre-wrap rounded-sm border border-border bg-canvas px-3 py-2 text-fg">
            {activeTemplate.body}
          </pre>
        )}
      </Surface>

      <Surface level={1} radius="md" className="p-4 space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted flex items-center gap-1.5">
          <IconUsers className="size-3" />
          Audience
        </h2>
        <textarea
          value={contactsValue}
          onChange={(e) => setContacts(e.target.value)}
          rows={5}
          placeholder="Phone numbers or contact IDs, comma- or line-separated"
          className="w-full font-mono text-sm rounded-sm border border-border bg-canvas px-3 py-2 outline-none focus:border-accent"
        />
        <div className="text-sm text-fg-muted">
          {recipientCount.toLocaleString()} recipient{recipientCount === 1 ? "" : "s"} parsed
        </div>
      </Surface>

      <Surface level={1} radius="md" className="p-4 space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-fg-muted flex items-center gap-1.5">
          <IconCalendar className="size-3" />
          Schedule
        </h2>
        <div className="grid grid-cols-2 gap-3">
          <Field label="Start">
            <Input
              type="datetime-local"
              value={isoToInput(startDateValue)}
              onChange={(e) => setStartDate(inputToIso(e.target.value))}
              className="h-8"
            />
          </Field>
          <Field label="End">
            <Input
              type="datetime-local"
              value={isoToInput(endDateValue)}
              onChange={(e) => setEndDate(inputToIso(e.target.value))}
              className="h-8"
            />
          </Field>
        </div>
        <Field label="Frequency">
          <select
            value={frequencyValue}
            onChange={(e) => setFrequency(e.target.value)}
            className="w-full h-8 rounded-sm border border-border bg-surface px-2 text-base"
          >
            <option value="once">Send once</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
          </select>
        </Field>
      </Surface>

      <div className="flex items-center justify-end gap-2 sticky bottom-0 bg-canvas py-2">
        <span className="text-sm text-fg-muted flex items-center gap-1">
          <IconClock className="size-3.5" />
          The proto currently only exposes Create/View/List for campaigns.
          Saving edits ships when the backend lands an UpdateCampaign RPC.
        </span>
        <button
          type="button"
          disabled
          onClick={() => toast.info("UpdateCampaign RPC not in proto yet")}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          <IconDeviceFloppy className="size-3.5" />
          Save
        </button>
      </div>
    </div>
  )
}

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1">
      <label className="block text-xs font-mono uppercase tracking-wider text-fg-muted">
        {label}
      </label>
      {children}
    </div>
  )
}

/** ISO 8601 → "YYYY-MM-DDTHH:mm" for <input type="datetime-local">. */
function isoToInput(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  const pad = (n: number) => n.toString().padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function inputToIso(local: string): string {
  if (!local) return ""
  const d = new Date(local)
  return Number.isNaN(d.getTime()) ? "" : d.toISOString()
}
