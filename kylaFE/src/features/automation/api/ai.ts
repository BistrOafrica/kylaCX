import { services, unary } from "@/lib/rpc"
import {
  ClassifyTextRequest,
  SummarizeTextRequest,
  GenerateReplyRequest,
} from "@/pb/ai"

/**
 * Wrappers around the canonical AIService — three unary skills.
 *
 * F1's inbox copilot uses the sentiment + summarization protos
 * (separate services); the AIService here is what the Automation
 * Studio playground talks to, and what the `run_ai_skill` action
 * runtime mirrors on the backend.
 */

export async function classifyText(text: string, labels: string[] = []) {
  const res = await unary(
    services.ai.classifyText(
      ClassifyTextRequest.create({ text, labels }) as ClassifyTextRequest,
    ),
  )
  return { label: res.label, confidence: res.confidence }
}

export async function summarizeText(text: string, maxSentences = 4) {
  const res = await unary(
    services.ai.summarizeText(
      SummarizeTextRequest.create({
        text,
        maxSentences,
      }) as SummarizeTextRequest,
    ),
  )
  return res.summary
}

export async function generateReply(prompt: string, history: string[] = []) {
  const res = await unary(
    services.ai.generateReply(
      GenerateReplyRequest.create({
        prompt,
        history,
      }) as GenerateReplyRequest,
    ),
  )
  return res.reply
}
