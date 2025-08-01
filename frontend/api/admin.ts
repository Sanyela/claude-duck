import { request } from "./request";

// 管理员接口定义
export interface AdminUser {
  id: number;
  username: string;
  email: string;
  is_admin: boolean;
  is_disabled: boolean;
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
  // 积分相关信息
  subscription?: {
    total_points: number;
    used_points: number;
    available_points: number;
  };
}

// 冻结积分记录类型
export interface FrozenPointsRecord {
  id: number;
  user_id: number;
  user?: AdminUser;
  banned_activation_code: string;
  banned_code_id: number;
  frozen_points: number;
  frozen_benefits: string;
  before_ban_wallet_state: string;
  before_ban_benefits: string;
  calculation_method: string;
  estimated_usage: number;
  status: "frozen" | "restored";
  ban_reason: string;
  admin_user_id?: number;
  admin_user?: AdminUser;
  created_at: string;
  updated_at: string;
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

// 管理员赠送记录
export interface GiftRecord {
  id: number;
  from_admin_id: number;
  to_user_id: number;
  subscription_plan_id: number;
  points_amount: number;
  validity_days: number;
  daily_max_points: number;
  reason: string;
  status: string;
  subscription_id?: number;
  error_message: string;
  created_at: string;
  updated_at: string;
  from_admin: AdminUser;
  to_user: AdminUser;
  plan: SubscriptionPlan;
}

// 数据看板类型定义
export interface DashboardOverview {
  total_users: number;
  active_subscription_users: number;
  total_subscriptions: number;
  today_new_users: number;
  user_growth_rate: number;
  today_points_used: number;
  points_growth_rate: number;
}

export interface PointsStats {
  total_points: number;
  used_points: number;
  available_points: number;
}

export interface DailyStats {
  date: string;
  count: number;
}

export interface PlanStats {
  plan_name: string;
  count: number;
}

export interface SourceStats {
  source_type: string;
  count: number;
  points: number;
}

export interface DashboardData {
  overview: DashboardOverview;
  points_stats: PointsStats;
  trends: {
    user_registration: DailyStats[];
    points_usage: DailyStats[];
  };
  distributions: {
    subscription_plans: PlanStats[];
    points_sources: SourceStats[];
  };
  generated_at: string;
}

// 用户管理API
export const adminAPI = {
  // 获取所有用户
  getUsers: async (params?: {
    page?: number;
    page_size?: number;
    search?: string;
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
      if (params?.search) queryParams.append('search', params.search);
      
      const url = `/admin/users${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
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
      await request.put(`/admin/users/${id}`, data);
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
      await request.delete(`/admin/users/${id}`);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "删除用户失败"
      };
    }
  },

  // 切换用户禁用状态
  toggleUserStatus: async (id: number, isDisabled: boolean): Promise<{ success: boolean; message?: string; user?: any }> => {
    try {
      const response = await request.put(`/admin/users/${id}/status`, { is_disabled: isDisabled });
      return { 
        success: true, 
        message: response.data?.message,
        user: response.data?.user
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "切换用户状态失败"
      };
    }
  },

  // 获取系统配置
  getSystemConfigs: async (): Promise<{ success: boolean; configs?: SystemConfig[]; message?: string }> => {
    try {
      const response = await request.get("/admin/system-configs");
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
      await request.put("/admin/system-config", data);
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
      const response = await request.get("/admin/subscription-plans");
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
      await request.post("/admin/subscription-plans", createData);
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
      await request.put(`/admin/subscription-plans/${id}`, updateData);
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
      await request.delete(`/admin/subscription-plans/${id}`);
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
    status?: 'unused' | 'used' | 'expired' | 'depleted';
    batch_number?: string;
    code?: string;
    username?: string;
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
      if (params?.code) queryParams.append('code', params.code);
      if (params?.username) queryParams.append('username', params.username);
      
      const url = `/admin/activation-codes${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
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
      await request.post("/admin/activation-codes", data);
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
      await request.delete(`/admin/activation-codes/${id}`);
      return { success: true };
    } catch (error: any) {
      console.error("删除激活码失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "删除激活码失败" 
      };
    }
  },

  // 获取订阅每日限制
  async getSubscriptionDailyLimit(activationCodeId: number): Promise<{ success: boolean; daily_limit?: number; message?: string }> {
    try {
      const response = await request.get(`/admin/activation-codes/${activationCodeId}/daily-limit`);
      return { success: true, daily_limit: response.data.daily_limit };
    } catch (error: any) {
      console.error("获取每日限制失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取每日限制失败" 
      };
    }
  },

  // 更新订阅每日限制
  async updateSubscriptionDailyLimit(activationCodeId: number, dailyLimit: number): Promise<{ success: boolean; message?: string }> {
    try {
      await request.put(`/admin/activation-codes/${activationCodeId}/daily-limit`, { daily_limit: dailyLimit });
      return { success: true };
    } catch (error: any) {
      console.error("更新每日限制失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "更新每日限制失败" 
      };
    }
  },

  // 获取公告列表
  async getAnnouncements(): Promise<{ success: boolean; announcements?: Announcement[]; message?: string }> {
    try {
      const response = await request.get("/admin/announcements");
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
      await request.post("/admin/announcements", announcement);
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
      await request.put(`/admin/announcements/${id}`, announcement);
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
      await request.delete(`/admin/announcements/${id}`);
      return { success: true };
    } catch (error: any) {
      console.error("删除公告失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "删除公告失败" 
      };
    }
  },

  // 获取用户订阅列表
  getUserSubscriptions: async (userId: number): Promise<{ success: boolean; subscriptions?: any[]; message?: string }> => {
    try {
      const response = await request.get(`/admin/users/${userId}/subscriptions`);
      return {
        success: true,
        subscriptions: response.data?.subscriptions || []
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "获取用户订阅失败"
      };
    }
  },

  // 赠送订阅给用户
  giftSubscription: async (userId: number, data: {
    subscription_plan_id: number;
    points_amount?: number;
    validity_days?: number;
    daily_max_points?: number;
    reason?: string;
  }): Promise<{ success: boolean; message?: string; gift_record?: any; subscription?: any }> => {
    try {
      const response = await request.post(`/admin/users/${userId}/gift`, data);
      return {
        success: true,
        message: response.data?.message,
        gift_record: response.data?.gift_record,
        subscription: response.data?.subscription
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "赠送订阅失败"
      };
    }
  },

  // 获取数据看板统计信息
  getDashboard: async (): Promise<{ success: boolean; data?: DashboardData; message?: string }> => {
    try {
      const response = await request.get("/admin/dashboard");
      return {
        success: true,
        data: response.data
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "获取数据看板失败"
      };
    }
  },

  // ===== 激活码封禁管理相关接口 =====

  // 预览封禁激活码的影响
  previewBanActivationCode: async (data: {
    user_id: number;
    activation_code: string;
  }): Promise<{ 
    success: boolean; 
    current_wallet?: any;
    ban_impact?: any;
    consumption_details?: any[];
    target_card?: any;
    benefits_change?: any;
    message?: string;
  }> => {
    try {
      const response = await request.post("/admin/activation-codes/ban-preview", data);
      return {
        success: true,
        ...response.data  // 返回后端响应的所有字段
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "获取封禁预览失败"
      };
    }
  },

  // 封禁激活码
  banActivationCode: async (data: {
    user_id: number;
    activation_code: string;
    reason?: string;
  }): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.post("/admin/activation-codes/ban", data);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "封禁激活码失败"
      };
    }
  },

  // 解禁激活码
  unbanActivationCode: async (data: {
    user_id: number;
    activation_code: string;
  }): Promise<{ success: boolean; message?: string }> => {
    try {
      await request.post("/admin/activation-codes/unban", data);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "解禁激活码失败"
      };
    }
  },

  // 获取冻结记录列表
  getFrozenRecords: async (params?: {
    page?: number;
    page_size?: number;
    user_id?: number;
    status?: string;
    activation_code?: string;
  }): Promise<{ 
    success: boolean; 
    data?: FrozenPointsRecord[];
    total?: number;
    page?: number;
    page_size?: number;
    total_pages?: number;
    message?: string;
  }> => {
    try {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.page_size) queryParams.append('page_size', params.page_size.toString());
      if (params?.user_id) queryParams.append('user_id', params.user_id.toString());
      if (params?.status) queryParams.append('status', params.status);
      if (params?.activation_code) queryParams.append('activation_code', params.activation_code);

      const url = `/admin/frozen-records${queryParams.toString() ? '?' + queryParams.toString() : ''}`;
      const response = await request.get(url);
      
      return {
        success: true,
        data: response.data?.data || [],
        total: response.data?.total || 0,
        page: response.data?.page || 1,
        page_size: response.data?.page_size || 10,
        total_pages: response.data?.total_pages || 0
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "获取冻结记录失败"
      };
    }
  },

  // 获取冻结记录详情
  getFrozenRecordDetail: async (recordId: number): Promise<{ 
    success: boolean; 
    data?: FrozenPointsRecord;
    message?: string;
  }> => {
    try {
      const response = await request.get(`/admin/frozen-records/${recordId}`);
      return {
        success: true,
        data: response.data
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.error || "获取冻结记录详情失败"
      };
    }
  }
}; 