package broker

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	produce atomic.Uint64
	consume atomic.Uint64
	mu      sync.Mutex
	lastP   uint64
	lastC   uint64
	lastT   time.Time
	pRate   float64
	cRate   float64
}

type MetricsSnapshot struct {
	Topics             int     `json:"topics"`
	Partitions         int     `json:"partitions"`
	MessagesTotal      int64   `json:"messages_total"`
	BytesOnDisk        int64   `json:"bytes_on_disk"`
	ProduceRate        float64 `json:"produce_rate"`
	ConsumeRate        float64 `json:"consume_rate"`
	ActiveConnections  int64   `json:"active_connections"`
	Groups             int     `json:"groups"`
	UptimeSec          int64   `json:"uptime_sec"`
}

func newMetrics() *Metrics {
	return &Metrics{lastT: time.Now()}
}

func (m *Metrics) AddProduce(n uint64) { m.produce.Add(n) }
func (m *Metrics) AddConsume(n uint64) { m.consume.Add(n) }

func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	dt := now.Sub(m.lastT).Seconds()
	p := m.produce.Load()
	c := m.consume.Load()
	if dt >= 0.5 {
		m.pRate = float64(p-m.lastP) / dt
		m.cRate = float64(c-m.lastC) / dt
		m.lastP, m.lastC, m.lastT = p, c, now
	}
	return MetricsSnapshot{ProduceRate: m.pRate, ConsumeRate: m.cRate}
}
