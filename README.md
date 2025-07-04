# Claude Duck 部署脚本

## 自动化部署流程

### 1. 创建并推送标签触发构建

```bash
# 创建标签 (格式: vyyyymmddhhmm)
git tag v202507101934

# 推送标签到远程仓库 (这将触发自动构建)
git push origin v202507101934
```

### 2. 部署步骤

**设置环境变量:**
```bash
export GITHUB_REPOSITORY=cloxl/claude-duck
```

**停止现有容器:**
```bash
docker-compose -f docker-compose.prod.yml down
```

**拉取最新镜像:**
```bash
docker-compose -f docker-compose.prod.yml pull
```

**启动服务:**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

**检查服务状态:**
```bash
docker-compose -f docker-compose.prod.yml ps
```

**部署完成！访问地址:** http://localhost:9998

## 使用说明

1. **标签格式**: `vyyyymmddhhmm`
   - `v`: 版本前缀
   - `yyyy`: 年份 (4位)
   - `mm`: 月份 (2位)
   - `dd`: 日期 (2位)
   - `hh`: 小时 (2位)
   - `mm`: 分钟 (2位)

2. **示例标签**: `v202507101934` (2025年7月10日19点34分)

3. **部署流程**:
   - 推送标签后自动触发构建
   - 构建完成后运行部署脚本
   - 服务将在 http://localhost:9998 可用

## 注意事项

- 确保 Docker 和 Docker Compose 已正确安装
- 确保有权限访问 GitHub 仓库
- 部署前请检查 `docker-compose.prod.yml` 配置文件