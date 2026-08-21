package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"minikafka/internal/logger"
	"minikafka/pkg/client"
)

func main() {
	base := flag.String("base", "http://127.0.0.1:18591", "broker base url")
	topic := flag.String("topic", "bench", "topic name")
	parts := flag.Int("partitions", 1, "partitions when creating topic")
	n := flag.Int("n", 20000, "messages")
	batch := flag.Int("batch", 200, "batch size")
	payload := flag.Int("size", 64, "value bytes")
	flag.Parse()

	logger.Init("info")
	c := client.New(*base)
	if err := c.Health(); err != nil {
		logger.L().Error("health", "err", err)
		os.Exit(1)
	}
	_ = c.CreateTopic(*topic, *parts)
	val := string(bytesN(*payload))
	start := time.Now()
	sent := 0
	for sent < *n {
		k := *batch
		if sent+k > *n {
			k = *n - sent
		}
		got, err := c.ProduceBatch(*topic, k, val)
		if err != nil {
			logger.L().Error("produce", "err", err)
			os.Exit(1)
		}
		sent += got
	}
	dt := time.Since(start).Seconds()
	rate := float64(sent) / dt
	fmt.Printf("sent=%d elapsed=%.3fs rate=%.0f msg/s\n", sent, dt, rate)
}

func bytesN(n int) []byte {
	if n < 1 {
		n = 1
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return b
}
