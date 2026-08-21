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

func TestMessagesDefaultPartition(t *testing.T) {
	b, err := broker.Open(config.Config{
		DataDir:            t.TempDir(),
		SegmentMaxBytes:    1 << 20,
		IndexIntervalBytes: 128,
		SyncMode:           "always",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	if _, err := b.CreateTopic("events", 1); err != nil {
		t.Fatal(err)
	}
	partition := 0
	if _, err := b.Produce("events", "key", "payload", &partition); err != nil {
		t.Fatal(err)
	}

	handler := httpapi.New(b, "").Handler()
	explicitReq := httptest.NewRequest(http.MethodGet, "/api/v1/topics/events/messages?partition=0&offset=0&limit=20", nil)
	explicitRes := httptest.NewRecorder()
	handler.ServeHTTP(explicitRes, explicitReq)
	if explicitRes.Code != http.StatusOK {
		t.Fatalf("explicit partition status=%d body=%s", explicitRes.Code, explicitRes.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/topics/events/messages?offset=0&limit=20", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Data struct {
			Messages []struct {
				Value string `json:"value"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.Messages) != 1 || body.Data.Messages[0].Value != "payload" {
		t.Fatalf("messages=%+v", body.Data.Messages)
	}
}
