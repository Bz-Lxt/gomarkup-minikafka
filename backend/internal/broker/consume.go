package broker

import (
	"minikafka/internal/clock"
	"minikafka/internal/validate"
)

func (b *Broker) Consume(topic, group, clientID string, max int, autoCommit bool) ([]int, []Message, error) {
	if err := validate.ResourceName("group", group); err != nil {
		return nil, nil, err
	}
	if clientID == "" {
		clientID = "anonymous"
	}
	max = validate.ConsumeLimit(max)
	t, err := b.topic(topic)
	if err != nil {
		return nil, nil, err
	}
	assigned := b.groups.assign(group, clientID, topic, len(t.parts))
	var msgs []Message
	for _, pid := range assigned {
		if len(msgs) >= max {
			break
		}
		p := t.parts[pid]
		from, ok := b.offsets.Get(group, topic, pid)
		if !ok {
			from = p.log.Earliest()
		}
		recs, _, err := p.log.Read(from, max-len(msgs))
		if err != nil {
			return assigned, msgs, err
		}
		last := from
		for _, r := range recs {
			msgs = append(msgs, Message{
				Topic: topic, Partition: pid, Offset: r.Offset,
				Key: string(r.Key), Value: string(r.Value),
				Timestamp: clock.Format(r.Timestamp),
			})
			last = r.Offset + 1
		}
		if autoCommit && last != from {
			if err := b.offsets.Commit(group, topic, pid, last); err != nil {
				return assigned, msgs, err
			}
		}
	}
	b.metrics.AddConsume(uint64(len(msgs)))
	return assigned, msgs, nil
}

func (b *Broker) Commit(group, topic string, partition int, off int64) error {
	if _, err := b.topic(topic); err != nil {
		return err
	}
	return b.offsets.Commit(group, topic, partition, off)
}

func (b *Broker) Reset(group, topic, to string) error {
	if err := validate.ResetTarget(to); err != nil {
		return err
	}
	t, err := b.topic(topic)
	if err != nil {
		return err
	}
	for _, p := range t.parts {
		target := p.log.Earliest()
		if to == "latest" {
			target = p.log.NextOffset()
		}
		if err := b.offsets.Commit(group, topic, p.ID, target); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) ReadMessages(topic string, partition int, from int64, limit int) ([]Message, int64, error) {
	t, err := b.topic(topic)
	if err != nil {
		return nil, 0, err
	}
	if err := validate.PartitionID(partition, len(t.parts)); err != nil {
		return nil, 0, err
	}
	limit = validate.BrowseLimit(limit)
	recs, _, err := t.parts[partition].log.Read(from, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]Message, 0, len(recs))
	next := from
	for _, r := range recs {
		out = append(out, Message{
			Topic: topic, Partition: partition, Offset: r.Offset,
			Key: string(r.Key), Value: string(r.Value), Timestamp: clock.Format(r.Timestamp),
		})
		next = r.Offset
	}
	if len(out) == 0 {
		next = from
	}
	return out, next, nil
}
