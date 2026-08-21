package broker

import (
	"testing"

	"minikafka/internal/config"
)

func TestProduceConsumeOffset(t *testing.T) {
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
	if _, err := b.CreateTopic("orders", 2); err != nil {
		t.Fatal(err)
	}
	p0 := 0
	for i := 0; i < 10; i++ {
		if _, err := b.Produce("orders", "", "m", &p0); err != nil {
			t.Fatal(err)
		}
	}
	_, msgs, err := b.Consume("orders", "billing", "c1", 5, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected messages")
	}
	off, ok := b.offsets.Get("billing", "orders", msgs[0].Partition)
	if !ok || off <= msgs[0].Offset {
		t.Fatalf("committed=%d first=%d ok=%v", off, msgs[0].Offset, ok)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	b2, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b2.Close()
	off2, ok := b2.offsets.Get("billing", "orders", msgs[0].Partition)
	if !ok || off2 != off {
		t.Fatalf("persist committed %d vs %d", off2, off)
	}
}
