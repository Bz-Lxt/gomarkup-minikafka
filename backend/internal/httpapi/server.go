package httpapi

import (
	"net/http"

	"minikafka/internal/broker"
)

type Server struct {
	b         *broker.Broker
	staticDir string
}

func New(b *broker.Broker, staticDir string) *Server {
	return &Server{b: b, staticDir: staticDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/metrics", s.metrics)
	mux.HandleFunc("GET /api/v1/topics", s.listTopics)
	mux.HandleFunc("POST /api/v1/topics", s.createTopic)
	mux.HandleFunc("GET /api/v1/topics/{name}", s.topicDetail)
	mux.HandleFunc("GET /api/v1/topics/{name}/messages", s.messages)
	mux.HandleFunc("POST /api/v1/produce", s.produce)
	mux.HandleFunc("POST /api/v1/produce/batch", s.produceBatch)
	mux.HandleFunc("POST /api/v1/consume", s.consume)
	mux.HandleFunc("GET /api/v1/groups", s.listGroups)
	mux.HandleFunc("GET /api/v1/groups/{group}", s.groupDetail)
	mux.HandleFunc("POST /api/v1/groups/{group}/commit", s.commit)
	mux.HandleFunc("POST /api/v1/groups/{group}/reset", s.reset)
	mux.HandleFunc("/", s.spa)
	return chain(mux, withRecover, withReqID, func(h http.Handler) http.Handler {
		return withConn(s.b, h)
	}, withLog)
}
