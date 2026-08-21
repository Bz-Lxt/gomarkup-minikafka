package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// decodeHandler echoes back the decoded "topic" field so we can exercise
// decodeJSON end-to-end through the HTTP stack.
func decodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic    string `json:"topic"`
		Messages []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"messages"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	writeData(w, 201, map[string]any{"topic": req.Topic, "count": len(req.Messages)})
}

// buildBatch builds a JSON batch payload whose serialized size is roughly
// targetBytes.
func buildBatch(topic string, targetBytes int) []byte {
	// Each entry like {"key":"k","value":"vvvv..."} adds ~25 bytes overhead
	// plus the value length. Use a generous value so we hit the target quickly.
	valueLen := 256
	perMsg := 25 + valueLen
	n := targetBytes/perMsg + 1
	type msg struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	value := strings.Repeat("v", valueLen)
	msgs := make([]msg, n)
	for i := range msgs {
		msgs[i] = msg{Key: "k", Value: value}
	}
	body := map[string]any{"topic": topic, "messages": msgs}
	raw, _ := json.Marshal(body)
	return raw
}

func TestDecodeJSON_LargeBatchUnderLimitSucceeds(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/batch", decodeHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// 16 KiB payload: previously rejected because of the 8 KiB LimitReader.
	body := buildBatch("orders", 16*1024)
	if len(body) <= 8*1024 {
		t.Fatalf("test payload too small: %d bytes", len(body))
	}

	resp, err := http.Post(srv.URL+"/batch", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(b))
	}
}

func TestDecodeJSON_OversizedBodyRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/batch", decodeHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Exceed maxBodyBytes (16 MiB) so the server should refuse, not truncate.
	// Use a single very large value field so the payload clearly exceeds the limit.
	type msg struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	over := strings.Repeat("z", int(maxBodyBytes)+256)
	body, _ := json.Marshal(map[string]any{"topic": "orders", "messages": []msg{{Key: "k", Value: over}}})
	resp, err := http.Post(srv.URL+"/batch", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 413 {
		t.Fatalf("expected 413 for oversized body, got %d", resp.StatusCode)
	}
}

func TestDecodeJSON_MalformedJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/batch", decodeHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/batch", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for malformed json, got %d", resp.StatusCode)
	}
}
