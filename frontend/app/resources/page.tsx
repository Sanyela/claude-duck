"use client"

import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Download, Copy, Terminal, Code2, Eye, EyeOff } from "lucide-react"
import { getConfig } from "@/lib/env"
import { Button } from "@/components/ui/button"
import { useState, useEffect } from "react"
import { Alert, AlertDescription } from "@/components/ui/alert"

function StepCard({ 
  stepNumber, 
  title, 
  description, 
  icon: Icon, 
  children,
  cardColor = "bg-card"
}: { 
  stepNumber: number
  title: string | React.ReactNode
  description: string
  icon: React.ElementType
  children?: React.ReactNode
  cardColor?: string
}) {
  return (
    <Card className={`shadow-lg ${cardColor} text-card-foreground border-border hover:shadow-xl transition-shadow`}>
      <CardHeader className="pb-3">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-sky-500 text-white">
            <span className="text-sm font-bold">{stepNumber}</span>
          </div>
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <Icon className="h-5 w-5 text-sky-500 dark:text-sky-400" />
              {title}
            </CardTitle>
            <CardDescription className="text-sm text-muted-foreground">
              {description}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {children}
      </CardContent>
    </Card>
  )
}

function CodeBlock({ code, language = "bash" }: { code: string; language?: string }) {
  const [copied, setCopied] = useState(false)
  
  const copyCode = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  
  return (
    <div className="relative group">
      <div className="bg-slate-900 dark:bg-slate-800 rounded-lg p-4 font-mono text-sm overflow-x-auto">
        <div className="flex items-center justify-between mb-2">
          <span className="text-slate-400 text-xs uppercase tracking-wide">{language}</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={copyCode}
            className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-white h-6 px-2"
          >
            <Copy className="h-3 w-3 mr-1" />
            {copied ? "已复制" : "复制"}
          </Button>
        </div>
        <pre className="text-green-400">
          <code>{code}</code>
        </pre>
      </div>
    </div>
  )
}

