#!/bin/bash
set -e

# 配置
BACKEND_IMAGE="ghcr.io/cloxl/claude-duck-backend:latest"
FRONTEND_IMAGE="ghcr.io/cloxl/claude-duck-frontend:latest"

echo "开始零停机部署..."

# 1. 拉取最新镜像
echo "拉取最新镜像..."
docker pull $BACKEND_IMAGE
docker pull $FRONTEND_IMAGE

# 2. 创建新的后端容器（不启动）
echo "创建新的后端容器..."
docker create \
  --name claude-duck-backend-new \
  --network host \
  -e TZ=Asia/Shanghai \
  -e PORT=9998 \
  -v $(pwd)/.env:/app/.env \
  -v $(pwd)/logs:/app/logs \
  $BACKEND_IMAGE

# 3. 创建新的前端容器（不启动）
echo "创建新的前端容器..."
docker create \
  --name claude-duck-frontend-new \
  --network host \
  -e TZ=Asia/Shanghai \
  -e API_URL=http://154.219.117.38:9998 \
  -e APP_NAME=Claude Duck \
  -e INSTALL_COMMAND="npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com" \
  -e DOCS_URL=https://swjqc4r0111.feishu.cn/docx/CJT6dbdUBofDlrxfwpNcp1klnCg \
  $FRONTEND_IMAGE

echo "容器创建完成，准备进行快速切换..."

# 4. 快速切换：停止旧容器，立即启动新容器
echo "执行快速切换..."

# 停止旧的前端服务
docker stop claude-duck-frontend 2>/dev/null || true

# 立即启动新的前端服务
docker start claude-duck-frontend-new

# 停止旧的后端服务
docker stop claude-duck-backend 2>/dev/null || true

# 立即启动新的后端服务
docker start claude-duck-backend-new

echo "服务切换完成，等待服务启动..."

# 6. 清理旧容器，重命名新容器
echo "清理旧容器..."
docker rm claude-duck-backend 2>/dev/null || true
docker rm claude-duck-frontend 2>/dev/null || true

echo "重命名新容器..."
docker rename claude-duck-backend-new claude-duck-backend
docker rename claude-duck-frontend-new claude-duck-frontend

# 7. 更新重启策略
echo "更新重启策略..."
docker update --restart unless-stopped claude-duck-backend
docker update --restart unless-stopped claude-duck-frontend

echo "部署完成！"
echo "后端服务: http://localhost:9998"
echo "前端服务: http://localhost:3000"
echo ""
echo "服务状态:"
docker ps | grep claude-duck