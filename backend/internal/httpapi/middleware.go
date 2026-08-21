package httpapi

import (
	"net/http"
	"time"

	"minikafka/internal/broker"
	"minikafka/internal/logger"
	"minikafka/internal/reqid"
)

func withConn(b *broker.Broker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.IncConn()
		defer b.DecConn()
		next.ServeHTTP(w, r)
	})
}

func withLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.L().Debug("http",
			"method", r.Method,
			"path", r.URL.Path,
			"req_id", reqid.From(r.Context()),
			"ms", time.Since(start).Milliseconds(),
		)
	})
}

func withReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = reqid.New()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(reqid.With(r.Context(), id)))
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.L().Error("panic", "err", rec, "req_id", reqid.From(r.Context()))
				writeErr(w, 500, "internal_error", "internal error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func chain(h http.Handler, wraps ...func(http.Handler) http.Handler) http.Handler {
	for i := len(wraps) - 1; i >= 0; i-- {
		h = wraps[i](h)
	}
	return h
}
