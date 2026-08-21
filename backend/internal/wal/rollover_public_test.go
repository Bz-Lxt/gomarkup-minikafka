package wal_test

import (
	"bytes"
	"testing"

	"minikafka/internal/wal"
)

func TestRolledLogKeepsEarlierRecordsReadable(t *testing.T) {
	log, err := wal.Open(wal.Options{
		Dir:                t.TempDir(),
		SegmentMaxBytes:    100,
		IndexIntervalBytes: 32,
		SyncMode:           wal.SyncAlways,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	values := [][]byte{
		bytes.Repeat([]byte("a"), 64),
		bytes.Repeat([]byte("b"), 64),
		bytes.Repeat([]byte("c"), 64),
	}
	for _, value := range values {
		if _, err := log.Append(wal.Record{Timestamp: 1, Value: value}); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	if log.SegmentCount() < 2 {
		t.Fatalf("expected log rollover, got %d segment", log.SegmentCount())
	}

	records, _, err := log.Read(0, len(values))
	if err != nil {
		t.Fatalf("read from first offset after rollover: %v", err)
	}
	if len(records) != len(values) {
		t.Fatalf("read %d records, want %d", len(records), len(values))
	}
	for i, record := range records {
		if record.Offset != int64(i) || !bytes.Equal(record.Value, values[i]) {
			t.Fatalf("record %d = offset %d value %q", i, record.Offset, record.Value)
		}
	}
}
