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
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = targetURL
			r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefix)
			r.Host = r.URL.Host
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
