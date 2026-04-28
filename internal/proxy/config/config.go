package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddress            string
	JWTSecret              string
	KeycloakURL            string
	KeycloakRealm          string
	KeycloakClientID       string
	KeycloakClientSecret   string
	KeycloakRedirectURI    string
	UserserviceGRPCAddress string
	FieldURL               string
	UserserviceURL         string
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	httpAddr := getenv("WEB_HTTP_ADDRESS")
	if httpAddr == "" {
		httpAddr = ":9092"
	}

	secret := getenv("WEB_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("WEB_JWT_SECRET is required")
	}

	keycloakURL := getenv("WEB_KEYCLOAK_URL")
	if keycloakURL == "" {
		return Config{}, fmt.Errorf("WEB_KEYCLOAK_URL is required")
	}

	realm := getenv("WEB_KEYCLOAK_REALM")
	if realm == "" {
		return Config{}, fmt.Errorf("WEB_KEYCLOAK_REALM is required")
	}

	clientID := getenv("WEB_KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		return Config{}, fmt.Errorf("WEB_KEYCLOAK_CLIENT_ID is required")
	}

	clientSecret := getenv("WEB_KEYCLOAK_CLIENT_SECRET")
	if clientSecret == "" {
		return Config{}, fmt.Errorf("WEB_KEYCLOAK_CLIENT_SECRET is required")
	}

	redirectURI := getenv("WEB_KEYCLOAK_REDIRECT_URI")
	if redirectURI == "" {
		return Config{}, fmt.Errorf("WEB_KEYCLOAK_REDIRECT_URI is required")
	}

	grpcAddr := getenv("WEB_USERSERVICE_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9091"
	}

	fieldURL := getenv("WEB_FIELD_URL")
	if fieldURL == "" {
		fieldURL = "http://localhost:8080"
	}

	userserviceURL := getenv("WEB_USERSERVICE_URL")
	if userserviceURL == "" {
		userserviceURL = "http://localhost:9090"
	}

	return Config{
		HTTPAddress:            httpAddr,
		JWTSecret:              secret,
		KeycloakURL:            keycloakURL,
		KeycloakRealm:          realm,
		KeycloakClientID:       clientID,
		KeycloakClientSecret:   clientSecret,
		KeycloakRedirectURI:    redirectURI,
		UserserviceGRPCAddress: grpcAddr,
		FieldURL:               fieldURL,
		UserserviceURL:         userserviceURL,
	}, nil
}

func MustLoadFromEnv() Config {
	cfg, err := LoadFromEnv(os.Getenv)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config: %s\n", err)
		os.Exit(1)
	}
	return cfg
}
