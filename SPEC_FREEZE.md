# 古法琉璃窑炉烧制质量协同规格冻结表

批次：`kiln-glaze-batch-control-20260824`  
业务方向：古法琉璃窑炉烧制批次追踪与质量协同（backend）  
目标合格题数：30；本阶段只冻结无已知 Bug 基线，不设计题面、私测、根因或任务分支。

| 领域 | 冻结规格 |
|---|---|
| 业务边界 | 窑场主管登记窑炉与釉料配方；工艺师编排烧制批次、灰度发布窑炉控制配置、执行工序；质检员复核温度曲线与出窑记录；后台 worker 处理延迟批次、健康告警与回滚；稳定 HTTP API 面向作坊微信工作台。禁止电商、支付、社交聊天、游戏、医疗、博彩、加密货币。 |
| 持久化 | PostgreSQL 16 + pgx，真实 SQL；DATABASE_URL 注入，生产状态不使用内存 map 替代。 |
| Migration | 001_initial、002_rollout_contract、003_console 版本化、顺序执行、记录 schema 版本，支持空库建表、重复启动与升级。 |
| 关联表 | devices、device_groups、firmware_releases、rollouts、rollout_batches、assignments、health_reports、alerts、audit_events、idempotency_keys、console_users、console_sessions、console_jobs 等至少 13 张有关联表，具备外键、唯一约束、索引、时间和状态字段。 |
| 事务 | 批次创建/分配、灰度推进、回滚、健康告警确认、审计写入在跨实体事务中完成；中途失败回滚并保留错误链。 |
| 状态机 | release draft→approved→scheduled→rolling→completed/rolled_back；batch pending→running→paused→succeeded/failed；alert open→acknowledged→resolved；非法迁移拒绝。 |
| 并发控制 | 行锁、版本号乐观锁、幂等键和设备唯一约束；并发灰度推进验证版本冲突与重复请求。 |
| context | HTTP request id、deadline、取消信号从 handler 传至 service/repository/worker；数据库操作检查 ctx.Err()；优雅停机取消后台循环。 |
| worker | 扫描过期批次、重试失败设备、生成健康告警；有限重试和退避，永久失败落库，支持停止与重启恢复。 |
| 错误传播 | sentinel/domain error 使用 errors.Is/As；HTTP 统一 JSON 错误码、可读消息和 request id，禁止吞错。 |
| HTTP | /healthz、/readyz；登录、会话信息、退出撤销；设备组、版本、发布、批次、健康、告警、审计 API；鉴权、panic recovery、JSON 契约。 |
| 身份与权限 | 可撤销且过期的服务端会话；角色 curator、operator、reviewer；登录/过期/logout 验证；受保护路由按角色授权。 |
| Docker | 多阶段 golang:1.26，真实 ./cmd/server 构建；复制 migrations；distroless nonroot 入口 /app/task-api；compose 提供 PostgreSQL。 |
| 测试 | 领域状态机、发布服务、事务回滚、真实 PostgreSQL 集成、HTTP 契约、并发/race、幂等、context、worker 重试/取消/恢复、权限、分页过滤排序均有测试，不依赖在线服务。 |
| 规模 | ≥5000 生产 Go 行、≥30 生产 Go 文件、≥10 实际 package、≥1500 测试 Go 行；以 measure_project.go -enforce 为准。 |
| 后续容量 | 30 个独立运行时边界覆盖灰度状态、版本冲突、设备锁、幂等、事务回滚、错误链、context、worker、会话、权限、审计、分页、时间窗和重启恢复；后续仅从 immutable baseline 设计单根因候选。 |

## 30 个运行时出题边界（容量规划）

1. 登录创建会话；2. 会话过期；3. logout 撤销；4. 角色授权；5. 设备登记；6. 设备组编排；7. 配方版本审批；8. 发布计划；9. 灰度批次创建；10. 批次设备分配；11. 批次状态迁移；12. 暂停与恢复；13. 回滚事务；14. 健康上报；15. 告警生成；16. 告警确认；17. 审计写入；18. 幂等请求；19. 乐观锁冲突；20. 行锁并发；21. 分页查询；22. 过滤排序；23. 时间窗校验；24. worker 扫描；25. worker 重试退避；26. 永久失败记录；27. context 超时；28. context 取消；29. 重启恢复；30. 健康/就绪探测。

## 禁止事项

基础版本不包含已知 Bug、gold/参考修复、任务私测、`tasks/<task_key>/red|green`、候选题面或 intake 数据；不调用遗留 seed-ai-batch、record-ai-batch、push-ai-batch 或含 gold 分支脚本。
