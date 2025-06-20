import request from './request';

// 认证相关接口类型定义
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UserData {
  id: number;
  username: string;
  email: string;
}

export interface AuthResponse {
  success: boolean;
  message: string;
  token?: string;
  user?: UserData;
}

export interface UserInfoResponse {
  user: UserData;
}

// API 函数
export function register(data: RegisterRequest): Promise<AuthResponse> {
  return request({
    url: '/auth/register',
    method: 'post',
    data,
  });
}

export function login(data: LoginRequest): Promise<AuthResponse> {
  return request({
    url: '/auth/login',
    method: 'post',
    data,
  });
}

export function logout(): Promise<{ success: boolean; message: string }> {
  return request({
    url: '/auth/logout',
    method: 'post',
  });
}

export function getUserInfo(): Promise<UserInfoResponse> {
  return request({
    url: '/auth/user',
    method: 'get',
  });
}
