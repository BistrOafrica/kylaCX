import { useTranslation } from "react-i18next"
import { IconSparkles, IconX } from "@tabler/icons-react"
import { AnimatePresence, motion } from "framer-motion"
import { useAIStore } from "@/lib/ai"
import { AISuggestionCard, EmptyState } from "@/design-system"
import { cn } from "@/lib/utils"

/**
 * AIRail — right-edge copilot panel.
 *
 * Always renderable; visibility controlled by useAIStore. Width is
 * fixed at 320px on ≥1440px viewports; on narrower screens the rail
 * overlays the primary surface as a drawer (F1 will refine this).
 *
 * F0 scope: empty-state pitch + suggestion list reading from the
 * store. F1 wires real AIService streaming behind the same UI.
 */
export function AIRail() {
  const { t } = useTranslation()
  const isOpen = useAIStore((s) => s.isOpen)
  const close = useAIStore((s) => s.close)
  const context = useAIStore((s) => s.context)
  const suggestions = useAIStore((s) => s.suggestions)
  const removeSuggestion = useAIStore((s) => s.removeSuggestion)

  return (
    <AnimatePresence initial={false}>
      {isOpen && (
        <motion.aside
          key="ai-rail"
          aria-label={t("ai.railTitle")}
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: 320, opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: 0.18, ease: [0.2, 0, 0, 1] }}
          className={cn(
            "shrink-0 overflow-hidden bg-surface border-s border-border",
            "flex flex-col",
          )}
        >
          <header className="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border">
            <span
              className={cn(
                "inline-flex items-center justify-center size-5 rounded-xs",
                "bg-accent-subtle text-accent",
              )}
              aria-hidden
            >
              <IconSparkles className="size-3.5" />
            </span>
            <span className="text-md font-medium text-fg flex-1">
              {t("ai.railTitle")}
            </span>
            <button
              type="button"
              onClick={close}
              aria-label={t("common.close")}
              className={cn(
                "inline-flex items-center justify-center size-5 rounded-xs",
                "text-fg-muted hover:text-fg hover:bg-subtle",
              )}
            >
              <IconX className="size-3.5" />
            </button>
          </header>

          <div className="flex-1 overflow-y-auto">
            {suggestions.length === 0 ? (
              <EmptyState
                icon={<IconSparkles className="size-4" />}
                title={t("ai.emptyTitle")}
                description={t("ai.emptyDescription")}
                size="sm"
                className="px-4"
              />
            ) : (
              <div className="p-3 space-y-2">
                {suggestions.map((s) => (
                  <AISuggestionCard
                    key={s.id}
                    title={s.title}
                    body={s.body}
                    streaming={s.status === "streaming"}
                    citations={s.citations}
                    onDismiss={() => removeSuggestion(s.id)}
                  />
                ))}
              </div>
            )}
          </div>

          {context.kind !== "none" && (
            <footer className="border-t border-border px-3 py-1.5">
              <div className="text-xs text-fg-muted font-mono uppercase">
                {context.kind}
              </div>
              {context.title && (
                <div className="text-sm text-fg-secondary truncate">
                  {context.title}
                </div>
              )}
            </footer>
          )}
        </motion.aside>
      )}
    </AnimatePresence>
  )
}
