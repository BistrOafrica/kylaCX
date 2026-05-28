import type { Meta, StoryObj } from "@storybook/react"
import { IconInbox, IconSparkles } from "@tabler/icons-react"
import { EmptyState } from "./EmptyState"

const meta: Meta<typeof EmptyState> = {
  title: "Patterns/EmptyState",
  component: EmptyState,
}
export default meta
type Story = StoryObj<typeof EmptyState>

export const Default: Story = {
  render: () => (
    <EmptyState
      icon={<IconInbox className="size-5" />}
      title="Inbox zero"
      description="Nothing waiting for you. New conversations will appear here."
    />
  ),
}

export const WithAction: Story = {
  render: () => (
    <EmptyState
      icon={<IconSparkles className="size-5" />}
      title="No suggestions yet"
      description="Open a conversation to let Kyla draft a reply."
      action={
        <button className="inline-flex items-center h-8 px-3 rounded-sm bg-accent text-accent-fg text-md font-medium">
          Open inbox
        </button>
      }
    />
  ),
}

export const Sizes: Story = {
  render: () => (
    <div className="space-y-8">
      <EmptyState size="sm" title="Small" description="Compact rows." />
      <EmptyState size="md" title="Medium" description="The default." />
      <EmptyState size="lg" title="Large" description="Hero empty state." />
    </div>
  ),
}
