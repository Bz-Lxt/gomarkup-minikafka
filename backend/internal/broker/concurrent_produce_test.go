package broker_test

import (
	"fmt"
	"sync"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func TestConcurrentProducePreservesPartitionLog(t *testing.T) {
	b, err := broker.Open(config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })

	if _, err := b.CreateTopic("events", 1); err != nil {
		t.Fatal(err)
	}

	const producers = 64
	start := make(chan struct{})
	offsets := make(chan int64, producers)
	errs := make(chan error, producers)
	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			msg, err := b.Produce("events", fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), nil)
			if err != nil {
				errs <- err
				return
			}
			offsets <- msg.Offset
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	close(offsets)

	for err := range errs {
		t.Errorf("concurrent produce failed: %v", err)
	}
	seen := make(map[int64]struct{}, producers)
	for off := range offsets {
		if _, exists := seen[off]; exists {
			t.Errorf("duplicate offset returned: %d", off)
		}
		seen[off] = struct{}{}
	}
	if len(seen) != producers {
		t.Fatalf("got %d unique offsets, want %d", len(seen), producers)
	}

	messages, _, err := b.ReadMessages("events", 0, 0, producers)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != producers {
		t.Fatalf("stored %d messages after %d successful produces", len(messages), producers)
	}
}
