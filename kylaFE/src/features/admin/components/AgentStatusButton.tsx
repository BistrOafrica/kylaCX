import { useTranslation } from "react-i18next"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu"
import { StatusDot, type StatusTone } from "@/design-system"
import { cn } from "@/lib/utils"
import { useAgentStatus, useSetAgentStatus } from "../hooks/queries"
import { StatusType } from "../api/agent-status"

/**
 * AgentStatusButton — drop-in replacement for the placeholder
 * "Available" pill in the StatusBar. Reads the current status from
 * AgentStatusService and writes via CreateStatusChange.
 */
const STATUS_OPTIONS: {
  value: StatusType
  i18nKey: `agentStatus.${string}` | "common.search"
  label: string
  tone: StatusTone
}[] = [
  { value: StatusType.AVAILABLE,   i18nKey: "agentStatus.online",  label: "Available", tone: "online"  },
  { value: StatusType.ONLINE,      i18nKey: "agentStatus.online",  label: "Online",    tone: "online"  },
  { value: StatusType.BUSY,        i18nKey: "agentStatus.busy",    label: "Busy",      tone: "busy"    },
  { value: StatusType.ON_A_BREAK,  i18nKey: "agentStatus.break",   label: "On a break", tone: "away"   },
  { value: StatusType.IN_A_MEETING, i18nKey: "agentStatus.online", label: "In a meeting", tone: "busy" },
  { value: StatusType.IN_A_CALL,   i18nKey: "agentStatus.online",  label: "In a call",  tone: "busy"   },
  { value: StatusType.OFFLINE,     i18nKey: "agentStatus.offline", label: "Offline",   tone: "offline" },
]

export function AgentStatusButton() {
  const { t } = useTranslation()
  const status = useAgentStatus()
  const set = useSetAgentStatus()

  const current =
    STATUS_OPTIONS.find((o) => o.value === status.data?.statusType) ??
    STATUS_OPTIONS[0]!

  // Use the localized label if a key exists for the current option.
  const currentLabel =
    current.i18nKey === "common.search" ? current.label : t(current.i18nKey)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={t("shell.agentStatus")}
        className={cn(
          "inline-flex items-center gap-1.5 h-5 px-1.5 rounded-xs",
          "hover:bg-subtle text-fg-secondary transition-colors",
        )}
      >
        <StatusDot tone={current.tone} size={6} />
        <span className="text-sm">{currentLabel}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent side="top" align="start" className="w-56">
        <DropdownMenuLabel className="text-xs font-mono uppercase tracking-wider text-fg-muted">
          Set status
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuRadioGroup
          value={String(current.value)}
          onValueChange={(v) =>
            set.mutate({ statusType: Number(v) as StatusType })
          }
        >
          {STATUS_OPTIONS.map((opt) => (
            <DropdownMenuRadioItem
              key={opt.value}
              value={String(opt.value)}
              className="flex items-center gap-2"
            >
              <StatusDot tone={opt.tone} size={6} />
              <span>{opt.label}</span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
