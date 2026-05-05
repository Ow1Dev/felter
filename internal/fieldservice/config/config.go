// Package config loads runtime configuration from environment variables
// with sensible defaults for local development.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds server runtime configuration values.
type Config struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load reads configuration from environment variables and returns a Config.
func Load() Config {
	addr := getAddr()
	return Config{
		Address:      addr,
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
	}
}

func getAddr() string {
	if a := os.Getenv("ADDRESS"); a != "" {
		return a
	}
	if p := os.Getenv("PORT"); p != "" {
		if _, err := strconv.Atoi(p); err == nil {
			return ":" + p
		}
		return p
	}
	return ":8080"
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
