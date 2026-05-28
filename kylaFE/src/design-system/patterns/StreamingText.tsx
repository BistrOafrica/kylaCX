import { cn } from "@/lib/utils"

/**
 * StreamingText — renders body text with a blinking caret while a
 * stream is in flight. Consumers update `text` as tokens arrive; the
 * component just pretty-prints with the caret affordance.
 *
 * Live region: `aria-live="polite"` so screen readers announce updates
 * without interrupting the user.
 */
export interface StreamingTextProps {
  text: string
  streaming?: boolean
  className?: string
}

export function StreamingText({
  text,
  streaming = false,
  className,
}: StreamingTextProps) {
  return (
    <div
      aria-live="polite"
      aria-atomic="false"
      className={cn("text-base text-fg whitespace-pre-wrap leading-relaxed", className)}
    >
      {text}
      {streaming && (
        <span
          aria-hidden
          className={cn(
            "inline-block w-1.5 h-3.5 ms-0.5 align-text-bottom",
            "bg-accent rounded-xs animate-pulse",
          )}
        />
      )}
    </div>
  )
}
