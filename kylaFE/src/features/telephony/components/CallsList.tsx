import { Link } from "react-router-dom"
import {
  IconPhoneIncoming,
  IconPhoneOutgoing,
  IconPhone,
  IconPhoneCall,
} from "@tabler/icons-react"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useCalls } from "../hooks/queries"
import { CallDirection } from "@/pb/call_session"
import {
  DIRECTION_LABEL,
  STATUS_LABEL,
  formatDuration,
  statusTone,
} from "../utils/formatters"
import { useSoftphoneStore } from "../store/softphone"
import type { CallLog } from "@/pb/call_log"

const TONE_CLS = {
  success: "bg-success-subtle text-success",
  warn:    "bg-warn-subtle text-warn",
  danger:  "bg-danger-subtle text-danger",
  info:    "bg-info-subtle text-info",
  muted:   "bg-subtle text-fg-muted",
}

export function CallsList() {
  const list = useCalls({})

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Calls"
        description="Recent call history across the workspace"
      />

      {list.isPending ? (
        <div className="p-2">
          <ListRowSkeleton count={8} />
        </div>
      ) : list.isError ? (
        <ErrorState
          title="Couldn't load calls"
          description={(list.error as Error).message}
          onRetry={() => void list.refetch()}
        />
      ) : (list.data?.items.length ?? 0) === 0 ? (
        <EmptyState
          icon={<IconPhone className="size-5" />}
          title="No calls yet"
          description="Place a call from the softphone — it will appear here once it ends."
        />
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {list.data!.items.map((c) => (
            <CallRow key={c.id} call={c} />
          ))}
        </ul>
      )}
    </div>
  )
}

function CallRow({ call }: { call: CallLog }) {
  const requestDial = useSoftphoneStore((s) => s.requestDial)
  const Icon =
    call.direction === CallDirection.INBOUND
      ? IconPhoneIncoming
      : call.direction === CallDirection.OUTBOUND
        ? IconPhoneOutgoing
        : IconPhoneCall
  const otherParty =
    call.direction === CallDirection.OUTBOUND
      ? call.destinationNumber || call.callerIdNumber
      : call.callerIdNumber || call.destinationNumber

  return (
    <li className="flex items-center gap-3 px-3 py-2 rounded-sm hover:bg-subtle transition-colors">
      <div
        className={cn(
          "size-7 rounded-sm flex items-center justify-center shrink-0",
          TONE_CLS[statusTone(call.status)],
        )}
        aria-hidden
      >
        <Icon className="size-3.5" />
      </div>
      <Link
        to={`/calls/${call.id}`}
        className="min-w-0 flex-1 group"
      >
        <div className="text-md font-medium text-fg truncate group-hover:underline">
          {call.callerIdName || otherParty || "(unknown)"}
        </div>
        <div className="text-sm text-fg-muted font-mono truncate">
          {otherParty} · {DIRECTION_LABEL[call.direction] ?? "—"}
        </div>
      </Link>
      <span
        className={cn(
          "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
          TONE_CLS[statusTone(call.status)],
        )}
      >
        {STATUS_LABEL[call.status] ?? "—"}
      </span>
      <span className="font-mono text-xs text-fg-muted w-16 text-end">
        {formatDuration(call.duration)}
      </span>
      <button
        type="button"
        onClick={() => otherParty && requestDial(otherParty, call.callerIdName)}
        disabled={!otherParty}
        aria-label="Call back"
        className="inline-flex items-center justify-center size-7 rounded-xs text-fg-muted hover:text-success hover:bg-canvas disabled:opacity-30"
      >
        <IconPhoneCall className="size-3.5" />
      </button>
    </li>
  )
}
