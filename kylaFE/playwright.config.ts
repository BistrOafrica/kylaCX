import { defineConfig, devices } from "@playwright/test"

/**
 * Playwright E2E config.
 *
 * Tests live in tests/e2e/ and run against a freshly-started dev
 * server. F0 specs cover login flow + shell shortcuts + i18n
 * direction switch. Backend RPC calls are mocked at the fetch
 * boundary in tests/e2e/mocks/ so specs are hermetic.
 */
export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? "github" : "html",
  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
  ],
  webServer: {
    command: "pnpm dev --port 5173",
    url: "http://localhost:5173",
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
  },
})
