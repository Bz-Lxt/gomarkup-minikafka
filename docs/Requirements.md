# MiniKafka — 需求文档（SSOT · WHAT）

> 版本：v1.0（冻结于 2026-08-20，PM 阶段产出）
> 原始 Prompt：见 `docs/.meta/original_prompt.md`
> 本文件定义 **做什么**；何时做、分几期做由 `docs/Roadmap.md` 定义。

---

## 1. 项目概述

使用 **Go 语言**实现一个轻量级消息队列系统 **MiniKafka**，核心是一个支持 Pub/Sub 模型的单节点 Broker，具备 **WAL 预写日志持久化**、**Segment 文件分片**、**消费位点管理**与**内存稀疏索引**四项核心能力，并附带一个**后台管理与监控 Web 页面**。

项目定位为教学级/轻量级 Kafka 简化实现，单机部署，Docker 一键交付。

## 2. 功能需求

### 2.1 Pub/Sub 模型（核心引擎）

| 编号 | 需求 | 说明 |
|---|---|---|
| F1 | Topic 管理 | 支持创建/列举 Topic；每个 Topic 可配置 Partition 数量（创建时指定，默认 1） |
| F2 | Partition 模型 | 消息按 Partition 组织；单 Partition 内消息严格有序，Offset 单调递增 |
| F3 | Producer 生产 | 多个 Producer 可并发向同一/不同 Topic 发送消息；支持指定 Key（按 Key 哈希分区）或由 Broker 轮询分配分区 |
| F4 | Consumer 消费 | 多个 Consumer 以**消费组（Consumer Group）**形式并发订阅 Topic；同组内 Partition 互斥分配，不同组之间互不影响 |
| F5 | 并发安全 | 所有并发收发路径无数据竞争，必须通过 `go test -race` 检测 |

### 2.2 持久化存储（WAL + Segment）

| 编号 | 需求 | 说明 |
|---|---|---|
| F6 | WAL 顺序追加 | 消息先写入 Write-Ahead Log（顺序追加磁盘文件），写入成功后才视为生产成功 |
| F7 | Segment 分片 | 日志文件按大小/条数阈值切分为 Segment（如 `00000000000000000000.log`），旧 Segment 只读，活跃 Segment 可写 |
| F8 | 崩溃恢复 | 进程重启后能从 WAL 完整恢复消息与索引；截断尾部可能损坏的不完整记录 |
| F9 | 刷盘策略 | fsync 策略可配置（每条消息 / 批量 / 定时），默认批量刷盘并在文档中说明持久化保证边界 |

### 2.3 消费位点管理（Offset）

| 编号 | 需求 | 说明 |
|---|---|---|
| F10 | Offset 提交 | Consumer 消费后向 Broker 提交 Offset；支持自动提交与手动提交两种模式 |
| F11 | 位点持久化 | Broker 将每个消费组 × Topic × Partition 的进度持久化到磁盘（独立元数据文件），重启后不丢失 |
| F12 | 位点查询/重置 | 支持查询消费组进度（Lag = 最新 Offset − 已提交 Offset）；支持将消费组位点重置到最早/最新 |

### 2.4 内存稀疏索引（Sparse Index）

| 编号 | 需求 | 说明 |
|---|---|---|
| F13 | 稀疏索引结构 | 每个 Segment 维护内存稀疏索引（Offset → 磁盘物理位置），索引间隔可配置（默认每 4KB 数据一条索引项） |
| F14 | 二分查找定位 | 消费时通过目标 Offset 在稀疏索引上二分查找，定位到最近的索引项后顺序扫描至目标消息 |
| F15 | 索引重建 | 重启时通过扫描 Segment 文件重建内存索引（或从配套 `.index` 文件加载），恢复时间须满足 N4 基线 |

### 2.5 后台管理与监控页面（Web UI）

| 编号 | 需求 | 说明 |
|---|---|---|
| F16 | 概览仪表盘 | 展示 Broker 运行状态：消息总量、生产/消费速率（msg/s）、Topic 数、活跃连接数、磁盘占用 |
| F17 | Topic 管理页 | Topic 列表、创建 Topic、查看每个 Partition 的 LEO（Log End Offset）与最早 Offset |
| F18 | 消费组监控页 | 消费组列表、每组每分区的已提交 Offset 与 Lag 实时展示 |
| F19 | 消息浏览 | 可按 Topic + Partition + Offset 范围查询并查看消息内容（Key/Value/时间戳） |
| F20 | 实时性 | 监控数据自动刷新（轮询 ≤ 5s 间隔或 SSE/WebSocket 推送） |

