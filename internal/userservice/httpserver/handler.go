// Package httpserver exposes the userservice over HTTP/JSON.
package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ow1Dev/felter/internal/userservice/store"
)

// NewServer creates the HTTP handler with all routes and middleware.
func NewServer(s store.Store) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, s)
	var handler http.Handler = mux
	handler = recoverer(handler)
	return handler
}

// addRoutes maps the entire HTTP API surface in one place.
func addRoutes(mux *http.ServeMux, s store.Store) {
	mux.Handle("/api/users", handleListUsers(s))
	mux.HandleFunc("/healthz", handleHealthz())
	mux.Handle("/", http.NotFoundHandler())
}

func handleListUsers(s store.Store) http.Handler {
	type response struct {
		ID          int64   `json:"id"`
		Email       string  `json:"email"`
		Username    string  `json:"username"`
		DisplayName *string `json:"display_name,omitempty"`
		CreatedAt   string  `json:"created_at"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		users, err := s.ListUsers(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := make([]response, len(users))
		for i, u := range users {
			resp[i] = response{
				ID:        u.ID,
				Email:     u.Email,
				Username:  u.Username,
				CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			if u.DisplayName != nil {
				resp[i].DisplayName = u.DisplayName
			}
		}

		_ = encode(w, r, http.StatusOK, resp)
	})
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// recoverer converts panics to JSON 500 responses.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("{\"error\":\"internal server error\"}\n"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
