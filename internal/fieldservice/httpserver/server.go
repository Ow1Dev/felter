// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Ow1Dev/felter/internal/fieldservice/config"
	"github.com/Ow1Dev/felter/internal/middleware"
)

// New builds the root handler with routes and middleware.
func New(cfg config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, cfg)

	var h http.Handler = mux
	h = middleware.Recoverer(logger, h)
	h = middleware.RequestLogger(logger, h)
	return h
}
