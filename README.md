# MiniKafka

带 WAL 的轻量消息队列（Go）+ 深海观测台管理页。

```bash
docker compose up --build -d
```

浏览器打开 http://localhost:18591

健康检查：`curl http://localhost:18591/health`

需求与契约见 `docs/Requirements.md`、`docs/API.md`。完整交付说明将在 `/deploy` 阶段写入本 README 的标准七章。
