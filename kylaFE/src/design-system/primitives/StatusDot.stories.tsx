import type { Meta, StoryObj } from "@storybook/react"
import { StatusDot, type StatusTone } from "./StatusDot"

const meta: Meta<typeof StatusDot> = {
  title: "Primitives/StatusDot",
  component: StatusDot,
}
export default meta
type Story = StoryObj<typeof StatusDot>

const TONES: StatusTone[] = [
  "unread", "read", "online", "offline", "busy", "away",
  "success", "warn", "danger", "info",
]

export const AllTones: Story = {
  render: () => (
    <div className="grid grid-cols-2 gap-3">
      {TONES.map((tone) => (
        <div key={tone} className="flex items-center gap-2 text-base">
          <StatusDot tone={tone} />
          <span className="font-mono text-fg-muted">{tone}</span>
        </div>
      ))}
    </div>
  ),
}

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <StatusDot tone="online" size={4} />
      <StatusDot tone="online" size={6} />
      <StatusDot tone="online" size={8} />
      <StatusDot tone="online" size={10} />
    </div>
  ),
}

export const Pulse: Story = {
  render: () => <StatusDot tone="online" size={10} pulse />,
}
