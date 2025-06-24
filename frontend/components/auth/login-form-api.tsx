"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { KeyRound, Mail, Timer, AlertCircle, CheckCircle } from "lucide-react"
import Link from "next/link"
import { useSearchParams, useRouter } from "next/navigation"
import { Separator } from "@/components/ui/separator"
import { login, register, sendVerificationCode, registerWithCode, loginWithCode } from "@/api/auth"
import { useToast } from "@/components/ui/use-toast"
import { useAuth } from "@/contexts/AuthContext"
import { getEmailValidationError, getSupportedDomainsText, isAllowedEmailDomain } from "@/lib/email-validator"

export function LoginForm() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { toast } = useToast()
  const { login: authLogin } = useAuth()
  const initialTab = searchParams.get("tab") || "login"
  
  const [isLoading, setIsLoading] = useState(false)
  const [isCodeSent, setIsCodeSent] = useState(false)
  const [countdown, setCountdown] = useState(0)
  
  // 邮箱验证状态
  const [emailValidation, setEmailValidation] = useState({
    login: { isValid: true, message: '' },
    register: { isValid: true, message: '' }
  })

  // 验证邮箱并更新状态
  const validateEmail = (email: string, type: 'login' | 'register') => {
    const error = getEmailValidationError(email)
    const isValid = error === null
    
    setEmailValidation(prev => ({
      ...prev,
      [type]: {
        isValid,
        message: error || ''
      }
    }))
    
    return isValid
  }

  // 倒计时功能
  const startCountdown = () => {
    setCountdown(60)
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer)
          return 0
        }
        return prev - 1
      })
    }, 1000)
  }

  // 发送验证码
  const handleSendCode = async (email: string, type: 'register' | 'login') => {
    // 验证邮箱
    if (!validateEmail(email, type)) {
      return
    }

    setIsLoading(true)
    try {
      const response = await sendVerificationCode({ email, type })
      
      if (response.success) {
        toast({
          title: "验证码已发送",
          description: "请查收邮件，验证码10分钟内有效",
          variant: "default",
        })
        setIsCodeSent(true)
        startCountdown()
      } else {
        toast({
          title: "发送失败",
          description: response.message || "验证码发送失败",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("发送验证码出错:", error)
      toast({
        title: "发送失败",
        description: "服务器错误，请稍后再试",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  // 处理登录
  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setIsLoading(true)
    
    const formData = new FormData(e.currentTarget)
    const email = formData.get("email") as string
    const password = formData.get("password") as string
    
    // 验证邮箱
    if (!validateEmail(email, 'login')) {
      setIsLoading(false)
      return
    }
    
    try {
      const response = await login({ email, password })
      
      if (response.success) {
        toast({
          title: "登录成功",
          description: "正在跳转到主页...",
          variant: "default",
        })
        
        // 更新认证上下文
        if (response.token && response.user) {
          authLogin(response.token, response.user)
        }
        
        // 等待状态更新后再跳转
        setTimeout(() => {
          router.push("/")
        }, 100)
      } else {
        toast({
          title: "登录失败",
          description: response.message || "邮箱或密码错误",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("登录出错:", error)
      toast({
        title: "登录失败",
        description: "服务器错误，请稍后再试",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  // 处理注册 - 必须使用验证码
  const handleSignup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setIsLoading(true)
    
    const formData = new FormData(e.currentTarget)
    const username = formData.get("username") as string
    const email = formData.get("email") as string
    const password = formData.get("password") as string
    const code = formData.get("code") as string
    
    // 验证邮箱
    if (!validateEmail(email, 'register')) {
      setIsLoading(false)
      return
    }
    
    try {
      // 注册必须使用验证码
      const response = await registerWithCode({ username, email, password, code })
      
      if (response.success) {
        toast({
          title: "注册成功",
          description: "正在跳转到主页...",
          variant: "default",
        })
        
        // 更新认证上下文
        if (response.token && response.user) {
          authLogin(response.token, response.user)
        }
        
        // 等待状态更新后再跳转
        setTimeout(() => {
          router.push("/")
        }, 100)
      } else {
        toast({
          title: "注册失败",
          description: response.message || "注册信息有误",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("注册出错:", error)
      toast({
        title: "注册失败",
        description: "服务器错误，请稍后再试",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card className="w-full max-w-md shadow-2xl bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-50">
      <CardHeader className="text-center">
        <KeyRound className="mx-auto h-12 w-12 text-sky-400" />
        <CardTitle className="text-3xl font-bold mt-2">欢迎回来 👋</CardTitle>
        <CardDescription className="text-slate-600 dark:text-slate-400">登录或注册以继续使用我们的服务</CardDescription>
      </CardHeader>

      <Tabs defaultValue={initialTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2 bg-slate-200 dark:bg-slate-700">
          <TabsTrigger
            value="login"
            className="data-[state=active]:bg-sky-500 data-[state=active]:text-white dark:data-[state=active]:text-slate-50 text-slate-700 dark:text-slate-300"
          >
            登录
          </TabsTrigger>
          <TabsTrigger
            value="register"
            className="data-[state=active]:bg-sky-500 data-[state=active]:text-white dark:data-[state=active]:text-slate-50 text-slate-700 dark:text-slate-300"
          >
            注册
          </TabsTrigger>
        </TabsList>
        
        <TabsContent value="login">
          <form onSubmit={handleLogin}>
            <CardContent className="space-y-4 pt-6">
              <div className="space-y-2">
                <Label htmlFor="email-login" className="text-slate-700 dark:text-slate-300">
                  邮箱
                </Label>
                <div className="relative">
                  <Input
                    id="email-login"
                    name="email"
                    type="email"
                    placeholder="you@example.com"
                    required
                    className={`bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50 pr-10 ${
                      !emailValidation.login.isValid ? 'border-red-500 dark:border-red-400' : ''
                    }`}
                    onChange={(e) => validateEmail(e.target.value, 'login')}
                  />
                  {emailValidation.login.message && (
                    <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                      {emailValidation.login.isValid ? (
                        <CheckCircle className="h-4 w-4 text-green-500" />
                      ) : (
                        <AlertCircle className="h-4 w-4 text-red-500" />
                      )}
                    </div>
                  )}
                </div>
                {!emailValidation.login.isValid && (
                  <p className="text-sm text-red-500 dark:text-red-400 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {emailValidation.login.message}
                  </p>
                )}
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  {getSupportedDomainsText()}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="password-login" className="text-slate-700 dark:text-slate-300">
                  密码
                </Label>
                <Input
                  id="password-login"
                  name="password"
                  type="password"
                  placeholder="••••••••"
                  required
                  className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50"
                />
              </div>
              
              <div className="flex items-center justify-between">
                <Link href="#" className="text-sm text-sky-500 hover:underline dark:text-sky-400" prefetch={false}>
                  忘记密码?
                </Link>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button 
                type="submit" 
                className="w-full bg-sky-500 hover:bg-sky-600 text-white dark:text-slate-900"
                disabled={isLoading || !emailValidation.login.isValid}
              >
                {isLoading ? "登录中..." : (
                  <>
                    <Mail className="mr-2 h-4 w-4" /> 登录
                  </>
                )}
              </Button>
            </CardFooter>
          </form>
        </TabsContent>
        
        <TabsContent value="register">
          <form onSubmit={handleSignup}>
            <CardContent className="space-y-4 pt-6">
              <div className="space-y-2">
                <Label htmlFor="username-register" className="text-slate-700 dark:text-slate-300">
                  用户名
                </Label>
                <Input
                  id="username-register"
                  name="username"
                  type="text"
                  placeholder="johndoe"
                  required
                  className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email-register" className="text-slate-700 dark:text-slate-300">
                  邮箱
                </Label>
                <div className="relative">
                  <Input
                    id="email-register"
                    name="email"
                    type="email"
                    placeholder="you@example.com"
                    required
                    className={`bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50 pr-10 ${
                      !emailValidation.register.isValid ? 'border-red-500 dark:border-red-400' : ''
                    }`}
                    onChange={(e) => validateEmail(e.target.value, 'register')}
                  />
                  {emailValidation.register.message && (
                    <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                      {emailValidation.register.isValid ? (
                        <CheckCircle className="h-4 w-4 text-green-500" />
                      ) : (
                        <AlertCircle className="h-4 w-4 text-red-500" />
                      )}
                    </div>
                  )}
                </div>
                {!emailValidation.register.isValid && (
                  <p className="text-sm text-red-500 dark:text-red-400 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {emailValidation.register.message}
                  </p>
                )}
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  {getSupportedDomainsText()}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="password-register" className="text-slate-700 dark:text-slate-300">
                  密码
                </Label>
                <Input
                  id="password-register"
                  name="password"
                  type="password"
                  placeholder="••••••••"
                  required
                  className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50"
                />
              </div>
              
              {/* 注册必须要验证码 */}
              <div className="space-y-2">
                <Label htmlFor="code-register" className="text-slate-700 dark:text-slate-300">
                  验证码 <span className="text-red-500">*</span>
                </Label>
                <div className="flex space-x-2">
                  <Input
                    id="code-register"
                    name="code"
                    type="text"
                    placeholder="请输入6位验证码"
                    maxLength={6}
                    required
                    className="flex-1 bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={countdown > 0 || isLoading || !emailValidation.register.isValid}
                    onClick={() => {
                      const emailInput = document.getElementById("email-register") as HTMLInputElement
                      handleSendCode(emailInput?.value, "register")
                    }}
                    className="whitespace-nowrap"
                  >
                    {countdown > 0 ? (
                      <>
                        <Timer className="h-4 w-4 mr-1" />
                        {countdown}s
                      </>
                    ) : (
                      "发送验证码"
                    )}
                  </Button>
                </div>
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  注册需要邮箱验证，请先输入有效邮箱后点击发送验证码
                </p>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button 
                type="submit" 
                className="w-full bg-sky-500 hover:bg-sky-600 text-white dark:text-slate-900"
                disabled={isLoading || !emailValidation.register.isValid}
              >
                {isLoading ? "注册中..." : "注册"}
              </Button>
            </CardFooter>
          </form>
        </TabsContent>
      </Tabs>
    </Card>
  )
}