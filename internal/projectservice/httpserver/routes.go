// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/projectservice/handlers"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, s store.Store) {
	mux.Handle("POST /{$}", handlers.HandleCreateProject(s))
	mux.Handle("GET /{$}", handlers.HandleListProjects(s))
	mux.Handle("GET /{slug}", handlers.HandleGetProject(s))
}
