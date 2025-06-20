import request from './request';

export interface Announcement {
  id: number;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  description: string; // 支持HTML
  language: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AnnouncementsResponse {
  announcements: Announcement[];
}

// 获取公告列表
export function getAnnouncements(): Promise<AnnouncementsResponse> {
  return request({
    url: '/announcements',
    method: 'get',
  });
}
