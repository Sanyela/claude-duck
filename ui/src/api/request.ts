import axios from 'axios';
import type { AxiosResponse, AxiosError, InternalAxiosRequestConfig } from 'axios';

// 创建 axios 实例
const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:9998/api', // API 的 base_url，使用环境变量
  timeout: 10000, // 请求超时时间
});

// 请求拦截器 (可选)
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig | Promise<InternalAxiosRequestConfig> => {
    // 自动添加token到请求头
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    console.error('Request error:', error); // for debug
    return Promise.reject(error);
  }
);

// 响应拦截器 (可选)
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data;
    // 如果自定义了 code，可以在这里判断，例如：
    // if (res.code !== 200 && res.code !== 0) {
    //   // 处理错误
    //   return Promise.reject(new Error(res.message || 'Error'));
    // }
    return res;
  },
  (error: AxiosError<any>) => {
    console.error('Response error:', error); // for debug
    
    // 处理401认证失败
    if (error.response?.status === 401) {
      // 清除本地存储的认证信息
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user_data');
      
      // 重定向到登录页面（如果不在登录页面）
      if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
        window.location.href = '/login';
      }
    }
    
    // 处理 HTTP 网络错误
    if (error.response && error.response.data && error.response.data.error) {
      console.error('API Error:', error.response.data.error);
    } else if (error.message) {
      console.error('Network Error:', error.message);
    }
    
    return Promise.reject(error);
  }
);

export { service as axiosInstance };
export default service; 