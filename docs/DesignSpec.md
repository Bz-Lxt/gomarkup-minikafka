# MiniKafka 设计规范

> 美学方向：**深海观测台（Deep-Sea Observatory）**
> 不是通用 SaaS 仪表盘。这是一台给消息队列值班员用的监控台：暗舱、扫描线、等宽数字、冷青热琥珀。

---

## 1. 概念

Broker 像一艘潜水器的声呐室。左侧是舰桥导航，主区是实时声呐回波（吞吐曲线、Lag 热度、Segment 条带）。数字一律等宽，状态用青（健康）/ 琥珀（滞后）/ 珊瑚红（异常）三色语义，避免彩虹盘。

记忆点：页面背景有极淡的网格与噪声颗粒；指标卡带扫描线高光；Topic 名用等宽字体像磁带标签。

## 2. 色彩

| Token | Hex | 用途 |
|---|---|---|
| `--bg-void` | `#070B12` | 页面底 |
| `--bg-hull` | `#0E1624` | 侧栏 / 顶栏 |
| `--bg-panel` | `#121C2C` | 卡片 |
| `--bg-inset` | `#0A1220` | 表格、代码块 |
| `--line` | `#243044` | 边框 |
| `--text` | `#E7EEF8` | 主文字 |
| `--muted` | `#8BA0B8` | 次级 |
| `--cyan` | `#3EE0C5` | 生产、健康、主操作 |
| `--amber` | `#F5B942` | Lag、警告、强调数字 |
| `--coral` | `#FF6B6B` | 错误、高 Lag |
| `--violet` | `#8B7CFF` | 消费曲线 |

禁止紫色渐变白底、禁止 Inter / Roboto 作为展示字体。

## 3. 字体

- Display / 导航：`Syne`（700）
- Body：`Figtree`
- 数字 / Offset / Topic：`IBM Plex Mono`

## 4. 布局

- 左侧固定舰桥 240px（≤768px 收起为顶栏抽屉）
- 主区 `w-full`，无 `max-w-*` 限宽
- 顶栏：集群名 MiniKafka、实时时钟（GMT+8 `yyyy-MM-dd HH:mm:ss`）、刷新倒计时
- 断点：768px 平板、480px 手机；表格横向滚动

## 5. 组件

- **MetricTile**：大号等宽数字 + 微型 sparkline + 扫描线
- **LagBar**：已提交 / LEO 双色条
- **SegmentStrip**：每个 Partition 的 segment 色块条
- **Modal**：自定义，禁止 `alert/confirm`
- **Toast**：可手动关闭 + 5s 自动消失
- **Select**：自定义箭头，禁止原生 appearance
- **EmptyState**：声呐空圈 + 引导创建 Topic

## 6. 动效

- 首屏指标卡 stagger 80ms 淡入上移
- 吞吐数字变化时轻微闪烁（cyan）
- 路由切换 180ms fade
- 刷新圆环倒计时（3s）

## 7. 页面

1. 概览 — 四块指标 + 双曲线（生产/消费）+ 磁盘
2. Topic — 列表、创建、进入分区详情
3. 消费组 — Lag 表、重置位点（确认弹窗）
4. 消息浏览 — Topic/Partition/Offset 查询
5. 实验室 — 发送测试消息（演示用，功能完整）
