import { request } from "./request";

// 积分余额接口
export interface PointBalance {
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

// 订阅信息接口
export interface ActiveSubscription {
  id: string;
  plan: {
    id: string;
    name: string;
    features: string[];
  };
  status: string;
  currentPeriodEnd: string;
  cancelAtPeriodEnd: boolean;
  availablePoints: number;
  totalPoints: number;
  usedPoints: number;
  activatedAt: string;
  detailedStatus: '有效' | '已用完' | '已过期';
  isCurrentUsing: boolean;
}

// API响应格式
export interface DashboardData {
  pointBalance: PointBalance | null;
  subscriptions: ActiveSubscription[];
}

// 获取仪表盘数据
export const dashboardAPI = {
  // 获取用户积分余额
  async getPointBalance(): Promise<{ success: boolean; data?: PointBalance; message?: string }> {
    try {
      const response = await request.get("/api/credits/balance");
      
      // 检查并处理后端返回的数据结构
      if (response.data && response.data.balance) {
        // 将后端返回的balance格式转换为前端需要的PointBalance格式
        const balance = response.data.balance;
        const pointBalance: PointBalance = {
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
        return { success: true, data: pointBalance };
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

  // 获取用户所有订阅
  async getActiveSubscription(): Promise<{ success: boolean; data?: ActiveSubscription[]; message?: string }> {
    try {
      const response = await request.get("/api/subscription/active");
      return { 
        success: true, 
        data: response.data.subscriptions || []
      };
    } catch (error: any) {
      console.error("获取订阅信息失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取订阅信息失败" 
      };
    }
  },

  // 获取完整仪表盘数据
  async getDashboardData(): Promise<{ success: boolean; data?: DashboardData; message?: string }> {
    try {
      const [pointBalanceResult, subscriptionResult] = await Promise.all([
        this.getPointBalance(),
        this.getActiveSubscription()
      ]);

      return {
        success: true,
        data: {
          pointBalance: pointBalanceResult.success ? pointBalanceResult.data || null : null,
          subscriptions: subscriptionResult.success ? subscriptionResult.data || [] : []
        }
      };
    } catch (error: any) {
      console.error("获取仪表盘数据失败:", error);
      return { 
        success: false, 
        message: "获取仪表盘数据失败" 
      };
    }
  }
};