package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/config"
	"github.com/Ow1Dev/felter/internal/handlers"
	"github.com/Ow1Dev/felter/internal/httputil"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, _ config.Config) {
	mux.Handle("/api/hello", handlers.HandleHello())

	// NotFound -> JSON
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
}
