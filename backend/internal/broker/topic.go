package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
			// meta.json missing — either a leftover directory or a topic
			// being created concurrently by another instance. Skip it; the
			// creator will register it or clean it up.
			continue
		}
		var meta topicMeta
		if err := json.Unmarshal(raw, &meta); err != nil {
			// Corrupt or partially-written meta.json (e.g. another instance
			// hasn't finished writing it yet). Skip rather than abort startup.
			logger.L().Warn("skipping topic with unreadable meta", "name", e.Name(), "err", err)
			continue
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

	// Re-check under the write lock so concurrent in-process callers can't
	// both pass the check.
	if _, ok := b.topics[name]; ok {
		return nil, ErrAlreadyExists
	}

	// os.Mkdir (not MkdirAll): atomically claims the topic directory. If
	// another process or goroutine got there first, this fails with
	// fs.ErrExist — the single source of truth for cross-instance uniqueness.
	dir := filepath.Join(b.cfg.DataDir, "topics", name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			// Either another creator won the race (registered or in-flight),
			// or a previous attempt left an orphaned directory. Either way we
			// must not claim the topic a second time.
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	meta := topicMeta{Partitions: n, CreatedAt: clock.NowMilli()}
	raw, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), raw, 0o644); err != nil {
		// Roll back the directory we just created so a retry isn't blocked.
		_ = os.RemoveAll(dir)
		return nil, err
	}
	t := &Topic{Name: name, CreatedAt: meta.CreatedAt, parts: make([]*Partition, n)}
	for i := 0; i < n; i++ {
		lg, err := b.openPartitionLog(name, i)
		if err != nil {
			// Roll back partial state: close logs opened so far and remove
			// the topic directory so a later retry can start clean.
			for j := 0; j < i; j++ {
				_ = t.parts[j].log.Close()
			}
			_ = os.RemoveAll(dir)
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
