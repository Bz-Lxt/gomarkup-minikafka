package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
	"minikafka/internal/httpapi"
)

func TestTopicDetailMissingTopicReturnsNotFound(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/topics/missing", nil)
	res := httptest.NewRecorder()
	httpapi.New(b, "").Handler().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusNotFound, res.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want %q", body.Error.Code, "not_found")
	}
}
