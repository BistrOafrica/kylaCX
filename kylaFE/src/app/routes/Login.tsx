import { useState, type FormEvent } from "react"
import { useNavigate, useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { IconLoader2 } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Surface } from "@/design-system"
import { login } from "@/lib/auth"
import { RpcError } from "@/lib/rpc"
import { cn } from "@/lib/utils"

/**
 * Login — wired to AuthService.Login via lib/auth/api.ts.
 *
 * On success the auth store updates and the user is redirected to
 * either `state.from` (where they originally tried to go) or `/`.
 * Failure shows an inline error message; no toasts here so the user
 * stays focused on the form.
 */
export function Login() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await login(email, password)
      const redirect =
        (location.state as { from?: string } | null)?.from ?? "/"
      navigate(redirect, { replace: true })
    } catch (err) {
      const message =
        err instanceof RpcError ? err.message : t("auth.login.error")
      setError(message || t("auth.login.error"))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-dvh flex items-center justify-center bg-canvas p-6">
      <div className="w-full max-w-sm space-y-6">
        <Header />

        <Surface level={2} radius="lg" className="p-6 space-y-5">
          <div className="space-y-1.5">
            <h1 className="text-xl font-semibold text-fg tracking-tight">
              {t("auth.login.title")}
            </h1>
            <p className="text-base text-fg-muted">
              {t("auth.login.subtitle")}
            </p>
          </div>

          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <Field
              id="email"
              label={t("auth.login.emailLabel")}
              type="email"
              value={email}
              onChange={setEmail}
              placeholder={t("auth.login.emailPlaceholder")}
              autoComplete="email"
              required
            />
            <Field
              id="password"
              label={t("auth.login.passwordLabel")}
              type="password"
              value={password}
              onChange={setPassword}
              placeholder={t("auth.login.passwordPlaceholder")}
              autoComplete="current-password"
              required
              trailing={
                <a
                  href="#"
                  className="text-sm text-fg-link hover:underline"
                >
                  {t("auth.login.forgot")}
                </a>
              }
            />

            {error && (
              <div
                role="alert"
                className={cn(
                  "rounded-sm border border-danger/30 bg-danger-subtle",
                  "px-3 py-2 text-base text-danger",
                )}
              >
                {error}
              </div>
            )}

            <Button
              type="submit"
              disabled={submitting || !email || !password}
              className="w-full h-9"
            >
              {submitting ? (
                <>
                  <IconLoader2 className="size-3.5 animate-spin" />
                  {t("auth.login.submitting")}
                </>
              ) : (
                t("auth.login.submit")
              )}
            </Button>
          </form>
        </Surface>

        <p className="text-center text-base text-fg-muted">
          {t("auth.login.noAccount")}{" "}
          <a href="/signup" className="text-fg-link hover:underline">
            {t("auth.login.createAccount")}
          </a>
        </p>
      </div>
    </div>
  )
}

function Header() {
  return (
    <div className="flex flex-col items-center gap-2">
      <span
        className={cn(
          "inline-flex items-center justify-center size-9 rounded-md",
          "bg-accent text-accent-fg text-lg font-semibold",
        )}
        aria-hidden
      >
        K
      </span>
      <span className="font-mono text-xs uppercase tracking-widest text-fg-muted">
        Kyla
      </span>
    </div>
  )
}

interface FieldProps {
  id: string
  label: string
  type?: string
  value: string
  onChange: (v: string) => void
  placeholder?: string
  autoComplete?: string
  required?: boolean
  trailing?: React.ReactNode
}

function Field({
  id,
  label,
  type = "text",
  value,
  onChange,
  placeholder,
  autoComplete,
  required,
  trailing,
}: FieldProps) {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <Label htmlFor={id} className="text-sm font-medium">
          {label}
        </Label>
        {trailing}
      </div>
      <Input
        id={id}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        autoComplete={autoComplete}
        required={required}
      />
    </div>
  )
}
