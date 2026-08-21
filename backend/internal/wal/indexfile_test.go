package wal

import (
	"path/filepath"
	"testing"
)

func TestIndexFileRoundtrip(t *testing.T) {
	idx := &SparseIndex{}
	idx.Add(100, 0)
	idx.Add(110, 4096)
	p := filepath.Join(t.TempDir(), "00000000000000000100.index")
	if err := idx.Save(p, 100); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSparseIndex(p, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got.Len() != 2 {
		t.Fatalf("len %d", got.Len())
	}
	pos, off, ok := got.Lookup(105)
	if !ok || pos != 0 || off != 100 {
		t.Fatalf("lookup %d %d %v", pos, off, ok)
	}
}
