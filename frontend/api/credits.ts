import request from './request';

// 类型定义
export interface CreditBalance {
  available: number;
  total: number;
}

export interface ModelCost {
  id: string;
  modelName: string;
  status: 'available' | 'unavailable' | 'limited';
  costFactor?: number;
  description?: string;
}

export interface CreditUsageRecord {
  id: string;
  description: string;
  amount: number; // 消耗的积分，可为负数表示退款/补偿
  timestamp: string; // ISO Date string
  relatedModel?: string;
}

export interface GetCreditBalanceResponse {
  balance: CreditBalance;
}

export interface GetModelCostsResponse {
  costs: ModelCost[];
}

export interface GetCreditUsageHistoryResponse {
  history: CreditUsageRecord[];
  totalPages: number;
  currentPage: number;
}

// API 函数
/**
 * 获取用户积分余额
 */
export function getCreditBalance(): Promise<GetCreditBalanceResponse> {
  return request({
    url: '/credits/balance',
    method: 'get',
  });
}

/**
 * 获取模型费率
 */
export function getModelCosts(): Promise<GetModelCostsResponse> {
  return request({
    url: '/credits/model-costs',
    method: 'get',
  });
}

/**
 * 获取积分使用历史
 * @param page 页码
 * @param pageSize 每页记录数
 */
export function getCreditUsageHistory(
  page: number = 1, 
  pageSize: number = 10
): Promise<GetCreditUsageHistoryResponse> {
  return request({
    url: '/credits/history',
    method: 'get',
    params: { page, pageSize },
  });
}