import type { Preview, Decorator } from "@storybook/react"
import { useEffect } from "react"

// Load fonts + tokens at the Storybook preview level so stories see
// the same visual environment as the real app.
import "@fontsource-variable/geist"
import "@fontsource-variable/geist-mono"
import "../src/index.css"

/**
 * Apply the active theme by toggling the `dark` class on <html>.
 * Storybook's globals control which mode is active.
 */
function ThemeFrame({
  theme,
  direction,
  children,
}: {
  theme: "light" | "dark"
  direction: "ltr" | "rtl"
  children: React.ReactNode
}) {
  useEffect(() => {
    const root = document.documentElement
    root.classList.toggle("dark", theme === "dark")
    root.setAttribute("dir", direction)
    return () => {
      root.classList.remove("dark")
      root.setAttribute("dir", "ltr")
    }
  }, [theme, direction])

  return <>{children}</>
}

const withTheme: Decorator = (Story, context) => {
  const theme = (context.globals.theme as "light" | "dark") ?? "light"
  const direction = (context.globals.direction as "ltr" | "rtl") ?? "ltr"
  return (
    <ThemeFrame theme={theme} direction={direction}>
      <Story />
    </ThemeFrame>
  )
}

const preview: Preview = {
  parameters: {
    backgrounds: { disable: true },
    layout: "padded",
    controls: {
      matchers: { color: /(background|color)$/i, date: /Date$/i },
    },
    a11y: { config: { rules: [{ id: "color-contrast", enabled: true }] } },
  },
  globalTypes: {
    theme: {
      name: "Theme",
      description: "Light or dark theme",
      defaultValue: "light",
      toolbar: {
        icon: "circlehollow",
        items: [
          { value: "light", title: "Light" },
          { value: "dark",  title: "Dark"  },
        ],
        dynamicTitle: true,
      },
    },
    direction: {
      name: "Direction",
      description: "Layout direction",
      defaultValue: "ltr",
      toolbar: {
        icon: "transfer",
        items: [
          { value: "ltr", title: "LTR" },
          { value: "rtl", title: "RTL" },
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [withTheme],
}

export default preview
