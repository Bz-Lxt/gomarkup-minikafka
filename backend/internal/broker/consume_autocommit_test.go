package broker_test

import (
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func TestConsumeAutoCommitAdvancesBeforeNextPoll(t *testing.T) {
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
	partition := 0
	for _, value := range []string{"one", "two", "three"} {
		if _, err := b.Produce("events", "", value, &partition); err != nil {
			t.Fatal(err)
		}
	}

	_, first, err := b.Consume("events", "indexer", "worker-1", 10, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 3 {
		t.Fatalf("first poll returned %d messages, want 3", len(first))
	}

	_, second, err := b.Consume("events", "indexer", "worker-1", 10, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("second poll repeated %d already consumed messages, want 0", len(second))
	}
}
