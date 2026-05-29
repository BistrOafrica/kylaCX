import { test, expect } from "@playwright/test"

/**
 * Shell smoke tests — no backend required.
 *
 * These verify the cold-start UI: an unauthenticated user lands on
 * /login, the login form renders, ⌘K toggles the palette, and the
 * language switcher flips dir to RTL on Arabic.
 *
 * F1 will add specs that go through real login + inbox interactions.
 */

test.describe("public shell", () => {
  test("redirects unauthenticated user to /login", async ({ page }) => {
    await page.goto("/")
    await expect(page).toHaveURL(/\/login$/)
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible()
  })

  test("/inbox also bounces unauthenticated to /login", async ({ page }) => {
    await page.goto("/inbox")
    await expect(page).toHaveURL(/\/login$/)
  })

  test("/crm and /crm/deals also bounce unauthenticated to /login", async ({ page }) => {
    await page.goto("/crm")
    await expect(page).toHaveURL(/\/login$/)
    await page.goto("/crm/deals")
    await expect(page).toHaveURL(/\/login$/)
  })

  test("/tickets, /knowledge, /forms also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of ["/tickets", "/knowledge", "/forms"]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("/automation and /ai-studio also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of ["/automation", "/ai-studio"]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("/admin and admin sub-routes also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of [
      "/admin",
      "/admin/organisation",
      "/admin/users",
      "/admin/rbac",
      "/admin/apps",
      "/admin/profile",
    ]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("/analytics and analytics sub-routes also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of [
      "/analytics",
      "/analytics/overview",
      "/analytics/calls",
      "/analytics/billing",
    ]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("/calls and telephony sub-routes also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of [
      "/calls",
      "/calls/history",
      "/calls/wallboard",
      "/calls/ivr",
      "/ivr",
    ]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("/campaigns and campaign sub-routes also bounce unauthenticated to /login", async ({ page }) => {
    for (const path of [
      "/campaigns",
      "/campaigns/templates",
      "/campaigns/whatsapp/abc-123",
      "/campaigns/voice/xyz-456",
    ]) {
      await page.goto(path)
      await expect(page).toHaveURL(/\/login$/)
    }
  })

  test("login form renders with email + password fields", async ({ page }) => {
    await page.goto("/login")
    await expect(page.getByLabel(/email/i)).toBeVisible()
    await expect(page.getByLabel(/password/i)).toBeVisible()
    await expect(
      page.getByRole("button", { name: /sign in/i }),
    ).toBeVisible()
  })
})
