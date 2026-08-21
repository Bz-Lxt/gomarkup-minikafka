package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func TestRecoveredTopicRecreatesMissingEmptyPartition(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:            dir,
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}

	b, err := broker.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.CreateTopic("events", 2); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "topics", "events", "p-1")); err != nil {
		t.Fatal(err)
	}

	b, err = broker.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer func() { _ = recover() }()
		_ = b.Close()
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/produce", bytes.NewBufferString(
		`{"topic":"events","partition":1,"key":"k","value":"v"}`,
	))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	New(b, "").Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("produce to recovered empty partition: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
