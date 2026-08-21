package partitioner

import (
	"sync/atomic"
	"testing"
)

func TestHashStable(t *testing.T) {
	a := Hash("user-1", 8)
	b := Hash("user-1", 8)
	if a != b || a < 0 || a >= 8 {
		t.Fatalf("hash %d %d", a, b)
	}
}

func TestChooseForcedAndRR(t *testing.T) {
	p := 2
	got, err := Choose(4, "k", &p, nil)
	if err != nil || got != 2 {
		t.Fatalf("forced %d %v", got, err)
	}
	bad := 9
	if _, err := Choose(4, "", &bad, nil); err == nil {
		t.Fatal("out of range")
	}
	var rr atomic.Uint64
	x := NextRR(&rr, 3)
	y := NextRR(&rr, 3)
	if x == y && x != 0 {
		// first two increments are 1 and 2, must differ
		t.Fatalf("rr %d %d", x, y)
	}
}
