package retention

import "testing"

func TestSelectKeepsTail(t *testing.T) {
	p := Policy{MaxBytes: 100}
	segs := []SegmentInfo{
		{Base: 0, Bytes: 80},
		{Base: 10, Bytes: 80},
		{Base: 20, Bytes: 10},
		{Base: 30, Bytes: 10, Active: true},
	}
	got := p.Select(0, segs)
	if len(got) == 0 || got[0] != 0 {
		t.Fatalf("want drop oldest, got %v", got)
	}
}

func TestDisabled(t *testing.T) {
	var p Policy
	if p.Enabled() {
		t.Fatal("zero policy")
	}
	if p.Select(1, []SegmentInfo{{}, {}, {}}) != nil {
		t.Fatal("disabled should drop nothing")
	}
}

func TestAge(t *testing.T) {
	p := Policy{MaxAgeMS: 1000}
	segs := []SegmentInfo{
		{Base: 0, Bytes: 1, CreatedMS: 1},
		{Base: 1, Bytes: 1, CreatedMS: 2},
		{Base: 2, Bytes: 1, CreatedMS: 9000},
		{Base: 3, Bytes: 1, CreatedMS: 9001, Active: true},
	}
	got := p.Select(10000, segs)
	if len(got) == 0 {
		t.Fatal("old segments should drop")
	}
}
