# 统一Docker构建 - SPA模式
# 第一阶段：构建前端SPA
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json frontend/pnpm-lock.yaml ./

# 安装 pnpm 并安装依赖
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# 复制前端源码
COPY frontend/ .

# 构建前端SPA（静态导出）
RUN pnpm build

# 第二阶段：构建后端
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates

# 复制 Go 模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY . .

# 删除前端目录（避免冲突）
RUN rm -rf frontend

# 构建后端应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 第三阶段：统一运行时镜像
FROM alpine:latest

WORKDIR /app

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata wget && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 复制构建好的后端应用
COPY --from=backend-builder /app/main .

# 创建前端静态文件目录
RUN mkdir -p ui/dist

# 复制构建好的前端SPA文件
COPY --from=frontend-builder /app/frontend/dist/ ui/dist/

# 创建非 root 用户
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser && \
    chown -R appuser:appgroup /app

USER appuser

# 暴露端口
EXPOSE 9998

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9998/health || exit 1

# 启动应用
CMD ["./main"] 