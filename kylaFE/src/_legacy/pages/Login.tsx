import { LoginForm } from "@/components/login-form"
import ThemeToggle from "@/components/theme-toggle"

export default function LoginPage() {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="fixed top-6 right-6 md:top-10 md:right-10">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-sm md:max-w-4xl">
        <LoginForm />
      </div>
    </div>
  )
}
