package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeErr(w, 404, "not_found", "unknown endpoint", nil)
		return
	}
	if s.staticDir == "" {
		writeJSON(w, 200, map[string]any{"service": "minikafka", "ui": "not mounted"})
		return
	}
	p := path.Clean(r.URL.Path)
	if p == "/" {
		p = "/index.html"
	}
	full := path.Join(s.staticDir, p)
	if _, err := os.Stat(full); err == nil {
		http.ServeFile(w, r, full)
		return
	}
	index := path.Join(s.staticDir, "index.html")
	if _, err := os.Stat(index); errors.Is(err, fs.ErrNotExist) {
		writeJSON(w, 200, map[string]any{"service": "minikafka", "ui": "index missing"})
		return
	}
	http.ServeFile(w, r, index)
}
