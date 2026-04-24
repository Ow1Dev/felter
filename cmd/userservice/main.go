// Package main runs the userservice gRPC + HTTP servers.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/Ow1Dev/felter/internal/db"
	"github.com/Ow1Dev/felter/internal/userservice/config"
	"github.com/Ow1Dev/felter/internal/userservice/grpcserver"
	"github.com/Ow1Dev/felter/internal/userservice/httpserver"
	"github.com/Ow1Dev/felter/internal/userservice/pb"
	"github.com/Ow1Dev/felter/internal/userservice/store"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, config.LoadFromEnv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	_ []string,
	loadConfig func(func(string) string) (config.Config, error),
	_ io.Reader,
	_, stderr io.Writer,
) error {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	pool, err := db.Open(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "db close: %v\n", err)
		}
	}()

	s := store.NewPostgresStore(pool)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	// gRPC
	grpcLis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcSrv, grpcserver.NewServer(s))
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("grpc listening on %s", cfg.GRPCAddress)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Printf("grpc error: %v", err)
		}
	}()

	// HTTP
	httpSrv := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: httpserver.NewServer(s),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("http listening on %s", cfg.HTTPAddress)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		_, _ = fmt.Fprintf(stderr, "http shutdown: %v\n", err)
	}

	wg.Wait()
	return nil
}
