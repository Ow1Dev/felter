// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/fieldservice/config"
)

// New builds the root handler with routes and middleware.
func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, cfg)

	// Wrap with middleware: recover -> cors -> logger
	var h http.Handler = mux
	h = recoverer(h)
	h = cors(cfg.AllowedOrigins)(h)
	h = logger(h)
	return h
}
