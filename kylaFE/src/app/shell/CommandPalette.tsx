import { Command as CommandKit } from "cmdk"
import { useEffect, useMemo } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { useTheme } from "@/components/theme-provider"
import { useCommandStore, type Command, type CommandSection } from "@/lib/command"
import { useAIStore } from "@/lib/ai"
import {
  changeLanguage,
  LOCALE_NAMES,
  SUPPORTED_LOCALES,
  type Locale,
} from "@/lib/i18n"
import { Kbd } from "@/design-system"
import { logout } from "@/lib/auth"
import { PRIMARY_NAV, SECONDARY_NAV } from "./nav-items"
import { cn } from "@/lib/utils"

/**
 * ⌘K command palette.
 *
 * The palette renders from useCommandStore — feature modules can
 * register commands via `registerMany` and they appear here. F0
 * ships a baseline set covering navigation, theme, language,
 * workspace, account, help.
 */

export function CommandPalette() {
  const { t } = useTranslation()
  const isOpen = useCommandStore((s) => s.isOpen)
  const close = useCommandStore((s) => s.close)
  const recordRun = useCommandStore((s) => s.recordRun)
  const commands = useBaselineCommands()

  // Close on Escape (cmdk handles internally but we also clear focus)
  useEffect(() => {
    if (!isOpen) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close()
    }
    document.addEventListener("keydown", onKey)
    return () => document.removeEventListener("keydown", onKey)
  }, [isOpen, close])

  // Group by section while preserving order
  const grouped = useMemo(() => {
    const map = new Map<CommandSection, Command[]>()
    for (const cmd of commands) {
      const arr = map.get(cmd.section) ?? []
      arr.push(cmd)
      map.set(cmd.section, arr)
    }
    return [...map.entries()]
  }, [commands])

  if (!isOpen) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t("shell.openCommandPalette")}
      className={cn(
        "fixed inset-0 z-50 flex items-start justify-center pt-[12vh] px-4",
        "bg-canvas/60 backdrop-blur-sm",
      )}
      onClick={(e) => {
        if (e.target === e.currentTarget) close()
      }}
    >
      <CommandKit
        label="Command palette"
        className={cn(
          "w-full max-w-2xl rounded-lg overflow-hidden",
          "bg-elevated border border-border shadow-elev-4",
          "animate-in fade-in zoom-in-95 duration-150",
        )}
      >
        <div className="border-b border-border">
          <CommandKit.Input
            placeholder={t("command.placeholder")}
            className={cn(
              "w-full h-11 px-4 bg-transparent",
              "text-md text-fg placeholder:text-fg-muted",
              "outline-none focus:outline-none",
            )}
          />
        </div>
        <CommandKit.List className="max-h-[60vh] overflow-y-auto p-1.5">
          <CommandKit.Empty className="py-8 text-center text-base text-fg-muted">
            {t("command.noResults")}
          </CommandKit.Empty>
          {grouped.map(([section, items]) => (
            <CommandKit.Group
              key={section}
              heading={t(`command.sections.${section}`)}
              className={cn(
                "[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5",
                "[&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-mono",
                "[&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:text-fg-muted",
                "[&_[cmdk-group-heading]]:tracking-wider",
              )}
            >
              {items.map((cmd) => (
                <CommandKit.Item
                  key={cmd.id}
                  value={`${cmd.label} ${cmd.keywords?.join(" ") ?? ""}`}
                  onSelect={() => {
                    recordRun(cmd.id)
                    close()
                    void cmd.run()
                  }}
                  className={cn(
                    "flex items-center gap-2 px-2 h-8 rounded-sm cursor-pointer",
                    "text-md text-fg",
                    "data-[selected=true]:bg-accent-subtle data-[selected=true]:text-fg",
                  )}
                >
                  {cmd.icon && (
                    <span className="text-fg-muted shrink-0" aria-hidden>
                      {cmd.icon}
                    </span>
                  )}
                  <span className="flex-1 truncate">{cmd.label}</span>
                  {cmd.shortcut && (
                    <Kbd keys={cmd.shortcut} size="sm" />
                  )}
                </CommandKit.Item>
              ))}
            </CommandKit.Group>
          ))}
        </CommandKit.List>
      </CommandKit>
    </div>
  )
}

/**
 * The baseline set of commands registered by the shell. Feature
 * modules can extend by calling useCommandStore().registerMany() in
 * a useEffect at mount time.
 */
function useBaselineCommands(): Command[] {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const { setTheme } = useTheme()
  const toggleAI = useAIStore((s) => s.toggle)

  return useMemo<Command[]>(() => {
    const allNav = [...PRIMARY_NAV, ...SECONDARY_NAV]
    const nav: Command[] = allNav.map((item) => ({
      id: `nav.${item.id}`,
      section: "navigation",
      label: t(item.i18nKey),
      keywords: [item.id, item.phase],
      run: () => navigate(item.href),
      hidden: item.disabled,
    }))

    const themeCmds: Command[] = [
      {
        id: "theme.light",
        section: "theme",
        label: t("command.actions.setThemeLight"),
        run: () => setTheme("light"),
      },
      {
        id: "theme.dark",
        section: "theme",
        label: t("command.actions.setThemeDark"),
        run: () => setTheme("dark"),
      },
      {
        id: "theme.system",
        section: "theme",
        label: t("command.actions.setThemeSystem"),
        run: () => setTheme("system"),
      },
    ]

    const langCmds: Command[] = SUPPORTED_LOCALES.map<Command>((locale) => ({
      id: `language.${locale}`,
      section: "language",
      label: LOCALE_NAMES[locale],
      keywords: [locale, t(`language.${locale}`)],
      run: () => void changeLanguage(locale as Locale),
      hidden: i18n.language.split("-")[0] === locale,
    }))

    const actions: Command[] = [
      {
        id: "action.ai-rail",
        section: "actions",
        label: t("shell.toggleAiRail"),
        shortcut: ["Cmd", "J"],
        run: toggleAI,
      },
      {
        id: "action.sign-out",
        section: "account",
        label: t("common.signOut"),
        run: () => void logout(),
      },
    ]

    return [...nav, ...actions, ...themeCmds, ...langCmds]
  }, [t, i18n.language, navigate, setTheme, toggleAI])
}
