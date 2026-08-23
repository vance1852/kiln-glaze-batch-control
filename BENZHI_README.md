工业设备固件灰度发布与回滚编排后端，负责发布活动、部署波次、安装回报、健康告警和审计追踪。

## 标准构建、运行和测试命令

进入容器后执行：

```bash
# 编译
cd '/app' && GOTOOLCHAIN=local go build ./...

# 启动
cd '/app' && GOTOOLCHAIN=local go run ./cmd/server

# 测试
cd '/app' && GOTOOLCHAIN=local go test ./...
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-task-169-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-169-arm64 linux/arm64
docker run -it benzhi-task-169-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-169-arm64:latest
```

容器启动时会自动初始化本地 PostgreSQL，并通过 `DATABASE_URL` 供集成测试和服务进程使用。
