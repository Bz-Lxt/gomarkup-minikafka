package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"minikafka/internal/apperror"
	"minikafka/internal/broker"
	"minikafka/internal/config"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
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
	t.Cleanup(func() { b.Close() })
	srv := httptest.NewServer(New(b, "").Handler())
	t.Cleanup(srv.Close)
	return srv
}

func doCreateTopic(t *testing.T, srv *httptest.Server, body string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/topics", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

// Regression test for create-topic requests returning 500 internal_error
// instead of a validation error when the name is invalid (e.g. contains spaces).
func TestCreateTopicInvalidNameReturnsValidationError(t *testing.T) {
	srv := newTestServer(t)

	cases := []string{
		`{"name":"bad name","partitions":1}`,   // space
		`{"name":"bad/slash","partitions":1}`,  // disallowed char
		`{"name":"bad@at","partitions":1}`,     // disallowed char
		`{"name":"","partitions":1}`,          // empty
	}
	for _, body := range cases {
		status, out := doCreateTopic(t, srv, body)
		if status == http.StatusInternalServerError {
			t.Fatalf("create topic body %q returned 500 internal_error, want 4xx validation_error; resp=%v", body, out)
		}
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("create topic body %q returned status %d, want 422; resp=%v", body, status, out)
		}
		errObj, _ := out["error"].(map[string]any)
		if errObj == nil {
			t.Fatalf("create topic body %q: missing error object; resp=%v", body, out)
		}
		if code, _ := errObj["code"].(string); code != "validation_error" {
			t.Fatalf("create topic body %q: error code=%q, want validation_error; resp=%v", body, code, out)
		}
	}
}

// Sanity check: a valid topic name still succeeds.
func TestCreateTopicValidNameSucceeds(t *testing.T) {
	srv := newTestServer(t)

	for _, body := range []string{
		`{"name":"orders","partitions":1}`,
		`{"name":"my-topic-1","partitions":1}`,
		`{"name":"user_2.v3","partitions":1}`,
	} {
		status, out := doCreateTopic(t, srv, body)
		if status != http.StatusCreated {
			t.Fatalf("create topic body %q returned status %d, want 201; resp=%v", body, status, out)
		}
	}
}

// Direct unit test for mapErr mapping a wrapped ErrInvalid to 422.
func TestMapErrWrappedInvalid(t *testing.T) {
	rec := httptest.NewRecorder()
	// simulate the kind of error validate.ResourceName returns: a wrapped ErrInvalid.
	wrapped := fmt.Errorf("%w: topic 仅允许字母数字 . _ - ，最长 64", apperror.ErrInvalid)
	mapErr(rec, wrapped)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want 422", rec.Code)
	}
	var out map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&out)
	errObj, _ := out["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "validation_error" {
		t.Fatalf("code=%q, want validation_error", code)
	}
}
