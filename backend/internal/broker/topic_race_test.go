package broker

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"minikafka/internal/config"
)

func TestCreateTopicConcurrentSameBroker(t *testing.T) {
	cfg := config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}
	b, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	const goroutines = 32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	created := make(chan *Topic, goroutines)
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tp, err := b.CreateTopic("race-topic", 2)
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
	if createdCount != 1 {
		t.Fatalf("expected exactly 1 created, got %d", createdCount)
	}
	if errCount != goroutines-1 {
		t.Fatalf("expected %d errors, got %d", goroutines-1, errCount)
	}
}

func TestCreateTopicConcurrentDifferentNames(t *testing.T) {
	cfg := config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}
	b, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	const goroutines = 32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("topic-%c%c", rune('a'+idx%26), rune('a'+idx/26))
			_, err := b.CreateTopic(name, 1)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	topics := b.ListTopics()
	if len(topics) != goroutines {
		t.Fatalf("expected %d topics, got %d", goroutines, len(topics))
	}
}

func TestCreateTopicDuplicateReturnsConflict(t *testing.T) {
	cfg := config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}
	b, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	if _, err := b.CreateTopic("dup", 1); err != nil {
		t.Fatal(err)
	}
	_, err = b.CreateTopic("dup", 1)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestCreateTopicMultipleDifferentNames(t *testing.T) {
	cfg := config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}
	b, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	// Create two topics with different names; both should succeed.
	if _, err := b.CreateTopic("good-topic", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := b.CreateTopic("another-topic", 2); err != nil {
		t.Fatal(err)
	}

	topics := b.ListTopics()
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
}
