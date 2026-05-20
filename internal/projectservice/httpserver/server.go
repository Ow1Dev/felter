// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Ow1Dev/felter/internal/middleware"
	"github.com/Ow1Dev/felter/internal/projectservice/config"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// New builds the root handler with routes and middleware.
func New(_ config.Config, s store.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, s)

	// Wrap with middleware: recover -> logger
	var h http.Handler = mux
	h = middleware.Recoverer(logger, h)
	h = middleware.RequestLogger(logger, h)
	return h
}
