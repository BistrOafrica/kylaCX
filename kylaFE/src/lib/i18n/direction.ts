/**
 * Document direction control.
 *
 * Sets `<html dir>` and `<html lang>` based on the active locale.
 * Tailwind v4 `rtl:` / `ltr:` variants (defined in design-system/tokens/theme.css)
 * read from the `[dir]` attribute, so flipping it cascades.
 */

export const LANGUAGE_DIRECTION: Record<string, "ltr" | "rtl"> = {
  en: "ltr",
  fr: "ltr",
  sw: "ltr",
  ar: "rtl",
  // Add more RTL locales here (he, fa, ur) as they ship.
}

export function getDirectionForLocale(locale: string): "ltr" | "rtl" {
  // Normalize "ar-EG" → "ar"
  const base = locale.split("-")[0]?.toLowerCase() ?? "en"
  return LANGUAGE_DIRECTION[base] ?? "ltr"
}

export function applyDirectionForLocale(locale: string) {
  if (typeof document === "undefined") return
  const dir = getDirectionForLocale(locale)
  const base = locale.split("-")[0]?.toLowerCase() ?? "en"
  document.documentElement.setAttribute("dir", dir)
  document.documentElement.setAttribute("lang", base)
}
