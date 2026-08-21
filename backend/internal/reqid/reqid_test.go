package reqid

import (
	"context"
	"testing"
)

func TestNewAndContext(t *testing.T) {
	id := New()
	if len(id) != 16 {
		t.Fatalf("id len %d", len(id))
	}
	ctx := With(context.Background(), id)
	if From(ctx) != id {
		t.Fatal(From(ctx))
	}
}
