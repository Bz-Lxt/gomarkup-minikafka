package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProduceBatchAgainstStub(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/api/v1/topics", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"name": "t"}})
	})
	mux.HandleFunc("/api/v1/produce/batch", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"count": 3}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := New(srv.URL)
	if err := c.Health(); err != nil {
		t.Fatal(err)
	}
	if err := c.CreateTopic("t", 1); err != nil {
		t.Fatal(err)
	}
	n, err := c.ProduceBatch("t", 3, "x")
	if err != nil || n != 3 {
		t.Fatalf("%d %v", n, err)
	}
}
