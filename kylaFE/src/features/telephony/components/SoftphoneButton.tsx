import { IconPhone, IconPhoneX } from "@tabler/icons-react"
import { useSoftphoneStore } from "../store/softphone"
import { cn } from "@/lib/utils"

/**
 * SoftphoneButton — floating launcher pinned to the bottom-right of
 * the shell. Click to open the softphone dialer; while a call is
 * active, badges the button with a pulsing accent.
 */
export function SoftphoneButton() {
  const isOpen = useSoftphoneStore((s) => s.isOpen)
  const state = useSoftphoneStore((s) => s.state)
  const open = useSoftphoneStore((s) => s.open)
  const close = useSoftphoneStore((s) => s.close)

  const onCall = state === "active" || state === "on_hold" || state === "ringing"

  return (
    <button
      type="button"
      onClick={() => (isOpen ? close() : open())}
      aria-label={isOpen ? "Close softphone" : "Open softphone"}
      className={cn(
        "fixed bottom-9 end-3 z-30 inline-flex items-center justify-center",
        "size-9 rounded-full shadow-elev-2",
        onCall
          ? "bg-success text-success-fg"
          : "bg-accent text-accent-fg hover:bg-accent-hover",
        "transition-colors",
      )}
    >
      {isOpen ? (
        <IconPhoneX className="size-4" />
      ) : (
        <IconPhone className={cn("size-4", onCall && "animate-pulse")} />
      )}
    </button>
  )
}
