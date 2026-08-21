# MiniKafka HTTP API

Base URL: `http://localhost:18591`  
版本前缀: `/api/v1`  
时间：存储为 Unix 毫秒；API 展示字段 `timestamp` 为 `yyyy-MM-dd HH:mm:ss`（GMT+8）。

统一成功信封：`{ "data": ... }`  
统一错误信封：`{ "error": { "code": "<code>", "message": "<msg>", "details": [] } }`

## 错误码

| HTTP | code | 含义 |
|------|------|------|
| 400 | invalid_json | 请求体不是合法 JSON |
| 422 | validation_error | 字段校验失败 |
| 404 | not_found | Topic / Group / Partition 不存在 |
| 409 | already_exists | Topic 已存在 |
| 500 | internal_error | 内部错误（不泄露堆栈） |

---

### GET /health

```json
{ "status": "ok", "uptime_sec": 12 }
```

### GET /api/v1/metrics

```json
{
  "data": {
    "topics": 2,
    "partitions": 6,
    "messages_total": 10240,
    "bytes_on_disk": 1048576,
    "produce_rate": 1234.5,
    "consume_rate": 980.2,
    "active_connections": 3,
    "groups": 1,
    "uptime_sec": 88
  }
}
```

### GET /api/v1/topics

```json
{
  "data": [
    {
      "name": "orders",
      "partitions": 3,
      "messages": 1000,
      "bytes": 204800,
      "created_at": "2026-08-20 16:00:00"
    }
  ]
}
```

### POST /api/v1/topics

Request:

```json
{ "name": "orders", "partitions": 3 }
```

校验：`name` 必填，`^[a-zA-Z0-9._-]{1,64}$`；`partitions` 1–16，默认 1。

Response `201`:

```json
{ "data": { "name": "orders", "partitions": 3, "created_at": "2026-08-20 16:00:00" } }
```

### GET /api/v1/topics/:name

```json
{
  "data": {
    "name": "orders",
    "partitions": [
      { "id": 0, "earliest": 0, "leo": 120, "bytes": 4096, "segments": 1 }
    ]
  }
}
```

### POST /api/v1/produce

Request:

```json
{
  "topic": "orders",
  "key": "user-1",
  "value": "hello",
  "partition": null
}
```

`partition` 省略时：有 key 则 hash 分区，无 key 则 round-robin。

Response `201`:

```json
{ "data": { "topic": "orders", "partition": 1, "offset": 42, "timestamp": "2026-08-20 16:01:00" } }
```

### POST /api/v1/produce/batch

Request:

```json
{ "topic": "orders", "messages": [ { "key": "a", "value": "1" } ] }
```

### POST /api/v1/consume

Request:

```json
{
  "topic": "orders",
  "group": "billing",
  "client_id": "c1",
  "max_messages": 50,
  "auto_commit": true
}
```

Response:

```json
{
  "data": {
    "assignments": [0, 1],
    "messages": [
      {
        "topic": "orders",
        "partition": 0,
        "offset": 10,
        "key": "a",
        "value": "1",
        "timestamp": "2026-08-20 16:01:00"
      }
    ]
  }
}
```

### POST /api/v1/groups/:group/commit

```json
{ "topic": "orders", "partition": 0, "offset": 11 }
```

提交的 offset 表示**下一条要读的位置**（与 Kafka 一致）。

### GET /api/v1/groups

```json
{
  "data": [
    {
      "group": "billing",
      "members": ["c1"],
      "lag_total": 12,
      "topics": ["orders"]
    }
  ]
}
```

### GET /api/v1/groups/:group

```json
{
  "data": {
    "group": "billing",
    "members": ["c1"],
    "partitions": [
      {
        "topic": "orders",
        "partition": 0,
        "committed": 10,
        "leo": 20,
        "lag": 10
      }
    ]
  }
}
```

### POST /api/v1/groups/:group/reset

```json
{ "topic": "orders", "to": "earliest" }
```

`to`: `earliest` | `latest`

### GET /api/v1/topics/:name/messages?partition=0&offset=0&limit=20

```json
{
  "data": {
    "messages": [ { "offset": 0, "key": "a", "value": "1", "timestamp": "..." } ],
    "next_offset": 20
  }
}
```
