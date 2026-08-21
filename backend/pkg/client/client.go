package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	base   string
	http   *http.Client
}

func New(base string) *Client {
	return &Client{
		base: base,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

type ProduceResult struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
}

type BatchResult struct {
	Count int `json:"count"`
}

func (c *Client) Health() error {
	resp, err := c.http.Get(c.base + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("health %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) CreateTopic(name string, partitions int) error {
	return c.post("/api/v1/topics", map[string]any{"name": name, "partitions": partitions}, 201, nil)
}

func (c *Client) Produce(topic, key, value string) (*ProduceResult, error) {
	var wrap struct {
		Data ProduceResult `json:"data"`
	}
	if err := c.post("/api/v1/produce", map[string]any{
		"topic": topic, "key": key, "value": value,
	}, 201, &wrap); err != nil {
		return nil, err
	}
	return &wrap.Data, nil
}

func (c *Client) ProduceBatch(topic string, n int, value string) (int, error) {
	msgs := make([]map[string]string, n)
	for i := 0; i < n; i++ {
		msgs[i] = map[string]string{"key": fmt.Sprintf("k-%d", i), "value": value}
	}
	var wrap struct {
		Data BatchResult `json:"data"`
	}
	if err := c.post("/api/v1/produce/batch", map[string]any{"topic": topic, "messages": msgs}, 201, &wrap); err != nil {
		return 0, err
	}
	return wrap.Data.Count, nil
}

func (c *Client) post(path string, body any, want int, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := c.http.Post(c.base+path, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want && !(want == 201 && resp.StatusCode == 409) {
		return fmt.Errorf("%s -> %d %s", path, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}
