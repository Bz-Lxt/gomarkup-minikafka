package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SyncMode string

const (
	SyncAlways SyncMode = "always"
	SyncBatch  SyncMode = "batch"
)

type Options struct {
	Dir                string
	SegmentMaxBytes    int64
	IndexIntervalBytes int
	SyncMode           SyncMode
	SyncInterval       time.Duration
}

// Log is a partitioned WAL: ordered segments, sparse index, crash recovery.
type Log struct {
	mu       sync.RWMutex
	opts     Options
	segments []*Segment
	stop     chan struct{}
	wg       sync.WaitGroup
	closed   sync.Once
}

func Open(opts Options) (*Log, error) {
	if opts.SegmentMaxBytes <= 0 {
		opts.SegmentMaxBytes = 64 * 1024 * 1024
	}
	if opts.IndexIntervalBytes <= 0 {
		opts.IndexIntervalBytes = 4096
	}
	if opts.SyncMode == "" {
		opts.SyncMode = SyncBatch
	}
	if opts.SyncInterval <= 0 {
		opts.SyncInterval = 100 * time.Millisecond
	}
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return nil, err
	}
	l := &Log{opts: opts, stop: make(chan struct{})}
	if err := l.recover(); err != nil {
		return nil, err
	}
	if opts.SyncMode == SyncBatch {
		l.wg.Add(1)
		go l.flushLoop()
	}
	return l, nil
}

func (l *Log) recover() error {
	ents, err := os.ReadDir(l.opts.Dir)
	if err != nil {
		return err
	}
	var bases []int64
	for _, e := range ents {
		name := e.Name()
		if !strings.HasSuffix(name, ".log") {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSuffix(name, ".log"), 10, 64)
		if err != nil {
			continue
		}
		bases = append(bases, n)
	}
	sort.Slice(bases, func(i, j int) bool { return bases[i] < bases[j] })
	for i, b := range bases {
		readOnly := i < len(bases)-1
		seg, err := recoverSegment(segmentPath(l.opts.Dir, b), l.opts.IndexIntervalBytes, readOnly)
		if err != nil {
			return fmt.Errorf("recover %020d: %w", b, err)
		}
		l.segments = append(l.segments, seg)
	}
	if len(l.segments) == 0 {
		seg, err := createSegment(l.opts.Dir, 0, l.opts.IndexIntervalBytes)
		if err != nil {
			return err
		}
		l.segments = []*Segment{seg}
	}
	return nil
}

func (l *Log) flushLoop() {
	defer l.wg.Done()
	t := time.NewTicker(l.opts.SyncInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			l.mu.Lock()
			_ = l.active().sync()
			l.mu.Unlock()
		case <-l.stop:
			return
		}
	}
}

func (l *Log) active() *Segment {
	return l.segments[len(l.segments)-1]
}

func (l *Log) Append(rec Record) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	act := l.active()
	if act.size >= l.opts.SegmentMaxBytes && act.size > 0 {
		if err := l.roll(); err != nil {
			return 0, err
		}
		act = l.active()
	}
	rec.Offset = act.nextOffset
	off, err := act.append(rec)
	if err != nil {
		return 0, err
	}
	if l.opts.SyncMode == SyncAlways {
		if err := act.sync(); err != nil {
			return 0, err
		}
	}
	return off, nil
}

func (l *Log) AppendBatch(recs []Record) ([]int64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]int64, 0, len(recs))
	for i := range recs {
		act := l.active()
		if act.size >= l.opts.SegmentMaxBytes && act.size > 0 {
			if err := l.roll(); err != nil {
				return out, err
			}
			act = l.active()
		}
		recs[i].Offset = act.nextOffset
		off, err := act.append(recs[i])
		if err != nil {
			return out, err
		}
		out = append(out, off)
	}
	if l.opts.SyncMode == SyncAlways {
		if err := l.active().sync(); err != nil {
			return out, err
		}
	} else {
		if err := l.active().flush(); err != nil {
			return out, err
		}
	}
	return out, nil
}

func (l *Log) roll() error {
	act := l.active()
	if err := act.freeze(); err != nil {
		return err
	}
	seg, err := createSegment(l.opts.Dir, act.nextOffset, l.opts.IndexIntervalBytes)
	if err != nil {
		return err
	}
	l.segments = append(l.segments, seg)
	return nil
}

func (l *Log) Read(from int64, max int) ([]Record, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if max <= 0 {
		max = 1
	}
	var out []Record
	seeks := 0
	for _, seg := range l.segments {
		if from >= seg.nextOffset {
			continue
		}
		start := from
		if start < seg.baseOffset {
			start = seg.baseOffset
		}
		recs, n, err := seg.readFrom(start, max-len(out))
		seeks += n
		if err != nil && err != ErrIncomplete {
			return out, seeks, err
		}
		out = append(out, recs...)
		if len(out) >= max {
			break
		}
	}
	return out, seeks, nil
}

func (l *Log) Earliest() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.segments) == 0 {
		return 0
	}
	return l.segments[0].baseOffset
}

func (l *Log) NextOffset() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active().nextOffset
}

func (l *Log) Bytes() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	var n int64
	for _, s := range l.segments {
		n += s.size
	}
	return n
}

func (l *Log) SegmentCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.segments)
}

func (l *Log) IndexSize() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for _, s := range l.segments {
		n += s.index.Len()
	}
	return n
}

func (l *Log) Close() error {
	var first error
	l.closed.Do(func() {
		close(l.stop)
		l.wg.Wait()
		l.mu.Lock()
		defer l.mu.Unlock()
		for _, s := range l.segments {
			if err := s.sync(); err != nil && first == nil {
				first = err
			}
			if err := s.close(); err != nil && first == nil {
				first = err
			}
		}
	})
	return first
}

func ListSegmentFiles(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".log") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}
