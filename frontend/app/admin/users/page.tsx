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
import { Switch } from "@/components/ui/switch"
import { Users, Edit, Trash2 } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type AdminUser } from "@/api/admin"

export default function AdminUsersPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null)
  const [isUserDialogOpen, setIsUserDialogOpen] = useState(false)

  const loadUsers = async () => {
    setLoading(true)
    const result = await adminAPI.getUsers()
    if (result.success && result.users) {
      setUsers(Array.isArray(result.users) ? result.users : [])
    } else {
      setUsers([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  useEffect(() => {
    loadUsers()
  }, [])

  const handleUpdateUser = async (user: AdminUser) => {
    const result = await adminAPI.updateUser(user.id, user)
    if (result.success) {
      toast({ title: "更新成功", variant: "default" })
      setIsUserDialogOpen(false)
      loadUsers()
    } else {
      toast({
        title: "更新失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const handleDeleteUser = async (userId: number) => {
    if (!confirm("确定要删除此用户吗？")) return
    
    const result = await adminAPI.deleteUser(userId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadUsers()
    } else {
      toast({
        title: "删除失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  return (
    <DashboardLayout>
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Users className="h-6 w-6 text-blue-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">用户管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            管理系统中的所有用户账户
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold">用户列表</h3>
            </div>
            
            {loading ? (
              <div className="text-center py-8">加载中...</div>
            ) : (
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>用户名</TableHead>
                      <TableHead>邮箱</TableHead>
                      <TableHead>管理员</TableHead>
                      <TableHead>保证不降级的数量</TableHead>
                      <TableHead>注册时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array.isArray(users) && users.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell>{user.id}</TableCell>
                        <TableCell>{user.username}</TableCell>
                        <TableCell>{user.email}</TableCell>
                        <TableCell>
                          <Badge variant={user.is_admin ? "default" : "secondary"}>
                            {user.is_admin ? "是" : "否"}
                          </Badge>
                        </TableCell>
                        <TableCell>{user.degradation_guaranteed}</TableCell>
                        <TableCell>{new Date(user.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="space-x-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingUser(user)
                              setIsUserDialogOpen(true)
                            }}
                          >
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDeleteUser(user.id)}
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

          {/* 用户编辑对话框 */}
          <Dialog open={isUserDialogOpen} onOpenChange={setIsUserDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>编辑用户</DialogTitle>
                <DialogDescription>修改用户信息和权限设置</DialogDescription>
              </DialogHeader>
              {editingUser && (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label>用户名</Label>
                    <Input
                      value={editingUser.username}
                      onChange={(e) => setEditingUser({...editingUser, username: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>邮箱</Label>
                    <Input
                      value={editingUser.email}
                      onChange={(e) => setEditingUser({...editingUser, email: e.target.value})}
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={editingUser.is_admin}
                      onCheckedChange={(checked) => setEditingUser({...editingUser, is_admin: checked})}
                    />
                    <Label>管理员权限</Label>
                  </div>
                  <div className="space-y-2">
                    <Label>10条保证不降级的数量</Label>
                    <Input
                      type="number"
                      value={editingUser.degradation_guaranteed}
                      onChange={(e) => setEditingUser({...editingUser, degradation_guaranteed: parseInt(e.target.value) || 0})}
                    />
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsUserDialogOpen(false)}>取消</Button>
                <Button onClick={() => editingUser && handleUpdateUser(editingUser)}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 