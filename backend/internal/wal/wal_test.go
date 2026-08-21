package wal

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	rec := Record{Offset: 7, Timestamp: 123, Key: []byte("k"), Value: []byte("hello")}
	buf := Encode(rec)
	got, n, err := Decode(bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(buf) {
		t.Fatalf("n=%d want %d", n, len(buf))
	}
	if got.Offset != 7 || string(got.Key) != "k" || string(got.Value) != "hello" {
		t.Fatalf("roundtrip %+v", got)
	}
}

func TestCorruptCRC(t *testing.T) {
	buf := Encode(Record{Offset: 1, Value: []byte("x")})
	buf[10] ^= 0xff
	_, _, err := Decode(bytes.NewReader(buf))
	if err != ErrCorrupt {
		t.Fatalf("want corrupt, got %v", err)
	}
}

func TestIncomplete(t *testing.T) {
	buf := Encode(Record{Offset: 1, Value: []byte("xyz")})
	_, _, err := Decode(bytes.NewReader(buf[:6]))
	if err != ErrIncomplete {
		t.Fatalf("want incomplete, got %v", err)
	}
}

func TestSparseLookup(t *testing.T) {
	idx := &SparseIndex{}
	idx.Add(0, 0)
	idx.Add(10, 100)
	idx.Add(20, 200)
	pos, off, ok := idx.Lookup(15)
	if !ok || pos != 100 || off != 10 {
		t.Fatalf("lookup 15 -> pos=%d off=%d ok=%v", pos, off, ok)
	}
	pos, off, ok = idx.Lookup(0)
	if !ok || pos != 0 || off != 0 {
		t.Fatalf("lookup 0 -> %d %d %v", pos, off, ok)
	}
	if _, _, ok := idx.Lookup(-1); ok {
		t.Fatal("negative should miss")
	}
}

func TestLogRecoverTruncatesTail(t *testing.T) {
	dir := t.TempDir()
	lg, err := Open(Options{Dir: dir, SegmentMaxBytes: 1 << 20, IndexIntervalBytes: 64, SyncMode: SyncAlways})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if _, err := lg.Append(Record{Timestamp: 1, Key: []byte("k"), Value: []byte("v")}); err != nil {
			t.Fatal(err)
		}
	}
	if err := lg.Close(); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(files) != 1 {
		t.Fatalf("files %v", files)
	}
	f, err := os.OpenFile(files[0], os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte{0, 1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	lg2, err := Open(Options{Dir: dir, SyncMode: SyncAlways, IndexIntervalBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer lg2.Close()
	if lg2.NextOffset() != 20 {
		t.Fatalf("next=%d want 20", lg2.NextOffset())
	}
	recs, seeks, err := lg2.Read(7, 1)
	if err != nil || len(recs) != 1 || recs[0].Offset != 7 {
		t.Fatalf("read %+v seeks=%d err=%v", recs, seeks, err)
	}
	if seeks > 2 {
		t.Fatalf("seeks=%d want <=2", seeks)
	}
}

func TestSegmentRoll(t *testing.T) {
	dir := t.TempDir()
	lg, err := Open(Options{Dir: dir, SegmentMaxBytes: 200, IndexIntervalBytes: 32, SyncMode: SyncAlways})
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	for i := 0; i < 30; i++ {
		if _, err := lg.Append(Record{Timestamp: 1, Value: bytes.Repeat([]byte("x"), 40)}); err != nil {
			t.Fatal(err)
		}
	}
	if lg.SegmentCount() < 2 {
		t.Fatalf("expected roll, segments=%d", lg.SegmentCount())
	}
}

func TestConcurrentAppend(t *testing.T) {
	dir := t.TempDir()
	lg, err := Open(Options{Dir: dir, SyncMode: SyncBatch, SyncInterval: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	const n = 20
	const each = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < each; j++ {
				if _, err := lg.Append(Record{Timestamp: 1, Value: []byte("m")}); err != nil {
					t.Errorf("append: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if lg.NextOffset() != n*each {
		t.Fatalf("next=%d want %d", lg.NextOffset(), n*each)
	}
}
