import { request } from "./request";

// 接口定义
export interface User {
  id: number;
  username: string;
  email: string;
  balance?: number;
  is_admin?: boolean;
  role?: string;
  avatar?: string;
  last_login_at?: string;
  created_at?: string;
}

export interface AuthResponse {
  success: boolean;
  message?: string;
  token?: string;
  user?: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

// 用户注册
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  try {
    const response = await request.post("/api/auth/register", data);
    return response.data;
  } catch (error: any) {
    console.error("注册请求失败:", error);
    return {
      success: false,
      message: error.response?.data?.message || "注册失败，请稍后再试"
    };
  }
};

// 用户登录
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  try {
    const response = await request.post("/api/auth/login", data);
    return response.data;
  } catch (error: any) {
    console.error("登录请求失败:", error);
    return {
      success: false,
      message: error.response?.data?.message || "邮箱或密码错误"
    };
  }
};

// 用户登出
export const logout = async (): Promise<{ success: boolean; message?: string }> => {
  try {
    const response = await request.post("/api/auth/logout");
    return response.data;
  } catch (error: any) {
    console.error("登出请求失败:", error);
    return {
      success: false,
      message: error.response?.data?.message || "登出失败，请稍后再试"
    };
  }
};

// 获取当前用户信息
export const getUserInfo = async (): Promise<{ success: boolean; user?: User; message?: string }> => {
  try {
    const response = await request.get("/api/auth/user");
    if (response.data && response.data.user) {
      return {
        success: true,
        user: response.data.user
      };
    } else {
      return {
        success: false,
        message: "用户信息格式错误"
      };
    }
  } catch (error: any) {
    console.error("获取用户信息失败:", error);
    return {
      success: false,
      message: error.response?.data?.message || "获取用户信息失败"
    };
  }
};

// OAuth 授权相关接口
export interface OAuthAuthorizeParams {
  client_id: string;
  redirect_uri: string;
  response_type: string;
  scope: string;
  state?: string;
  device_flow?: boolean;
}

export const authorize = async (params: OAuthAuthorizeParams): Promise<{
  success: boolean;
  code?: string;
  token?: string;
  device_flow?: boolean;
  message?: string;
}> => {
  try {
    const response = await request.post("/api/sso/authorize", {
      client_id: params.client_id,
      redirect_uri: params.redirect_uri,
      state: params.state,
      device_flow: params.device_flow || false
    });
    return {
      success: true,
      ...response.data
    };
  } catch (error: any) {
    console.error("OAuth授权失败:", error);
    return {
      success: false,
      message: error.response?.data?.message || error.response?.data?.error || "授权失败"
    };
  }
};

// 验证token
export const verifyToken = async (token: string): Promise<{
  success: boolean;
  valid: boolean;
  message?: string;
}> => {
  try {
    const response = await request.post("/api/sso/verify-token", { token });
    return response.data;
  } catch (error: any) {
    return {
      success: false,
      valid: false,
      message: error.response?.data?.message || "Token验证失败"
    };
  }
};

// 验证授权码
export const verifyCode = async (code: string): Promise<{
  success: boolean;
  valid: boolean;
  message?: string;
}> => {
  try {
    const response = await request.post("/api/sso/verify-code", { code });
    return response.data;
  } catch (error: any) {
    return {
      success: false,
      valid: false,
      message: error.response?.data?.message || "授权码验证失败"
    };
  }
};