import { useTranslation } from "react-i18next"
import { IconLogout, IconUser, IconLanguage } from "@tabler/icons-react"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu"
import { useTheme } from "@/components/theme-provider"
import { logout } from "@/lib/auth"
import { changeLanguage, LOCALE_NAMES, SUPPORTED_LOCALES, type Locale } from "@/lib/i18n"
import type { AuthIdentity } from "@/lib/auth"
import { cn } from "@/lib/utils"
import { useTranslation as useT } from "react-i18next"

export function ProfileMenu({ identity }: { identity: AuthIdentity | null }) {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  const { i18n } = useT()

  const initials = (identity?.displayName ?? identity?.email ?? "?")
    .split(/\s+/)
    .map((part: string) => part[0] ?? "")
    .slice(0, 2)
    .join("")
    .toUpperCase()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={t("common.profile")}
        className={cn(
          "inline-flex items-center justify-center size-7 rounded-full",
          "bg-accent-subtle text-fg font-medium text-sm",
          "hover:bg-accent-muted transition-colors",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent",
        )}
      >
        {initials}
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" sideOffset={8} className="w-56">
        <DropdownMenuLabel className="font-normal">
          <div className="space-y-0.5">
            <div className="text-md font-medium truncate">
              {identity?.displayName ?? identity?.email ?? identity?.userId}
            </div>
            {identity?.email && identity.displayName && (
              <div className="text-sm text-fg-muted truncate">
                {identity.email}
              </div>
            )}
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        <DropdownMenuItem>
          <IconUser className="size-3.5" />
          <span>{t("common.profile")}</span>
        </DropdownMenuItem>

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <IconLanguage className="size-3.5" />
            <span>{t("command.actions.changeLanguage")}</span>
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuRadioGroup
              value={i18n.language.split("-")[0]}
              onValueChange={(value) => void changeLanguage(value as Locale)}
            >
              {SUPPORTED_LOCALES.map((locale) => (
                <DropdownMenuRadioItem key={locale} value={locale}>
                  {LOCALE_NAMES[locale]}
                </DropdownMenuRadioItem>
              ))}
            </DropdownMenuRadioGroup>
          </DropdownMenuSubContent>
        </DropdownMenuSub>

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <span className="text-sm font-mono uppercase">A</span>
            <span>Theme</span>
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuRadioGroup
              value={theme}
              onValueChange={(v) => setTheme(v as "light" | "dark" | "system")}
            >
              <DropdownMenuRadioItem value="light">
                {t("command.actions.setThemeLight")}
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="dark">
                {t("command.actions.setThemeDark")}
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="system">
                {t("command.actions.setThemeSystem")}
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </DropdownMenuSubContent>
        </DropdownMenuSub>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          onClick={() => void logout()}
          className="text-danger focus:bg-danger-subtle focus:text-danger"
        >
          <IconLogout className="size-3.5" />
          <span>{t("common.signOut")}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
