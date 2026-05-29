import { IconBolt, IconFilter, IconLayoutGrid } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { ACTIONS, type ActionType } from "../utils/actions"

/**
 * NodePalette — left-side strip the user drags from to add nodes.
 *
 * F4 ships a click-to-add interaction (cleaner than HTML5 drag for the
 * keyboard-driven workflow): the consuming canvas registers an
 * `onAdd(type)` handler and the palette buttons fire it.
 */
export function NodePalette({
  onAddTrigger,
  onAddCondition,
  onAddAction,
}: {
  onAddTrigger: () => void
  onAddCondition: () => void
  onAddAction: (type: ActionType) => void
}) {
  return (
    <aside
      aria-label="Workflow building blocks"
      className="w-56 shrink-0 overflow-y-auto bg-surface border-e border-border"
    >
      <Section title="Building blocks">
        <PaletteButton
          icon={<IconBolt className="size-3.5" />}
          label="Trigger"
          description="What starts the workflow"
          onClick={onAddTrigger}
        />
        <PaletteButton
          icon={<IconFilter className="size-3.5" />}
          label="If condition"
          description="Branch on event data"
          onClick={onAddCondition}
        />
      </Section>

      <Section title="Actions">
        {ACTIONS.map((spec) => {
          const Icon = spec.icon
          return (
            <PaletteButton
              key={spec.type}
              icon={<Icon className="size-3.5" />}
              label={spec.label}
              description={spec.description}
              accent={spec.color}
              onClick={() => onAddAction(spec.type)}
            />
          )
        })}
      </Section>
    </aside>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="border-b border-border-subtle">
      <div className="flex items-center gap-1.5 px-3 h-7 text-xs font-mono uppercase tracking-wider text-fg-muted">
        <IconLayoutGrid className="size-3" aria-hidden />
        {title}
      </div>
      <div className="p-1 space-y-px">{children}</div>
    </section>
  )
}

function PaletteButton({
  icon,
  label,
  description,
  accent,
  onClick,
}: {
  icon: React.ReactNode
  label: string
  description?: string
  accent?: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full text-start flex items-start gap-2 px-2 py-1.5 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <span
        className="mt-0.5 inline-flex items-center justify-center size-5 rounded-xs bg-accent-subtle"
        style={accent ? { color: accent } : undefined}
        aria-hidden
      >
        {icon}
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-md font-medium text-fg truncate">{label}</span>
        {description && (
          <span className="block text-sm text-fg-muted truncate">{description}</span>
        )}
      </span>
    </button>
  )
}
