# Firmware Rollout Control

面向工业设备集群的固件灰度发布与回滚编排后端，使用 Go 1.26、PostgreSQL 16 和 pgx。系统覆盖发布操作员登录与会话生命周期、设备和发布活动管理、部署作业、灰度波次审批、安装回报、激活事件、健康告警、审计链和后台重试。

## 启动

```bash
docker compose up -d postgres
go test ./... -count=1
go run ./cmd/server
```

默认数据库连接为 `postgres://firmware:firmware@localhost:5432/firmware_rollout_control?sslmode=disable`，可使用 `DATABASE_URL` 覆盖。`GET /healthz` 检查进程，`GET /readyz` 检查 PostgreSQL。开发账号为 `admin / admin123`，登录、当前身份、退出撤销、角色授权和业务 API 均有自动化测试覆盖。
