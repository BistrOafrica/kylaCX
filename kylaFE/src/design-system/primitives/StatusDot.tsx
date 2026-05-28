import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * StatusDot — small filled circle used for unread / presence / status
 * indicators in dense rows (inbox, contacts, conversation headers).
 *
 *   <StatusDot tone="unread" />
 *   <StatusDot tone="online" pulse />
 */
export type StatusTone =
  | "unread"
  | "read"
  | "online"
  | "offline"
  | "busy"
  | "away"
  | "success"
  | "warn"
  | "danger"
  | "info"

const TONE: Record<StatusTone, string> = {
  unread:  "bg-accent",
  read:    "bg-transparent border border-border-strong",
  online:  "bg-success",
  offline: "bg-border-strong",
  busy:    "bg-danger",
  away:    "bg-warn",
  success: "bg-success",
  warn:    "bg-warn",
  danger:  "bg-danger",
  info:    "bg-info",
}

export interface StatusDotProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: StatusTone
  size?: 4 | 6 | 8 | 10
  pulse?: boolean
}

export function StatusDot({
  tone = "unread",
  size = 6,
  pulse = false,
  className,
  ...rest
}: StatusDotProps) {
  return (
    <span
      role="presentation"
      className={cn(
        "inline-flex shrink-0 rounded-full",
        TONE[tone],
        size === 4 && "size-1",
        size === 6 && "size-1.5",
        size === 8 && "size-2",
        size === 10 && "size-2.5",
        pulse && "animate-pulse",
        className,
      )}
      {...rest}
    />
  )
}