function ConfigSection() {
  const [copied, setCopied] = useState(false)
  const [jwtToken, setJwtToken] = useState("your-jwt-token-here")
  const [screenshotMode, setScreenshotMode] = useState(false)
  const [apiUrl, setApiUrl] = useState("http://localhost:9998")
  const [claudeUrl, setClaudeUrl] = useState("https://api.anthropic.com")
  
  // 获取JWT token和API URL
  useEffect(() => {
    const loadData = async () => {
      if (typeof window !== "undefined") {
        const token = localStorage.getItem("auth_token")
        if (token) {
          setJwtToken(token)
        }
      }
      
      const config = await getConfig()
      setApiUrl(config.apiUrl)
      setClaudeUrl(config.claudeUrl)
    }
    loadData()
  }, [])
  
  // 实际的配置（包含真实的JWT token）
  const realConfig = `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "${jwtToken}",
    "ANTHROPIC_BASE_URL": "${claudeUrl}",
    "CLAUDE_CODE_MAX_OUTPUT_TOKENS": 64000,
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": 1
  },
  "permissions": {
    "allow": [],
    "deny": []
  },
  "apiKeyHelper": "echo '${jwtToken}'"
}`
  
  const copyConfig = () => {
    navigator.clipboard.writeText(realConfig)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  
  return (
    <div className="space-y-4">
      <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
        <h4 className="font-semibold text-amber-800 dark:text-amber-200 mb-2">配置文件路径</h4>
        <div className="space-y-2 text-sm text-amber-700 dark:text-amber-300">
          <p><strong>macOS/Linux:</strong> <code className="bg-amber-100 dark:bg-amber-900/50 px-1 rounded">~/.claude/settings.json</code></p>
          <p><strong>Windows:</strong> <code className="bg-amber-100 dark:bg-amber-900/50 px-1 rounded">C:\Users\[username]\.claude\settings.json</code></p>
        </div>
      </div>
      
      <div>
        <div className="flex items-center justify-between mb-3">
          <h4 className="font-semibold">配置文件内容</h4>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setScreenshotMode(!screenshotMode)}
              className="text-xs"
            >
              {screenshotMode ? <Eye className="h-3 w-3 mr-1" /> : <EyeOff className="h-3 w-3 mr-1" />}
              {screenshotMode ? "显示敏感信息" : "截图模式"}
            </Button>
            <Button
              variant="default"
              size="sm"
              onClick={copyConfig}
              disabled={copied}
              className="bg-sky-500 hover:bg-sky-600 text-white text-xs"
            >
              <Copy className="h-3 w-3 mr-1" />
              {copied ? "已复制" : "复制完整配置"}
            </Button>
          </div>
        </div>
        
        <div className="relative group">
          <div className="bg-slate-900 dark:bg-slate-800 rounded-lg p-4 font-mono text-sm overflow-x-auto">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-xs uppercase tracking-wide">json</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={copyConfig}
                className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-white h-6 px-2"
              >
                <Copy className="h-3 w-3 mr-1" />
                {copied ? "已复制" : "复制"}
              </Button>
            </div>
            <pre className="text-green-400">
              <code>
{`{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "`}<span className="group/token">
                  <span className={`transition-opacity ${screenshotMode ? 'group-hover/token:hidden' : ''}`}>
                    {screenshotMode ? "••••••••••••••••••••" : jwtToken}
                  </span>
                  {screenshotMode && (
                    <span className="hidden group-hover/token:inline transition-opacity">
                      {jwtToken}
                    </span>
                  )}
                </span>{`",
    "ANTHROPIC_BASE_URL": "`}<span className="group/url">
                  <span className={`transition-opacity ${screenshotMode ? 'group-hover/url:hidden' : ''}`}>
                    {screenshotMode ? "••••••••••••••••••••••••••••••" : claudeUrl}
                  </span>
                  {screenshotMode && (
                    <span className="hidden group-hover/url:inline transition-opacity">
                      {claudeUrl}
                    </span>
                  )}
                </span>{`",
    "CLAUDE_CODE_MAX_OUTPUT_TOKENS": 64000,
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": 1
  },
  "permissions": {
    "allow": [],
    "deny": []
  },
  "apiKeyHelper": "echo '`}<span className="group/helper">
                  <span className={`transition-opacity ${screenshotMode ? 'group-hover/helper:hidden' : ''}`}>
                    {screenshotMode ? "••••••••••••••••••••" : jwtToken}
                  </span>
                  {screenshotMode && (
                    <span className="hidden group-hover/helper:inline transition-opacity">
                      {jwtToken}
                    </span>
                  )}
                </span>{`'"
}`}
              </code>
            </pre>
          </div>
        </div>
        
        <Alert className="bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 mt-3">
          <AlertDescription className="text-red-700 dark:text-red-300">
            <strong>重要提示：</strong> 如果使用 Opus 模型，请将 <code className="bg-red-100 dark:bg-red-900/50 px-1 rounded">CLAUDE_CODE_MAX_OUTPUT_TOKENS</code> 设置为 <code className="bg-red-100 dark:bg-red-900/50 px-1 rounded">32000</code>
          </AlertDescription>
        </Alert>
      </div>
    </div>
  )
}

export default function ResourcesPage() {
  const [appName, setAppName] = useState('Duck Code')
  const [installCommand, setInstallCommand] = useState('npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com')
  const [docsUrl, setDocsUrl] = useState('https://github.com/anthropics/claude-code')
  
  useEffect(() => {
    const loadConfig = async () => {
      const config = await getConfig()
      setAppName(config.appName)
      setInstallCommand(config.installCommand)
      setDocsUrl(config.docsUrl)
    }
    loadConfig()
  }, [])
  
  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-2">{appName} 安装教程</h1>
        </div>

        <div className="space-y-6">
          {/* 步骤 1: 安装一键包 */}
          <StepCard
            stepNumber={1}
            title={
              <>
                安装 {appName} 一键包（
                <a 
                  href={docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  官方文档
                </a>
                ）
              </>
            }
            description={`安装 ${appName} 专用一键包`}
            icon={Download}
            cardColor="bg-amber-100 dark:bg-amber-900/20"
          >
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                使用 npm 全局安装 {appName} 一键包，确保您的网络连接正常
              </p>
              <CodeBlock 
                code={installCommand}
              />
            </div>
          </StepCard>

          {/* 步骤 2: 安装官方客户端 */}
          <StepCard
            stepNumber={2}
            title="安装官方客户端"
            description="安装 Claude Code 官方客户端并配置"
            icon={Terminal}
            cardColor="bg-sky-200 dark:bg-sky-900/20"
          >
            <div className="space-y-6">
              <div>
                <p className="text-sm text-muted-foreground mb-3">
                  安装 Claude Code 官方客户端：
                </p>
                <CodeBlock 
                  code="npm install -g @anthropic-ai/claude-code"
                />
              </div>
              
              <div>
                <h4 className="font-semibold mb-3">配置客户端</h4>
                <ConfigSection />
              </div>
            </div>
          </StepCard>
        </div>
      </div>
    </DashboardLayout>
  )
}
