import { useState } from "react"
import { IconBolt, IconPlus, IconLoader2 } from "@tabler/icons-react"
import { toast } from "sonner"
import { Surface } from "@/design-system"
import { cn } from "@/lib/utils"
import { useMacros, useApplyMacro } from "../hooks/queries"
import type { Macro } from "../api/ticketing"

/**
 * MacroPicker — apply a canned response to a ticket.
 *
 * Lists macros for the active workspace; selecting one calls
 * ApplyMacro which posts a message (when roomId is provided) and
 * patches ticket data fields per the macro's `actions` JSON.
 */
export function MacroPicker({
  ticketId,
  roomId,
}: {
  ticketId: string
  roomId?: string
}) {
  const macros = useMacros()
  const apply = useApplyMacro()
  const [filter, setFilter] = useState("")
  const [openId, setOpenId] = useState<string | null>(null)

  const items = (macros.data ?? []).filter((m) =>
    m.name.toLowerCase().includes(filter.toLowerCase()),
  )

  const onApply = async (macro: Macro) => {
    setOpenId(macro.id)
    try {
      await apply.mutateAsync({ macroId: macro.id, ticketId, roomId })
      toast.success(`Macro "${macro.name}" applied`)
    } finally {
      setOpenId(null)
    }
  }

  return (
    <Surface level={1} radius="md" className="flex flex-col h-full">
      <header className="flex items-center gap-2 h-9 shrink-0 px-3 border-b border-border">
        <IconBolt className="size-3.5 text-fg-muted" aria-hidden />
        <span className="text-md font-medium text-fg flex-1">Macros</span>
        <button
          type="button"
          className="inline-flex items-center justify-center size-6 rounded-xs text-fg-muted hover:text-fg hover:bg-subtle"
          aria-label="Create macro"
        >
          <IconPlus className="size-3.5" />
        </button>
      </header>

      <div className="p-2 border-b border-border-subtle">
        <input
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Filter macros…"
          className="w-full h-7 px-2 rounded-sm bg-subtle border border-border text-base outline-none focus:border-accent placeholder:text-fg-muted"
        />
      </div>

      <div className="flex-1 overflow-y-auto p-1.5 space-y-px">
        {macros.isPending && (
          <div className="text-sm text-fg-muted text-center py-4">Loading…</div>
        )}
        {!macros.isPending && items.length === 0 && (
          <div className="text-sm text-fg-muted text-center py-4">
            {filter ? "No matches" : "No macros yet"}
          </div>
        )}
        {items.map((m) => (
          <button
            key={m.id}
            type="button"
            onClick={() => void onApply(m)}
            disabled={apply.isPending && openId === m.id}
            className={cn(
              "w-full text-start px-2 py-1.5 rounded-sm",
              "hover:bg-subtle disabled:opacity-50 transition-colors",
            )}
          >
            <div className="flex items-center gap-2">
              <span className="text-md font-medium text-fg truncate flex-1">
                {m.name}
              </span>
              {apply.isPending && openId === m.id && (
                <IconLoader2 className="size-3 animate-spin text-fg-muted" />
              )}
            </div>
            <div className="text-sm text-fg-muted line-clamp-2">
              {m.content}
            </div>
          </button>
        ))}
      </div>
    </Surface>
  )
}
