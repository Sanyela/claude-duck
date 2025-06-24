"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { Megaphone, Plus, Edit, Trash2 } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type Announcement } from "@/api/admin"

export default function AdminAnnouncementsPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [editingAnnouncement, setEditingAnnouncement] = useState<Announcement | null>(null)
  const [isAnnouncementDialogOpen, setIsAnnouncementDialogOpen] = useState(false)
  const [isCreateAnnouncementMode, setIsCreateAnnouncementMode] = useState(false)

  const loadAnnouncements = async () => {
    setLoading(true)
    const result = await adminAPI.getAnnouncements()
    if (result.success && result.announcements) {
      setAnnouncements(Array.isArray(result.announcements) ? result.announcements : [])
    } else {
      setAnnouncements([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  useEffect(() => {
    loadAnnouncements()
  }, [])

  const handleSaveAnnouncement = async (announcement: Announcement) => {
    const result = isCreateAnnouncementMode
      ? await adminAPI.createAnnouncement(announcement)
      : await adminAPI.updateAnnouncement(announcement.id, announcement)
    
    if (result.success) {
      toast({ title: isCreateAnnouncementMode ? "创建成功" : "更新成功", variant: "default" })
      setIsAnnouncementDialogOpen(false)
      loadAnnouncements()
    } else {
      toast({
        title: isCreateAnnouncementMode ? "创建失败" : "更新失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const handleDeleteAnnouncement = async (announcementId: number) => {
    if (!confirm("确定要删除此公告吗？")) return
    
    const result = await adminAPI.deleteAnnouncement(announcementId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadAnnouncements()
    } else {
      toast({
        title: "删除失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case "info": return "bg-blue-500 text-white"
      case "warning": return "bg-yellow-500 text-white"
      case "error": return "bg-red-500 text-white"
      case "success": return "bg-green-500 text-white"
      default: return "bg-gray-500 text-white"
    }
  }

  const getTypeName = (type: string) => {
    switch (type) {
      case "info": return "信息"
      case "warning": return "警告"
      case "error": return "错误"
      case "success": return "成功"
      default: return "未知"
    }
  }

  return (
    <DashboardLayout>
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Megaphone className="h-6 w-6 text-blue-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">公告管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            管理系统公告和通知消息
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold">公告列表</h3>
              <Button
                onClick={() => {
                  setEditingAnnouncement({
                    title: "",
                    description: "",
                    type: "info",
                    language: "zh",
                    active: true
                  } as Announcement)
                  setIsCreateAnnouncementMode(true)
                  setIsAnnouncementDialogOpen(true)
                }}
                className="bg-blue-500 hover:bg-blue-600 text-white"
              >
                <Plus className="w-4 h-4 mr-2" />
                新建公告
              </Button>
            </div>
            
            {loading ? (
              <div className="text-center py-8">加载中...</div>
            ) : (
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>标题</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>语言</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array.isArray(announcements) && announcements.map((announcement) => (
                      <TableRow key={announcement.id}>
                        <TableCell className="font-medium">{announcement.title}</TableCell>
                        <TableCell>
                          <Badge className={getTypeColor(announcement.type)}>
                            {getTypeName(announcement.type)}
                          </Badge>
                        </TableCell>
                        <TableCell>{announcement.language === "zh" ? "中文" : "英文"}</TableCell>
                        <TableCell>
                          <Badge variant={announcement.active ? "default" : "secondary"}>
                            {announcement.active ? "启用" : "禁用"}
                          </Badge>
                        </TableCell>
                        <TableCell>{new Date(announcement.created_at).toLocaleString()}</TableCell>
                        <TableCell className="space-x-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingAnnouncement(announcement)
                              setIsCreateAnnouncementMode(false)
                              setIsAnnouncementDialogOpen(true)
                            }}
                          >
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDeleteAnnouncement(announcement.id)}
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          {/* 公告编辑对话框 */}
          <Dialog open={isAnnouncementDialogOpen} onOpenChange={setIsAnnouncementDialogOpen}>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{isCreateAnnouncementMode ? "创建公告" : "编辑公告"}</DialogTitle>
                <DialogDescription>设置公告的详细信息</DialogDescription>
              </DialogHeader>
              {editingAnnouncement && (
                <div className="space-y-4 max-h-96 overflow-y-auto px-2">
                  <div className="space-y-2">
                    <Label>标题</Label>
                    <Input
                      value={editingAnnouncement.title}
                      onChange={(e) => setEditingAnnouncement({...editingAnnouncement, title: e.target.value})}
                      placeholder="输入公告标题"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>内容描述</Label>
                    <Textarea
                      value={editingAnnouncement.description}
                      onChange={(e) => setEditingAnnouncement({...editingAnnouncement, description: e.target.value})}
                      placeholder="输入公告内容，支持HTML标签"
                      className="resize-none min-h-[100px]"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>公告类型</Label>
                      <Select
                        value={editingAnnouncement.type}
                        onValueChange={(value) => setEditingAnnouncement({...editingAnnouncement, type: value})}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="info">信息</SelectItem>
                          <SelectItem value="warning">警告</SelectItem>
                          <SelectItem value="error">错误</SelectItem>
                          <SelectItem value="success">成功</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label>语言</Label>
                      <Select
                        value={editingAnnouncement.language}
                        onValueChange={(value) => setEditingAnnouncement({...editingAnnouncement, language: value})}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="zh">中文</SelectItem>
                          <SelectItem value="en">英文</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={editingAnnouncement.active}
                      onCheckedChange={(checked) => setEditingAnnouncement({...editingAnnouncement, active: checked})}
                    />
                    <Label>启用公告</Label>
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsAnnouncementDialogOpen(false)}>取消</Button>
                <Button onClick={() => editingAnnouncement && handleSaveAnnouncement(editingAnnouncement)}>
                  {isCreateAnnouncementMode ? "创建" : "保存"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 