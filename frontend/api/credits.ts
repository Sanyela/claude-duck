import { request } from "./request";

// 计费详情接口
export interface BillingDetails {
  input_multiplier: number;         // 输入token倍率
  output_multiplier: number;        // 输出token倍率  
  cache_multiplier: number;         // 缓存token倍率
  weighted_input_tokens: number;    // 加权后的输入tokens
  weighted_output_tokens: number;   // 加权后的输出tokens
  weighted_cache_tokens: number;    // 加权后的缓存tokens
  total_weighted_tokens: number;    // 总加权tokens
  final_points: number;             // 最终扣除积分
  pricing_table_used: boolean;      // 是否使用了阶梯计费表
}

// 积分使用历史接口
export interface CreditUsageHistory {
  id: string;
  description?: string;
  amount: number;
  timestamp: string;
  relatedModel: string;
  input_tokens: number;
  output_tokens: number;
  cache_creation_tokens?: number;
  cache_read_tokens?: number;
  total_cache_tokens?: number;
  billing_details?: BillingDetails;
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
  free_model_usage_count: number;
  checkin_points: number;          // 签到积分
  admin_gift_points: number;       // 管理员赠送积分
  updated_at: string;
}

// 模型价格接口
export interface ModelCost {
  model: string;
  prompt_multiplier: number;
  completion_multiplier: number;
  description: string;
}

// 计费表接口
export interface PricingTable {
  pricing_table: Record<string, number>; // token阈值 -> 积分的映射
  description: string; // 说明
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
          free_model_usage_count: balance.free_model_usage_count || 0,
          checkin_points: balance.checkin_points || 0,
          admin_gift_points: balance.admin_gift_points || 0,
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
  },

  // 获取计费表
  async getPricingTable(): Promise<{ success: boolean; data?: PricingTable; message?: string }> {
    try {
      const response = await request.get("/api/credits/pricing-table");
      return { success: true, data: response.data };
    } catch (error: any) {
      console.error("获取计费表失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取计费表失败" 
      };
    }
  }
};