import i18n from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import { initReactI18next } from "react-i18next"

import en from "@/locales/en.json"
import ar from "@/locales/ar.json"
import fr from "@/locales/fr.json"
import sw from "@/locales/sw.json"
import { applyDirectionForLocale } from "./direction"

/**
 * i18next configuration.
 *
 * Locales shipped in F0: en (default), ar (RTL), fr, sw.
 * Direction is managed by `applyDirectionForLocale` which sets
 * `<html dir="rtl|ltr">` on language change.
 *
 * Adding a new locale:
 *   1. Drop a JSON file in src/locales/
 *   2. Import it here and add it to `resources` + `supportedLngs`
 *   3. Update `LANGUAGE_DIRECTION` in direction.ts if it's RTL
 *   4. Add a name entry under `language.*` in every locale file
 */

export const SUPPORTED_LOCALES = ["en", "ar", "fr", "sw"] as const
export type Locale = (typeof SUPPORTED_LOCALES)[number]

export const DEFAULT_LOCALE: Locale = "en"

export const LOCALE_NAMES: Record<Locale, string> = {
  en: "English",
  ar: "العربية",
  fr: "Français",
  sw: "Kiswahili",
}

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      ar: { translation: ar },
      fr: { translation: fr },
      sw: { translation: sw },
    },
    fallbackLng: DEFAULT_LOCALE,
    supportedLngs: SUPPORTED_LOCALES,
    nonExplicitSupportedLngs: true,
    interpolation: { escapeValue: false },
    detection: {
      order: ["localStorage", "navigator", "htmlTag"],
      caches: ["localStorage"],
      lookupLocalStorage: "kyla.lang",
    },
    returnNull: false,
  })

// Apply direction on boot and whenever language changes.
applyDirectionForLocale(i18n.language)
i18n.on("languageChanged", (lng) => applyDirectionForLocale(lng))

export { i18n }

export async function changeLanguage(locale: Locale) {
  await i18n.changeLanguage(locale)
}
