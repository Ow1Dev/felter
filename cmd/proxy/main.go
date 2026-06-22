// Package main runs the authentication proxy server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Ow1Dev/felter/internal/log"
	"github.com/Ow1Dev/felter/internal/middleware"
	"github.com/Ow1Dev/felter/internal/proxy/cache"
	"github.com/Ow1Dev/felter/internal/proxy/config"
	"github.com/Ow1Dev/felter/internal/proxy/httpserver"
)

func main() {
	cfg := config.MustLoadFromEnv()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, cfg); err != nil {
		slog.Default().Error("fatal", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	logger := log.New()
	slog.SetDefault(logger)

	userCache := cache.NewMemoryCache(cfg.CacheTTL)

	srv := httpserver.New(cfg, logger, userCache)
	defer srv.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/login", srv.HandleLogin())
	mux.HandleFunc("/api/auth/callback", srv.HandleCallback())
	mux.HandleFunc("/api/auth/logout", srv.HandleLogout())
	mux.HandleFunc("/api/auth/me", srv.HandleMe())

	registerProxy(srv, mux, cfg.FieldURL, "/api/field")
	registerProxy(srv, mux, cfg.UserserviceURL, "/api/users")
	registerProxy(srv, mux, cfg.ProjectserviceURL, "/api/projects")

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var handler http.Handler = mux
	handler = middleware.CorrelationID(handler)
	handler = middleware.RequestLogger(logger, handler)
	handler = middleware.CORS(handler)

	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: handler,
	}

	go func() {
		logger.Info("proxy listening", slog.String("addr", cfg.HTTPAddress))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("err", err.Error()))
	}
	return nil
}

// registerProxy creates a reverse-proxy handler and registers both the exact
// path and its trailing-slash prefix so POSTs to the bare path are not redirected.
func registerProxy(srv *httpserver.Server, mux *http.ServeMux, targetURL, pathPrefix string) {
	handler := srv.HandleProxy(targetURL, pathPrefix)
	mux.Handle(pathPrefix, handler)
	if !strings.HasSuffix(pathPrefix, "/") {
		mux.Handle(pathPrefix+"/", handler)
	}
}
