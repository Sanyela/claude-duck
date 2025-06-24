import request from './request';

// 类型定义
export interface Announcement {
  id: number;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  description: string;
  language: string;
}

export interface GetAnnouncementsResponse {
  announcements: Announcement[];
}

/**
 * 获取系统公告
 * @param language 语言代码
 */
export function getAnnouncements(language: string = 'zh'): Promise<GetAnnouncementsResponse> {
  return request({
    url: '/announcements',
    method: 'get',
    params: { language },
  });
}