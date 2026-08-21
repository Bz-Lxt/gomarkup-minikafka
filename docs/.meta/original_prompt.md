# 原始需求 Prompt（存档）

> 存档时间：2026-08-20 16:19 (GMT+8)
> 来源：`/pm` 命令用户输入

---

使用Go语言实现一个带有 WAL 日志的轻量级消息队列，需要带一个后台管理和监控页面。

具体功能：

- **Pub/Sub 模型**：支持 Topic 和 Partition（分区）概念，多个 Producer 和 Consumer 并发收发消息。
- **持久化存储**：实现 WAL（Write-Ahead Log）预写日志，消息顺序追加写入磁盘文件，并维护 Segment 文件分片。
- **消费位点管理**：Consumer 消费后需提交 Offset（位点），服务器记录并持久化每个消费组的进度。
- **内存索引**：稀疏索引（Sparse Index）设计，通过 Offset 快速二分查找到磁盘中的消息。
