// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/httputil"
	"github.com/Ow1Dev/felter/internal/projectservice/config"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// New builds the root handler with routes and middleware.
func New(_ config.Config, s store.Store) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, s)

	// Wrap with middleware: recover -> logger
	var h http.Handler = mux
	h = httputil.Recoverer(h)
	h = httputil.Logger(h)
	return h
}
