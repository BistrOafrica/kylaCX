import { SignupForm } from "@/components/signup-form";
import ThemeToggle from "@/components/theme-toggle";

export default function SignupPage() {
  return (
    <div className="min-h-svh grid place-items-center bg-muted p-6 md:p-10">
      <div className="fixed top-6 right-6 md:top-10 md:right-10">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-sm md:max-w-4xl">
        <SignupForm />
      </div>
    </div>
  );
}
