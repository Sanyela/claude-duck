"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { User, getUserInfo } from "@/api/auth";
import { useRouter } from "next/navigation";

// 认证上下文接口定义
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, userData: User) => void;
  logout: () => void;
  refreshUserInfo: () => Promise<void>;
}

// 创建认证上下文
const AuthContext = createContext<AuthContextType>({
  user: null,
  isAuthenticated: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
  refreshUserInfo: async () => {},
});

// 认证提供者组件属性
interface AuthProviderProps {
  children: ReactNode;
}

// 认证提供者组件
export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // 登录方法
  const login = (token: string, userData: User) => {
    localStorage.setItem("auth_token", token);
    localStorage.setItem("user_data", JSON.stringify(userData));
    setUser(userData);
  };

  // 登出方法
  const logout = () => {
    localStorage.removeItem("auth_token");
    localStorage.removeItem("user_data");
    setUser(null);
    router.push("/login");
  };

  // 刷新用户信息
  const refreshUserInfo = async () => {
    try {
      setIsLoading(true);
      const response = await getUserInfo();
      
      if (response.success && response.user) {
        setUser(response.user);
        localStorage.setItem("user_data", JSON.stringify(response.user));
      } else {
        throw new Error("获取用户信息失败");
      }
    } catch (error) {
      console.error("刷新用户信息失败:", error);
      // 如果获取用户信息失败，清除本地存储的用户信息
      logout();
    } finally {
      setIsLoading(false);
    }
  };

  // 页面加载时初始化认证状态
  useEffect(() => {
    // 从本地存储中获取认证数据
    const initAuth = async () => {
      setIsLoading(true);
      
      try {
        const token = localStorage.getItem("auth_token");
        const userData = localStorage.getItem("user_data");
        
        if (token && userData) {
          // 暂时使用本地存储的用户数据
          setUser(JSON.parse(userData));
          
          // 然后尝试从服务器获取最新的用户数据
          const response = await getUserInfo();
          
          if (response.success && response.user) {
            setUser(response.user);
            localStorage.setItem("user_data", JSON.stringify(response.user));
          } else {
            // 如果获取用户信息失败，清除本地存储的用户信息
            localStorage.removeItem("auth_token");
            localStorage.removeItem("user_data");
            setUser(null);
          }
        }
      } catch (error) {
        console.error("初始化认证状态失败:", error);
        // 出错时清除本地存储的用户信息
        localStorage.removeItem("auth_token");
        localStorage.removeItem("user_data");
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    // 当在客户端环境时，初始化认证状态
    if (typeof window !== "undefined") {
      initAuth();
    } else {
      setIsLoading(false);
    }
  }, []);

  // 提供上下文值
  const contextValue: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    refreshUserInfo,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

// 使用认证上下文的钩子
export const useAuth = () => useContext(AuthContext);