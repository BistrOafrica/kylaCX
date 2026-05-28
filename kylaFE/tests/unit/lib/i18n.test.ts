import { describe, it, expect } from "vitest"
import {
  getDirectionForLocale,
  LANGUAGE_DIRECTION,
} from "@/lib/i18n/direction"

describe("i18n direction", () => {
  it("Arabic is RTL", () => {
    expect(getDirectionForLocale("ar")).toBe("rtl")
    expect(getDirectionForLocale("ar-EG")).toBe("rtl")
  })

  it("English, French, Swahili are LTR", () => {
    expect(getDirectionForLocale("en")).toBe("ltr")
    expect(getDirectionForLocale("fr")).toBe("ltr")
    expect(getDirectionForLocale("sw")).toBe("ltr")
  })

  it("falls back to LTR for unknown locales", () => {
    expect(getDirectionForLocale("xx")).toBe("ltr")
    expect(getDirectionForLocale("")).toBe("ltr")
  })

  it("registry only flags known RTL locales", () => {
    expect(LANGUAGE_DIRECTION.ar).toBe("rtl")
    expect(LANGUAGE_DIRECTION.en).toBe("ltr")
  })
})
