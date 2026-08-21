package client_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"minikafka/pkg/client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type delayedBody struct {
	ctx     context.Context
	payload *bytes.Reader
	delay   time.Duration
	waited  bool
}

func (b *delayedBody) Read(p []byte) (int, error) {
	if !b.waited {
		b.waited = true
		timer := time.NewTimer(b.delay)
		defer timer.Stop()
		select {
		case <-b.ctx.Done():
			return 0, b.ctx.Err()
		case <-timer.C:
		}
	}
	return b.payload.Read(p)
}

func (b *delayedBody) Close() error { return nil }

func TestProduceBatchReadsDelayedResponseBody(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	payload := []byte(`{"data":{"count":3}}`)
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     make(http.Header),
			Body: &delayedBody{
				ctx:     r.Context(),
				payload: bytes.NewReader(payload),
				delay:   50 * time.Millisecond,
			},
			ContentLength: int64(len(payload)),
			Request:       r,
		}, nil
	})

	got, err := client.New("http://broker.test").ProduceBatch("events", 3, "payload")
	if err != nil {
		t.Fatalf("ProduceBatch returned an error after a successful response: %v", err)
	}
	if got != 3 {
		t.Fatalf("ProduceBatch count = %d, want 3", got)
	}
}
