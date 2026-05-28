import { useTranslation } from "react-i18next"
import {
  IconChevronDown,
  IconSearch,
  IconSparkles,
  IconBell,
} from "@tabler/icons-react"
import { Kbd } from "@/design-system"
import { cn } from "@/lib/utils"
import { useCommandStore } from "@/lib/command"
import { useAIStore } from "@/lib/ai"
import { useWorkspaceStore } from "@/lib/workspace"
import { useAuthStore } from "@/lib/auth"
import { ProfileMenu } from "./ProfileMenu"

/**
 * TopBar — 40px strip across the top of the authenticated shell.
 *
 *   [Workspace ▾]   [⌘K Search anything…]            [✨] [🔔] [Profile ▾]
 */
export function TopBar() {
  const { t } = useTranslation()
  const openPalette = useCommandStore((s) => s.open)
  const toggleAI = useAIStore((s) => s.toggle)
  const identity = useAuthStore((s) => s.identity)
  const workspace = useWorkspaceStore((s) => s.workspace)
  const organisation = useWorkspaceStore((s) => s.organisation)

  return (
    <header
      className={cn(
        "h-10 shrink-0 flex items-center gap-2 px-2",
        "bg-canvas border-b border-border",
      )}
      role="banner"
    >
      <WorkspaceButton
        org={organisation?.name}
        workspace={workspace?.name}
        ariaLabel={t("shell.workspaceSwitcher")}
      />

      <button
        type="button"
        onClick={openPalette}
        aria-label={t("shell.openCommandPalette")}
        className={cn(
          "flex items-center gap-2 h-7 max-w-md flex-1 px-2.5 rounded-sm",
          "bg-subtle text-fg-muted hover:text-fg hover:bg-muted",
          "border border-border transition-colors",
        )}
      >
        <IconSearch className="size-3.5" aria-hidden />
        <span className="flex-1 text-start text-base truncate">
          {t("common.searchAnything")}
        </span>
        <Kbd keys={["Cmd", "K"]} size="sm" />
      </button>

      <div className="ms-auto flex items-center gap-1">
        <IconButton
          icon={<IconSparkles className="size-4" />}
          label={t("shell.toggleAiRail")}
          onClick={toggleAI}
          shortcut={["Cmd", "J"]}
        />
        <IconButton
          icon={<IconBell className="size-4" />}
          label={t("shell.notifications")}
        />
        <ProfileMenu identity={identity} />
      </div>
    </header>
  )
}

function WorkspaceButton({
  org,
  workspace,
  ariaLabel,
}: {
  org?: string
  workspace?: string
  ariaLabel: string
}) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={cn(
        "flex items-center gap-1.5 h-7 px-2 rounded-sm",
        "hover:bg-subtle text-fg transition-colors",
        "max-w-[14rem]",
      )}
    >
      <span
        className={cn(
          "inline-flex items-center justify-center size-5 rounded-xs",
          "bg-accent text-accent-fg font-semibold text-[11px]",
        )}
        aria-hidden
      >
        {(org ?? workspace ?? "K").charAt(0).toUpperCase()}
      </span>
      <span className="text-md font-medium truncate">
        {workspace ?? org ?? "Kyla"}
      </span>
      <IconChevronDown className="size-3 text-fg-muted shrink-0" aria-hidden />
    </button>
  )
}

function IconButton({
  icon,
  label,
  onClick,
  shortcut,
}: {
  icon: React.ReactNode
  label: string
  onClick?: () => void
  shortcut?: string[]
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={shortcut ? `${label}  ${shortcut.join("+")}` : label}
      className={cn(
        "inline-flex items-center justify-center size-7 rounded-sm",
        "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
      )}
    >
      {icon}
    </button>
  )
}
