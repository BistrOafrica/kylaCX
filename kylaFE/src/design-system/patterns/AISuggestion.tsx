import * as React from "react"
import { IconSparkles, IconCheck, IconX, IconCopy } from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { StreamingText } from "./StreamingText"

/**
 * AISuggestionCard — the canonical surface for any AI output in the app.
 *
 * Used by the AI rail (Phase F0 scaffold), inbox suggested replies (F1),
 * KB drafts (F3), workflow explanations (F4), etc. Components compose
 * this rather than building their own AI affordances so the look stays
 * uniform.
 *
 *   <AISuggestionCard
 *     title="Suggested reply"
 *     body={text}
 *     streaming={isStreaming}
 *     onAccept={() => apply(text)}
 *     onDismiss={() => discard()}
 *   />
 */
export interface AISuggestionCardProps {
  title: React.ReactNode
  body: string
  streaming?: boolean
  citations?: { label: string; href?: string }[]
  onAccept?: () => void
  onDismiss?: () => void
  onCopy?: () => void
  acceptLabel?: React.ReactNode
  dismissLabel?: React.ReactNode
  className?: string
}

export function AISuggestionCard({
  title,
  body,
  streaming = false,
  citations,
  onAccept,
  onDismiss,
  onCopy,
  acceptLabel = "Accept",
  dismissLabel = "Discard",
  className,
}: AISuggestionCardProps) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-md border border-border",
        "bg-surface",
        "before:absolute before:inset-x-0 before:top-0 before:h-px",
        "before:bg-gradient-to-r before:from-accent/40 before:via-accent before:to-accent/40",
        className,
      )}
    >
      <header className="flex items-center gap-2 px-3 pt-2.5 pb-1.5">
        <span
          className={cn(
            "inline-flex items-center justify-center size-5 rounded-xs",
            "bg-accent-subtle text-accent",
          )}
          aria-hidden
        >
          <IconSparkles className="size-3.5" />
        </span>
        <span className="text-sm font-medium text-fg flex-1">{title}</span>
        {onCopy && (
          <button
            type="button"
            onClick={onCopy}
            className="text-fg-muted hover:text-fg p-1 rounded-xs hover:bg-subtle"
            aria-label="Copy"
          >
            <IconCopy className="size-3.5" />
          </button>
        )}
      </header>

      <div className="px-3 pb-3">
        <StreamingText text={body} streaming={streaming} />
        {citations && citations.length > 0 && (
          <ul className="mt-2 flex flex-wrap gap-1">
            {citations.map((c, i) => (
              <li key={i}>
                {c.href ? (
                  <a
                    href={c.href}
                    className={cn(
                      "inline-flex items-center px-1.5 h-4 rounded-xs",
                      "bg-subtle text-fg-muted text-[10px]",
                      "hover:bg-muted hover:text-fg",
                    )}
                  >
                    {c.label}
                  </a>
                ) : (
                  <span
                    className={cn(
                      "inline-flex items-center px-1.5 h-4 rounded-xs",
                      "bg-subtle text-fg-muted text-[10px]",
                    )}
                  >
                    {c.label}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {(onAccept || onDismiss) && !streaming && (
        <footer className="flex items-center gap-2 border-t border-border bg-subtle px-3 py-1.5">
          {onAccept && (
            <button
              type="button"
              onClick={onAccept}
              className={cn(
                "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm font-medium",
                "bg-accent text-accent-fg hover:bg-accent-hover",
              )}
            >
              <IconCheck className="size-3" />
              {acceptLabel}
            </button>
          )}
          {onDismiss && (
            <button
              type="button"
              onClick={onDismiss}
              className={cn(
                "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm",
                "text-fg-muted hover:text-fg hover:bg-canvas",
              )}
            >
              <IconX className="size-3" />
              {dismissLabel}
            </button>
          )}
        </footer>
      )}
    </div>
  )
}
