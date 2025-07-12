import { request } from "./request";

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
  }).then(response => response.data);
}

// 公告接口定义
export interface PublicAnnouncement {
  id: number;
  type: string;
  title: string;
  description: string;
  language: string;
  created_at: string;
}

// 获取公告数据
export const announcementsAPI = {
  // 获取活跃公告列表（用户端）
  async getActiveAnnouncements(language: string = "zh"): Promise<{ success: boolean; data?: PublicAnnouncement[]; message?: string }> {
    try {
      const response = await request.get(`/api/announcements?language=${language}&active=true`);
      return { success: true, data: response.data.announcements };
    } catch (error: any) {
      console.error("获取公告失败:", error);
      return { 
        success: false, 
        message: error.response?.data?.error || "获取公告失败" 
      };
    }
  }
};