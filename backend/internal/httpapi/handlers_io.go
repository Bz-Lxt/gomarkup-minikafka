package httpapi

import "net/http"

func (s *Server) produce(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic     string `json:"topic"`
		Key       string `json:"key"`
		Value     string `json:"value"`
		Partition *int   `json:"partition"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" {
		writeErr(w, 422, "validation_error", "topic 必填", []map[string]string{{"field": "topic", "message": "必填"}})
		return
	}
	partition := *req.Partition
	msg, err := s.b.Produce(req.Topic, req.Key, req.Value, &partition)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 201, msg)
}

func (s *Server) produceBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic    string `json:"topic"`
		Messages []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"messages"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" || len(req.Messages) == 0 {
		writeErr(w, 422, "validation_error", "topic 与 messages 必填", nil)
		return
	}
	if len(req.Messages) > 10000 {
		writeErr(w, 422, "validation_error", "单批最多 10000 条", nil)
		return
	}
	pairs := make([][2]string, len(req.Messages))
	for i, m := range req.Messages {
		pairs[i] = [2]string{m.Key, m.Value}
	}
	msgs, err := s.b.ProduceBatch(req.Topic, pairs)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 201, map[string]any{"count": len(msgs), "messages": msgs})
}

func (s *Server) consume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic       string `json:"topic"`
		Group       string `json:"group"`
		ClientID    string `json:"client_id"`
		MaxMessages int    `json:"max_messages"`
		AutoCommit  bool   `json:"auto_commit"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" || req.Group == "" {
		writeErr(w, 422, "validation_error", "topic 与 group 必填", nil)
		return
	}
	assigned, msgs, err := s.b.Consume(req.Topic, req.Group, req.ClientID, req.MaxMessages, req.AutoCommit)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"assignments": assigned, "messages": msgs})
}
