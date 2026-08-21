package broker

import (
	"minikafka/internal/clock"
	"minikafka/internal/partitioner"
	"minikafka/internal/wal"
)

func (b *Broker) pickPartition(t *Topic, key string, forced *int) (int, error) {
	return partitioner.Choose(len(t.parts), key, forced, &t.rr)
}

func (b *Broker) Produce(topic, key, value string, partition *int) (Message, error) {
	t, err := b.topic(topic)
	if err != nil {
		return Message{}, err
	}
	pid, err := b.pickPartition(t, key, partition)
	if err != nil {
		return Message{}, err
	}
	p := t.parts[pid]
	ts := clock.NowMilli()
	off, err := p.log.Append(wal.Record{Timestamp: ts, Key: []byte(key), Value: []byte(value)})
	if err != nil {
		return Message{}, err
	}
	b.metrics.AddProduce(1)
	return Message{
		Topic: topic, Partition: pid, Offset: off,
		Key: key, Value: value, Timestamp: clock.Format(ts),
	}, nil
}

func (b *Broker) ProduceBatch(topic string, msgs [][2]string) ([]Message, error) {
	t, err := b.topic(topic)
	if err != nil {
		return nil, err
	}
	grouped := make(map[int][]wal.Record)
	meta := make(map[int][][2]string)
	for _, m := range msgs {
		pid, err := b.pickPartition(t, m[0], nil)
		if err != nil {
			return nil, err
		}
		ts := clock.NowMilli()
		grouped[pid] = append(grouped[pid], wal.Record{Timestamp: ts, Key: []byte(m[0]), Value: []byte(m[1])})
		meta[pid] = append(meta[pid], m)
	}
	var out []Message
	for pid, recs := range grouped {
		offs, err := t.parts[pid].log.AppendBatch(recs)
		if err != nil {
			return out, err
		}
		b.metrics.AddProduce(uint64(len(offs)))
		for i, off := range offs {
			out = append(out, Message{
				Topic: topic, Partition: pid, Offset: off,
				Key: meta[pid][i][0], Value: meta[pid][i][1],
				Timestamp: clock.Format(recs[i].Timestamp),
			})
		}
	}
	return out, nil
}
