import {
  Channel,
  ConversationStatus,
  ConversationPriority,
  SenderType,
  MessageStatus,
} from "@/pb/conversations"
import type { Channel as ChannelToken } from "@/design-system"

/**
 * Map between the generated proto enums (numeric) and the
 * design-system token names + i18n keys we use in the UI.
 *
 * Components should never branch on raw `Channel.WHATSAPP` —
 * they consume `channelMeta(c)` and get the colored badge token,
 * i18n key, and ARIA label in one shot.
 */

export interface ChannelMeta {
  token: ChannelToken
  protoEnum: Channel
  label: string         // raw English label, also used as the i18n key suffix
}

export const CHANNEL_META: ChannelMeta[] = [
  { protoEnum: Channel.WHATSAPP,  token: "whatsapp",  label: "WhatsApp"  },
  { protoEnum: Channel.EMAIL,     token: "email",     label: "Email"     },
  { protoEnum: Channel.SMS,       token: "sms",       label: "SMS"       },
  { protoEnum: Channel.VOICE,     token: "voice",     label: "Voice"     },
  { protoEnum: Channel.WEBCHAT,   token: "webchat",   label: "Web Chat"  },
  { protoEnum: Channel.INSTAGRAM, token: "instagram", label: "Instagram" },
  { protoEnum: Channel.MESSENGER, token: "messenger", label: "Messenger" },
]

export function channelMeta(c: Channel): ChannelMeta {
  return CHANNEL_META.find((m) => m.protoEnum === c) ?? CHANNEL_META[0]
}

export interface StatusMeta {
  protoEnum: ConversationStatus
  label: string
  tone: "info" | "warn" | "success" | "muted"
}

export const STATUS_META: StatusMeta[] = [
  { protoEnum: ConversationStatus.OPEN,     label: "Open",     tone: "info"    },
  { protoEnum: ConversationStatus.PENDING,  label: "Pending",  tone: "warn"    },
  { protoEnum: ConversationStatus.RESOLVED, label: "Resolved", tone: "success" },
  { protoEnum: ConversationStatus.SNOOZED,  label: "Snoozed",  tone: "muted"   },
]

export function statusMeta(s: ConversationStatus): StatusMeta {
  return STATUS_META.find((m) => m.protoEnum === s) ?? STATUS_META[0]
}

export interface PriorityMeta {
  protoEnum: ConversationPriority
  label: string
  tone: "muted" | "info" | "warn" | "danger"
  rank: number   // 0 = least urgent
}

export const PRIORITY_META: PriorityMeta[] = [
  { protoEnum: ConversationPriority.LOW,    label: "Low",    tone: "muted",  rank: 0 },
  { protoEnum: ConversationPriority.NORMAL, label: "Normal", tone: "info",   rank: 1 },
  { protoEnum: ConversationPriority.HIGH,   label: "High",   tone: "warn",   rank: 2 },
  { protoEnum: ConversationPriority.URGENT, label: "Urgent", tone: "danger", rank: 3 },
]

export function priorityMeta(p: ConversationPriority): PriorityMeta {
  return PRIORITY_META.find((m) => m.protoEnum === p) ?? PRIORITY_META[1]
}

export { Channel, ConversationStatus, ConversationPriority, SenderType, MessageStatus }