## 3. 非功能需求与验收基线（可测量）

| 编号 | 基线 | 验收标准 |
|---|---|---|
| N1 | 生产吞吐 | 单 Partition、1KB 消息、批量追加模式下 ≥ **50,000 msg/s**（本地容器环境基准测试脚本实测） |
| N2 | 消费吞吐 | 顺序消费吞吐不低于生产吞吐的 **70%** |
| N3 | 持久化保证 | fsync 已确认的消息在进程崩溃（kill -9）后**零丢失**；崩溃恢复自动截断不完整尾部记录 |
| N4 | 恢复时间 | 100 万条消息规模的 Segment 数据，重启恢复（含索引重建）≤ **5 秒** |
| N5 | 索引效率 | 稀疏索引定位单条消息的平均磁盘读取次数 ≤ **2 次**；索引内存占用 ≤ 消息数据的 1%（默认 4KB 间隔） |
| N6 | 位点可靠性 | 消费组 Offset 提交后重启不丢进度；Lag 计算准确（与实测收发计数一致） |
| N7 | 并发 | ≥ 10 Producer + ≥ 10 Consumer（≥ 3 个消费组）并发运行 60s 无死锁、无消息错乱（Partition 内有序性 100%） |
| N8 | Docker 交付 | `docker compose up --build -d` 一键启动，监控页面可通过 `localhost` 访问；镜像支持 ARM64 + AMD64 |
| N9 | 代码质量 | 后端核心引擎（WAL/Segment/索引/Offset）单元测试覆盖；提供统一 Logger（含 level 控制），无散落 `fmt.Println` 调试输出 |

## 4. 技术栈决策（PM 预研结论）

| 层 | 选型 | 理由 |
|---|---|---|
| Broker 核心 | Go（标准库为主） | 原生命令要求；goroutine 天然契合并发收发；静态编译利于 Docker 多架构交付 |
| 对外协议 | HTTP/JSON API（生产/消费/位点/管理） | 简化客户端实现与测试；监控页面与 Broker 同源通信 |
| 管理前端 | Vue 3 + Vite + Element Plus（或同级现代框架） | 满足"美学卓越"红线；构建产物由 Broker 或独立容器托管 |
| 存储 | 本地磁盘文件（WAL Segment + `.index` + offset 元数据文件） | 无外部依赖，符合轻量定位 |
| 部署 | Docker Compose（单服务或 broker+web 双服务） | 满足 Docker 交付红线 |

> 无外部 API 依赖 → **不涉及** Contract Gate 与成本护栏；无需 Mock Provider。

## 5. 兼容性与逻辑检查记录

- **矛盾检测**：Prompt 内部无冲突。WAL 顺序追加与并发生产的矛盾点（多 Producer 并发写同一 Partition 时的顺序性）已在 F2/F5 中以"Partition 内有序 + 并发安全"约束明确。
- **Docker 交付可行性**：✅ Go 静态编译交叉构建 ARM64/AMD64 无障碍；Web 页面经 `localhost` 暴露，满足红线 1。
- **规模评估**：预估 Go 后端 ~4,000–6,000 LoC + 前端 ~1,500–2,500 LoC + 测试 ~1,500 LoC，**总量 < 10,000 LoC → 直接接受**。

## 6. 范围边界（Out of Scope）

以下内容**不属于**本项目范围，防止范围蔓延：

1. 多 Broker 集群、副本复制（Replication）、Leader 选举
2. 消息事务、Exactly-Once 语义（仅保证 At-Least-Once + 位点管理）
3. 消息过期清理（Retention）之外的复杂生命周期管理（Retention 可做可选项）
4. 认证/鉴权体系（监控页可留简单只读访问，不接 SSO）
5. 多语言客户端 SDK（仅提供 HTTP API + 文档示例）

---

**冻结声明**：本需求文档经 PM Agent 评估通过后冻结，后续变更须重新走 `/pm` 流程修订。
