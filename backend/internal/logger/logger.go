package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu   sync.Mutex
	inst *slog.Logger
)

func Init(level string) *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	inst = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
	slog.SetDefault(inst)
	return inst
}

func L() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	if inst == nil {
		inst = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		slog.SetDefault(inst)
	}
	return inst
}
