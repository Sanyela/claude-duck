export const dynamic = 'force-dynamic'

import { AuthorizeFlow } from "@/components/oauth/authorize-flow"
import { Suspense } from "react"

export default function OAuthAuthorizePage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-100 dark:bg-slate-900 p-4">
      <Suspense fallback={<div className="text-slate-700 dark:text-slate-200">加载中...</div>}>
        <AuthorizeFlow />
      </Suspense>
    </div>
  )
}
