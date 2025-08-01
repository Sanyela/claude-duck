import axios from "axios";

const getBaseURL = () => {
  // 根据环境动态设置API地址
  if (typeof window === 'undefined') {
    // 服务端环境，使用默认地址
    return "http://localhost:9998/api";
  }
  
  // 客户端环境，根据当前域名判断
  const hostname = window.location.hostname;
  
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    // 本地开发环境
    return "http://localhost:9998/api";
  } else if (hostname === 'www.duckcode.top') {
    // 生产环境
    return "https://api.duckcode.top/api";
  } else {
    // 其他环境，使用相对路径
    return "/api";
  }
};

// 创建axios实例
export const request = axios.create({
  baseURL: getBaseURL(),
  timeout: 300000, // 增加到30秒
  headers: {
    "Content-Type": "application/json",
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 仅在客户端环境中获取token，防止SSR状态泄露
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("auth_token");
      
      // 如果存在token，则在请求头中添加
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
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
      
      // 仅在客户端环境中操作localStorage和location
      if (typeof window !== "undefined") {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("user_data");
        
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