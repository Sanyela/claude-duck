"use client"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { useSearchParams } from "next/navigation"
import { useEffect, useState } from "react"
import { ShieldCheck, Copy, AlertTriangle, ExternalLink } from "lucide-react"
import { useAuth } from "@/contexts/AuthContext"
import { authorize } from "@/api/auth"

export function AuthorizeFlow() {
  const searchParams = useSearchParams()
  const { user, isAuthenticated } = useAuth()
  const [clientId, setClientId] = useState("")
  const [redirectUri, setRedirectUri] = useState("")
  const [state, setState] = useState("")
  const [deviceFlow, setDeviceFlow] = useState(false)
  const [authCode, setAuthCode] = useState("")
  const [token, setToken] = useState("")
  const [error, setError] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [showCodeCopied, setShowCodeCopied] = useState(false)
  const [showToken, setShowToken] = useState(false)

  useEffect(() => {
    setClientId(searchParams.get("client_id") || "示例应用")
    setRedirectUri(searchParams.get("redirect_uri") || "")
    setState(searchParams.get("state") || "")
    setDeviceFlow(searchParams.get("device_flow") === "true")
  }, [searchParams])

  // 如果用户未登录，显示登录提示
  if (!isAuthenticated) {
    return (
      <Card className="w-full max-w-md shadow-lg bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader className="text-center">
          <AlertTriangle className="mx-auto h-10 w-10 text-yellow-500" />
          <CardTitle className="text-2xl text-slate-900 dark:text-slate-100">需要登录</CardTitle>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            请先登录您的账户以完成OAuth授权。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            onClick={() => window.location.href = '/login'}
            className="w-full bg-sky-500 hover:bg-sky-600 text-white"
          >
            前往登录
          </Button>
        </CardContent>
      </Card>
    )
  }

  const handleAuthorize = async () => {
    setIsLoading(true)
    setError("")
    setAuthCode("")
    setToken("")
    setShowToken(false)

    try {
      const response = await authorize({
        client_id: clientId,
        redirect_uri: redirectUri,
        response_type: "code",
        scope: "read",
        state: state,
        device_flow: deviceFlow
      })

      if (response.success) {
        if (deviceFlow) {
          setAuthCode(response.code!)
        } else {
          setToken(response.token!)
          // Attempt redirect
          if (redirectUri) {
            try {
              const callbackUrl = new URL(redirectUri)
              callbackUrl.searchParams.set("token", response.token!)
              if (state) {
                callbackUrl.searchParams.set("state", state)
              }

              // Try postMessage if opener exists (e.g. popup)
              if (window.opener) {
                window.opener.postMessage(
                  { type: "oauth_callback", success: true, data: { token: response.token, state: state } },
                  new URL(redirectUri).origin,
                )
                // Optionally close popup after postMessage
                // window.close();
                // return;
              }

              // Standard redirect
              window.location.href = callbackUrl.toString()
            } catch (e) {
              console.error("重定向失败:", e)
              setError("重定向失败。请手动复制 Token。")
              setShowToken(true)
            }
          } else {
            setError("缺少 redirect_uri，无法自动重定向。")
            setShowToken(true)
          }
        }
      } else {
        setError(response.message || "授权失败，请重试。")
      }
    } catch (error) {
      console.error("授权请求失败:", error)
      setError("授权请求失败，请重试。")
    } finally {
      setIsLoading(false)
    }
  }

  const handleDeny = () => {
    // Redirect to redirect_uri with error, or close window, or show message
    if (redirectUri) {
      const callbackUrl = new URL(redirectUri)
      callbackUrl.searchParams.set("error", "access_denied")
      if (state) callbackUrl.searchParams.set("state", state)
      window.location.href = callbackUrl.toString()
    } else {
      setError("授权已拒绝。您可以关闭此页面。")
    }
  }

  const handleCopyCode = async () => {
    if (!authCode) return
    try {
      await navigator.clipboard.writeText(authCode)
      setShowCodeCopied(true)
      setTimeout(() => setShowCodeCopied(false), 2000)
    } catch (err) {
      console.error("无法复制授权码: ", err)
      alert("无法复制授权码，请手动复制。")
    }
  }

  if (authCode) {
    return (
      <Card className="w-full max-w-md shadow-lg bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader className="text-center">
          <ShieldCheck className="mx-auto h-10 w-10 text-green-500" />
          <CardTitle className="text-2xl text-slate-900 dark:text-slate-100">授权码已生成</CardTitle>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            请复制以下授权码并粘贴到您的 {deviceFlow ? "CLI/终端应用" : "应用"} 中：
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="relative">
            <Input
              readOnly
              value={authCode}
              className="text-lg font-mono p-4 pr-12 text-center bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-200 border-slate-300 dark:border-slate-600"
            />
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-sky-500 dark:text-slate-400 dark:hover:text-sky-400"
              onClick={handleCopyCode}
            >
              <Copy className="h-5 w-5" />
              <span className="sr-only">复制授权码</span>
            </Button>
          </div>
          {showCodeCopied && (
            <p className="text-sm text-green-600 dark:text-green-400 text-center">授权码已复制到剪贴板！</p>
          )}
          <p className="text-xs text-slate-500 dark:text-slate-400 text-center">此授权码是一次性的，请妥善保管。</p>
        </CardContent>
      </Card>
    )
  }

  if (showToken && token) {
    return (
      <Card className="w-full max-w-md shadow-lg bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader className="text-center">
          <AlertTriangle className="mx-auto h-10 w-10 text-yellow-500" />
          <CardTitle className="text-2xl text-slate-900 dark:text-slate-100">重定向可能失败</CardTitle>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            自动重定向似乎未成功。请手动复制以下 Token 并粘贴到您的应用中，或尝试通过以下链接访问：
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="relative">
            <Input
              readOnly
              value={token}
              className="text-sm font-mono p-3 pr-12 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-200 border-slate-300 dark:border-slate-600 break-all"
            />
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-sky-500 dark:text-slate-400 dark:hover:text-sky-400"
              onClick={async () => {
                await navigator.clipboard.writeText(token)
                alert("Token 已复制!")
              }}
            >
              <Copy className="h-5 w-5" />
              <span className="sr-only">复制Token</span>
            </Button>
          </div>
          {redirectUri && (
            <Button
              variant="outline"
              className="w-full border-slate-300 text-slate-700 hover:bg-slate-100 dark:bg-slate-700 dark:border-slate-600 dark:hover:bg-slate-600 dark:text-slate-50"
              asChild
            >
              <a
                href={`${redirectUri}?token=${encodeURIComponent(token)}&state=${encodeURIComponent(state)}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                尝试手动打开回调链接 <ExternalLink className="ml-2 h-4 w-4" />
              </a>
            </Button>
          )}
          {error && <p className="text-sm text-red-600 dark:text-red-400 text-center">{error}</p>}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="w-full max-w-md shadow-lg bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
      <CardHeader>
        <ShieldCheck className="mx-auto h-10 w-10 text-sky-500" />
        <CardTitle className="text-center text-2xl text-slate-900 dark:text-slate-100">授权请求</CardTitle>
        <CardDescription className="text-center text-slate-600 dark:text-slate-400">
          应用 <strong className="text-sky-400">{clientId}</strong> 请求访问您的账户。
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-slate-700 dark:text-slate-300">此应用将能够：</p>
        <ul className="list-disc list-inside space-y-1 text-sm text-slate-600 dark:text-slate-400">
          <li>读取您的基本用户信息</li>
          <li>代表您执行操作</li>
        </ul>
        {deviceFlow && (
          <div className="rounded-md border border-yellow-500 bg-yellow-50 p-3 dark:bg-yellow-900/30 dark:border-yellow-700">
            <p className="text-sm text-yellow-700 dark:text-yellow-300">
              <AlertTriangle className="inline h-4 w-4 mr-1" />
              您正在使用设备流程。授权后，您将获得一个授权码，需要手动输入到您的设备或应用中。
            </p>
          </div>
        )}
        {user && (
          <div className="rounded-md border border-blue-500 bg-blue-50 p-3 dark:bg-blue-900/30 dark:border-blue-700">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              当前登录用户：<strong>{user.username ? user.username : (user.email || '未知用户')}</strong>
            </p>
          </div>
        )}
        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
      </CardContent>
      <CardFooter className="flex flex-col sm:flex-row gap-2">
        <Button
          onClick={handleAuthorize}
          className="w-full sm:w-auto flex-1 bg-sky-500 hover:bg-sky-600 text-white dark:text-slate-900"
          disabled={isLoading}
        >
          {isLoading ? "处理中..." : "授权"}
        </Button>
        <Button
          onClick={handleDeny}
          variant="outline"
          className="w-full sm:w-auto flex-1 border-slate-300 text-slate-700 hover:bg-slate-100 dark:bg-slate-700 dark:border-slate-600 dark:hover:bg-slate-600 dark:text-slate-50"
          disabled={isLoading}
        >
          拒绝
        </Button>
      </CardFooter>
    </Card>
  )
}
