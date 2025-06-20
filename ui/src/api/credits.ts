import request from './request';

export interface CreditBalance {
  available: number;
  total: number;
  rechargeRatePerHour: number;
  canRequestReset: boolean; // 是否可以申请重置
  nextResetTime?: string; // ISO Date string, 如果有限制的话
}

export interface ModelCost {
  id: string;
  modelName: string;
  status: 'available' | 'unavailable' | 'limited';
  costFactor?: number; // 相对成本因子，或具体价格
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
export function getCreditBalance(): Promise<GetCreditBalanceResponse> {
  return request({
    url: '/credits/balance',
    method: 'get',
  });
}

export function getModelCosts(): Promise<GetModelCostsResponse> {
  return request({
    url: '/credits/model-costs',
    method: 'get',
  });
}

export function getCreditUsageHistory(page = 1, pageSize = 10): Promise<GetCreditUsageHistoryResponse> {
  return request({
    url: '/credits/history',
    method: 'get',
    params: { page, pageSize },
  });
}

export function requestCreditReset(): Promise<{success: boolean, message: string, nextAvailableTime?: string}> {
    return request({
        url: '/credits/request-reset',
        method: 'post'
    });
}
