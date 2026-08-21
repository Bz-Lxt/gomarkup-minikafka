# QA Record

## Round 1 — 2026-08-20 17:18 (GMT+8)

**Cost**: ¥0（无外部计费 API，全程 Mock/离线本地 WAL）

### 环境
- `docker compose up --build -d`（开发端口 18591）
- 单元测试：宿主机 `go test ./...` 与 `-race` 通过；镜像构建阶段曾执行 `go test` 亦通过
- API Smoke：`docker compose --profile qa run --rm qa`（容器内，`BASE_URL=http://minikafka:8080`）

### 结果
```
[PASS] Docker Build
[PASS] Health Check
[PASS] SPA
[PASS] Topic / Produce / Consume / Offset / Metrics
[PASS] go test ./internal/wal ./internal/offset ./internal/broker
[PASS] go test -race
```

Playwright `tests/e2e_flow.spec.ts` 已落地，本轮未安装浏览器运行时；等价路径由 Docker 内 API smoke + SPA 壳（index 含 MiniKafka）覆盖。

### Round 1 失败（已修复后复测）
1. **Topic 名硬编码导致 409**：首次宿主机 smoke 创建 `smoke-orders` 后，QA 容器复跑 `POST /topics` expect 201 失败。
   - 修复：topic/group 使用时间戳唯一名。
   - 复测：PASS
2. **`docker compose --profile qa up --abort-on-container-exit` 连带停掉 Broker**：不符合「在运行中的服务上测」。
   - 修复：改为 `docker compose --profile qa run --rm qa`。
   - 复测：PASS
3. **端口 18481 被 minigate-gateway 占用**：compose 启动失败。
   - 修复：改随机端口 18591。
   - 复测：PASS

### 结论
Round 1 最终 **PASS**，进入审计。
