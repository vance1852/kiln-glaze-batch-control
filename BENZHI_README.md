# BENZHI_README

## 项目说明

- 项目：vance1852/kiln-glaze-batch-control
- 项目用途：面向工业设备集群的固件灰度发布与回滚编排后端，使用 Go 1.26、PostgreSQL 16 和 pgx。系统覆盖发布操作员登录与会话生命周期、设备和发布活动管理、部署作业、灰度波次审批、安装回报、激活事件、健康告警、审计链和后台重试。
- Go 工具链：`golang:1.26`
- 前端工具链：无

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
./build_benzhi_docker.sh benzhi-task-172-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-172-arm64 linux/arm64
docker run -it benzhi-task-172-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-172-arm64:latest
```

## 题目验证命令

1. 预期退出码 0：`go test ./internal/worker -run '^TestLeaseConcurrentAcquireHasSingleOwner$' -count=1`
2. 预期退出码 0：`go test ./...`
3. 预期退出码 0：`GOTOOLCHAIN=local go build -buildvcs=false ./... && GOTOOLCHAIN=local go vet ./...`
