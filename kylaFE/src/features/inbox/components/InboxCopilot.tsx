import { useEffect } from "react"
import { useTranslation } from "react-i18next"
import {
  IconSparkles,
  IconMessage2Heart,
  IconLanguage,
  IconWand,
} from "@tabler/icons-react"
import { toast } from "sonner"
import { useAIStore } from "@/lib/ai"
import { streamMockResponse } from "@/lib/ai"
import { useMessages } from "../hooks/queries"
import {
  analyzeSentiment,
  awaitSummary,
  startSummary,
  transcriptFromMessages,
} from "../api/ai"
import { readContentText } from "../utils/format"
import { cn } from "@/lib/utils"
import type { Conversation, Message } from "@/pb/conversations"

/**
 * Action strip shown inside the AI rail when a conversation is open.
 *
 * Provides four affordances:
 *   - Summarize thread → real Summarization service (async polling)
 *   - Sentiment        → real SentimentAnalysis service
 *   - Suggest reply    → mock streamer (real GenerateReply RPC pending
 *                        backend regen; same UI flow so swap is one
 *                        function once proto lands)
 *   - Translate        → mock streamer (same caveat)
 *
 * Each action emits an AISuggestion into the rail store, which the
 * AIRail in app/shell renders.
 */
export function InboxCopilot({ conv }: { conv: Conversation }) {
  const { t } = useTranslation()
  const messages = useMessages(conv.id)
  const setContext = useAIStore((s) => s.setContext)
  const open = useAIStore((s) => s.open)
  const pushSuggestion = useAIStore((s) => s.pushSuggestion)
  const appendToSuggestion = useAIStore((s) => s.appendToSuggestion)
  const finalizeSuggestion = useAIStore((s) => s.finalizeSuggestion)

  // Tell the rail what we're contextual to so it can render the chip.
  useEffect(() => {
    setContext({
      kind: "conversation",
      resourceId: conv.id,
      title: conv.subject || "Conversation",
    })
  }, [conv.id, conv.subject, setContext])

  const onSummarize = async () => {
    if (!messages.data?.length) {
      toast.info("No messages to summarize yet")
      return
    }
    open()
    const id = `summary-${Date.now()}`
    pushSuggestion({
      id,
      kind: "summary",
      title: "Thread summary",
      body: "",
      status: "streaming",
      createdAt: Date.now(),
    })

    try {
      const transcript = transcriptFromMessages(messages.data)
      const { id: jobId } = await startSummary(transcript, {
        minSentences: 2,
        maxSentences: 5,
      })
      const summary = await awaitSummary(jobId)
      // We don't have a real token stream here, but we can still
      // animate by chunking the result word-by-word for visual parity.
      await streamMockResponse(summary, {
        onToken: (chunk) => appendToSuggestion(id, chunk),
      })
      finalizeSuggestion(id, "complete")
    } catch (err) {
      finalizeSuggestion(id, "error")
      toast.error(err instanceof Error ? err.message : "Summary failed")
    }
  }

  const onSentiment = async () => {
    const lastFromContact = lastContactMessage(messages.data ?? [])
    if (!lastFromContact) {
      toast.info("No customer messages to analyze yet")
      return
    }
    open()
    const id = `sentiment-${Date.now()}`
    pushSuggestion({
      id,
      kind: "classify",
      title: "Sentiment",
      body: "",
      status: "streaming",
      createdAt: Date.now(),
    })
    try {
      const result = await analyzeSentiment(readContentText(lastFromContact.content))
      const label = result?.label ?? "neutral"
      const score = result ? Math.round(result.score * 100) : 0
      appendToSuggestion(id, `${label.toUpperCase()} · ${score}% confidence`)
      finalizeSuggestion(id, "complete")
    } catch (err) {
      finalizeSuggestion(id, "error")
      toast.error(err instanceof Error ? err.message : "Sentiment failed")
    }
  }

  const onSuggestReply = async () => {
    const lastFromContact = lastContactMessage(messages.data ?? [])
    if (!lastFromContact) {
      toast.info("Reply suggestions need a customer message to react to")
      return
    }
    open()
    const id = `reply-${Date.now()}`
    pushSuggestion({
      id,
      kind: "reply",
      title: "Suggested reply",
      body: "",
      status: "streaming",
      createdAt: Date.now(),
    })

    // Until the backend AI proto ships GenerateReply, draft via mock
    // streamer — keeps the UX flow real for the agent.
    const mockReply =
      "Thanks for reaching out. I can see this in your account and I'm looking into it now — I'll have an update within the next 30 minutes. In the meantime, can you confirm the order number you used at checkout?"
    await streamMockResponse(mockReply, {
      onToken: (chunk) => appendToSuggestion(id, chunk),
    })
    finalizeSuggestion(id, "complete")
  }

  const onTranslate = async () => {
    const lastFromContact = lastContactMessage(messages.data ?? [])
    if (!lastFromContact) return
    open()
    const id = `translate-${Date.now()}`
    pushSuggestion({
      id,
      kind: "translate",
      title: "Translation",
      body: "",
      status: "streaming",
      createdAt: Date.now(),
    })
    // Mock — replace with real translate RPC when proto ships.
    await streamMockResponse(
      "(translate.proto not yet wired — message would be translated to your active locale here.)",
      { onToken: (chunk) => appendToSuggestion(id, chunk) },
    )
    finalizeSuggestion(id, "complete")
  }

  return (
    <div className="flex flex-col border-s border-border bg-surface w-12">
      <Tool
        icon={<IconSparkles className="size-4" />}
        label={t("ai.suggestReply")}
        onClick={() => void onSuggestReply()}
      />
      <Tool
        icon={<IconWand className="size-4" />}
        label={t("ai.summarize")}
        onClick={() => void onSummarize()}
      />
      <Tool
        icon={<IconMessage2Heart className="size-4" />}
        label={t("ai.classify")}
        onClick={() => void onSentiment()}
      />
      <Tool
        icon={<IconLanguage className="size-4" />}
        label={t("ai.translate")}
        onClick={() => void onTranslate()}
      />
    </div>
  )
}

function Tool({
  icon,
  label,
  onClick,
}: {
  icon: React.ReactNode
  label: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className={cn(
        "size-10 inline-flex items-center justify-center",
        "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
        "border-b border-border-subtle last:border-b-0",
      )}
    >
      {icon}
    </button>
  )
}

function lastContactMessage(messages: Message[]): Message | null {
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i]!.senderType === 2 /* CONTACT */) return messages[i]!
  }
  return null
}
