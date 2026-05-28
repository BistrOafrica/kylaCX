import type { Meta, StoryObj } from "@storybook/react"
import { ChannelBadge } from "./ChannelBadge"

const meta: Meta<typeof ChannelBadge> = {
  title: "Primitives/ChannelBadge",
  component: ChannelBadge,
}
export default meta
type Story = StoryObj<typeof ChannelBadge>

const CHANNELS = ["whatsapp", "email", "sms", "voice", "webchat", "instagram", "messenger"] as const

export const Subtle: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      {CHANNELS.map((c) => <ChannelBadge key={c} channel={c} />)}
    </div>
  ),
}

export const Solid: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      {CHANNELS.map((c) => <ChannelBadge key={c} channel={c} variant="solid" />)}
    </div>
  ),
}

export const WithLabel: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <ChannelBadge channel="whatsapp" label="Support" />
      <ChannelBadge channel="email"    label="Inbound" />
      <ChannelBadge channel="sms"      label="Marketing" />
    </div>
  ),
}
