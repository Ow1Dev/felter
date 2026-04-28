package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ow1Dev/felter/internal/web/config"
	"github.com/Ow1Dev/felter/internal/web/httpserver"
)

func main() {
	cfg := config.MustLoadFromEnv()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, cfg); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	srv := httpserver.New(cfg)
	defer srv.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/login", srv.HandleLogin())
	mux.HandleFunc("/api/auth/callback", srv.HandleCallback())
	mux.HandleFunc("/api/auth/userinfo", srv.HandleUserInfo())
	mux.HandleFunc("/api/auth/logout", srv.HandleLogout())

	mux.Handle("/api/field/", srv.HandleFieldProxy())
	mux.Handle("/api/users/", srv.HandleUserserviceProxy())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: mux,
	}

	go func() {
		log.Printf("web listening on %s", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	return nil
}
