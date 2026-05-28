import { services, unary } from "@/lib/rpc"
import {
  AnalyzeSentimentRequest,
  type Sentiment,
} from "@/pb/text_classification"
import {
  TextSummaryRequest,
  RetrieveTextSummaryRequest,
} from "@/pb/summarization"
import type { Message } from "@/pb/conversations"

/**
 * Wrappers around the existing AI proto surfaces.
 *
 * The current frontend pb contains:
 *   - SentimentAnalysis.analyzeSentiment  → sentiment badges
 *   - Summarization.createTextSummary +
 *     Summarization.readTextSummary       → async polling text summary
 *
 * The roadmap calls for an `AIService.GenerateReply` (streaming) for
 * suggested replies, but that proto isn't regenerated in kylaFE yet.
 * `suggestReplyMock` here uses the F0 mock streamer so the UI flow
 * works end-to-end; swap the implementation when the proto lands.
 */

export async function analyzeSentiment(text: string): Promise<Sentiment | null> {
  if (!text.trim()) return null
  const res = await unary(
    services.sentiment.analyzeSentiment(
      AnalyzeSentimentRequest.create({ text }) as AnalyzeSentimentRequest,
    ),
  )
  return res.sentiment ?? null
}

/**
 * Compact conversation text into a string the text summarizer can chew.
 * Limits to ~12 most-recent messages, ~6KB total, so we stay under
 * reasonable token budgets without forcing the backend to handle truncation.
 */
export function transcriptFromMessages(messages: Message[]): string {
  const recent = messages.slice(-12)
  const text = recent
    .map((m) => {
      const role = m.senderType === 1 /* AGENT */ ? "Agent" : "Customer"
      let body = ""
      try {
        body = (JSON.parse(m.content) as { text?: string }).text ?? m.content
      } catch {
        body = m.content
      }
      return `${role}: ${body}`
    })
    .join("\n")
  return text.length > 6000 ? text.slice(-6000) : text
}

/**
 * Submit a summarization job. Returns the job id; poll with `readTextSummary`
 * until status is "completed". Most jobs complete in <2s.
 */
export async function startSummary(
  text: string,
  opts: { minSentences?: number; maxSentences?: number } = {},
): Promise<{ id: string }> {
  const res = await unary(
    services.summarization.createTextSummary(
      TextSummaryRequest.create({
        text,
        minSentences: String(opts.minSentences ?? 2),
        maxSentences: String(opts.maxSentences ?? 4),
      }) as TextSummaryRequest,
    ),
  )
  return { id: res.id }
}

export async function readSummary(id: string) {
  const res = await unary(
    services.summarization.readTextSummary(
      RetrieveTextSummaryRequest.create({ id }) as RetrieveTextSummaryRequest,
    ),
  )
  if (res.response.oneofKind === "textSummary") {
    return res.response.textSummary
  }
  if (res.response.oneofKind === "notification") {
    return {
      id: res.response.notification.id,
      summaryText: "",
      status: res.response.notification.status,
      createdOn: res.response.notification.createdOn,
      completedOn: "",
    }
  }
  return null
}

/**
 * Poll until a text summary is complete. Backs off exponentially up to
 * `maxMs` (default 20s) and returns the resolved summary text.
 */
export async function awaitSummary(id: string, maxMs = 20_000): Promise<string> {
  const start = Date.now()
  let delay = 400
  while (Date.now() - start < maxMs) {
    const summary = await readSummary(id)
    if (summary?.status === "completed" && summary.summaryText) {
      return summary.summaryText
    }
    if (summary?.status === "failed") {
      throw new Error("Summary failed")
    }
    await new Promise((r) => setTimeout(r, delay))
    delay = Math.min(2000, Math.round(delay * 1.5))
  }
  throw new Error("Summary timed out")
}
