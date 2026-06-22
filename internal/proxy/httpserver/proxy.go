package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/Ow1Dev/felter/internal/log"
	"github.com/Ow1Dev/felter/internal/userservice/pb"
)

// HandleProxy creates a reverse proxy handler that validates auth and forwards to the target service.
func (s *Server) HandleProxy(targetURL, pathPrefix string) http.HandlerFunc {
	host := strings.TrimPrefix(targetURL, "http://")
	host = strings.TrimPrefix(host, "https://")

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = host
			path := strings.TrimPrefix(r.URL.Path, pathPrefix)
			if path == "" {
				path = "/"
			}
			r.URL.Path = path
			r.Host = host
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.validateAuth(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		userID, err := s.resolveUserID(r.Context(), claims.Sub)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, claims)
		r = r.WithContext(ctx)
		r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
		r.Header.Set("X-Correlation-ID", log.CorrelationID(r.Context()))

		proxy.ServeHTTP(w, r)
	}
}

// resolveUserID resolves the Keycloak provider ID to the internal userservice ID,
// using the cache when available.
func (s *Server) resolveUserID(ctx context.Context, providerID string) (int64, error) {
	provider := s.provider.Type()

	// Try cache first.
	if userID, found, err := s.userCache.Get(ctx, provider, providerID); err != nil {
		s.logger.Warn("cache get failed", slog.String("err", err.Error()))
	} else if found {
		return userID, nil
	}

	// Fetch from userservice.
	resp, err := s.grpcClient.GetUserFromProvider(ctx, &pb.GetUserFromProviderRequest{
		Provider:   provider,
		ProviderId: providerID,
	})
	if err != nil {
		return 0, err
	}

	// Cache the result.
	_ = s.userCache.Set(ctx, provider, providerID, resp.Id)
	return resp.Id, nil
}
