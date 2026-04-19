// Package config loads runtime configuration from environment variables
// with sensible defaults for local development.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds server runtime configuration values.
type Config struct {
	Address        string
	AllowedOrigins []string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
}

// Load reads configuration from environment variables and returns a Config.
func Load() Config {
	addr := getAddr()
	origins := getOrigins()
	return Config{
		Address:        addr,
		AllowedOrigins: origins,
		ReadTimeout:    getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:   getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:    getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
	}
}

func getAddr() string {
	// Support PORT env (common on platforms) and ADDRESS override
	if a := os.Getenv("ADDRESS"); a != "" {
		return a
	}
	if p := os.Getenv("PORT"); p != "" {
		// if PORT is digits, prefix with ':'
		if _, err := strconv.Atoi(p); err == nil {
			return ":" + p
		}
		return p
	}
	return ":8080"
}

func getOrigins() []string {
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, s := range parts {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return []string{"http://localhost:4200", "http://127.0.0.1:4200"}
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
