import type { Meta, StoryObj } from "@storybook/react"

const meta: Meta = {
  title: "Tokens/Typography",
  parameters: { layout: "fullscreen" },
}
export default meta

export const Scale: StoryObj = {
  render: () => (
    <div className="p-8 bg-canvas min-h-dvh space-y-6">
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-3xl · 28 / 36</div>
        <div className="text-3xl font-semibold text-fg">Kyla agent workbench</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-2xl · 22 / 30</div>
        <div className="text-2xl font-semibold text-fg">Customer experience operations</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-xl · 18 / 26</div>
        <div className="text-xl font-semibold text-fg">Inbox · CRM · Tickets · Knowledge</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-lg · 16 / 24</div>
        <div className="text-lg text-fg">A workbench for the eight-hour shift.</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-md · 14 / 20</div>
        <div className="text-md text-fg">Default UI weight. Buttons, inputs, menu items.</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-base · 13 / 20 (dense default)</div>
        <div className="text-base text-fg">Body copy. Most labels. Most metadata.</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-sm · 12 / 18</div>
        <div className="text-sm text-fg">Table cells. Captions. Secondary information.</div>
      </div>
      <div className="space-y-1">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">text-xs · 11 / 16</div>
        <div className="text-xs text-fg">Micro labels. Badges. Kbd.</div>
      </div>

      <div className="border-t border-border pt-6 space-y-2">
        <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">font-mono</div>
        <code className="font-mono text-md text-fg block">
          #1284 · org_2x9k · TKL-3892 · $1,240.00
        </code>
      </div>
    </div>
  ),
}
