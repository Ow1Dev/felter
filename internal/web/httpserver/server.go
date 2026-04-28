package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ow1Dev/felter/internal/userservice/pb"
	"github.com/Ow1Dev/felter/internal/web/config"
	"github.com/Ow1Dev/felter/internal/web/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type contextKey struct{}

type Server struct {
	cfg        config.Config
	httpClient *http.Client
	grpcConn   *grpc.ClientConn
	grpcClient pb.UserServiceClient
	provider   Provider
}

func New(cfg config.Config) *Server {
	conn, err := grpc.NewClient(cfg.UserserviceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to userservice: %v", err)
	}

	provider := NewKeycloakProvider(
		cfg.KeycloakURL,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakClientSecret,
		cfg.KeycloakRedirectURI,
	)

	return &Server{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		grpcConn:   conn,
		grpcClient: pb.NewUserServiceClient(conn),
		provider:   provider,
	}
}

func (s *Server) Close() {
	if s.grpcConn != nil {
		_ = s.grpcConn.Close()
	}
}

func (s *Server) HandleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()
		redirectURI := s.provider.BuildAuthURL(state, s.cfg.KeycloakRedirectURI)
		http.Redirect(w, r, redirectURI, http.StatusFound)
	}
}

func (s *Server) HandleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
			http.Error(w, `{"error":"missing code"}`, http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		accessToken, err := s.provider.ExchangeCode(ctx, body.Code, s.cfg.KeycloakRedirectURI)
		if err != nil {
			log.Printf("callback: exchange code failed: %v", err)
			http.Error(w, "failed to exchange code", http.StatusInternalServerError)
			return
		}

		userInfo, err := s.provider.GetUserInfo(ctx, accessToken)
		if err != nil {
			log.Printf("callback: get user info failed: %v", err)
			http.Error(w, "failed to get user info", http.StatusInternalServerError)
			return
		}

		username := deriveUsername(userInfo.Email)

		getResp, err := s.grpcClient.GetUserFromProvider(ctx, &pb.GetUserFromProviderRequest{
			Provider:   s.provider.Type(),
			ProviderId: userInfo.Sub,
		})
		if err == nil {
			jwtToken, err := jwt.Generate(userInfo.Sub, userInfo.Email, s.cfg.JWTSecret, 1*time.Hour)
			if err != nil {
				http.Error(w, "failed to generate token", http.StatusInternalServerError)
				return
			}
			s.writeJWT(w, jwtToken, getResp.Email)
			return
		}

		createResp, err := s.grpcClient.CreateUserFromProvider(ctx, &pb.CreateUserFromProviderRequest{
			Provider:   s.provider.Type(),
			ProviderId: userInfo.Sub,
			Email:      userInfo.Email,
			Username:   username,
		})
		if err != nil {
			log.Printf("callback: create user failed: %v", err)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		jwtToken, err := jwt.Generate(userInfo.Sub, userInfo.Email, s.cfg.JWTSecret, 1*time.Hour)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}
		s.writeJWT(w, jwtToken, createResp.Email)
	}
}

func (s *Server) HandleUserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.validateAuth(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"sub":   claims.Sub,
			"email": claims.Email,
		})
	}
}

func (s *Server) HandleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := s.validateAuth(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
	}
}

func (s *Server) validateAuth(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}
	return jwt.Validate(parts[1], s.cfg.JWTSecret)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) writeJWT(w http.ResponseWriter, token, email string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"email": email,
	})
}

func deriveUsername(email string) string {
	if idx := strings.Index(email, "@"); idx != -1 {
		return email[:idx]
	}
	return email
}
