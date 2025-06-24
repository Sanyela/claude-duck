"use client"

import { useState } from "react"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { LinkIcon, Copy, Check } from "lucide-react"
import Link from "next/link"

export function OAuthLinkGenerator() {
  const [clientId, setClientId] = useState("my-awesome-app")
  const [redirectUri, setRedirectUri] = useState("https://oauth.pstmn.io/v1/callback")
  const [state, setState] = useState("random_state_string")
  const [deviceFlow, setDeviceFlow] = useState(false)
  const [generatedUrl, setGeneratedUrl] = useState("")
  const [isCopied, setIsCopied] = useState(false)

  const handleGenerateLink = () => {
    const params = new URLSearchParams()
    if (clientId) params.append("client_id", clientId)
    if (redirectUri) params.append("redirect_uri", redirectUri)
    if (state) params.append("state", state)
    if (deviceFlow) params.append("device_flow", "true")

    const url = `/oauth/authorize?${params.toString()}`
    setGeneratedUrl(url)
    setIsCopied(false)
  }

  const handleCopy = () => {
    if (!generatedUrl) return
    const fullUrl = window.location.origin + generatedUrl
    navigator.clipboard.writeText(fullUrl).then(() => {
      setIsCopied(true)
      setTimeout(() => setIsCopied(false), 2000)
    })
  }

  return (
    <div className="max-w-2xl mx-auto">
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader>
          <CardTitle className="text-slate-900 dark:text-slate-100">OAuth 授权链接生成器</CardTitle>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            使用此工具生成不同参数的 OAuth 授权页面链接，以测试其显示样式。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="client_id" className="text-slate-700 dark:text-slate-300">
              Client ID
            </Label>
            <Input
              id="client_id"
              value={clientId}
              onChange={(e) => setClientId(e.target.value)}
              placeholder="例如：my-awesome-app"
              className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="redirect_uri" className="text-slate-700 dark:text-slate-300">
              Redirect URI
            </Label>
            <Input
              id="redirect_uri"
              value={redirectUri}
              onChange={(e) => setRedirectUri(e.target.value)}
              placeholder="例如：https://myapp.com/callback"
              className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="state" className="text-slate-700 dark:text-slate-300">
              State
            </Label>
            <Input
              id="state"
              value={state}
              onChange={(e) => setState(e.target.value)}
              placeholder="随机状态字符串"
              className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600"
            />
          </div>
          <div className="flex items-center space-x-2">
            <Switch
              id="device_flow"
              checked={deviceFlow}
              onCheckedChange={setDeviceFlow}
              className="data-[state=checked]:bg-sky-500"
            />
            <Label htmlFor="device_flow" className="text-slate-700 dark:text-slate-300">
              启用设备流程 (Device Flow)
            </Label>
          </div>
        </CardContent>
        <CardFooter className="flex-col items-start gap-4 border-t border-slate-200 dark:border-slate-700 pt-4">
          <Button onClick={handleGenerateLink} className="w-full sm:w-auto bg-sky-500 hover:bg-sky-600 text-white">
            <LinkIcon className="mr-2 h-4 w-4" /> 生成链接
          </Button>
          {generatedUrl && (
            <div className="w-full space-y-2 p-4 bg-slate-100 dark:bg-slate-900 rounded-md">
              <Label className="text-slate-700 dark:text-slate-300">生成的链接:</Label>
              <div className="flex items-center gap-2">
                <Input
                  readOnly
                  value={generatedUrl}
                  className="bg-white dark:bg-slate-800 border-slate-300 dark:border-slate-700 font-mono text-sm"
                />
                <Button variant="outline" size="icon" onClick={handleCopy}>
                  {isCopied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                  <span className="sr-only">复制</span>
                </Button>
              </div>
              <Button
                variant="default"
                asChild
                className="bg-slate-900 text-white hover:bg-slate-700 dark:bg-slate-50 dark:text-slate-900 dark:hover:bg-slate-200"
              >
                <Link href={generatedUrl} target="_blank" rel="noopener noreferrer">
                  在新标签页中打开
                </Link>
              </Button>
            </div>
          )}
        </CardFooter>
      </Card>
    </div>
  )
}
