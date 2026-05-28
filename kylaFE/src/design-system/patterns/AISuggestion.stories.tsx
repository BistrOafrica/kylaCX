import type { Meta, StoryObj } from "@storybook/react"
import { useEffect, useState } from "react"
import { AISuggestionCard } from "./AISuggestion"
import { streamMockResponse } from "@/lib/ai"

const meta: Meta<typeof AISuggestionCard> = {
  title: "Patterns/AI Suggestion",
  component: AISuggestionCard,
}
export default meta
type Story = StoryObj<typeof AISuggestionCard>

const SAMPLE =
  "Thanks for reaching out — I can see your order was dispatched yesterday and is currently in transit. Tracking number TKL-3892. It should arrive by Thursday."

export const Static: Story = {
  render: () => (
    <div className="max-w-sm">
      <AISuggestionCard
        title="Suggested reply"
        body={SAMPLE}
        citations={[{ label: "Order #1284" }, { label: "Tracking" }]}
        onAccept={() => alert("Accepted")}
        onDismiss={() => alert("Dismissed")}
      />
    </div>
  ),
}

function StreamingDemo() {
  const [text, setText] = useState("")
  const [streaming, setStreaming] = useState(true)

  useEffect(() => {
    const controller = new AbortController()
    void streamMockResponse(SAMPLE, {
      onToken: (chunk) => setText((t) => t + chunk),
      onDone: () => setStreaming(false),
      signal: controller.signal,
    })
    return () => controller.abort()
  }, [])

  return (
    <div className="max-w-sm">
      <AISuggestionCard
        title="Kyla is thinking…"
        body={text}
        streaming={streaming}
      />
    </div>
  )
}

export const Streaming: Story = {
  render: () => <StreamingDemo />,
}
