"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { KeyRound, Mail, Timer, AlertCircle, CheckCircle, Loader2 } from "lucide-react"
import Link from "next/link"
import { useSearchParams, useRouter } from "next/navigation"
import { login, sendVerificationCode, registerWithCode } from "@/api/auth"
import { useToast } from "@/components/ui/use-toast"
import { useAuth } from "@/contexts/AuthContext"
import { getEmailValidationError, getSupportedDomainsText } from "@/lib/email-validator"
import { Alert, AlertDescription } from "@/components/ui/alert"

export function LoginForm() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { toast } = useToast()
  const { login: authLogin } = useAuth()
  const initialTab = searchParams.get("tab") || "login"
  
  const [isLoading, setIsLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [errors, setErrors] = useState<{
    login?: string
    register?: string
    verification?: string
  }>({})
  
  // 邮箱验证状态
  const [emailValidation, setEmailValidation] = useState({
    login: { isValid: true, message: '' },
    register: { isValid: true, message: '' }
  })

  // 用户名验证状态
  const [usernameValidation, setUsernameValidation] = useState({
    isValid: true,
    message: ''
  })

  // 验证码输入状态
  const [verificationCode, setVerificationCode] = useState('')

  // 清除错误信息
  const clearErrors = () => {
    setErrors({})
  }

  // 验证验证码格式 (6位数字)
  const validateVerificationCode = (code: string): boolean => {
    const codeRegex = /^\d{6}$/
    return codeRegex.test(code)
  }

  // 处理错误消息，提供更友好的提示
  const getErrorMessage = (error: string): string => {
    const errorMap: { [key: string]: string } = {
      "邮箱或密码错误": "邮箱或密码错误，请检查后重试",
      "用户名或邮箱已存在": "该用户名或邮箱已被注册，请使用其他信息",
      "邮箱已被注册": "该邮箱已被注册，请使用其他邮箱或尝试登录",
      "邮箱未注册": "该邮箱尚未注册，请先注册账号",
      "验证码错误或已过期": "验证码错误或已过期，请重新获取验证码",
      "发送过于频繁，请稍后再试": "验证码发送过于频繁，请等待1分钟后重试",
      "不支持的邮箱域名": "不支持该邮箱域名，请使用支持的邮箱服务",
      "请求格式错误": "输入信息格式不正确，请检查后重试",
      "密码加密失败": "系统繁忙，请稍后重试",
      "创建用户失败": "注册失败，请稍后重试",
      "生成访问令牌失败": "登录成功但令牌生成失败，请重新登录",
      "验证码存储失败": "系统繁忙，验证码发送失败，请稍后重试",
      "验证码发送失败": "邮件发送失败，请检查邮箱地址或稍后重试"
    }
    
    // 尝试从错误映射中找到匹配的错误信息
    for (const [key, value] of Object.entries(errorMap)) {
      if (error.includes(key)) {
        return value
      }
    }
    
    return error || "操作失败，请稍后重试"
  }

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

  // 验证用户名
  const validateUsername = (username: string) => {
    let error = null
    let isValid = true
    
    if (!username) {
      error = "用户名不能为空"
      isValid = false
    } else if (username.length < 5) {
      error = "用户名至少需要5个字符"
      isValid = false
    } else if (username.length > 20) {
      error = "用户名不能超过20个字符"
      isValid = false
    } else if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
      error = "用户名只能包含字母、数字、下划线和连字符"
      isValid = false
    }
    
    setUsernameValidation({
      isValid,
      message: error || ''
    })
    
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
    if (!validateEmail(email, type)) {
      return
    }

    setIsLoading(true)
    clearErrors()
    
    try {
      const response = await sendVerificationCode({ email, type })
      
      if (response.success) {
        toast({
          title: "验证码已发送",
          description: "请查收邮件，验证码10分钟内有效",
          variant: "default",
        })
        startCountdown()
        setErrors(prev => ({ ...prev, verification: undefined }))
      } else {
        const errorMessage = getErrorMessage(response.message || "")
        setErrors(prev => ({ ...prev, verification: errorMessage }))
        toast({
          title: "验证码发送失败",
          description: errorMessage,
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("发送验证码出错:", error)
      const errorMessage = "网络错误，请检查网络连接后重试"
      setErrors(prev => ({ ...prev, verification: errorMessage }))
      toast({
        title: "网络错误",
        description: errorMessage,
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
    clearErrors()
    
    const formData = new FormData(e.currentTarget)
    const email = formData.get("email") as string
    const password = formData.get("password") as string
    
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
        
        if (response.token && response.user) {
          authLogin(response.token, response.user)
        }
        
        setTimeout(() => {
          router.push("/")
        }, 100)
      } else {
        const errorMessage = getErrorMessage(response.message || "")
        setErrors(prev => ({ ...prev, login: errorMessage }))
        toast({
          title: "登录失败",
          description: errorMessage,
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("登录出错:", error)
      const errorMessage = "网络错误，请检查网络连接后重试"
      setErrors(prev => ({ ...prev, login: errorMessage }))
      toast({
        title: "登录失败",
        description: errorMessage,
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  // 处理注册
  const handleSignup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setIsLoading(true)
    clearErrors()
    
    const formData = new FormData(e.currentTarget)
    const username = formData.get("username") as string
    const email = formData.get("email") as string
    const password = formData.get("password") as string
    const code = formData.get("code") as string
    
    const isUsernameValid = validateUsername(username)
    const isEmailValid = validateEmail(email, 'register')
    const isCodeValid = validateVerificationCode(code)
    
    if (!isUsernameValid || !isEmailValid) {
      setIsLoading(false)
      return
    }

    if (!isCodeValid) {
      setErrors(prev => ({ ...prev, register: "请输入6位数字验证码" }))
      setIsLoading(false)
      return
    }
    
    try {
      const response = await registerWithCode({ username, email, password, code })
      
      if (response.success) {
        toast({
          title: "注册成功",
          description: "正在跳转到主页...",
          variant: "default",
        })
        
        if (response.token && response.user) {
          authLogin(response.token, response.user)
        }
        
        setTimeout(() => {
          router.push("/")
        }, 100)
      } else {
        const errorMessage = getErrorMessage(response.message || "")
        setErrors(prev => ({ ...prev, register: errorMessage }))
        toast({
          title: "注册失败",
          description: errorMessage,
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error("注册出错:", error)
      const errorMessage = "网络错误，请检查网络连接后重试"
      setErrors(prev => ({ ...prev, register: errorMessage }))
      toast({
        title: "注册失败",
        description: errorMessage,
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="w-full max-w-md">
      <Card className="shadow-lg bg-card text-card-foreground border-border">
        <CardHeader className="space-y-4 text-center pb-6">
          <div className="flex justify-center">
            <div className="p-3 rounded-full bg-sky-100 dark:bg-sky-900/20">
              <KeyRound className="h-8 w-8 text-sky-600 dark:text-sky-400" />
            </div>
          </div>
          <div className="space-y-2">
            <CardTitle className="text-2xl font-bold tracking-tight">欢迎回来 👋</CardTitle>
            <CardDescription className="text-muted-foreground">
              登录或注册以继续使用我们的服务
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent>
          <Tabs defaultValue={initialTab} className="w-full" onValueChange={() => {
            clearErrors()
            setVerificationCode('')
          }}>
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">登录</TabsTrigger>
              <TabsTrigger value="register">注册</TabsTrigger>
            </TabsList>
            
            <TabsContent value="login" className="space-y-4">
              {errors.login && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{errors.login}</AlertDescription>
                </Alert>
              )}
              
              <form onSubmit={handleLogin} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email-login" className="text-sm font-medium">
                    邮箱
                  </Label>
                  <div className="relative">
                    <Input
                      id="email-login"
                      name="email"
                      type="email"
                      placeholder="you@example.com"
                      required
                      className={`h-10 ${
                        !emailValidation.login.isValid ? 'border-destructive focus-visible:ring-destructive' : ''
                      }`}
                      onChange={(e) => {
                        validateEmail(e.target.value, 'login')
                        clearErrors()
                      }}
                    />
                    {emailValidation.login.message && (
                      <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                        {emailValidation.login.isValid ? (
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        ) : (
                          <AlertCircle className="h-4 w-4 text-destructive" />
                        )}
                      </div>
                    )}
                  </div>
                  {!emailValidation.login.isValid && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      {emailValidation.login.message}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {getSupportedDomainsText()}
                  </p>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="password-login" className="text-sm font-medium">
                    密码
                  </Label>
                  <Input
                    id="password-login"
                    name="password"
                    type="password"
                    placeholder="••••••••"
                    required
                    className="h-10"
                    onChange={clearErrors}
                  />
                </div>
                
                <div className="flex items-center justify-end">
                  <Link 
                    href="#" 
                    className="text-sm text-sky-600 hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300 hover:underline" 
                    prefetch={false}
                  >
                    忘记密码?
                  </Link>
                </div>

                <Button 
                  type="submit" 
                  className="w-full h-10"
                  disabled={isLoading || !emailValidation.login.isValid}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      登录中...
                    </>
                  ) : (
                    <>
                      <Mail className="mr-2 h-4 w-4" /> 
                      登录
                    </>
                  )}
                </Button>
              </form>
            </TabsContent>
            
            <TabsContent value="register" className="space-y-4">
              {errors.register && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{errors.register}</AlertDescription>
                </Alert>
              )}
              
              {errors.verification && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{errors.verification}</AlertDescription>
                </Alert>
              )}
              
              <form onSubmit={handleSignup} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="username-register" className="text-sm font-medium">
                    用户名
                  </Label>
                  <div className="relative">
                    <Input
                      id="username-register"
                      name="username"
                      type="text"
                      placeholder="johndoe"
                      required
                      minLength={5}
                      maxLength={20}
                      pattern="^[a-zA-Z0-9_-]+$"
                      className={`h-10 ${
                        !usernameValidation.isValid ? 'border-destructive focus-visible:ring-destructive' : ''
                      }`}
                      onChange={(e) => {
                        validateUsername(e.target.value)
                        clearErrors()
                      }}
                    />
                    {usernameValidation.message && (
                      <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                        {usernameValidation.isValid ? (
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        ) : (
                          <AlertCircle className="h-4 w-4 text-destructive" />
                        )}
                      </div>
                    )}
                  </div>
                  {!usernameValidation.isValid && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      {usernameValidation.message}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    用户名长度为5-20个字符，只能包含字母、数字、下划线和连字符
                  </p>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="email-register" className="text-sm font-medium">
                    邮箱
                  </Label>
                  <div className="relative">
                    <Input
                      id="email-register"
                      name="email"
                      type="email"
                      placeholder="you@example.com"
                      required
                      className={`h-10 ${
                        !emailValidation.register.isValid ? 'border-destructive focus-visible:ring-destructive' : ''
                      }`}
                      onChange={(e) => {
                        validateEmail(e.target.value, 'register')
                        clearErrors()
                      }}
                    />
                    {emailValidation.register.message && (
                      <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                        {emailValidation.register.isValid ? (
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        ) : (
                          <AlertCircle className="h-4 w-4 text-destructive" />
                        )}
                      </div>
                    )}
                  </div>
                  {!emailValidation.register.isValid && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      {emailValidation.register.message}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {getSupportedDomainsText()}
                  </p>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="password-register" className="text-sm font-medium">
                    密码
                  </Label>
                  <Input
                    id="password-register"
                    name="password"
                    type="password"
                    placeholder="••••••••"
                    required
                    className="h-10"
                    onChange={clearErrors}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="code-register" className="text-sm font-medium">
                    验证码 <span className="text-destructive">*</span>
                  </Label>
                  <div className="flex gap-2">
                    <div className="relative flex-1">
                      <Input
                        id="code-register"
                        name="code"
                        type="text"
                        placeholder="请输入6位验证码"
                        maxLength={6}
                        required
                        className={`h-10 ${
                          verificationCode && !validateVerificationCode(verificationCode) 
                            ? 'border-destructive focus-visible:ring-destructive' 
                            : ''
                        }`}
                        value={verificationCode}
                        onChange={(e) => {
                          const value = e.target.value.replace(/\D/g, '') // 只允许数字
                          setVerificationCode(value)
                          clearErrors()
                        }}
                      />
                      {verificationCode && (
                        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                          {validateVerificationCode(verificationCode) ? (
                            <CheckCircle className="h-4 w-4 text-green-600" />
                          ) : (
                            <AlertCircle className="h-4 w-4 text-destructive" />
                          )}
                        </div>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={countdown > 0 || isLoading || !emailValidation.register.isValid}
                      onClick={() => {
                        const emailInput = document.getElementById("email-register") as HTMLInputElement
                        handleSendCode(emailInput?.value, "register")
                      }}
                      className="h-10 px-3 whitespace-nowrap"
                    >
                      {countdown > 0 ? (
                        <>
                          <Timer className="h-4 w-4 mr-1" />
                          {countdown}s
                        </>
                      ) : isLoading ? (
                        <>
                          <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                          发送中
                        </>
                      ) : (
                        "发送验证码"
                      )}
                    </Button>
                  </div>
                  {verificationCode && !validateVerificationCode(verificationCode) && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      验证码必须是6位数字
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    注册需要邮箱验证，请先输入有效邮箱后点击发送验证码
                  </p>
                </div>

                <Button 
                  type="submit" 
                  className="w-full h-10"
                  disabled={
                    isLoading || 
                    !emailValidation.register.isValid || 
                    !usernameValidation.isValid ||
                    !validateVerificationCode(verificationCode)
                  }
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      注册中...
                    </>
                  ) : (
                    "注册"
                  )}
                </Button>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}