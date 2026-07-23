package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Ow1Dev/felter/internal/fieldservice/handlers"
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/httputil"
)

// addRoutes maps the entire API surface in one place.
func addRoutes(mux *http.ServeMux, schemaStore store.SchemaStore, recordStore store.RecordStore, logger *slog.Logger) {
	mux.Handle("GET /schemas", handlers.HandleListSchemas(logger, schemaStore))
	mux.Handle("POST /schemas", handlers.HandleCreateSchema(logger, schemaStore))
	mux.Handle("GET /schemas/{schemaKey}", handlers.HandleGetSchema(logger, schemaStore))
	mux.Handle("DELETE /schemas/{schemaKey}", handlers.HandleDeleteSchema(logger, schemaStore))
	mux.Handle("POST /schemas/{schemaKey}/fields", handlers.HandleCreateSchemaField(logger, schemaStore))
	mux.Handle("DELETE /schemas/{schemaKey}/fields/{fieldKey}", handlers.HandleDeleteSchemaField(logger, schemaStore))
	mux.Handle("POST /mutate", handlers.HandleMutateFieldValue(logger, recordStore))
	mux.Handle("POST /query", handlers.HandleQueryFieldValues(logger, recordStore))
	mux.Handle("GET /definition", handlers.HandleFieldValueDefinition(logger, recordStore))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
}
