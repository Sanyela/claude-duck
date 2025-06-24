"use client"

import { useEffect, useState } from "react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Bell, InfoIcon, AlertTriangle, AlertCircle, CheckCircle2 } from "lucide-react"
import { getAnnouncements } from "@/api/announcements"

interface Announcement {
  id: number
  type: 'info' | 'warning' | 'error' | 'success'
  title: string
  description: string
  language: string
}

export function AnnouncementsCard() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  useEffect(() => {
    const fetchAnnouncements = async () => {
      try {
        const response = await getAnnouncements("zh")
        setAnnouncements(response.announcements || [])
      } catch (err) {
        console.error("获取公告失败:", err)
        setError("无法加载公告信息")
      } finally {
        setLoading(false)
      }
    }
    
    fetchAnnouncements()
  }, [])
  
  const getAnnouncementIcon = (type: string) => {
    switch (type) {
      case 'info':
        return <InfoIcon className="h-4 w-4" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4" />
      case 'error':
        return <AlertCircle className="h-4 w-4" />
      case 'success':
        return <CheckCircle2 className="h-4 w-4" />
      default:
        return <InfoIcon className="h-4 w-4" />
    }
  }
  
  const getAnnouncementVariant = (type: string): "default" | "destructive" | null => {
    switch (type) {
      case 'error':
        return "destructive"
      default:
        return "default"
    }
  }
  
  return (
    <Card className="overflow-hidden border border-slate-200 dark:border-slate-700">
      <CardHeader className="bg-gradient-to-r from-amber-500 to-orange-600 text-white">
        <CardTitle className="flex items-center gap-2">
          <Bell className="h-5 w-5" />
          系统公告
        </CardTitle>
        <CardDescription className="text-amber-100">重要通知与更新</CardDescription>
      </CardHeader>
      <CardContent className="pt-6">
        {loading ? (
          <div className="space-y-3">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        ) : error ? (
          <div className="text-red-500">{error}</div>
        ) : announcements.length === 0 ? (
          <div className="text-center py-6 text-slate-500 dark:text-slate-400">
            暂无系统公告
          </div>
        ) : (
          <div className="space-y-4">
            {announcements.map((announcement) => (
              <Alert 
                key={announcement.id}
                variant={getAnnouncementVariant(announcement.type)}
              >
                <div className="flex items-center gap-2">
                  {getAnnouncementIcon(announcement.type)}
                  <AlertTitle>{announcement.title}</AlertTitle>
                </div>
                <AlertDescription 
                  className="mt-2" 
                  dangerouslySetInnerHTML={{ __html: announcement.description }}
                />
              </Alert>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}