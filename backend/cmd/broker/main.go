package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"minikafka/internal/broker"
	"minikafka/internal/config"
	"minikafka/internal/httpapi"
	"minikafka/internal/logger"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)
	log := logger.L()

	b, err := broker.Open(cfg)
	if err != nil {
		log.Error("open broker", "err", err)
		os.Exit(1)
	}
	defer b.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.New(b, cfg.StaticDir).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("minikafka listening", "addr", cfg.HTTPAddr, "data", cfg.DataDir)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("shutdown complete")
}
