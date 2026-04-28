package httpserver

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"
)

func (s *Server) HandleFieldProxy() http.HandlerFunc {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = s.cfg.FieldURL
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/field")
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

		proxy.ServeHTTP(w, r)
	}
}

func (s *Server) HandleUserserviceProxy() http.HandlerFunc {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = s.cfg.UserserviceURL
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/users")
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

		proxy.ServeHTTP(w, r)
	}
}
