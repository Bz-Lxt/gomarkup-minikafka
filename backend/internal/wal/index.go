package wal

import "sort"

// IndexEntry maps a message offset to a byte position inside a segment file.
type IndexEntry struct {
	Offset   int64
	Position int64
}

// SparseIndex is an in-memory sparse map, kept sorted by Offset.
type SparseIndex struct {
	entries []IndexEntry
}

func (s *SparseIndex) Add(offset, pos int64) {
	s.entries = append(s.entries, IndexEntry{Offset: offset, Position: pos})
}

func (s *SparseIndex) Len() int { return len(s.entries) }

func (s *SparseIndex) Entries() []IndexEntry { return s.entries }

// Lookup returns the greatest entry with Offset <= target (binary search).
func (s *SparseIndex) Lookup(offset int64) (pos int64, startOff int64, ok bool) {
	if len(s.entries) == 0 {
		return 0, 0, false
	}
	i := sort.Search(len(s.entries), func(i int) bool {
		return s.entries[i].Offset > offset
	}) - 1
	if i < 0 {
		return 0, 0, false
	}
	e := s.entries[i]
	return e.Position, e.Offset, true
}
