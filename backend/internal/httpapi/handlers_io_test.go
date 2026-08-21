package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func newTestServer(t *testing.T) (*Server, *broker.Broker) {
	t.Helper()
	cfg := config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
		SyncIntervalMS:     50,
	}
	b, err := broker.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })
	if _, err := b.CreateTopic("orders", 3); err != nil {
		t.Fatal(err)
	}
	return New(b, ""), b
}

// TestProduceAutoPartition reproduces the regression where posting a message
// without an explicit partition caused a nil-pointer dereference and was
// converted into a 500 by the recovery middleware. Auto-partition selection
// (key hash / round-robin) must keep working when partition is omitted.
func TestProduceAutoPartition(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Case 1: no key, no partition -> round-robin should pick a valid partition.
	body, _ := json.Marshal(map[string]any{"topic": "orders", "value": "m1"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/produce", bytes.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("no-partition produce: got %d body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Partition int `json:"partition"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Partition < 0 || resp.Data.Partition > 2 {
		t.Fatalf("auto partition out of range: %d", resp.Data.Partition)
	}

	// Case 2: with key but no partition -> key hash routing should pick a valid partition.
	body, _ = json.Marshal(map[string]any{"topic": "orders", "key": "user-1", "value": "m2"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/produce", bytes.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("key-only produce: got %d body %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Partition < 0 || resp.Data.Partition > 2 {
		t.Fatalf("key-hash partition out of range: %d", resp.Data.Partition)
	}
}

// TestProduceExplicitPartition ensures manual partition selection still works.
func TestProduceExplicitPartition(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	body, _ := json.Marshal(map[string]any{"topic": "orders", "value": "m", "partition": 1})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/produce", bytes.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("explicit partition produce: got %d body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Partition int `json:"partition"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Partition != 1 {
		t.Fatalf("expected partition 1, got %d", resp.Data.Partition)
	}
}