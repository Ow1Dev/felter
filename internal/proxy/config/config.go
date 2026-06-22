// Package config loads proxy runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds proxy runtime configuration values.
type Config struct {
	HTTPAddress            string
	JWTSecret              string
	KeycloakURL            string
	KeycloakPublicURL      string
	KeycloakRealm          string
	KeycloakClientID       string
	KeycloakClientSecret   string
	KeycloakRedirectURI    string
	UserserviceGRPCAddress string
	FieldURL               string
	UserserviceURL         string
	ProjectserviceURL      string
	CacheTTL               time.Duration
}

// LoadFromEnv reads configuration from environment variables.
func LoadFromEnv(getenv func(string) string) (Config, error) {
	httpAddr := getenv("PROXY_HTTP_ADDRESS")
	if httpAddr == "" {
		return Config{}, fmt.Errorf("PROXY_HTTP_ADDRESS is required")
	}

	secret := getenv("PROXY_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("PROXY_JWT_SECRET is required")
	}

	keycloakURL := getenv("PROXY_KEYCLOAK_URL")
	if keycloakURL == "" {
		return Config{}, fmt.Errorf("PROXY_KEYCLOAK_URL is required")
	}

	keycloakPublicURL := getenv("PROXY_KEYCLOAK_PUBLIC_URL")
	if keycloakPublicURL == "" {
		keycloakPublicURL = keycloakURL
	}

	realm := getenv("PROXY_KEYCLOAK_REALM")
	if realm == "" {
		return Config{}, fmt.Errorf("PROXY_KEYCLOAK_REALM is required")
	}

	clientID := getenv("PROXY_KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		return Config{}, fmt.Errorf("PROXY_KEYCLOAK_CLIENT_ID is required")
	}

	clientSecret := getenv("PROXY_KEYCLOAK_CLIENT_SECRET")
	if clientSecret == "" {
		return Config{}, fmt.Errorf("PROXY_KEYCLOAK_CLIENT_SECRET is required")
	}

	redirectURI := getenv("PROXY_KEYCLOAK_REDIRECT_URI")
	if redirectURI == "" {
		return Config{}, fmt.Errorf("PROXY_KEYCLOAK_REDIRECT_URI is required")
	}

	grpcAddr := getenv("PROXY_USERSERVICE_GRPC_ADDR")
	if grpcAddr == "" {
		return Config{}, fmt.Errorf("PROXY_USERSERVICE_GRPC_ADDR is required")
	}

	fieldURL := getenv("PROXY_FIELD_URL")
	if fieldURL == "" {
		return Config{}, fmt.Errorf("PROXY_FIELD_URL is required")
	}

	userserviceURL := getenv("PROXY_USERSERVICE_URL")
	if userserviceURL == "" {
		return Config{}, fmt.Errorf("PROXY_USERSERVICE_URL is required")
	}

	projectserviceURL := getenv("PROXY_PROJECTSERVICE_URL")
	if projectserviceURL == "" {
		return Config{}, fmt.Errorf("PROXY_PROJECTSERVICE_URL is required")
	}

	cacheTTL := 1 * time.Hour
	if v := getenv("PROXY_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cacheTTL = d
		}
	}

	return Config{
		HTTPAddress:            httpAddr,
		JWTSecret:              secret,
		KeycloakURL:            keycloakURL,
		KeycloakPublicURL:      keycloakPublicURL,
		KeycloakRealm:          realm,
		KeycloakClientID:       clientID,
		KeycloakClientSecret:   clientSecret,
		KeycloakRedirectURI:    redirectURI,
		UserserviceGRPCAddress: grpcAddr,
		FieldURL:               fieldURL,
		UserserviceURL:         userserviceURL,
		ProjectserviceURL:      projectserviceURL,
		CacheTTL:               cacheTTL,
	}, nil
}

// MustLoadFromEnv reads configuration and exits on error.
func MustLoadFromEnv() Config {
	cfg, err := LoadFromEnv(os.Getenv)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config: %s\n", err)
		os.Exit(1)
	}
	return cfg
}
