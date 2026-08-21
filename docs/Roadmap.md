# MiniKafka — 实施路线图（WHEN）

> 权威需求见 `docs/Requirements.md`。本文件定义构建顺序与完成判据。
> 预估规模 < 10,000 LoC，单期交付，无 MVP / V1 / V2 切分必要。

---

## 阶段顺序决策

**选择：UI-First（默认）**

**理由**：管理监控页是仪表盘/CRUD 形态，组件结构可对照 API schema 草图（Topic / Partition / Group / Offset / Metrics）先行搭建，不必等 WAL 引擎落地。

> 非 Logic-First：UI 不是从数据模型推导出的编辑器/时间线/画布。

---

## 目录结构

```
MiniKafka/
├── backend/                 # Go Broker（WAL / 索引 / Offset / HTTP）
├── frontend-admin/          # Vue 3 管理监控台
├── frontend-mp/             # N/A（无小程序端）
├── frontend-user/           # N/A（无 C 端）
├── docs/
├── tests/                   # API smoke + Playwright E2E
├── docker-compose.yml
└── Dockerfile               # 多阶段：前端构建 → Go 编译 → Alpine 运行
```

---

## Phase 1 — 架构骨架  [x]

- [x] `git init` + `.gitignore`
- [x] `docker-compose.yml`（开发随机端口 **18591**）
- [x] 标准目录落地
- [x] `docs/API.md` 契约草图（供 UI-First）

## Phase 2 — 管理监控 UI  [x]

- [x] `docs/DesignSpec.md`
- [x] 概览仪表盘（F16, F20）
- [x] Topic 管理页（F17）
- [x] 消费组监控页（F18）
- [x] 消息浏览页（F19）
- [x] 创建 Topic / 发送测试消息 / 位点重置（带确认弹窗与表单校验）

## Phase 3 — Broker 核心  [x]

- [x] WAL Record 编解码 + CRC + 截断恢复（F6, F8）
- [x] Segment 分片滚动（F7）
- [x] 稀疏索引 + 二分查找（F13, F14, F15）
- [x] Topic / Partition / Producer / Consumer Group（F1–F5）
- [x] Offset 持久化 / Lag / 重置（F10–F12）
- [x] 可配置 fsync（F9）
- [x] HTTP API 对齐 `docs/API.md`
- [x] Dockerfile 多阶段 + 时区 `Asia/Shanghai`
- [x] 核心引擎单元测试（N9）

## Phase 4 — QA  [x]

- [x] `go test ./...` + `-race`
- [x] `tests/api_smoke.py`（Health / Topic / Produce / Consume / Offset）
- [x] `tests/e2e_flow.spec.ts` 关键用户路径
- [x] 记录 `docs/QA_Record.md`（Cost ¥0）

## Phase 5 — 审计  [x]

- [x] `docs/AuditReport.md`
- [x] Knowledge Harvest `/learn`

---

## 完成判据

全部 F1–F20 可在 `docker compose up --build -d` 后通过 `localhost:18591` 演示；N1–N9 由测试与文档覆盖。
