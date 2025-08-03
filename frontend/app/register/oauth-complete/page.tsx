"use client"

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Loader2, User, Mail, CheckCircle, AlertTriangle, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";
import { useAuth } from "@/contexts/AuthContext";

interface LinuxDoUserInfo {
  id: number;
  username: string;
  name: string;
  avatar_template: string;
  active: boolean;
  trust_level: number;
}

interface TemporaryLinuxDoUser {
  user_info: LinuxDoUserInfo;
  created_at: string;
}

export default function OAuthCompletePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  ;
  const { login } = useAuth();
  
  const [tempToken, setTempToken] = useState<string>("");
  const [userInfo, setUserInfo] = useState<LinuxDoUserInfo | null>(null);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 获取临时token并验证
  useEffect(() => {
    const token = searchParams.get("temp_token");
    if (!token) {
      setError("缺少临时令牌");
      setLoading(false);
      return;
    }

    setTempToken(token);
    fetchTemporaryUserInfo(token);
  }, [searchParams]);

  // 动态获取API基础URL
  const getApiBaseURL = () => {
    const hostname = window.location.hostname;
    
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
      return "http://localhost:9998/api";
    } else if (hostname === 'www.duckcode.top') {
      return "https://api.duckcode.top/api";
    } else {
      return "/api";
    }
  };

  // 获取临时用户信息
  const fetchTemporaryUserInfo = async (token: string) => {
    try {
      const apiBaseURL = getApiBaseURL();
      const response = await fetch(`${apiBaseURL}/oauth/linuxdo/temp-user?temp_token=${encodeURIComponent(token)}`);
      const data = await response.json();

      if (response.ok && data.success) {
        const tempUser: TemporaryLinuxDoUser = data.data;
        setUserInfo(tempUser.user_info);
        
        // 预填充建议的用户名和邮箱
        setUsername(tempUser.user_info.username || "");
        setEmail(`${tempUser.user_info.username}@linux.do`);
      } else {
        setError(data.message || "获取用户信息失败");
      }
    } catch (err) {
      console.error("获取临时用户信息失败:", err);
      setError("网络错误，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  // 完成注册
  const handleCompleteRegistration = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!username.trim() || !email.trim()) {
      toast({
        title: "输入错误",
        description: "请填写完整的用户名和邮箱",
        variant: "destructive",
      });
      return;
    }

    // 简单的邮箱格式验证
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      toast({
        title: "邮箱格式错误",
        description: "请输入有效的邮箱地址",
        variant: "destructive",
      });
      return;
    }

    setSubmitting(true);

    try {
      const apiBaseURL = getApiBaseURL();
      const response = await fetch(`${apiBaseURL}/oauth/linuxdo/complete-registration`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          temp_token: tempToken,
          username: username.trim(),
          email: email.trim(),
        }),
      });

      const data = await response.json();

      if (response.ok && data.success) {
        // 注册成功，使用返回的token登录
        if (data.token && data.user) {
          login(data.token, data.user);
          
          toast({
            title: "注册成功",
            description: "欢迎加入！正在跳转到主页...",
            variant: "default",
          });

          setTimeout(() => {
            router.replace("/");
          }, 1500);
        } else {
          throw new Error("服务器响应数据不完整");
        }
      } else {
        setError(data.message || "注册失败");
        toast({
          title: "注册失败",
          description: data.message || "注册过程中出现错误",
          variant: "destructive",
        });
      }
    } catch (err) {
      console.error("完成注册失败:", err);
      setError("网络错误，请稍后重试");
      toast({
        title: "注册失败",
        description: "网络错误，请稍后重试",
        variant: "destructive",
      });
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardContent className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin mr-3" />
            <span>正在验证信息...</span>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error && !userInfo) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center text-red-600">
              <AlertTriangle className="h-6 w-6 mr-2" />
              验证失败
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>错误</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
            <div className="flex gap-2">
              <Button 
                variant="outline" 
                className="flex-1"
                onClick={() => router.back()}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                返回
              </Button>
              <Button 
                variant="default" 
                className="flex-1"
                asChild
              >
                <Link href="/login">
                  重新登录
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl flex items-center justify-center">
            <CheckCircle className="h-7 w-7 mr-2 text-green-600" />
            完成注册
          </CardTitle>
          <p className="text-sm text-muted-foreground mt-2">
            Linux Do 授权成功！请完善您的账户信息
          </p>
        </CardHeader>
        
        <CardContent className="space-y-6">
          {/* Linux Do 用户信息展示 */}
          {userInfo && (
            <div className="bg-muted/50 rounded-lg p-4 border">
              <h3 className="font-medium mb-3 text-center">Linux Do 账户信息</h3>
              <div className="flex items-center space-x-4">
                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-lg">
                  {userInfo.name?.charAt(0) || userInfo.username?.charAt(0) || "U"}
                </div>
                <div className="flex-1">
                  <p className="font-medium">{userInfo.name || userInfo.username}</p>
                  <p className="text-sm text-muted-foreground">@{userInfo.username}</p>
                  <p className="text-xs text-muted-foreground">
                    信任等级: {userInfo.trust_level} | ID: {userInfo.id}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* 注册表单 */}
          <form onSubmit={handleCompleteRegistration} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="输入您的用户名"
                  className="pl-10"
                  required
                />
              </div>
              <p className="text-xs text-muted-foreground">
                可以与 Linux Do 用户名不同，用于本站显示
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">邮箱地址</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="输入您的邮箱地址"
                  className="pl-10"
                  required
                />
              </div>
              <p className="text-xs text-muted-foreground">
                用于接收重要通知
              </p>
            </div>

            {error && (
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>注册失败</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="flex gap-3 pt-2">
              <Button 
                type="button" 
                variant="outline" 
                className="flex-1"
                onClick={() => router.push("/login")}
                disabled={submitting}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                取消
              </Button>
              <Button 
                type="submit" 
                className="flex-1"
                disabled={submitting}
              >
                {submitting ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    注册中...
                  </>
                ) : (
                  <>
                    <CheckCircle className="h-4 w-4 mr-2" />
                    完成注册
                  </>
                )}
              </Button>
            </div>
          </form>

          {/* 说明文字 */}
          <div className="text-center text-xs text-muted-foreground space-y-1">
            <p>完成注册后，您的 Linux Do 账户将与本站账户绑定</p>
            <p>下次可直接使用 Linux Do 一键登录</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}