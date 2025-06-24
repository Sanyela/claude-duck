"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import { LoadingSpinner } from "@/components/ui/loading-spinner";

interface ProtectedRouteProps {
  children: React.ReactNode;
  adminOnly?: boolean;
}

export function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // 如果已完成加载且用户未登录，则重定向到登录页
    if (!isLoading && !isAuthenticated) {
      router.replace("/login");
    }

    // 如果需要管理员权限但用户不是管理员，则重定向到首页
    if (!isLoading && adminOnly && isAuthenticated && user?.is_admin !== true) {
      router.replace("/");
    }
  }, [isLoading, isAuthenticated, router, adminOnly, user]);

  // 如果正在加载，显示加载状态
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // 如果用户已认证，显示子组件
  if (isAuthenticated && (!adminOnly || (adminOnly && user?.is_admin))) {
    return <>{children}</>;
  }

  // 默认情况下，显示空白以等待重定向
  return null;
}