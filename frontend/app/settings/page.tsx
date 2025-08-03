"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { User, Lock, Mail, Timer } from "lucide-react"
import { useAuth } from "@/contexts/AuthContext"
import { toast } from "sonner"
import { request } from "@/api/request"

export default function SettingsPage() {
  const { user, logout } = useAuth()
  const [isLoading, setIsLoading] = useState(false)
  
  // 用户名修改相关状态
  const [newUsername, setNewUsername] = useState("")
  const [usernameVerificationCode, setUsernameVerificationCode] = useState("")
  const [usernameCodeSent, setUsernameCodeSent] = useState(false)
  const [usernameCountdown, setUsernameCountdown] = useState(0)
  const [usernameValid, setUsernameValid] = useState(false)
  const [usernameChecking, setUsernameChecking] = useState(false)
  const [usernameError, setUsernameError] = useState("")
  
  // 密码修改相关状态
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [passwordVerificationCode, setPasswordVerificationCode] = useState("")
  const [passwordCodeSent, setPasswordCodeSent] = useState(false)
  const [passwordCountdown, setPasswordCountdown] = useState(0)
  const [passwordValid, setPasswordValid] = useState(false)
  const [passwordError, setPasswordError] = useState("")

  // 检查用户名合法性
  const checkUsername = async (username: string) => {
    if (!username.trim()) {
      setUsernameValid(false)
      setUsernameError("")
      return
    }
    
    if (username === user?.username) {
      setUsernameValid(false)
      setUsernameError("新用户名不能与当前用户名相同")
      return
    }

    try {
      setUsernameChecking(true)
      setUsernameError("")
      
      await request.post("/auth/check-username", {
        username: username
      })
      
      setUsernameValid(true)
      setUsernameError("")
    } catch (error: any) {
      setUsernameValid(false)
      setUsernameError(error.response?.data?.message || "用户名不可用")
    } finally {
      setUsernameChecking(false)
    }
  }

  // 检查密码合法性
  const checkPassword = () => {
    if (!newPassword) {
      setPasswordValid(false)
      setPasswordError("")
      return
    }
    
    if (newPassword.length < 6) {
      setPasswordValid(false)
      setPasswordError("密码长度至少为6位")
      return
    }
    
    if (newPassword !== confirmPassword) {
      setPasswordValid(false)
      setPasswordError("新密码和确认密码不匹配")
      return
    }
    
    setPasswordValid(true)
    setPasswordError("")
  }

  // 用户名输入变化时的处理
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (newUsername) {
        checkUsername(newUsername)
      }
    }, 500) // 500ms 防抖

    return () => clearTimeout(timeoutId)
  }, [newUsername, user?.username])

  // 密码输入变化时的处理
  useEffect(() => {
    checkPassword()
  }, [newPassword, confirmPassword])

  // 发送用户名修改验证码
  const sendUsernameVerificationCode = async () => {
    if (!usernameValid) {
      toast.error("请输入有效的用户名")
      return
    }

    try {
      setIsLoading(true)
      
      // 发送验证码
      await request.post("/auth/send-verification-code-settings", {
        type: "change_username",
        new_username: newUsername
      })
      
      setUsernameCodeSent(true)
      setUsernameCountdown(60)
      toast.success("验证码已发送到您的邮箱")

      // 开始倒计时
      const timer = setInterval(() => {
        setUsernameCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            setUsernameCodeSent(false)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch (error: any) {
      toast.error(error.response?.data?.message || "操作失败")
    } finally {
      setIsLoading(false)
    }
  }

  // 发送密码修改验证码
  const sendPasswordVerificationCode = async () => {
    if (!passwordValid) {
      toast.error("请输入有效的密码")
      return
    }

    try {
      setIsLoading(true)
      await request.post("/auth/send-verification-code-settings", {
        type: "change_password"
      })
      
      setPasswordCodeSent(true)
      setPasswordCountdown(60)
      toast.success("验证码已发送到您的邮箱")

      // 开始倒计时
      const timer = setInterval(() => {
        setPasswordCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            setPasswordCodeSent(false)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch (error: any) {
      toast.error(error.response?.data?.message || "发送验证码失败")
    } finally {
      setIsLoading(false)
    }
  }

  // 提交用户名修改
  const handleUsernameChange = async () => {
    if (!usernameVerificationCode) {
      toast.error("请输入验证码")
      return
    }

    try {
      setIsLoading(true)
      await request.post("/auth/change-username", {
        new_username: newUsername,
        verification_code: usernameVerificationCode
      })
      
      toast.success("用户名修改成功")
      setNewUsername("")
      setUsernameVerificationCode("")
      setUsernameCodeSent(false)
      setUsernameCountdown(0)
      // 刷新用户信息
      window.location.reload()
    } catch (error: any) {
      toast.error(error.response?.data?.message || "用户名修改失败")
    } finally {
      setIsLoading(false)
    }
  }

  // 提交密码修改
  const handlePasswordChange = async () => {
    if (!passwordVerificationCode) {
      toast.error("请输入验证码")
      return
    }

    try {
      setIsLoading(true)
      await request.post("/auth/change-password", {
        new_password: newPassword,
        verification_code: passwordVerificationCode
      })
      
      toast.success("密码修改成功，请重新登录")
      setNewPassword("")
      setConfirmPassword("")
      setPasswordVerificationCode("")
      setPasswordCodeSent(false)
      setPasswordCountdown(0)
      
      // 密码修改成功后，3秒后自动退出登录
      setTimeout(() => {
        logout()
      }, 3000)
    } catch (error: any) {
      toast.error(error.response?.data?.message || "密码修改失败")
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <DashboardLayout>
      <div className="space-y-8 max-w-3xl mx-auto">
        {/* 用户名修改 */}
        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <div className="flex items-center gap-3">
              <User className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              <CardTitle>修改用户名</CardTitle>
            </div>
            <CardDescription>修改您的用户名需要邮箱验证。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="new-username">新用户名</Label>
              <Input
                id="new-username"
                value={newUsername}
                onChange={(e) => setNewUsername(e.target.value)}
                placeholder={`当前用户名: ${user?.username || ""}`}
                className={`bg-input border-border placeholder:text-muted-foreground ${
                  usernameError ? 'border-red-500' : usernameValid && newUsername ? 'border-green-500' : ''
                }`}
              />
              {usernameChecking && (
                <p className="text-sm text-muted-foreground">检查中...</p>
              )}
              {usernameError && (
                <p className="text-sm text-red-500">{usernameError}</p>
              )}
              {usernameValid && newUsername && !usernameError && (
                <p className="text-sm text-green-600">用户名可用</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="username-code">验证码</Label>
              <div className="flex gap-2">
                <Input
                  id="username-code"
                  value={usernameVerificationCode}
                  onChange={(e) => setUsernameVerificationCode(e.target.value)}
                  placeholder="请输入邮箱验证码"
                  className="bg-input border-border placeholder:text-muted-foreground"
                />
                <Button 
                  variant="outline" 
                  onClick={sendUsernameVerificationCode}
                  disabled={isLoading || usernameCodeSent || !usernameValid || usernameChecking || !newUsername.trim()}
                  className="min-w-32 shrink-0"
                >
                  {usernameCodeSent ? (
                    <div className="flex items-center gap-2">
                      <Timer className="h-4 w-4" />
                      {usernameCountdown}s
                    </div>
                  ) : (
                    <div className="flex items-center gap-2">
                      <Mail className="h-4 w-4" />
                      发送验证码
                    </div>
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
          <CardFooter className="border-t border-border pt-4">
            <Button 
              onClick={handleUsernameChange}
              disabled={isLoading || !usernameVerificationCode}
              className="bg-sky-500 hover:bg-sky-600 text-primary-foreground"
            >
              {isLoading ? "修改中..." : "确认修改"}
            </Button>
          </CardFooter>
        </Card>

        {/* 密码修改 */}
        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <div className="flex items-center gap-3">
              <Lock className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              <CardTitle>修改密码</CardTitle>
            </div>
            <CardDescription>修改您的密码需要邮箱验证，定期更新密码可保护账户安全。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="new-password">新密码</Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="请输入新密码（至少6位）"
                className={`bg-input border-border placeholder:text-muted-foreground ${
                  passwordError && newPassword ? 'border-red-500' : passwordValid && newPassword ? 'border-green-500' : ''
                }`}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">确认新密码</Label>
              <Input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="请再次输入新密码"
                className={`bg-input border-border placeholder:text-muted-foreground ${
                  passwordError && confirmPassword ? 'border-red-500' : passwordValid && confirmPassword ? 'border-green-500' : ''
                }`}
              />
              {passwordError && (newPassword || confirmPassword) && (
                <p className="text-sm text-red-500">{passwordError}</p>
              )}
              {passwordValid && newPassword && confirmPassword && !passwordError && (
                <p className="text-sm text-green-600">密码设置有效</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password-code">验证码</Label>
              <div className="flex gap-2">
                <Input
                  id="password-code"
                  value={passwordVerificationCode}
                  onChange={(e) => setPasswordVerificationCode(e.target.value)}
                  placeholder="请输入邮箱验证码"
                  className="bg-input border-border placeholder:text-muted-foreground"
                />
                <Button 
                  variant="outline" 
                  onClick={sendPasswordVerificationCode}
                  disabled={isLoading || passwordCodeSent || !passwordValid || !newPassword.trim() || !confirmPassword.trim()}
                  className="min-w-32 shrink-0"
                >
                  {passwordCodeSent ? (
                    <div className="flex items-center gap-2">
                      <Timer className="h-4 w-4" />
                      {passwordCountdown}s
                    </div>
                  ) : (
                    <div className="flex items-center gap-2">
                      <Mail className="h-4 w-4" />
                      发送验证码
                    </div>
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
          <CardFooter className="border-t border-border pt-4">
            <Button 
              onClick={handlePasswordChange}
              disabled={isLoading || !passwordVerificationCode}
              className="bg-sky-500 hover:bg-sky-600 text-primary-foreground"
            >
              {isLoading ? "修改中..." : "确认修改"}
            </Button>
          </CardFooter>
        </Card>
      </div>
    </DashboardLayout>
  )
}