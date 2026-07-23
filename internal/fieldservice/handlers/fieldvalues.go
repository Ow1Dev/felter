// Package handlers contains HTTP handler constructors for fieldservice API endpoints.
package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/fieldservice/validation"
	"github.com/Ow1Dev/felter/internal/httputil"
)

// HandleMutateFieldValue returns a handler for POST /fields/mutate.
func HandleMutateFieldValue(logger *slog.Logger, s store.RecordStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		req, err := httputil.Decode[fieldvalue.MutateRequest](r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.SchemaKey) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "schema_key is required")
			return
		}

		createdBy, ok := userIDFromHeader(r)
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "missing user id")
			return
		}

		var recordID *string
		if req.RecordId != nil {
			rid := req.RecordId.String()
			recordID = &rid
		}

		var values map[string]any
		if req.Values != nil {
			values = *req.Values
		}

		rec, err := s.MutateRecord(r.Context(), projectSlug, req.SchemaKey, recordID, values, req.DeleteRecord != nil && *req.DeleteRecord, createdBy)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema_not_found")
				return
			}
			if err == store.ErrRecordNotFound {
				httputil.WriteError(w, http.StatusNotFound, "record_not_found")
				return
			}
			if fe, ok := err.(validation.FieldError); ok {
				_ = httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{
					"error": "validation_failed",
					"details": []any{map[string]string{
						"field":         fe.Field,
						"expected_type": fe.ExpectedType,
						"got":           fe.Got,
					}},
				})
				return
			}
			logger.Error("mutate record", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if rec == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if recordID == nil {
			_ = httputil.WriteJSON(w, http.StatusCreated, rec)
		} else {
			_ = httputil.WriteJSON(w, http.StatusOK, rec)
		}
	})
}

// HandleQueryFieldValues returns a handler for POST /fields/query.
func HandleQueryFieldValues(logger *slog.Logger, s store.RecordStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		req, err := httputil.Decode[fieldvalue.QueryRequest](r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.SchemaKey) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "schema_key is required")
			return
		}

		var filter *store.FilterNode
		if req.Filter != nil {
			filter = convertFilterNode(*req.Filter)
		}

		records, err := s.QueryRecords(r.Context(), projectSlug, req.SchemaKey, filter)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema_not_found")
				return
			}
			logger.Error("query records", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := fieldvalue.QueryResponse{Records: records}
		_ = httputil.WriteJSON(w, http.StatusOK, resp)
	})
}

// HandleFieldValueDefinition returns a handler for GET /fields/definition.
func HandleFieldValueDefinition(logger *slog.Logger, s store.RecordStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectSlug := r.URL.Query().Get("project_slug")
		if projectSlug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "project_slug is required")
			return
		}

		schemaKey := r.URL.Query().Get("schema_key")
		if schemaKey == "" {
			httputil.WriteError(w, http.StatusBadRequest, "schema_key is required")
			return
		}

		def, err := s.GetSchemaDefinition(r.Context(), projectSlug, schemaKey)
		if err != nil {
			if err == store.ErrSchemaNotFound {
				httputil.WriteError(w, http.StatusNotFound, "schema_not_found")
				return
			}
			logger.Error("get schema definition", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		_ = httputil.WriteJSON(w, http.StatusOK, def)
	})
}

func convertFilterNode(n fieldvalue.FilterNode) *store.FilterNode {
	out := &store.FilterNode{
		Op:    store.FilterOp(n.Op),
		Field: "",
		Value: n.Value,
	}
	if n.Field != nil {
		out.Field = *n.Field
	}
	if n.Conditions != nil {
		out.Conditions = make([]store.FilterNode, 0, len(*n.Conditions))
		for _, c := range *n.Conditions {
			out.Conditions = append(out.Conditions, *convertFilterNode(c))
		}
	}
	return out
}
