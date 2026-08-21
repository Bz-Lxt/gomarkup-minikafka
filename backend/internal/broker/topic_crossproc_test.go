package broker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"minikafka/internal/config"
)

// TestCreateTopicConcurrentTwoBrokers simulates the exact scenario from the
// bug report: two broker instances sharing the same data directory both try
// to create the same topic at startup. Exactly one should succeed; the other
// must receive ErrAlreadyExists.
func TestCreateTopicConcurrentTwoBrokers(t *testing.T) {
	dataDir := t.TempDir()

	cfg := config.Config{
		DataDir:            dataDir,
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}

	// Pre-create the topics root so both brokers start cleanly.
	if err := os.MkdirAll(filepath.Join(dataDir, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}

	const pairs = 16
	var wg sync.WaitGroup
	wg.Add(pairs * 2)
	created := make(chan *Topic, pairs*2)
	errs := make(chan error, pairs*2)

	for i := 0; i < pairs; i++ {
		// Two independent broker instances sharing the same DataDir, both
		// trying to create the same topic name.
		b1, err := Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		b2, err := Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		topicName := fmt.Sprintf("cross-%d", i)
		go func() {
			defer wg.Done()
			tp, err := b1.CreateTopic(topicName, 1)
			if err != nil {
				errs <- err
				return
			}
			created <- tp
		}()
		go func() {
			defer wg.Done()
			tp, err := b2.CreateTopic(topicName, 1)
			if err != nil {
				errs <- err
				return
			}
			created <- tp
		}()
	}
	wg.Wait()
	close(created)
	close(errs)

	createdCount := 0
	for range created {
		createdCount++
	}
	errCount := 0
	for err := range errs {
		if !errors.Is(err, ErrAlreadyExists) {
			t.Fatalf("unexpected error type: %v", err)
		}
		errCount++
	}
	if createdCount != pairs {
		t.Fatalf("expected exactly %d created, got %d", pairs, createdCount)
	}
	if errCount != pairs {
		t.Fatalf("expected %d errors, got %d", pairs, errCount)
	}
}
