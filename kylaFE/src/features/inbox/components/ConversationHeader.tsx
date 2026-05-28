import { useTranslation } from "react-i18next"
import {
  IconUserPlus,
  IconCircleCheck,
  IconAlarm,
  IconTag,
} from "@tabler/icons-react"
import { ChannelBadge, Kbd } from "@/design-system"
import { cn } from "@/lib/utils"
import {
  channelMeta,
  statusMeta,
  ConversationStatus,
} from "../utils/enums"
import { SlaBadge } from "./SlaBadge"
import {
  useResolveConversation,
  useUpdateStatus,
} from "../hooks/queries"
import type { Conversation } from "@/pb/conversations"

/**
 * ConversationHeader — channel badge + subject + actions.
 *
 *   [WA] Maria Khalil — "Order not arrived"   [Open] [SLA 28m] [Assign▾] [Snooze▾] [Resolve]
 *
 * Actions are scoped to permissions in F5; F1 wires Resolve + Snooze
 * via the existing UpdateConversationStatus RPC.
 */
export function ConversationHeader({ conv }: { conv: Conversation }) {
  const { t } = useTranslation()
  const ch = channelMeta(conv.channel)
  const status = statusMeta(conv.status)
  const resolve = useResolveConversation()
  const updateStatus = useUpdateStatus()

  const onResolve = () =>
    resolve.mutate({ conversationId: conv.id, reason: "" })

  const onSnooze = () => {
    const until = new Date(Date.now() + 4 * 60 * 60 * 1000).toISOString()
    updateStatus.mutate({
      conversationId: conv.id,
      status: ConversationStatus.SNOOZED,
      snoozedUntil: until,
    })
  }

  return (
    <header className="flex items-center gap-2 h-11 shrink-0 px-3 border-b border-border bg-canvas">
      <ChannelBadge channel={ch.token} />
      <div className="min-w-0 flex-1 flex items-center gap-2">
        <h1 className="text-md font-semibold text-fg truncate">
          {conv.subject || "(no subject)"}
        </h1>
        <StatusChip tone={status.tone} label={status.label} />
        <SlaBadge deadline={conv.slaDeadline ?? null} />
      </div>
      <div className="flex items-center gap-1">
        <HeaderButton icon={<IconUserPlus className="size-3.5" />} label={t("common.more")}>
          Assign
        </HeaderButton>
        <HeaderButton icon={<IconTag className="size-3.5" />} label="Label">
          Label
        </HeaderButton>
        <HeaderButton
          icon={<IconAlarm className="size-3.5" />}
          label="Snooze"
          onClick={onSnooze}
          shortcut={["S"]}
          disabled={updateStatus.isPending}
        >
          Snooze
        </HeaderButton>
        <button
          type="button"
          onClick={onResolve}
          disabled={
            resolve.isPending || conv.status === ConversationStatus.RESOLVED
          }
          className={cn(
            "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md font-medium",
            "bg-success text-success-fg hover:opacity-90",
            "disabled:opacity-40 disabled:pointer-events-none transition-opacity",
          )}
        >
          <IconCircleCheck className="size-3.5" />
          Resolve
          <Kbd keys={["E"]} size="sm" />
        </button>
      </div>
    </header>
  )
}

function HeaderButton({
  icon,
  label,
  children,
  onClick,
  shortcut,
  disabled,
}: {
  icon: React.ReactNode
  label: string
  children: React.ReactNode
  onClick?: () => void
  shortcut?: string[]
  disabled?: boolean
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={shortcut ? `${label}  ${shortcut.join("+")}` : label}
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "inline-flex items-center gap-1.5 h-7 px-2 rounded-sm text-md",
        "text-fg-secondary hover:text-fg hover:bg-subtle",
        "disabled:opacity-40 disabled:pointer-events-none transition-colors",
      )}
    >
      {icon}
      <span className="hidden md:inline">{children}</span>
    </button>
  )
}

function StatusChip({ tone, label }: { tone: "info" | "warn" | "success" | "muted"; label: string }) {
  const cls =
    tone === "info"
      ? "bg-info-subtle text-info"
      : tone === "warn"
        ? "bg-warn-subtle text-warn"
        : tone === "success"
          ? "bg-success-subtle text-success"
          : "bg-subtle text-fg-muted"
  return (
    <span className={cn("inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium", cls)}>
      {label}
    </span>
  )
}
