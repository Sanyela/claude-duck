import { OAuthCallbackHandler } from "@/components/oauth/callback-handler"
import { Suspense } from "react"

export default function OAuthCallbackPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-100 dark:bg-slate-900 p-4">
      <Suspense fallback={<div className="text-slate-700 dark:text-slate-200">处理回调...</div>}>
        <OAuthCallbackHandler />
      </Suspense>
    </div>
  )
}
