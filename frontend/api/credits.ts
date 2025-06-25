import { request } from "./request";

// 积分使用历史接口
export interface CreditUsageHistory {
  id: string;
  description: string;
  amount: number;
  timestamp: string;
  relatedModel: string;
  input_tokens: number;
  output_tokens: number;
}

// 积分余额接口
export interface CreditBalance {
  id: number;
  user_id: number;
  total_points: number;
  used_points: number;
  available_points: number;
  expired_points: number;
  is_current_subscription: boolean;
  updated_at: string;
}

// 模型价格接口
export interface ModelCost {
  model: string;
  prompt_multiplier: number;
  completion_multiplier: number;
  description: string;
}

// 积分API
export const creditsAPI = {
  // 获取积分余额
  async getBalance(): Promise<{ success: boolean; data?: CreditBalance; message?: string }> {
    try {
      const response = await request.get("/api/credits/balance");
      
      // 检查并处理后端返回的数据结构
      if (response.data && response.data.balance) {
        // 将后端返回的balance格式转换为前端需要的CreditBalance格式
        const balance = response.data.balance;
        const creditBalance: CreditBalance = {
          id: 0, // 这个字段前端实际没有使用
          user_id: 0, // 这个字段前端实际没有使用
          total_points: balance.total || 0,
          used_points: balance.used || 0,  // 使用后端返回的实际已使用积分
          available_points: balance.available || 0,
          expired_points: balance.expired || 0,  // 使用后端返回的已过期积分
          is_current_subscription: balance.is_current_subscription || false,
          updated_at: new Date().toISOString()
        };
        return { success: true, data: creditBalance };
      }

      return { success: true, data: response.data };
    } catch (error: any) {
      console.error("获取积分余额失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取积分余额失败" 
      };
    }
  },

  // 获取积分使用历史
  async getUsageHistory(params?: { 
    page?: number; 
    page_size?: number;
    start_date?: string;
    end_date?: string;
  }): Promise<{ success: boolean; data?: { history: CreditUsageHistory[]; totalPages: number; currentPage: number } | null; message?: string }> {
    try {
      const response = await request.get("/api/credits/history", { params });
      return { 
        success: true, 
        data: response.data
      };
    } catch (error: any) {
      console.error("获取积分使用历史失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取积分使用历史失败" 
      };
    }
  },

  // 获取模型价格
  async getModelCosts(): Promise<{ success: boolean; data?: ModelCost[]; message?: string }> {
    try {
      const response = await request.get("/api/credits/model-costs");
      return { success: true, data: response.data };
    } catch (error: any) {
      console.error("获取模型价格失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取模型价格失败" 
      };
    }
  }
};