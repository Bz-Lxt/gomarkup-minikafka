package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func TestConcurrentCreateTopicRejectsDuplicates(t *testing.T) {
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
	h := New(b, "").Handler()

	for round := 0; round < 8; round++ {
		name := fmt.Sprintf("simultaneous-%d", round)
		start := make(chan struct{})
		statuses := make(chan int, 16)
		var wg sync.WaitGroup
		for i := 0; i < cap(statuses); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				body := []byte(fmt.Sprintf(`{"name":%q,"partitions":16}`, name))
				req := httptest.NewRequest(http.MethodPost, "/api/v1/topics", bytes.NewReader(body))
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)
				statuses <- rec.Code
			}()
		}
		close(start)
		wg.Wait()
		close(statuses)

		created, conflicts := 0, 0
		for status := range statuses {
			switch status {
			case http.StatusCreated:
				created++
			case http.StatusConflict:
				conflicts++
			default:
				t.Fatalf("round %d: unexpected HTTP status %d", round, status)
			}
		}
		if created != 1 || conflicts != 15 {
			t.Fatalf("round %d: got %d created and %d conflicts; want 1 created and 15 conflicts", round, created, conflicts)
		}
	}
}
