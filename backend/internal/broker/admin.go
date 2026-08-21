package broker

import (
	"sort"
	"time"

	"minikafka/internal/offset"
)

func (b *Broker) ListGroups() []GroupInfo {
	names := map[string]struct{}{}
	for _, g := range b.offsets.Groups() {
		names[g] = struct{}{}
	}
	for _, g := range b.groups.names() {
		names[g] = struct{}{}
	}
	var out []GroupInfo
	for name := range names {
		info, _ := b.GroupDetail(name)
		out = append(out, GroupInfo{
			Group:    name,
			Members:  b.groups.members(name),
			LagTotal: lagSum(info),
			Topics:   uniqueTopics(info),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Group < out[j].Group })
	return out
}

func lagSum(ps []GroupPartition) int64 {
	var n int64
	for _, p := range ps {
		n += p.Lag
	}
	return n
}

func uniqueTopics(ps []GroupPartition) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range ps {
		if _, ok := seen[p.Topic]; ok {
			continue
		}
		seen[p.Topic] = struct{}{}
		out = append(out, p.Topic)
	}
	return out
}

func (b *Broker) GroupDetail(group string) ([]GroupPartition, error) {
	committed := b.offsets.ByGroup(group)
	b.mu.RLock()
	defer b.mu.RUnlock()
	topics := map[string]struct{}{}
	for k := range committed {
		topics[k.Topic] = struct{}{}
	}
	for _, t := range b.groups.topicsOf(group) {
		topics[t] = struct{}{}
	}
	var out []GroupPartition
	for name := range topics {
		t, ok := b.topics[name]
		if !ok {
			continue
		}
		for _, p := range t.parts {
			leo := p.log.NextOffset()
			c, ok := committed[offset.Key{Group: group, Topic: name, Partition: p.ID}]
			if !ok {
				c = p.log.Earliest()
			}
			lag := leo - c
			if lag < 0 {
				lag = 0
			}
			out = append(out, GroupPartition{
				Topic: name, Partition: p.ID, Committed: c, LEO: leo, Lag: lag,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Topic == out[j].Topic {
			return out[i].Partition < out[j].Partition
		}
		return out[i].Topic < out[j].Topic
	})
	return out, nil
}

func (b *Broker) GroupMembers(group string) []string {
	return b.groups.members(group)
}

func (b *Broker) ListGroupsUnlocked() []string {
	names := map[string]struct{}{}
	for _, g := range b.offsets.Groups() {
		names[g] = struct{}{}
	}
	for _, g := range b.groups.names() {
		names[g] = struct{}{}
	}
	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}
	return out
}

func (b *Broker) Snapshot() MetricsSnapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var msgs, bytes int64
	parts := 0
	for _, t := range b.topics {
		parts += len(t.parts)
		for _, p := range t.parts {
			msgs += p.log.NextOffset() - p.log.Earliest()
			bytes += p.log.Bytes()
		}
	}
	s := b.metrics.Snapshot()
	s.Topics = len(b.topics)
	s.Partitions = parts
	s.MessagesTotal = msgs
	s.BytesOnDisk = bytes
	s.ActiveConnections = b.conns.Load()
	s.Groups = len(b.ListGroupsUnlocked())
	s.UptimeSec = int64(time.Since(b.started).Seconds())
	return s
}
