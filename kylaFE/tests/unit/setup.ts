// Module-level: ensure localStorage is a real Storage shim before any
// zustand-persist store is created at import time. setupFiles run
// before test modules are imported, so this runs first.
if (
  typeof window !== "undefined" &&
  (typeof window.localStorage === "undefined" ||
    typeof window.localStorage.setItem !== "function")
) {
  const memory = new Map<string, string>()
  const fake: Storage = {
    getItem: (k) => memory.get(k) ?? null,
    setItem: (k, v) => void memory.set(k, String(v)),
    removeItem: (k) => void memory.delete(k),
    clear: () => memory.clear(),
    key: (i) => Array.from(memory.keys())[i] ?? null,
    get length() {
      return memory.size
    },
  }
  Object.defineProperty(window, "localStorage", {
    value: fake,
    writable: true,
    configurable: true,
  })
}

import "@testing-library/jest-dom/vitest"
import { afterEach } from "vitest"
import { cleanup } from "@testing-library/react"

afterEach(() => {
  cleanup()
})
