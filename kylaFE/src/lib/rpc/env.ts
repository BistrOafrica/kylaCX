/**
 * Backend endpoints — single source of truth.
 *
 * F0 collapses the historical multi-hostname layout (VITE_ORG_HOSTNAME,
 * VITE_CRM_HOSTNAME, ...) into a single Envoy gateway URL. Set
 * `VITE_API_URL` in `.env.local`. The legacy multi-hostname variables
 * still work as a fallback so the demo pages keep loading.
 */

const fallback = "http://localhost:8000"

function readEnv(key: string): string | undefined {
  const value = import.meta.env[key]
  return typeof value === "string" && value.length > 0 ? value : undefined
}

export const env = {
  apiUrl:
    readEnv("VITE_API_URL") ??
    readEnv("VITE_ORG_HOSTNAME") ??
    fallback,

  // Per-domain overrides for the legacy multi-hostname layout.
  // New code should not branch on these — they only exist so the
  // legacy GlobalClients.ts keeps working until features replace it.
  legacy: {
    org:        readEnv("VITE_ORG_HOSTNAME"),
    crm:        readEnv("VITE_CRM_HOSTNAME"),
    call:       readEnv("VITE_CALL_HOSTNAME"),
    chatdesk:   readEnv("VITE_CHATDESK_HOSTNAME"),
    integ:      readEnv("VITE_INTEGRATION_GRPC_HOSTNAME"),
    qa:         readEnv("VITE_QA_HOSTNAME"),
    nia:        readEnv("VITE_NIA_HOSTNAME"),
    billing:    readEnv("VITE_BILLING_URL"),
  },

  sentry: {
    dsn:         readEnv("VITE_SENTRY_DSN"),
    environment: readEnv("VITE_ENVIRONMENT") ?? "local",
  },

  isDev: import.meta.env.DEV,
} as const
