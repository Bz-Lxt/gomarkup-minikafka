package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "uptime_sec": s.b.Snapshot().UptimeSec})
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	writeData(w, 200, s.b.Snapshot())
}

func (s *Server) listTopics(w http.ResponseWriter, _ *http.Request) {
	writeData(w, 200, s.b.ListTopics())
}

func (s *Server) createTopic(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		Partitions int    `json:"partitions"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	var details []map[string]string
	if req.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "必填"})
	}
	if req.Partitions == 0 {
		req.Partitions = 1
	}
	if req.Partitions < 1 || req.Partitions > 16 {
		details = append(details, map[string]string{"field": "partitions", "message": "须为 1–16"})
	}
	if len(details) > 0 {
		writeErr(w, 422, "validation_error", "字段校验失败", details)
		return
	}
	t, err := s.b.CreateTopic(req.Name, req.Partitions)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 201, map[string]any{
		"name": t.Name, "partitions": t.PartitionCount(), "created_at": t.CreatedAtText(),
	})
}

func (s *Server) topicDetail(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	n, parts, err := s.b.TopicDetail(name)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"name": n, "partitions": parts})
}

func (s *Server) messages(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var partition *int
	if raw := r.URL.Query().Get("partition"); raw != "" {
		p, _ := strconv.Atoi(raw)
		partition = &p
	}
	off, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	msgs, next, err := s.b.ReadMessages(name, *partition, off, limit)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"messages": msgs, "next_offset": next})
}

func (s *Server) listGroups(w http.ResponseWriter, _ *http.Request) {
	writeData(w, 200, s.b.ListGroups())
}

func (s *Server) groupDetail(w http.ResponseWriter, r *http.Request) {
	g := r.PathValue("group")
	parts, err := s.b.GroupDetail(g)
	if err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"group": g, "members": s.b.GroupMembers(g), "partitions": parts})
}

func (s *Server) commit(w http.ResponseWriter, r *http.Request) {
	g := r.PathValue("group")
	var req struct {
		Topic     string `json:"topic"`
		Partition int    `json:"partition"`
		Offset    int64  `json:"offset"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" {
		writeErr(w, 422, "validation_error", "topic 必填", nil)
		return
	}
	if err := s.b.Commit(g, req.Topic, req.Partition, req.Offset); err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"ok": true})
}

func (s *Server) reset(w http.ResponseWriter, r *http.Request) {
	g := r.PathValue("group")
	var req struct {
		Topic string `json:"topic"`
		To    string `json:"to"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" || (req.To != "earliest" && req.To != "latest") {
		writeErr(w, 422, "validation_error", "topic 必填且 to 须为 earliest|latest", nil)
		return
	}
	if err := s.b.Reset(g, req.Topic, req.To); err != nil {
		mapErr(w, err)
		return
	}
	writeData(w, 200, map[string]any{"ok": true})
}
