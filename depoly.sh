#!/bin/bash
set -e

# SPA统一架构部署脚本
# 配置
IMAGE="ghcr.io/cloxl/claude-duck-backend:latest"

echo "开始SPA统一服务零停机部署..."

# 1. 拉取最新镜像
echo "拉取最新统一镜像..."
docker pull $IMAGE

# 2. 创建新的统一容器（不启动）
echo "创建新的统一容器..."
docker create \
  --name claude-duck-new \
  --network host \
  -e TZ=Asia/Shanghai \
  -e PORT=9998 \
  -v $(pwd)/.env:/app/.env \
  -v $(pwd)/logs:/app/logs \
  $IMAGE

echo "容器创建完成，准备进行快速切换..."

# 3. 快速切换：停止旧容器，立即启动新容器
echo "执行快速服务切换..."

# 停止旧的统一服务
docker stop claude-duck 2>/dev/null || true

# 立即启动新的统一服务
docker start claude-duck-new

echo "服务切换完成，等待服务启动..."

# 等待服务启动
sleep 5

# 4. 健康检查
echo "执行健康检查..."
if curl -f http://localhost:9998/health >/dev/null 2>&1; then
    echo "✅ 服务健康检查通过"
else
    echo "❌ 服务健康检查失败，正在回滚..."
    docker stop claude-duck-new 2>/dev/null || true
    docker start claude-duck 2>/dev/null || true
    echo "回滚完成"
    exit 1
fi

# 5. 清理旧容器，重命名新容器
echo "清理旧容器..."
docker rm claude-duck 2>/dev/null || true

echo "重命名新容器..."
docker rename claude-duck-new claude-duck

# 6. 更新重启策略
echo "更新重启策略..."
docker update --restart unless-stopped claude-duck

echo "SPA统一部署完成！"
echo "统一服务地址:"
echo "  - 前端页面: http://localhost:9998/"
echo "  - API服务: http://localhost:9998/api"
echo "  - 健康检查: http://localhost:9998/health"
echo ""
echo "服务状态:"
docker ps | grep claude-duck