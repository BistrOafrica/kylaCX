import { useState } from "react"
import {
  IconLoader2,
  IconSparkles,
  IconWand,
  IconMessage2Heart,
  IconBrain,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  Surface,
  AISuggestionCard,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useMutation } from "@tanstack/react-query"
import * as aiApi from "../api/ai"

/**
 * AIStudio — playground for the three AIService skills.
 *
 * Each tab calls the real RPC and renders the response inline. Token
 * counts and cost estimates aren't returned by the proto yet, so the
 * stats panel uses character counts as a proxy until the proto adds
 * usage fields.
 */
export function AIStudio() {
  const [tab, setTab] = useState<"classify" | "summarize" | "generate">("classify")

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="AI Studio"
        description="Test the LLM skills that automation workflows can invoke"
      />
      <nav
        role="tablist"
        className="flex items-center gap-px px-3 h-9 border-b border-border bg-canvas"
      >
        {(
          [
            { id: "classify",  label: "Classify",      icon: IconMessage2Heart },
            { id: "summarize", label: "Summarize",     icon: IconWand },
            { id: "generate",  label: "Generate reply", icon: IconBrain },
          ] as const
        ).map((t) => (
          <button
            key={t.id}
            role="tab"
            aria-selected={tab === t.id}
            onClick={() => setTab(t.id)}
            className={cn(
              "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md",
              "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
              tab === t.id && "bg-accent-subtle text-fg font-medium",
            )}
          >
            <t.icon className="size-3.5" />
            {t.label}
          </button>
        ))}
      </nav>

      <div className="flex-1 overflow-y-auto p-6 max-w-4xl">
        {tab === "classify"  && <ClassifyTab />}
        {tab === "summarize" && <SummarizeTab />}
        {tab === "generate"  && <GenerateTab />}
      </div>
    </div>
  )
}

// ── Classify ─────────────────────────────────────────────────────────────────

function ClassifyTab() {
  const [text, setText] = useState(
    "I've been waiting three days for a refund on order #1284 and nobody has responded.",
  )
  const [labels, setLabels] = useState("billing, support, sales, complaint")
  const mutation = useMutation({
    mutationFn: () =>
      aiApi.classifyText(
        text,
        labels.split(",").map((s) => s.trim()).filter(Boolean),
      ),
    onError: (err) => toast.error((err as Error).message),
  })

  return (
    <Surface level={1} radius="md" className="p-4 space-y-3">
      <Field label="Text to classify">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={5}
          className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent"
        />
      </Field>
      <Field label="Candidate labels (comma-separated)">
        <Input
          value={labels}
          onChange={(e) => setLabels(e.target.value)}
          className="h-8"
        />
      </Field>
      <RunButton onClick={() => mutation.mutate()} pending={mutation.isPending} />
      {mutation.data && (
        <div className="rounded-md border border-border bg-canvas p-3 space-y-2">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm font-medium bg-accent-subtle text-fg">
              <IconSparkles className="size-3.5" />
              {mutation.data.label}
            </span>
            <span className="font-mono text-sm text-fg-muted">
              {Math.round(mutation.data.confidence * 100)}% confidence
            </span>
          </div>
          <StatsRow inputChars={text.length} />
        </div>
      )}
    </Surface>
  )
}

// ── Summarize ────────────────────────────────────────────────────────────────

function SummarizeTab() {
  const [text, setText] = useState(
    "Customer reported recurring 502 errors on the checkout endpoint. " +
      "We traced the issue to a database connection pool exhaustion under load. " +
      "Mitigation: scaled the pool from 50 to 200 and added connection-leak monitoring. " +
      "Customer confirmed errors resolved at 14:32 UTC.",
  )
  const [maxSentences, setMaxSentences] = useState(2)
  const mutation = useMutation({
    mutationFn: () => aiApi.summarizeText(text, maxSentences),
    onError: (err) => toast.error((err as Error).message),
  })

  return (
    <Surface level={1} radius="md" className="p-4 space-y-3">
      <Field label="Text to summarize">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={6}
          className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent"
        />
      </Field>
      <Field label="Max sentences">
        <Input
          type="number"
          value={maxSentences}
          onChange={(e) => setMaxSentences(Number(e.target.value) || 2)}
          className="h-8 w-24"
          min={1}
          max={20}
        />
      </Field>
      <RunButton onClick={() => mutation.mutate()} pending={mutation.isPending} />
      {mutation.data && (
        <AISuggestionCard title="Summary" body={mutation.data} />
      )}
      {mutation.data && <StatsRow inputChars={text.length} outputChars={mutation.data.length} />}
    </Surface>
  )
}

// ── Generate reply ───────────────────────────────────────────────────────────

function GenerateTab() {
  const [history, setHistory] = useState(
    "Customer: My order is late and I'm leaving on a trip tomorrow.\n" +
      "Agent: Sorry to hear that — let me check the tracking.\n" +
      "Customer: Please, I really need this before 6pm.",
  )
  const [prompt, setPrompt] = useState(
    "Draft a polite, empathetic reply that offers expedited shipping at no charge.",
  )
  const mutation = useMutation({
    mutationFn: () =>
      aiApi.generateReply(
        prompt,
        history.split("\n").map((s) => s.trim()).filter(Boolean),
      ),
    onError: (err) => toast.error((err as Error).message),
  })

  return (
    <Surface level={1} radius="md" className="p-4 space-y-3">
      <Field label="Conversation history (one turn per line)">
        <textarea
          value={history}
          onChange={(e) => setHistory(e.target.value)}
          rows={6}
          className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent font-mono text-sm"
        />
      </Field>
      <Field label="Instruction">
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={3}
          className="w-full rounded-sm border border-border bg-surface px-2 py-1.5 text-base outline-none focus:border-accent"
        />
      </Field>
      <RunButton onClick={() => mutation.mutate()} pending={mutation.isPending} />
      {mutation.data && (
        <AISuggestionCard title="Generated reply" body={mutation.data} />
      )}
      {mutation.data && (
        <StatsRow
          inputChars={history.length + prompt.length}
          outputChars={mutation.data.length}
        />
      )}
    </Surface>
  )
}

// ── Shared ───────────────────────────────────────────────────────────────────

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1">
      <label className="block text-xs font-mono uppercase tracking-wider text-fg-muted">
        {label}
      </label>
      {children}
    </div>
  )
}

function RunButton({
  onClick,
  pending,
}: {
  onClick: () => void
  pending: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={pending}
      className={cn(
        "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
        "bg-accent text-accent-fg hover:bg-accent-hover",
        "disabled:opacity-40 disabled:pointer-events-none transition-colors",
      )}
    >
      {pending ? (
        <IconLoader2 className="size-3.5 animate-spin" />
      ) : (
        <IconSparkles className="size-3.5" />
      )}
      {pending ? "Running…" : "Run"}
    </button>
  )
}

/**
 * Quick approximation of token usage from character count. Backend
 * adds explicit token counts in a future proto revision; until then
 * agents see *something*.
 */
function StatsRow({
  inputChars,
  outputChars,
}: {
  inputChars: number
  outputChars?: number
}) {
  const inToks = Math.ceil(inputChars / 4)
  const outToks = outputChars ? Math.ceil(outputChars / 4) : 0
  return (
    <div className="flex items-center gap-4 text-xs font-mono text-fg-muted">
      <span>≈ {inToks} input tok</span>
      {outputChars !== undefined && <span>≈ {outToks} output tok</span>}
      <span className="text-fg-disabled">
        approx — backend doesn&apos;t report exact usage yet
      </span>
    </div>
  )
}
