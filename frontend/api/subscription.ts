import request from './request';

// 类型定义
export interface SubscriptionPlan {
  id: string;
  name: string;
  pricePerMonth: number;
  currency: string;
  features: string[];
}

export interface ActiveSubscription {
  id: string;
  plan: SubscriptionPlan;
  status: 'active' | 'canceled' | 'past_due';
  currentPeriodEnd: string; // ISO Date string
  cancelAtPeriodEnd: boolean;
}

export interface SubscriptionHistoryItem {
  id: string;
  planName: string;
  amount: number;
  currency: string;
  date: string; // ISO Date string
  status: 'paid' | 'failed';
  invoiceUrl?: string;
}

export interface GetActiveSubscriptionResponse {
  subscription: ActiveSubscription | null;
}

export interface GetSubscriptionHistoryResponse {
  history: SubscriptionHistoryItem[];
}

export interface RedeemCouponRequest {
  couponCode: string;
}

export interface RedeemCouponResponse {
  success: boolean;
  message: string;
  newSubscription?: ActiveSubscription;
}

// API 函数
/**
 * 获取活跃订阅
 */
export function getActiveSubscription(): Promise<GetActiveSubscriptionResponse> {
  return request({
    url: '/subscription/active',
    method: 'get',
  });
}

/**
 * 获取订阅历史
 */
export function getSubscriptionHistory(): Promise<GetSubscriptionHistoryResponse> {
  return request({
    url: '/subscription/history',
    method: 'get',
  });
}

/**
 * 兑换激活码
 */
export function redeemCoupon(data: RedeemCouponRequest): Promise<RedeemCouponResponse> {
  return request({
    url: '/subscription/redeem',
    method: 'post',
    data,
  });
}