package broker

import (
	"sort"
	"sync"
	"time"

	"minikafka/internal/assignor"
)

type groupHub struct {
	mu     sync.Mutex
	groups map[string]*liveGroup
}

type liveGroup struct {
	members map[string]time.Time
	topics  map[string]int // topic -> partition count last seen
}

func newGroupHub() *groupHub {
	return &groupHub{groups: map[string]*liveGroup{}}
}

func (h *groupHub) assign(group, client, topic string, nParts int) []int {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, ok := h.groups[group]
	if !ok {
		g = &liveGroup{members: map[string]time.Time{}, topics: map[string]int{}}
		h.groups[group] = g
	}
	g.members[client] = time.Now()
	g.topics[topic] = nParts
	h.expire(g)
	ids := make([]string, 0, len(g.members))
	for id := range g.members {
		ids = append(ids, id)
	}
	return assignor.ForMember(assignor.StrategyRoundRobin, ids, nParts, client)
}

func (h *groupHub) expire(g *liveGroup) {
	for id, t := range g.members {
		if time.Since(t) > 30*time.Second {
			delete(g.members, id)
		}
	}
}

func (h *groupHub) names() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.groups))
	for n, g := range h.groups {
		h.expire(g)
		out = append(out, n)
	}
	return out
}

func (h *groupHub) members(group string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, ok := h.groups[group]
	if !ok {
		return nil
	}
	h.expire(g)
	out := make([]string, 0, len(g.members))
	for id := range g.members {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func (h *groupHub) topicsOf(group string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, ok := h.groups[group]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(g.topics))
	for t := range g.topics {
		out = append(out, t)
	}
	return out
}
