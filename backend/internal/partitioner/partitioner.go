package partitioner

import (
	"hash/fnv"
	"sync/atomic"

	"minikafka/internal/apperror"
	"minikafka/internal/validate"
)

// Hash maps a key onto [0, n) using FNV-1a, matching Kafka-style sticky key routing.
func Hash(key string, n int) int {
	if n <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32() % uint32(n))
}

func NextRR(counter *atomic.Uint64, n int) int {
	if n <= 0 {
		return 0
	}
	return int(counter.Add(1) % uint64(n))
}

// Choose picks a partition: forced id, else key hash, else round-robin.
func Choose(n int, key string, forced *int, rr *atomic.Uint64) (int, error) {
	if n <= 0 {
		return 0, apperror.Wrap(apperror.ErrInvalid, "no partitions")
	}
	if forced != nil {
		if err := validate.PartitionID(*forced, n); err != nil {
			return 0, err
		}
		return *forced, nil
	}
	if key != "" {
		return Hash(key, n), nil
	}
	return NextRR(rr, n), nil
}
