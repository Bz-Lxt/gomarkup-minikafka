package broker

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"minikafka/internal/clock"
	"minikafka/internal/config"
	"minikafka/internal/offset"
)

type Broker struct {
	cfg     config.Config
	mu      sync.RWMutex
	topics  map[string]*Topic
	groups  *groupHub
	offsets *offset.Store
	metrics *Metrics
	started time.Time
	conns   atomic.Int64
}

func Open(cfg config.Config) (*Broker, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}
	off, err := offset.Open(filepath.Join(cfg.DataDir, "offsets"))
	if err != nil {
		return nil, err
	}
	b := &Broker{
		cfg:     cfg,
		topics:  map[string]*Topic{},
		groups:  newGroupHub(),
		offsets: off,
		metrics: newMetrics(),
		started: clock.Now(),
	}
	if err := b.loadTopics(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range b.topics {
		for _, p := range t.parts {
			_ = p.log.Close()
		}
	}
	return nil
}

func (b *Broker) IncConn() { b.conns.Add(1) }
func (b *Broker) DecConn() { b.conns.Add(-1) }
