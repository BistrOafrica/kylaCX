import type { Meta, StoryObj } from "@storybook/react"
import { Surface } from "./Surface"

const meta: Meta<typeof Surface> = {
  title: "Primitives/Surface",
  component: Surface,
}
export default meta
type Story = StoryObj<typeof Surface>

export const Levels: Story = {
  render: () => (
    <div className="space-y-3">
      {[0, 1, 2, 3].map((level) => (
        <Surface key={level} level={level as 0 | 1 | 2 | 3} inset radius="md">
          <div className="text-md font-medium text-fg">Level {level}</div>
          <div className="text-base text-fg-muted">
            Hairline border + escalating shadow.
          </div>
        </Surface>
      ))}
    </div>
  ),
}
