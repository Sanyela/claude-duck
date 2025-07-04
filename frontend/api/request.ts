import axios from "axios";

// 智能检测API基础URL
const getBaseURL = () => {
  // 如果有显式设置环境变量，优先使用
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  
  // 在浏览器环境中
  if (typeof window !== "undefined") {
    const { protocol, hostname, port } = window.location;
    
    // 分离部署模式：前端在3000端口，后端在9998端口
    if (port === "3000") {
      return `${protocol}//${hostname}:9998`;
    }
    
    // 单体部署模式：如果是从9998端口访问，说明是单点应用
    if (port === "9998" || !port) {
      return `${protocol}//${hostname}${port ? `:${port}` : ""}`;
    }
  }
  
  // 服务端渲染时的默认值
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