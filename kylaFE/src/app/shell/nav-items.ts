import {
  IconInbox,
  IconUsersGroup,
  IconTicket,
  IconBook,
  IconFileText,
  IconBolt,
  IconSpeakerphone,
  IconPhone,
  IconChartBar,
  IconSettings,
  type Icon,
} from "@tabler/icons-react"

/**
 * Top-level navigation registry — the sidebar reads from here.
 *
 * Each entry references an i18n key under "nav.*" and a route path.
 * Items can be marked `disabled` so the link renders dimmed for
 * surfaces not yet implemented (F5+ phases). The phase tag is used by
 * Storybook docs and by the help overlay to explain availability.
 */

export interface NavItem {
  id: string
  i18nKey: `nav.${string}`
  icon: Icon
  href: string
  badgeKey?: string         // optional badge: live count slot
  disabled?: boolean
  phase: "F0" | "F1" | "F2" | "F3" | "F4" | "F5" | "F6" | "F7" | "F8"
}

export const PRIMARY_NAV: NavItem[] = [
  { id: "inbox",      i18nKey: "nav.inbox",      icon: IconInbox,        href: "/inbox",      phase: "F1" },
  { id: "crm",        i18nKey: "nav.crm",        icon: IconUsersGroup,   href: "/crm",        phase: "F2" },
  { id: "tickets",    i18nKey: "nav.tickets",    icon: IconTicket,       href: "/tickets",    phase: "F3" },
  { id: "knowledge",  i18nKey: "nav.knowledge",  icon: IconBook,         href: "/knowledge",  phase: "F3" },
  { id: "forms",      i18nKey: "nav.forms",      icon: IconFileText,     href: "/forms",      phase: "F3" },
  { id: "automation", i18nKey: "nav.automation", icon: IconBolt,         href: "/automation", phase: "F4" },
  { id: "campaigns",  i18nKey: "nav.campaigns",  icon: IconSpeakerphone, href: "/campaigns",  phase: "F8" },
  { id: "calls",      i18nKey: "nav.calls",      icon: IconPhone,        href: "/calls",      phase: "F7" },
  { id: "analytics",  i18nKey: "nav.analytics",  icon: IconChartBar,     href: "/analytics",  phase: "F6" },
]

export const SECONDARY_NAV: NavItem[] = [
  { id: "admin", i18nKey: "nav.admin", icon: IconSettings, href: "/admin", phase: "F5" },
]
