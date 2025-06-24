import { LoginForm } from "@/components/auth/login-form"

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-100 dark:bg-gradient-to-br dark:from-slate-900 dark:to-slate-800 p-4">
      <LoginForm />
    </div>
  )
}
