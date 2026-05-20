// Package main runs the HTTP projectservice server.
//
//nolint:revive
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ow1Dev/felter/internal/db"
	"github.com/Ow1Dev/felter/internal/log"
	"github.com/Ow1Dev/felter/internal/projectservice/config"
	"github.com/Ow1Dev/felter/internal/projectservice/httpserver"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, cfg); err != nil {
		slog.Default().Error("fatal", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

// run wires the server and blocks until context cancellation, shutting down gracefully.
func run(ctx context.Context, cfg config.Config) error {
	logger := log.New()
	slog.SetDefault(logger)

	pool, err := db.Open(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer func() { _ = pool.Close() }()

	s := store.NewPostgresStore(pool)
	handler := httpserver.New(cfg, s, logger)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		logger.Info("projectservice listening", slog.String("addr", cfg.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("err", err.Error()))
		if err := srv.Close(); err != nil {
			logger.Error("force close error", slog.String("err", err.Error()))
		}
	}
	return nil
}
