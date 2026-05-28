import type { StorybookConfig } from "@storybook/react-vite"
import path from "path"

/**
 * Storybook 8 + Vite 7 config for the design system.
 *
 * Stories live alongside the components they document under
 * src/design-system/. Token + density pages live under
 * src/design-system/stories/ as MDX.
 */
const config: StorybookConfig = {
  stories: [
    "../src/design-system/**/*.stories.@(ts|tsx|mdx)",
    "../src/design-system/**/*.mdx",
  ],
  addons: [
    "@storybook/addon-essentials",
    "@storybook/addon-a11y",
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  typescript: {
    reactDocgen: "react-docgen-typescript",
  },
  async viteFinal(viteConfig) {
    viteConfig.resolve ??= {}
    viteConfig.resolve.alias = {
      ...(viteConfig.resolve.alias as object),
      "@": path.resolve(__dirname, "../src"),
    }
    return viteConfig
  },
}

export default config
