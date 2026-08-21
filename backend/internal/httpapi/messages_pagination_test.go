package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func TestMessagesPaginationDoesNotRepeatBoundaryMessage(t *testing.T) {
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
	defer b.Close()

	if _, err := b.CreateTopic("events", 1); err != nil {
		t.Fatal(err)
	}
	partition := 0
	for _, value := range []string{"first", "second", "third"} {
		if _, err := b.Produce("events", "", value, &partition); err != nil {
			t.Fatal(err)
		}
	}

	handler := New(b, "").Handler()

	type message struct {
		Offset int64  `json:"offset"`
		Value  string `json:"value"`
	}
	type page struct {
		Data struct {
			Messages   []message `json:"messages"`
			NextOffset int64     `json:"next_offset"`
		} `json:"data"`
	}

	fetch := func(offset int64) page {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/topics/events/messages?partition=0&offset=%d&limit=2", offset), nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.Code)
		}
		var got page
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		return got
	}

	first := fetch(0)
	if len(first.Data.Messages) != 2 {
		t.Fatalf("first page has %d messages, want 2", len(first.Data.Messages))
	}
	second := fetch(first.Data.NextOffset)
	if len(second.Data.Messages) != 1 {
		t.Fatalf("second page has %d messages, want 1", len(second.Data.Messages))
	}
	if second.Data.Messages[0].Offset <= first.Data.Messages[len(first.Data.Messages)-1].Offset {
		t.Fatalf("page boundary repeated offset %d", second.Data.Messages[0].Offset)
	}
	if second.Data.Messages[0].Value != "third" {
		t.Fatalf("second page value = %q, want third", second.Data.Messages[0].Value)
	}
}
