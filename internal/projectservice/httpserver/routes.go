// Package httpserver wires the HTTP server's routing and middleware.
package httpserver

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/httputil"
	"github.com/Ow1Dev/felter/internal/projectservice/handlers"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, s store.Store) {
	mux.Handle("/projects", handlers.HandleProjects(s))
	mux.Handle("/projects/", handlers.HandleProject(s))

	// NotFound -> JSON
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
}
