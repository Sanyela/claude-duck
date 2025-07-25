export const dynamic = 'force-dynamic'

import { LoginForm } from "@/components/auth/login-form-api"

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <LoginForm />
    </div>
  )
}
