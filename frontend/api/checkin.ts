import { request } from './request';

export interface CheckinPointsRange {
  minPoints: number;
  maxPoints: number;
  hasValid: boolean;
}

// 签到状态响应
export interface CheckinStatusResponse {
  canCheckin: boolean;      // 是否可以签到
  todayChecked: boolean;    // 今天是否已签到
  lastCheckinDate: string;  // 最后签到日期
  pointsRange: CheckinPointsRange;
}

// 签到响应
export interface CheckinResponse {
  success: boolean;
  message: string;
  rewardPoints: number;     // 获得的奖励积分
}

// 签到API
export const checkinAPI = {
  // 获取签到状态
  getStatus: async (): Promise<CheckinStatusResponse> => {
    const response = await request.get('/api/checkin/status');
    return response.data;
  },

  // 执行签到
  checkin: async (): Promise<CheckinResponse> => {
    const response = await request.post('/api/checkin');
    return response.data;
  }
}; 