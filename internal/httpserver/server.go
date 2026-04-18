package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/config"
	"github.com/Ow1Dev/felter/internal/handlers"
)

// New builds the root handler with routes and middleware.
func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/api/hello", handlers.Hello)

	// NotFound -> JSON
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// ServeMux calls this handler for unmatched patterns we attach; for true 404s, we ensure JSON error
		writeError(w, http.StatusNotFound, "not found")
	})

	// Wrap with middleware: recover -> cors -> logger
	var h http.Handler = mux
	h = recoverer(h)
	h = cors(cfg.AllowedOrigins)(h)
	h = logger(h)
	return h
}
