package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/server"
	"danmakustream/content-service/internal/svc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	serviceContext, err := svc.New(cfg)
	if err != nil {
		logger.Error("initialize service", "error", err)
		os.Exit(1)
	}
	defer serviceContext.Close()

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           server.Router(serviceContext, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		logger.Info("content-service started", "address", server.Addr, "version", cfg.ServiceVersion, "commit", cfg.CommitSHA)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server stopped", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown", "error", err)
	}
}
