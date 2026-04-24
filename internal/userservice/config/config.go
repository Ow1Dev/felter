// Package config loads runtime configuration for the userservice.
package config

// Config holds userservice runtime configuration values.
type Config struct {
	DatabaseDSN string
	GRPCAddress string
	HTTPAddress string
}

// LoadFromEnv reads configuration from the provided getenv function.
func LoadFromEnv(getenv func(string) string) Config {
	dsn := getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://felter:felter@localhost:5432/felter?sslmode=disable"
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
	}
}
