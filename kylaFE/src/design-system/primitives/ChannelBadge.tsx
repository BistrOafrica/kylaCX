import * as React from "react"
import { cn } from "@/lib/utils"
import type { Channel } from "../tokens/tokens"

/**
 * ChannelBadge — Linear-dense tag showing which communication channel
 * a conversation came in on. Uses the channel-* hues from the palette
 * (palette.css → theme.css → bg-channel-*).
 *
 *   <ChannelBadge channel="whatsapp" />
 *   <ChannelBadge channel="email" label="Support" />
 */
export interface ChannelBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  channel: Channel
  label?: string
  variant?: "solid" | "subtle"
}

const ABBR: Record<Channel, string> = {
  whatsapp:  "WA",
  email:     "✉",
  sms:       "SMS",
  voice:     "☎",
  webchat:   "WC",
  instagram: "IG",
  messenger: "MS",
}

const BG_SOLID: Record<Channel, string> = {
  whatsapp:  "bg-channel-whatsapp text-white",
  email:     "bg-channel-email text-white",
  sms:       "bg-channel-sms text-white",
  voice:     "bg-channel-voice text-zinc-950",
  webchat:   "bg-channel-webchat text-white",
  instagram: "bg-channel-instagram text-white",
  messenger: "bg-channel-messenger text-white",
}

const BG_SUBTLE: Record<Channel, string> = {
  whatsapp:  "bg-channel-whatsapp/15 text-channel-whatsapp",
  email:     "bg-channel-email/15 text-channel-email",
  sms:       "bg-channel-sms/15 text-channel-sms",
  voice:     "bg-channel-voice/20 text-channel-voice",
  webchat:   "bg-channel-webchat/15 text-channel-webchat",
  instagram: "bg-channel-instagram/15 text-channel-instagram",
  messenger: "bg-channel-messenger/15 text-channel-messenger",
}

export function ChannelBadge({
  channel,
  label,
  variant = "subtle",
  className,
  ...rest
}: ChannelBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 px-1.5 h-4 rounded-xs",
        "text-[10px] font-medium uppercase tracking-wider",
        variant === "solid" ? BG_SOLID[channel] : BG_SUBTLE[channel],
        className,
      )}
      aria-label={`Channel: ${channel}`}
      {...rest}
    >
      <span>{ABBR[channel]}</span>
      {label && <span className="normal-case tracking-normal">{label}</span>}
    </span>
  )
}
