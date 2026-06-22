package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Ow1Dev/felter/internal/fieldservice/handlers"
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/httputil"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, s store.Store, logger *slog.Logger) {
	mux.Handle("GET /schemas", handlers.HandleListSchemas(logger, s))
	mux.Handle("POST /schemas", handlers.HandleCreateSchema(logger, s))
	mux.Handle("GET /schemas/{schemaKey}", handlers.HandleGetSchema(logger, s))
	mux.Handle("DELETE /schemas/{schemaKey}", handlers.HandleDeleteSchema(logger, s))
	mux.Handle("POST /schemas/{schemaKey}/fields", handlers.HandleCreateSchemaField(logger, s))
	mux.Handle("DELETE /schemas/{schemaKey}/fields/{fieldKey}", handlers.HandleDeleteSchemaField(logger, s))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
}
