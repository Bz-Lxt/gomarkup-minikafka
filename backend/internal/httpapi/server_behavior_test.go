package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
	"minikafka/internal/httpapi"
)

func TestCreateTopicMapsWrappedValidationError(t *testing.T) {
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

	body := bytes.NewBufferString(`{"name":"order events","partitions":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	httpapi.New(b, "").Handler().ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusUnprocessableEntity, res.Body.String())
	}
	var got struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Error.Code != "validation_error" {
		t.Fatalf("error code = %q, want %q", got.Error.Code, "validation_error")
	}
}
