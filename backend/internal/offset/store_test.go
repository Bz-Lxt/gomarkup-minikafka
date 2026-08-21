package offset

import "testing"

func TestCommitPersist(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Commit("g1", "orders", 0, 42); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := s2.Get("g1", "orders", 0)
	if !ok || v != 42 {
		t.Fatalf("got %d %v", v, ok)
	}
}
