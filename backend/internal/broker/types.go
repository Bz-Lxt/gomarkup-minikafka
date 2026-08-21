package broker

import (
	"sync/atomic"

	"minikafka/internal/wal"
)

type Topic struct {
	Name      string
	CreatedAt int64
	parts     []*Partition
	rr        atomic.Uint64
}

type Partition struct {
	ID  int
	log *wal.Log
}

type Message struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp string `json:"timestamp"`
}

type TopicInfo struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
	Messages   int64  `json:"messages"`
	Bytes      int64  `json:"bytes"`
	CreatedAt  string `json:"created_at"`
}

type PartitionInfo struct {
	ID       int   `json:"id"`
	Earliest int64 `json:"earliest"`
	LEO      int64 `json:"leo"`
	Bytes    int64 `json:"bytes"`
	Segments int   `json:"segments"`
}

type GroupInfo struct {
	Group    string   `json:"group"`
	Members  []string `json:"members"`
	LagTotal int64    `json:"lag_total"`
	Topics   []string `json:"topics"`
}

type GroupPartition struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Committed int64  `json:"committed"`
	LEO       int64  `json:"leo"`
	Lag       int64  `json:"lag"`
}

func (t *Topic) PartitionCount() int { return len(t.parts) }
