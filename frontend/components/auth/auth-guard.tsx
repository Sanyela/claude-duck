"use client";

import { ReactNode, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import { LoadingSpinner } from "@/components/ui/loading-spinner";

// 定义不需要登录就可以访问的路由
const PUBLIC_ROUTES = ["/login"];

export function AuthGuard({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);

  useEffect(() => {
    // 如果已完成加载, 不是公共路由, 并且用户未登录，则重定向到登录页
    if (!isLoading && !isPublicRoute && !isAuthenticated) {
      router.replace(`/login?redirect=${encodeURIComponent(pathname)}`);
    }
    
    // 如果已完成加载, 是登录页, 并且用户已经登录, 则重定向到首页
    if (!isLoading && isPublicRoute && isAuthenticated) {
      router.replace("/");
    }
  }, [isLoading, isAuthenticated, isPublicRoute, pathname, router]);

  // 如果页面正在加载认证状态，显示加载状态
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // 如果是公共路由或者用户已认证，显示子组件
  if (isPublicRoute || isAuthenticated) {
    return <>{children}</>;
  }

  // 默认情况下，显示加载状态（可能处于重定向过程中）
  return (
    <div className="flex items-center justify-center min-h-screen">
      <LoadingSpinner size="lg" />
    </div>
  );
}