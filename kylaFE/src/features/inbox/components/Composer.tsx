import { useState, useRef, type KeyboardEvent } from "react"
import { useTranslation } from "react-i18next"
import { IconSend, IconNote, IconMessage, IconLoader2 } from "@tabler/icons-react"
import { toast } from "sonner"
import { cn } from "@/lib/utils"
import { Kbd } from "@/design-system"
import { useSendMessage } from "../hooks/queries"

/**
 * Composer — reply field at the bottom of the conversation thread.
 *
 * Currently a plain textarea. F1.10 will swap to Tiptap for rich text
 * + macro insertion; the API and key-handling stay the same so callers
 * don't break.
 *
 * Keys:
 *   ⌘+Enter   send
 *   ⌘+Shift+N toggle internal note mode (drafted, no backend yet)
 */
export function Composer({ conversationId }: { conversationId: string }) {
  const { t } = useTranslation()
  const [text, setText] = useState("")
  const [mode, setMode] = useState<"reply" | "note">("reply")
  const ref = useRef<HTMLTextAreaElement>(null)
  const send = useSendMessage()

  const handleSend = async () => {
    const body = text.trim()
    if (!body) return
    if (mode === "note") {
      // Internal notes aren't a proto RPC yet — surface the limitation
      // clearly rather than swallowing the click. F1.x replaces this
      // once the backend exposes an internal-note channel.
      toast.info("Internal notes ship in a later F1 pass")
      return
    }
    try {
      await send.mutateAsync({ conversationId, body })
      setText("")
      ref.current?.focus()
    } catch {
      /* mutationCache toasts the error */
    }
  }

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault()
      void handleSend()
    }
  }

  return (
    <div className="border-t border-border bg-surface">
      <div className="flex items-center gap-1 px-3 pt-2">
        <ModeButton
          active={mode === "reply"}
          onClick={() => setMode("reply")}
          icon={<IconMessage className="size-3.5" />}
          label="Reply"
        />
        <ModeButton
          active={mode === "note"}
          onClick={() => setMode("note")}
          icon={<IconNote className="size-3.5" />}
          label="Internal note"
        />
      </div>

      <textarea
        ref={ref}
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder={
          mode === "note"
            ? "Internal note. Not visible to the customer."
            : "Reply…  ⌘+Enter to send"
        }
        rows={3}
        className={cn(
          "w-full resize-none px-3 py-2 bg-transparent outline-none",
          "text-base text-fg placeholder:text-fg-muted",
          mode === "note" && "bg-warn-subtle",
        )}
        aria-label={mode === "note" ? "Internal note" : "Reply"}
      />

      <div className="flex items-center gap-2 px-3 pb-2">
        <span className="text-xs text-fg-muted ms-auto">
          {t("common.search") /* placeholder; macros land in F1.10 */}
        </span>
        <button
          type="button"
          onClick={() => void handleSend()}
          disabled={send.isPending || !text.trim()}
          className={cn(
            "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none transition-colors",
          )}
        >
          {send.isPending ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconSend className="size-3.5" />
          )}
          Send
          <Kbd keys={["Cmd", "Enter"]} size="sm" />
        </button>
      </div>
    </div>
  )
}

function ModeButton({
  active,
  onClick,
  icon,
  label,
}: {
  active: boolean
  onClick: () => void
  icon: React.ReactNode
  label: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "inline-flex items-center gap-1 h-6 px-2 rounded-xs text-sm",
        "text-fg-muted hover:text-fg hover:bg-subtle",
        active && "bg-accent-subtle text-fg",
      )}
    >
      {icon}
      {label}
    </button>
  )
}
