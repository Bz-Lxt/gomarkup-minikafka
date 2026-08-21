package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

const indexMagic = "MKIX"

func indexPath(logPath string) string {
	return strings.TrimSuffix(logPath, ".log") + ".index"
}

// Save writes the sparse index next to a segment log.
// Layout: magic(4) version(1) count(4) {relOff u32, pos u32}*
func (s *SparseIndex) Save(path string, baseOffset int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(indexMagic)); err != nil {
		return err
	}
	hdr := make([]byte, 5)
	hdr[0] = 1
	binary.BigEndian.PutUint32(hdr[1:5], uint32(len(s.entries)))
	if _, err := f.Write(hdr); err != nil {
		return err
	}
	buf := make([]byte, 8)
	for _, e := range s.entries {
		rel := uint32(e.Offset - baseOffset)
		binary.BigEndian.PutUint32(buf[0:4], rel)
		binary.BigEndian.PutUint32(buf[4:8], uint32(e.Position))
		if _, err := f.Write(buf); err != nil {
			return err
		}
	}
	return f.Sync()
}

func LoadSparseIndex(path string, baseOffset int64) (*SparseIndex, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < 9 || string(raw[:4]) != indexMagic || raw[4] != 1 {
		return nil, fmt.Errorf("bad index header")
	}
	count := binary.BigEndian.Uint32(raw[5:9])
	need := 9 + int(count)*8
	if len(raw) < need {
		return nil, io.ErrUnexpectedEOF
	}
	idx := &SparseIndex{entries: make([]IndexEntry, 0, count)}
	off := 9
	for i := 0; i < int(count); i++ {
		rel := binary.BigEndian.Uint32(raw[off : off+4])
		pos := binary.BigEndian.Uint32(raw[off+4 : off+8])
		idx.Add(baseOffset+int64(rel), int64(pos))
		off += 8
	}
	return idx, nil
}

func (s *Segment) persistIndex() error {
	if s.index == nil {
		return nil
	}
	return s.index.Save(indexPath(s.path), s.baseOffset)
}
