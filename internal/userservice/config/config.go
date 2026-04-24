// Package config loads runtime configuration for the userservice.
package config

import (
	"fmt"
	"os"
)

// Config holds userservice runtime configuration values.
type Config struct {
	DatabaseDSN string
	GRPCAddress string
	HTTPAddress string
}

// LoadFromEnv reads configuration from the provided getenv function.
// It returns an error if required environment variables are missing.
func LoadFromEnv(getenv func(string) string) (Config, error) {
	dsn := getenv("DATABASE_DSN")
	if dsn == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}

	grpc := getenv("GRPC_ADDRESS")
	if grpc == "" {
		grpc = ":9091"
	}

	http := getenv("HTTP_ADDRESS")
	if http == "" {
		http = ":9090"
	}

	return Config{
		DatabaseDSN: dsn,
		GRPCAddress: grpc,
		HTTPAddress: http,
	}, nil
}

// MustLoadFromEnv is a convenience wrapper that calls LoadFromEnv with
// os.Getenv and exits the program on error.
func MustLoadFromEnv() Config {
	cfg, err := LoadFromEnv(os.Getenv)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config: %s\n", err)
		os.Exit(1)
	}
	return cfg
}
