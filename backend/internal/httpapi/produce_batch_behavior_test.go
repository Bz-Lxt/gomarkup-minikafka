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

func TestProduceBatchPersistsEveryMessage(t *testing.T) {
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

	if _, err := b.CreateTopic("imports", 1); err != nil {
		t.Fatal(err)
	}

	handler := httpapi.New(b, "").Handler()

	body := []byte(`{"topic":"imports","messages":[{"key":"a","value":"first"},{"key":"b","value":"middle"},{"key":"c","value":"last"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/produce/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("batch status = %d, want %d", resp.Code, http.StatusCreated)
	}
	var produced struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&produced); err != nil {
		t.Fatal(err)
	}
	if produced.Data.Count != 3 {
		t.Errorf("batch count = %d, want 3", produced.Data.Count)
	}

	browseReq := httptest.NewRequest(http.MethodGet, "/api/v1/topics/imports/messages?partition=0&offset=0&limit=10", nil)
	browse := httptest.NewRecorder()
	handler.ServeHTTP(browse, browseReq)
	if browse.Code != http.StatusOK {
		t.Fatalf("browse status = %d, want %d", browse.Code, http.StatusOK)
	}
	var got struct {
		Data struct {
			Messages []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.NewDecoder(browse.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Data.Messages) != 3 {
		t.Fatalf("stored messages = %d, want 3", len(got.Data.Messages))
	}
	last := got.Data.Messages[2]
	if last.Key != "c" || last.Value != "last" {
		t.Fatalf("last stored message = %q/%q, want c/last", last.Key, last.Value)
	}
}
