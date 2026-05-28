import type { Meta, StoryObj } from "@storybook/react"

/**
 * Tokens — visual reference for every semantic token in the design system.
 *
 * Toggle the theme in the toolbar to see both light and dark mappings.
 * Every component in the app reads from these variables; if a color
 * needs adding, edit src/design-system/tokens/*.css, not this file.
 */

const meta: Meta = {
  title: "Tokens/Reference",
  parameters: { layout: "fullscreen" },
}
export default meta

const BG_TOKENS = ["canvas", "surface", "elevated", "subtle", "muted", "inverse"]
const TEXT_TOKENS = ["fg", "fg-secondary", "fg-muted", "fg-disabled", "fg-inverse", "fg-link"]
const ACCENT_TOKENS = ["accent", "accent-hover", "accent-active", "accent-subtle", "accent-muted"]
const STATUS_TOKENS = ["success", "warn", "danger", "info"]
const CHANNEL_TOKENS = [
  "channel-whatsapp", "channel-email", "channel-sms",
  "channel-voice", "channel-webchat", "channel-instagram", "channel-messenger",
]
const RADIUS_TOKENS = ["xs", "sm", "md", "lg", "xl"]

function Swatch({ name, tw }: { name: string; tw: string }) {
  return (
    <div className="flex items-center gap-3 p-2 border border-border rounded-sm">
      <div
        className={`size-10 rounded-sm border border-border ${tw}`}
        aria-hidden
      />
      <div className="font-mono text-xs text-fg-secondary truncate">{name}</div>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="text-md font-semibold text-fg font-mono uppercase tracking-wider">
        {title}
      </h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
        {children}
      </div>
    </section>
  )
}

export const All: StoryObj = {
  render: () => (
    <div className="p-6 bg-canvas text-fg space-y-8 min-h-dvh">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Tokens</h1>
        <p className="text-base text-fg-muted">
          Switch theme in the toolbar to compare light + dark mappings.
        </p>
      </header>

      <Section title="Backgrounds">
        {BG_TOKENS.map((t) => <Swatch key={t} name={t} tw={`bg-${t}`} />)}
      </Section>

      <Section title="Text">
        {TEXT_TOKENS.map((t) => (
          <div
            key={t}
            className="flex items-center gap-3 p-2 border border-border rounded-sm"
          >
            <div className={`text-lg font-semibold text-${t}`}>Aa</div>
            <div className="font-mono text-xs text-fg-secondary">{t}</div>
          </div>
        ))}
      </Section>

      <Section title="Accent">
        {ACCENT_TOKENS.map((t) => <Swatch key={t} name={t} tw={`bg-${t}`} />)}
      </Section>

      <Section title="Status">
        {STATUS_TOKENS.map((t) => (
          <>
            <Swatch key={`${t}-solid`} name={`${t} solid`} tw={`bg-${t}`} />
            <Swatch key={`${t}-subtle`} name={`${t} subtle`} tw={`bg-${t}-subtle`} />
          </>
        ))}
      </Section>

      <Section title="Channels">
        {CHANNEL_TOKENS.map((t) => <Swatch key={t} name={t} tw={`bg-${t}`} />)}
      </Section>

      <Section title="Radius">
        {RADIUS_TOKENS.map((t) => (
          <div
            key={t}
            className="flex items-center gap-3 p-2 border border-border rounded-sm"
          >
            <div className={`size-10 bg-accent rounded-${t}`} aria-hidden />
            <div className="font-mono text-xs text-fg-secondary">{t}</div>
          </div>
        ))}
      </Section>
    </div>
  ),
}
