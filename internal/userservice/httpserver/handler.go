// Package httpserver exposes the userservice over HTTP/JSON.
package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/httputil"
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
			httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		users, err := s.ListUsers(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
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

		_ = httputil.WriteJSON(w, http.StatusOK, resp)
	})
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// recoverer converts panics to JSON 500 responses.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
