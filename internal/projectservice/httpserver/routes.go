// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Ow1Dev/felter/internal/projectservice/handlers"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, s store.Store, logger *slog.Logger) {
	mux.Handle("POST /{$}", handlers.HandleCreateProject(logger, s))
	mux.Handle("GET /{$}", handlers.HandleListProjects(logger, s))
	mux.Handle("GET /{slug}", handlers.HandleGetProject(logger, s))
}
