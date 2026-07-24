// Package handlers contains HTTP handler constructors for fieldservice API endpoints.
package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/httputil"
)

func userIDFromHeader(r *http.Request) (int64, bool) {
	v := r.Header.Get("X-User-ID")
	if v == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// HandleListSchemas returns a handler for GET /schemas.
func HandleListSchemas(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		schemas, err := s.ListSchemasByProject(r.Context(), projectSlug)
		if err != nil {
			logger.Error("list schemas", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = httputil.WriteJSON(w, http.StatusOK, schemas)
	})
}

// HandleCreateSchema returns a handler for POST /schemas.
func HandleCreateSchema(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[api.CreateSchemaRequest](r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.ProjectSlug) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}
		if strings.TrimSpace(req.Key) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "key is required")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "name is required")
			return
		}

		createdBy, ok := userIDFromHeader(r)
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "missing user id")
			return
		}

		schema, err := s.CreateSchema(r.Context(), req.ProjectSlug, req.Key, req.Name, createdBy)
		if err != nil {
			if err == store.ErrSchemaAlreadyExists {
				httputil.WriteError(w, http.StatusConflict, "schema already exists")
				return
			}
			logger.Error("create schema", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = httputil.WriteJSON(w, http.StatusCreated, schema)
	})
}

// HandleGetSchema returns a handler for GET /schemas/{schemaKey}.
func HandleGetSchema(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schemaKey := r.PathValue("schemaKey")
		if schemaKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing schema key")
			return
		}
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		schema, fields, err := s.GetSchemaWithFields(r.Context(), projectSlug, schemaKey)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema not found")
				return
			}
			logger.Error("get schema", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		resp := api.SchemaWithFields{
			ProjectSlug: schema.ProjectSlug,
			Key:         schema.Key,
			Name:        schema.Name,
			CreatedBy:   schema.CreatedBy,
			CreatedAt:   schema.CreatedAt,
			UpdatedAt:   schema.UpdatedAt,
			Fields:      &fields,
		}
		_ = httputil.WriteJSON(w, http.StatusOK, resp)
	})
}

// HandleDeleteSchema returns a handler for DELETE /schemas/{schemaKey}.
func HandleDeleteSchema(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schemaKey := r.PathValue("schemaKey")
		if schemaKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing schema key")
			return
		}
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		err := s.DeleteSchema(r.Context(), projectSlug, schemaKey)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema not found")
				return
			}
			logger.Error("delete schema", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// HandleCreateSchemaField returns a handler for POST /schemas/{schemaKey}/fields.
func HandleCreateSchemaField(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schemaKey := r.PathValue("schemaKey")
		if schemaKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing schema key")
			return
		}
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		req, err := httputil.Decode[api.CreateSchemaFieldRequest](r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(req.Key) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "key is required")
			return
		}
		if req.Type == "" {
			httputil.WriteError(w, http.StatusBadRequest, "type is required")
			return
		}

		createdBy, ok := userIDFromHeader(r)
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "missing user id")
			return
		}

		field, err := s.CreateSchemaField(r.Context(), projectSlug, schemaKey, req.Key, req.Type, createdBy)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema not found")
				return
			}
			if err == store.ErrFieldAlreadyExists {
				httputil.WriteError(w, http.StatusConflict, "field already exists")
				return
			}
			logger.Error("create field", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = httputil.WriteJSON(w, http.StatusCreated, field)
	})
}

// HandleDeleteSchemaField returns a handler for DELETE /schemas/{schemaKey}/fields/{fieldKey}.
func HandleDeleteSchemaField(logger *slog.Logger, s store.SchemaStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schemaKey := r.PathValue("schemaKey")
		if schemaKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing schema key")
			return
		}
		fieldKey := r.PathValue("fieldKey")
		if fieldKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing field key")
			return
		}
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		err := s.DeleteSchemaField(r.Context(), projectSlug, schemaKey, fieldKey)
		if err != nil {
			if err == store.ErrSchemaNotFound || err == store.ErrFieldNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema or field not found")
				return
			}
			logger.Error("delete field", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
