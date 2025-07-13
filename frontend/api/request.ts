import axios from "axios";

const getBaseURL = () => {
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:9998";
};

// 创建axios实例
export const request = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从localStorage获取token
    const token = localStorage.getItem("auth_token");
    
    // 如果存在token，则在请求头中添加
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    return config;
  },
  (error) => {
    console.error("请求拦截器错误:", error);
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    // 成功响应
    return response;
  },
  (error) => {
    // 处理错误
    const { response } = error;
    
    // 如果是401未授权错误，清除本地token并重定向到登录页
    if (response && response.status === 401) {
      console.warn("登录已过期，请重新登录");
      localStorage.removeItem("auth_token");
      localStorage.removeItem("user_data");
      
      // 判断当前是否在客户端环境
      if (typeof window !== "undefined") {
        // 避免无限重定向，如果当前不是登录页，才重定向到登录页
        const currentPath = window.location.pathname;
        if (!currentPath.includes("/login")) {
          window.location.href = "/login";
        }
      }
    }
    
    return Promise.reject(error);
  }
);