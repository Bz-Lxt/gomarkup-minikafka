package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr            string
	DataDir             string
	StaticDir           string
	LogLevel            string
	SegmentMaxBytes     int64
	IndexIntervalBytes  int
	SyncMode            string
	SyncIntervalMS      int
}

func Load() Config {
	return Config{
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		DataDir:            env("DATA_DIR", "./data"),
		StaticDir:          env("STATIC_DIR", ""),
		LogLevel:           env("LOG_LEVEL", "info"),
		SegmentMaxBytes:    envInt64("SEGMENT_MAX_BYTES", 64*1024*1024),
		IndexIntervalBytes: int(envInt64("INDEX_INTERVAL_BYTES", 4096)),
		SyncMode:           strings.ToLower(env("SYNC_MODE", "batch")),
		SyncIntervalMS:     int(envInt64("SYNC_INTERVAL_MS", 100)),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt64(k string, def int64) int64 {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return def
}
