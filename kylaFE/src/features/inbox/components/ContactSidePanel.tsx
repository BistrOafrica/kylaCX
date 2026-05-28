import { IconMail, IconPhone, IconBuilding } from "@tabler/icons-react"
import { Surface, CardSkeleton } from "@/design-system"
import { useContact, useConversationTimeline } from "../hooks/queries"
import { relativeTime } from "../utils/format"
import { cn } from "@/lib/utils"
import type { Conversation } from "@/pb/conversations"

/**
 * Right-side context panel. Shows the contact card, related objects,
 * and recent timeline events for the active conversation.
 *
 * F2 adds related deals / tickets via Object Core relations; F1 ships
 * with just the contact + timeline so the layout is locked in.
 */
export function ContactSidePanel({ conv }: { conv: Conversation }) {
  const contact = useContact(conv.contactId || null)
  const timeline = useConversationTimeline(conv.id || null)

  return (
    <aside
      aria-label="Conversation context"
      className="w-80 shrink-0 overflow-y-auto border-s border-border bg-surface"
    >
      <Section title="Contact">
        {contact.isPending ? (
          <CardSkeleton lines={3} />
        ) : !contact.data ? (
          <div className="text-base text-fg-muted">
            {conv.contactId
              ? "Contact not found"
              : "No contact linked to this conversation"}
          </div>
        ) : (
          <ContactCard
            name={fullName(contact.data)}
            email={contact.data.email}
            phone={contact.data.phone}
            company={contact.data.company}
            jobTitle={contact.data.jobTitle}
            isVip={contact.data.isVip}
          />
        )}
      </Section>

      <Section title="Conversation">
        <KeyValue label="ID" value={conv.id.slice(0, 12) + "…"} mono />
        <KeyValue label="Channel ref" value={conv.channelRef || "—"} mono />
        <KeyValue
          label="Assigned to"
          value={conv.assignedTo || "Unassigned"}
          mono={Boolean(conv.assignedTo)}
        />
        <KeyValue label="Team" value={conv.teamId || "—"} mono={Boolean(conv.teamId)} />
        <KeyValue label="Updated" value={relativeTime(conv.updatedAt)} />
        <KeyValue label="Created" value={relativeTime(conv.createdAt)} />
      </Section>

      <Section title="Timeline">
        {timeline.isPending ? (
          <div className="text-base text-fg-muted">Loading…</div>
        ) : !timeline.data?.length ? (
          <div className="text-base text-fg-muted">No events yet.</div>
        ) : (
          <ol className="space-y-2">
            {timeline.data.slice(0, 12).map((event) => (
              <li key={event.id} className="flex flex-col gap-0.5">
                <span className="text-sm font-medium text-fg">
                  {event.eventType}
                </span>
                <span className="text-xs font-mono text-fg-muted">
                  {relativeTime(event.createdAt)} · {event.actorId.slice(0, 8)}
                </span>
              </li>
            ))}
          </ol>
        )}
      </Section>
    </aside>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="border-b border-border p-3 space-y-2">
      <h3 className="text-xs font-mono uppercase tracking-wider text-fg-muted">
        {title}
      </h3>
      <div className="space-y-2">{children}</div>
    </section>
  )
}

function KeyValue({
  label,
  value,
  mono = false,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className="flex items-baseline gap-2 text-base">
      <span className="text-fg-muted w-24 shrink-0 text-sm">{label}</span>
      <span className={cn("text-fg truncate", mono && "font-mono text-sm")}>
        {value}
      </span>
    </div>
  )
}

function ContactCard({
  name,
  email,
  phone,
  company,
  jobTitle,
  isVip,
}: {
  name: string
  email: string
  phone: string
  company: string
  jobTitle: string
  isVip: boolean
}) {
  return (
    <Surface level={0} radius="sm" className="p-3 space-y-2">
      <div className="flex items-center gap-2">
        <div className="size-9 rounded-full bg-accent-subtle text-fg-secondary flex items-center justify-center font-medium text-md">
          {initialsFrom(name)}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="text-md font-medium text-fg truncate">{name}</span>
            {isVip && (
              <span className="text-[10px] font-mono uppercase tracking-wider px-1 rounded-xs bg-warn-subtle text-warn">
                VIP
              </span>
            )}
          </div>
          {(company || jobTitle) && (
            <div className="text-sm text-fg-muted truncate">
              {[jobTitle, company].filter(Boolean).join(" · ")}
            </div>
          )}
        </div>
      </div>

      {(email || phone) && (
        <div className="space-y-1 pt-1 border-t border-border">
          {email && <IconLine icon={<IconMail className="size-3.5" />} text={email} />}
          {phone && <IconLine icon={<IconPhone className="size-3.5" />} text={phone} />}
          {company && (
            <IconLine icon={<IconBuilding className="size-3.5" />} text={company} />
          )}
        </div>
      )}
    </Surface>
  )
}

function IconLine({ icon, text }: { icon: React.ReactNode; text: string }) {
  return (
    <div className="flex items-center gap-2 text-base text-fg-secondary">
      <span className="text-fg-muted" aria-hidden>{icon}</span>
      <span className="truncate">{text}</span>
    </div>
  )
}

function fullName(c: {
  firstName: string
  lastName: string
  otherName: string
  email: string
}): string {
  const name = [c.firstName, c.otherName, c.lastName].filter(Boolean).join(" ").trim()
  return name || c.email || "Unknown contact"
}

function initialsFrom(name: string): string {
  const parts = name.split(/\s+/).filter(Boolean)
  if (!parts.length) return "?"
  return parts
    .slice(0, 2)
    .map((p) => p[0]!.toUpperCase())
    .join("")
}
