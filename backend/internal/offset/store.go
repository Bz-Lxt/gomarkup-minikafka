package offset

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Key struct {
	Group     string
	Topic     string
	Partition int
}

type Store struct {
	mu   sync.Mutex
	dir  string
	mem  map[Key]int64
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, mem: map[Key]int64{}}
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) != 3 {
			return nil
		}
		p, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		off, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
		if err != nil {
			return nil
		}
		s.mem[Key{Group: parts[0], Topic: parts[1], Partition: p}] = off
		return nil
	})
	return s, nil
}

func (s *Store) Get(group, topic string, partition int) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.mem[Key{group, topic, partition}]
	return v, ok
}

func (s *Store) Commit(group, topic string, partition int, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := Key{group, topic, partition}
	s.mem[key] = offset
	dir := filepath.Join(s.dir, group, topic)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, strconv.Itoa(partition))
	tmp := path + ".tmp"
	body := []byte(fmt.Sprintf("%d\n", offset))
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return err
	}
	f, err := os.OpenFile(tmp, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()
	return os.Rename(tmp, path)
}

func (s *Store) Groups() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen := map[string]struct{}{}
	var out []string
	for k := range s.mem {
		if _, ok := seen[k.Group]; ok {
			continue
		}
		seen[k.Group] = struct{}{}
		out = append(out, k.Group)
	}
	return out
}

func (s *Store) ByGroup(group string) map[Key]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[Key]int64{}
	for k, v := range s.mem {
		if k.Group == group {
			out[k] = v
		}
	}
	return out
}

func (s *Store) TopicsOf(group string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen := map[string]struct{}{}
	var out []string
	for k := range s.mem {
		if k.Group != group {
			continue
		}
		if _, ok := seen[k.Topic]; ok {
			continue
		}
		seen[k.Topic] = struct{}{}
		out = append(out, k.Topic)
	}
	return out
}
