import { request } from "./request";

// 管理员接口定义
export interface AdminUser {
  id: number;
  username: string;
  email: string;
  is_admin: boolean;
  degradation_guaranteed: number;
  degradation_source: string;
  degradation_locked: boolean;
  degradation_counter: number;
  created_at: string;
  updated_at: string;
}

export interface SystemConfig {
  id: number;
  config_key: string;
  config_value: string;
  description: string;
  updated_at: string;
}

export interface SubscriptionPlan {
  id: number;
  title: string;
  description: string;
  point_amount: number;
  price: number;
  currency: string;
  validity_days: number;
  degradation_guaranteed: number;
  daily_checkin_points: number;
  daily_checkin_points_max: number;
  features: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ActivationCode {
  id: number;
  code: string;
  subscription_plan_id: number;
  plan?: SubscriptionPlan;
  status: string;
  used_by_user_id?: number;
  used_by?: AdminUser;
  used_at?: string;
  expires_at?: string;
  batch_number: string;
  created_at: string;
}

export interface Announcement {
  id: number;
  type: string;
  title: string;
  description: string;
  language: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

// 分页响应接口
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 用户管理API
export const adminAPI = {
  // 获取所有用户
  getUsers: async (params?: {
    page?: number;
    page_size?: number;
  }): Promise<{ 
    success: boolean; 
    users?: AdminUser[]; 
    total?: number;
    page?: number;
    page_size?: number;
    total_pages?: number;
    message?: string;
  }> => {
    try {
      // 构建查询参数
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.page_size) queryParams.append('page_size', params.page_size.toString());
      
      const url = `/api/admin/users${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
      const response = await request.get(url);
      // 后端返回分页格式，数据在data字段中
      const paginatedData = response.data as PaginatedResponse<AdminUser>;
      return {
        success: true,
        users: paginatedData.data || [],
        total: paginatedData.total,
        page: paginatedData.page,
        page_size: paginatedData.page_size,
        total_pages: paginatedData.total_pages
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "获取用户列表失败"
      };
    }
  },

  // 更新用户
  updateUser: async (id: number, data: Partial<AdminUser>): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.put(`/api/admin/users/${id}`, data);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "更新用户失败"
      };
    }
  },

  // 删除用户
  deleteUser: async (id: number): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.delete(`/api/admin/users/${id}`);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "删除用户失败"
      };
    }
  },

  // 获取系统配置
  getSystemConfigs: async (): Promise<{ success: boolean; configs?: SystemConfig[]; message?: string }> => {
    try {
      const response = await request.get("/api/admin/system-configs");
      return {
        success: true,
        configs: Array.isArray(response.data) ? response.data : []
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "获取系统配置失败"
      };
    }
  },

  // 更新系统配置
  updateSystemConfig: async (data: { config_key: string; config_value: string }): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.put("/api/admin/system-config", data);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "更新系统配置失败"
      };
    }
  },

  // 获取订阅计划
  getSubscriptionPlans: async (): Promise<{ success: boolean; plans?: SubscriptionPlan[]; message?: string }> => {
    try {
      const response = await request.get("/api/admin/subscription-plans");
      // 后端返回分页格式，数据在data字段中
      const paginatedData = response.data as PaginatedResponse<SubscriptionPlan>;
      return {
        success: true,
        plans: paginatedData.data || []
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "获取订阅计划失败"
      };
    }
  },

  // 创建订阅计划
  createSubscriptionPlan: async (data: Partial<SubscriptionPlan>): Promise<{ success: boolean; message?: string }> => {
    try {
      // 过滤掉不应该发送的字段
      const { id, created_at, updated_at, ...createData } = data;
      await request.post("/api/admin/subscription-plans", createData);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "创建订阅计划失败"
      };
    }
  },

  // 更新订阅计划
  updateSubscriptionPlan: async (id: number, data: Partial<SubscriptionPlan>): Promise<{ success: boolean; message?: string }> => {
    try {
      // 过滤掉不应该发送的字段
      const { id: planId, created_at, updated_at, ...updateData } = data;
      await request.put(`/api/admin/subscription-plans/${id}`, updateData);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "更新订阅计划失败"
      };
    }
  },

  // 删除订阅计划
  deleteSubscriptionPlan: async (id: number): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.delete(`/api/admin/subscription-plans/${id}`);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "删除订阅计划失败"
      };
    }
  },

  // 获取激活码
  getActivationCodes: async (params?: {
    page?: number;
    page_size?: number;
    status?: 'unused' | 'used' | 'expired';
    batch_number?: string;
  }): Promise<{ 
    success: boolean; 
    codes?: ActivationCode[]; 
    total?: number;
    page?: number;
    page_size?: number;
    total_pages?: number;
    message?: string;
  }> => {
    try {
      // 构建查询参数
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.page_size) queryParams.append('page_size', params.page_size.toString());
      if (params?.status) queryParams.append('status', params.status);
      if (params?.batch_number) queryParams.append('batch_number', params.batch_number);
      
      const url = `/api/admin/activation-codes${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
      const response = await request.get(url);
      // 后端返回分页格式，数据在data字段中
      const paginatedData = response.data as PaginatedResponse<ActivationCode>;
      return {
        success: true,
        codes: paginatedData.data || [],
        total: paginatedData.total,
        page: paginatedData.page,
        page_size: paginatedData.page_size,
        total_pages: paginatedData.total_pages
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "获取激活码失败"
      };
    }
  },

  // 创建激活码
  createActivationCodes: async (data: { 
    subscription_plan_id: number; 
    count: number; 
    batch_number?: string;
  }): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.post("/api/admin/activation-codes", data);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "创建激活码失败"
      };
    }
  },

  // 删除激活码
  async deleteActivationCode(id: number): Promise<{ success: boolean; message?: string }> {
    try {
      await request.delete(`/api/admin/activation-codes/${id}`);
      return { success: true };
    } catch (error: any) {
      console.error("删除激活码失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "删除激活码失败" 
      };
    }
  },

  // 获取公告列表
  async getAnnouncements(): Promise<{ success: boolean; announcements?: Announcement[]; message?: string }> {
    try {
      const response = await request.get("/api/admin/announcements");
      return { success: true, announcements: response.data.data || response.data };
    } catch (error: any) {
      console.error("获取公告列表失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取公告列表失败" 
      };
    }
  },

  // 创建公告
  async createAnnouncement(announcement: Omit<Announcement, 'id' | 'created_at' | 'updated_at'>): Promise<{ success: boolean; message?: string }> {
    try {
      await request.post("/api/admin/announcements", announcement);
      return { success: true };
    } catch (error: any) {
      console.error("创建公告失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "创建公告失败" 
      };
    }
  },

  // 更新公告
  async updateAnnouncement(id: number, announcement: Partial<Announcement>): Promise<{ success: boolean; message?: string }> {
    try {
      await request.put(`/api/admin/announcements/${id}`, announcement);
      return { success: true };
    } catch (error: any) {
      console.error("更新公告失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "更新公告失败" 
      };
    }
  },

  // 删除公告
  async deleteAnnouncement(id: number): Promise<{ success: boolean; message?: string }> {
    try {
      await request.delete(`/api/admin/announcements/${id}`);
      return { success: true };
    } catch (error: any) {
      console.error("删除公告失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "删除公告失败" 
      };
    }
  }
}; 