import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

// Self-hosted Geist Sans + Mono variable fonts (via Fontsource). The
// side-effect imports inject @font-face declarations; we read them via
// --font-sans / --font-mono in design-system/tokens/typography.css.
import "@fontsource-variable/geist"
import "@fontsource-variable/geist-mono"

// Tokens + base stylesheet. Order is documented in src/index.css.
import "./index.css"

// i18n must be initialized before any component renders so first paint
// has the active locale. Side-effect import only.
import "./lib/i18n/config"

import { AppProviders } from "./app/providers/AppProviders"
import { App } from "./App"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppProviders>
      <App />
    </AppProviders>
  </StrictMode>,
)
