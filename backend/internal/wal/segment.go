package wal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const defaultBufSize = 256 * 1024

// Segment is one WAL file slice. The filename is "{baseOffset:020d}.log".
type Segment struct {
	baseOffset int64
	nextOffset int64
	size       int64
	path       string
	file       *os.File
	buf        *bufio.Writer
	index      *SparseIndex
	bytesSince int
	interval   int
	readOnly   bool
}

func segmentPath(dir string, base int64) string {
	return filepath.Join(dir, fmt.Sprintf("%020d.log", base))
}

func createSegment(dir string, base int64, interval int) (*Segment, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := segmentPath(dir, base)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	return &Segment{
		baseOffset: base,
		nextOffset: base,
		path:       path,
		file:       f,
		buf:        bufio.NewWriterSize(f, defaultBufSize),
		index:      &SparseIndex{},
		interval:   interval,
	}, nil
}

func recoverSegment(path string, interval int, readOnly bool) (*Segment, error) {
	flag := os.O_RDWR
	if readOnly {
		flag = os.O_RDONLY
	}
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, err
	}
	base, err := parseBase(filepath.Base(path))
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	s := &Segment{
		baseOffset: base,
		nextOffset: base,
		path:       path,
		file:       f,
		index:      &SparseIndex{},
		interval:   interval,
		readOnly:   readOnly,
	}
	if !readOnly {
		s.buf = bufio.NewWriterSize(f, defaultBufSize)
	}
	var pos int64
	bytesSince := interval // force first record into index
	for {
		rec, n, err := Decode(io.NewSectionReader(f, pos, 1<<62))
		if err == io.EOF {
			break
		}
		if err == ErrIncomplete || err == ErrCorrupt {
			if !readOnly {
				if terr := f.Truncate(pos); terr != nil {
					_ = f.Close()
					return nil, terr
				}
				if _, serr := f.Seek(pos, io.SeekStart); serr != nil {
					_ = f.Close()
					return nil, serr
				}
			}
			break
		}
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		if bytesSince >= interval {
			s.index.Add(rec.Offset, pos)
			bytesSince = 0
		}
		bytesSince += n
		pos += int64(n)
		s.nextOffset = rec.Offset + 1
	}
	s.size = pos
	s.bytesSince = bytesSince
	if !readOnly {
		if _, err := f.Seek(pos, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	return s, nil
}

func parseBase(name string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(name, "%d.log", &n)
	return n, err
}

func (s *Segment) append(rec Record) (int64, error) {
	if s.readOnly {
		return 0, fmt.Errorf("append to read-only segment")
	}
	buf := Encode(rec)
	if s.bytesSince >= s.interval || s.size == 0 {
		s.index.Add(rec.Offset, s.size)
		s.bytesSince = 0
	}
	if _, err := s.buf.Write(buf); err != nil {
		return 0, err
	}
	n := int64(len(buf))
	s.size += n
	s.bytesSince += int(n)
	s.nextOffset = rec.Offset + 1
	return rec.Offset, nil
}

func (s *Segment) flush() error {
	if s.buf != nil {
		return s.buf.Flush()
	}
	return nil
}

func (s *Segment) sync() error {
	if err := s.flush(); err != nil {
		return err
	}
	if s.file != nil {
		return s.file.Sync()
	}
	return nil
}

func (s *Segment) close() error {
	_ = s.flush()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *Segment) freeze() error {
	defer s.file.Close()
	if err := s.sync(); err != nil {
		return err
	}
	_ = s.persistIndex()
	s.readOnly = true
	s.buf = nil
	return nil
}

func (s *Segment) contains(offset int64) bool {
	return offset >= s.baseOffset && offset < s.nextOffset
}

func (s *Segment) readFrom(offset int64, max int) ([]Record, int, error) {
	if err := s.flush(); err != nil {
		return nil, 0, err
	}
	if offset < s.baseOffset || offset >= s.nextOffset {
		return nil, 0, nil
	}
	pos := int64(0)
	if p, _, ok := s.index.Lookup(offset); ok {
		pos = p
	}
	out := make([]Record, 0, max)
	seeks := 1
	sr := io.NewSectionReader(s.file, pos, s.size-pos)
	for len(out) < max {
		rec, n, err := Decode(sr)
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, seeks, err
		}
		pos += int64(n)
		if rec.Offset < offset {
			continue
		}
		if rec.Offset > offset && len(out) == 0 {
			// sparse jump landed before a hole; keep scanning
		}
		if rec.Offset >= offset {
			out = append(out, rec)
		}
	}
	return out, seeks, nil
}
