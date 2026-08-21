package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
	"minikafka/internal/httpapi"
)

func TestProduceBatchAcceptsPayloadLargerThanEightKiB(t *testing.T) {
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
	if _, err := b.CreateTopic("imports", 1); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(map[string]any{
		"topic": "imports",
		"messages": []map[string]string{{
			"key": "large-record", "value": strings.Repeat("x", 16<<10),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/produce/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	httpapi.New(b, "").Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
