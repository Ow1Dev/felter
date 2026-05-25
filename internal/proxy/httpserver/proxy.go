package httpserver

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/Ow1Dev/felter/internal/log"
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

		ctx := context.WithValue(r.Context(), contextKey{}, claims)
		r = r.WithContext(ctx)
		r.Header.Set("X-User-ID", claims.Sub)
		r.Header.Set("X-Correlation-ID", log.CorrelationID(r.Context()))

		proxy.ServeHTTP(w, r)
	}
}
