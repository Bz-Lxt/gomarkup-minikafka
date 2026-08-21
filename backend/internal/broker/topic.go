package broker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"minikafka/internal/clock"
	"minikafka/internal/logger"
	"minikafka/internal/validate"
	"minikafka/internal/wal"
)

type topicMeta struct {
	Partitions int   `json:"partitions"`
	CreatedAt  int64 `json:"created_at"`
}

func (b *Broker) openPartitionLog(topic string, id int) (*wal.Log, error) {
	return wal.Open(wal.Options{
		Dir:                filepath.Join(b.cfg.DataDir, "topics", topic, fmt.Sprintf("p-%d", id)),
		SegmentMaxBytes:    b.cfg.SegmentMaxBytes,
		IndexIntervalBytes: b.cfg.IndexIntervalBytes,
		SyncMode:           wal.SyncMode(b.cfg.SyncMode),
		SyncInterval:       time.Duration(b.cfg.SyncIntervalMS) * time.Millisecond,
	})
}

func (b *Broker) loadTopics() error {
	root := filepath.Join(b.cfg.DataDir, "topics")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, e.Name(), "meta.json"))
		if err != nil {
			continue
		}
		var meta topicMeta
		if err := json.Unmarshal(raw, &meta); err != nil {
			return fmt.Errorf("topic %s meta: %w", e.Name(), err)
		}
		n, err := validate.PartitionCount(meta.Partitions)
		if err != nil {
			return fmt.Errorf("topic %s invalid partition count", e.Name())
		}
		t := &Topic{Name: e.Name(), CreatedAt: meta.CreatedAt, parts: make([]*Partition, n)}
		for i := 0; i < n; i++ {
			lg, err := b.openPartitionLog(e.Name(), i)
			if err != nil {
				return err
			}
			t.parts[i] = &Partition{ID: i, log: lg}
		}
		b.topics[e.Name()] = t
		logger.L().Info("recovered topic", "name", e.Name(), "partitions", n)
	}
	return nil
}

func (b *Broker) CreateTopic(name string, partitions int) (*Topic, error) {
	if err := validate.ResourceName("topic", name); err != nil {
		return nil, err
	}
	n, err := validate.PartitionCount(partitions)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.topics[name]; ok {
		return nil, ErrAlreadyExists
	}
	dir := filepath.Join(b.cfg.DataDir, "topics", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	meta := topicMeta{Partitions: n, CreatedAt: clock.NowMilli()}
	raw, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), raw, 0o644); err != nil {
		return nil, err
	}
	t := &Topic{Name: name, CreatedAt: meta.CreatedAt, parts: make([]*Partition, n)}
	for i := 0; i < n; i++ {
		lg, err := b.openPartitionLog(name, i)
		if err != nil {
			return nil, err
		}
		t.parts[i] = &Partition{ID: i, log: lg}
	}
	b.topics[name] = t
	logger.L().Info("topic created", "name", name, "partitions", n)
	return t, nil
}

func (t *Topic) CreatedAtText() string { return clock.Format(t.CreatedAt) }

func (b *Broker) ListTopics() []TopicInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]TopicInfo, 0, len(b.topics))
	for _, t := range b.topics {
		info := TopicInfo{
			Name:       t.Name,
			Partitions: len(t.parts),
			CreatedAt:  clock.Format(t.CreatedAt),
		}
		for _, p := range t.parts {
			info.Messages += p.log.NextOffset() - p.log.Earliest()
			info.Bytes += p.log.Bytes()
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (b *Broker) TopicDetail(name string) (string, []PartitionInfo, error) {
	t, err := b.topic(name)
	if err != nil {
		return "", nil, err
	}
	parts := make([]PartitionInfo, 0, len(t.parts))
	for _, p := range t.parts {
		parts = append(parts, PartitionInfo{
			ID:       p.ID,
			Earliest: p.log.Earliest(),
			LEO:      p.log.NextOffset(),
			Bytes:    p.log.Bytes(),
			Segments: p.log.SegmentCount(),
		})
	}
	return t.Name, parts, nil
}

func (b *Broker) topic(name string) (*Topic, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.topics[name]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}
