import type { Meta, StoryObj } from "@storybook/react"
import { Kbd } from "./Kbd"

const meta: Meta<typeof Kbd> = {
  title: "Primitives/Kbd",
  component: Kbd,
  parameters: { docs: { description: { component: "Keyboard key affordance used in menus, tooltips, and the command palette." } } },
}
export default meta
type Story = StoryObj<typeof Kbd>

export const Single: Story = {
  render: () => <Kbd>K</Kbd>,
}

export const Chord: Story = {
  render: () => <Kbd keys={["Cmd", "K"]} />,
}

export const ChordLarge: Story = {
  render: () => <Kbd keys={["Cmd", "Shift", "Enter"]} />,
}

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Kbd keys={["Cmd", "K"]} size="sm" />
      <Kbd keys={["Cmd", "K"]} size="md" />
    </div>
  ),
}
