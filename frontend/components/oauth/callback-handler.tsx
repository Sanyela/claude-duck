"use client"

import { useSearchParams } from "next/navigation"
import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { CheckCircle, XCircle, Loader2 } from "lucide-react"

export function OAuthCallbackHandler() {
  const searchParams = useSearchParams()
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading")
  const [message, setMessage] = useState("正在处理授权回调...")
  const [token, setToken] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const tokenParam = searchParams.get("token")
    const stateParam = searchParams.get("state")
    const errorParam = searchParams.get("error")
    const errorDescriptionParam = searchParams.get("error_description")

    if (errorParam) {
      setError(errorParam)
      setMessage(errorDescriptionParam || `授权失败: ${errorParam}`)
      setStatus("error")
    } else if (tokenParam) {
      setToken(tokenParam)
      setMessage("授权成功！正在完成设置...")
      setStatus("success")
      // Save token to localStorage (conditionally, e.g. for specific ports or dev environments)
      // This is a simplified example. In production, consider security implications.
      if (window.location.port === "3000" || window.location.port === "3001") {
        // Example condition
        localStorage.setItem("oauth_token", tokenParam)
        if (stateParam) localStorage.setItem("oauth_state", stateParam)
      }
    } else {
      setMessage("无效的回调参数。")
      setStatus("error")
    }

    // Try to notify opener window (if this was a popup)
    if (window.opener) {
      window.opener.postMessage(
        {
          type: "oauth_callback",
          success: !errorParam && !!tokenParam,
          data: { token: tokenParam, state: stateParam, error: errorParam },
        },
        "*", // In production, specify the target origin
      )

      // Auto-close popup after a delay
      setTimeout(() => {
        // window.close(); // This might be blocked by browsers if not user-initiated
      }, 3000)
    }
  }, [searchParams])

  return (
    <Card className="w-full max-w-md shadow-lg bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
      <CardHeader className="text-center">
        {status === "loading" && <Loader2 className="mx-auto h-12 w-12 animate-spin text-sky-500" />}
        {status === "success" && <CheckCircle className="mx-auto h-12 w-12 text-green-500" />}
        {status === "error" && <XCircle className="mx-auto h-12 w-12 text-red-500" />}
        <CardTitle className="mt-4 text-2xl text-slate-900 dark:text-slate-100">
          {status === "success" ? "授权成功" : status === "error" ? "授权失败" : "处理中"}
        </CardTitle>
        <CardDescription className="text-slate-600 dark:text-slate-400">{message}</CardDescription>
      </CardHeader>
      <CardContent>
        {status === "success" && (
          <p className="text-sm text-center text-slate-700 dark:text-slate-300">
            您现在可以关闭此窗口。如果窗口没有自动关闭，请手动关闭。
          </p>
        )}
        {status === "error" && (
          <p className="text-sm text-center text-red-500 dark:text-red-400">请尝试重新授权或联系支持。</p>
        )}
        {/* For debugging, you might want to display the token or error */}
        {/* {token && <p className="text-xs break-all">Token: {token}</p>} */}
        {/* {error && <p className="text-xs break-all">Error: {error}</p>} */}
      </CardContent>
    </Card>
  )
}
