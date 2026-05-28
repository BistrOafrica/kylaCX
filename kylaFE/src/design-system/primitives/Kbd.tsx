import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * Kbd — keyboard key affordance.
 *
 * Renders a key cap for inline hints in menus, tooltips, the command
 * palette, and the shortcuts overlay. Accepts either text content or
 * a list of keys (renders chord with `+` separators).
 *
 *   <Kbd>K</Kbd>
 *   <Kbd keys={["Cmd", "K"]} />
 *
 * On macOS the platform glyph (⌘) is used; on Windows/Linux "Ctrl"
 * is rendered.
 */
export interface KbdProps extends React.HTMLAttributes<HTMLElement> {
  keys?: string[]
  size?: "sm" | "md"
}

const PLATFORM_KEY: Record<string, string> = {
  Cmd: typeof navigator !== "undefined" && /Mac/.test(navigator.platform) ? "⌘" : "Ctrl",
  Mod: typeof navigator !== "undefined" && /Mac/.test(navigator.platform) ? "⌘" : "Ctrl",
  Shift: "⇧",
  Alt: typeof navigator !== "undefined" && /Mac/.test(navigator.platform) ? "⌥" : "Alt",
  Option: "⌥",
  Enter: "↵",
  ArrowUp: "↑",
  ArrowDown: "↓",
  ArrowLeft: "←",
  ArrowRight: "→",
  Escape: "Esc",
  Backspace: "⌫",
}

const renderKey = (key: string) => PLATFORM_KEY[key] ?? key

export function Kbd({
  children,
  keys,
  size = "md",
  className,
  ...rest
}: KbdProps) {
  const content = keys
    ? keys.map((k, i) => (
        <React.Fragment key={i}>
          <span>{renderKey(k)}</span>
          {i < keys.length - 1 && <span className="opacity-50">+</span>}
        </React.Fragment>
      ))
    : children

  return (
    <kbd
      className={cn(
        "inline-flex items-center gap-0.5 font-mono font-medium",
        "rounded-xs border border-border bg-subtle text-fg-secondary",
        "shadow-elev-1",
        size === "sm" && "h-4 min-w-4 px-1 text-[10px]",
        size === "md" && "h-5 min-w-5 px-1.5 text-[11px]",
        className,
      )}
      {...rest}
    >
      {content}
    </kbd>
  )
}
