package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"minikafka/internal/apperror"
	"minikafka/internal/broker"
	"minikafka/internal/logger"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeErr(w http.ResponseWriter, status int, code, msg string, details any) {
	writeJSON(w, status, map[string]any{"error": apiErr{Code: code, Message: msg, Details: details}})
}

func mapErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, broker.ErrNotFound), apperror.IsNotFound(err):
		writeErr(w, 404, "not_found", err.Error(), nil)
	case errors.Is(err, broker.ErrAlreadyExists), apperror.IsAlreadyExists(err):
		writeErr(w, 409, "already_exists", err.Error(), nil)
	case errors.Is(err, broker.ErrInvalid), apperror.IsInvalid(err):
		writeErr(w, 422, "validation_error", err.Error(), nil)
	default:
		logger.L().Error("internal", "err", err)
		writeErr(w, 500, "internal_error", "internal error", nil)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(io.LimitReader(r.Body, 8<<10))
	if err := dec.Decode(dst); err != nil {
		writeErr(w, 400, "invalid_json", "请求体不是合法 JSON", nil)
		return false
	}
	return true
}
