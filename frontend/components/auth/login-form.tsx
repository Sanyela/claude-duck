"use client"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Github, KeyRound, Mail, LogIn } from "lucide-react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useRouter } from "next/navigation"
import { Separator } from "@/components/ui/separator"

export function LoginForm() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const initialTab = searchParams.get("tab") || "login"

  // Placeholder action handlers
  const handleLogin = async (formData: FormData) => {
    alert("登录逻辑（占位符）")
  }

  const handleSignup = async (formData: FormData) => {
    alert("注册逻辑（占位符）")
  }

  const handleOneClickLogin = () => {
    // In a real app, this would set a mock session. For UI, we just redirect.
    router.push("/")
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
          <form action={handleLogin}>
            <CardContent className="space-y-4 pt-6">
              <div className="space-y-2">
                <Label htmlFor="email-login" className="text-slate-700 dark:text-slate-300">
                  邮箱
                </Label>
                <Input
                  id="email-login"
                  name="email"
                  type="email"
                  placeholder="you@example.com"
                  required
                  className="bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-slate-900 dark:text-slate-50"
                />
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
              <Button type="submit" className="w-full bg-sky-500 hover:bg-sky-600 text-white dark:text-slate-900">
                <Mail className="mr-2 h-4 w-4" /> 登录 🚀
              </Button>
              <Button
                variant="outline"
                className="w-full bg-slate-900 hover:bg-slate-800 text-white dark:bg-slate-700 dark:border-slate-600 dark:hover:bg-slate-600 dark:text-slate-50"
              >
                <Github className="mr-2 h-4 w-4" /> 使用 GitHub 登录
              </Button>
            </CardFooter>
          </form>
        </TabsContent>
        <TabsContent value="register">
          <form action={handleSignup}>
            <CardContent className="space-y-4 pt-6">{/* Registration form fields */}</CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button type="submit" className="w-full bg-sky-500 hover:bg-sky-600 text-white dark:text-slate-900">
                注册 ✨
              </Button>
            </CardFooter>
          </form>
        </TabsContent>
      </Tabs>
      <div className="p-6 pt-0">
        <Separator className="my-4 bg-slate-200 dark:bg-slate-700" />
        <Button
          variant="secondary"
          className="w-full bg-green-100 text-green-700 hover:bg-green-200 dark:bg-green-900/50 dark:text-green-300 dark:hover:bg-green-900"
          onClick={handleOneClickLogin}
        >
          <LogIn className="mr-2 h-4 w-4" />
          一键登录 (测试)
        </Button>
      </div>
    </Card>
  )
}
