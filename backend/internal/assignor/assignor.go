package assignor

import "sort"

// Strategy is a consumer group partition assignor.
type Strategy string

const (
	StrategyRoundRobin Strategy = "roundrobin"
	StrategyRange      Strategy = "range"
)

func Normalize(members []string) []string {
	cp := append([]string(nil), members...)
	sort.Strings(cp)
	return cp
}

// RoundRobin spreads partitions across members: p % n == memberIndex.
func RoundRobin(members []string, nParts int) map[string][]int {
	ids := Normalize(members)
	out := make(map[string][]int, len(ids))
	if len(ids) == 0 || nParts <= 0 {
		return out
	}
	for i, id := range ids {
		var parts []int
		for p := 0; p < nParts; p++ {
			if p%len(ids) == i {
				parts = append(parts, p)
			}
		}
		out[id] = parts
	}
	return out
}

// Range gives each member a consecutive slice of partitions.
func Range(members []string, nParts int) map[string][]int {
	ids := Normalize(members)
	out := make(map[string][]int, len(ids))
	if len(ids) == 0 || nParts <= 0 {
		return out
	}
	base := nParts / len(ids)
	extra := nParts % len(ids)
	start := 0
	for i, id := range ids {
		n := base
		if i < extra {
			n++
		}
		parts := make([]int, 0, n)
		for p := start; p < start+n; p++ {
			parts = append(parts, p)
		}
		out[id] = parts
		start += n
	}
	return out
}

func Assign(strategy Strategy, members []string, nParts int) map[string][]int {
	if strategy == StrategyRange {
		return Range(members, nParts)
	}
	return RoundRobin(members, nParts)
}

func ForMember(strategy Strategy, members []string, nParts int, client string) []int {
	m := Assign(strategy, members, nParts)
	return m[client]
}
