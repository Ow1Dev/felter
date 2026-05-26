// Package handlers contains HTTP handler constructors for API endpoints.
package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ow1Dev/felter/internal/httputil"
	"github.com/Ow1Dev/felter/internal/projectservice/api"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// HandleCreateProject returns a handler for POST /.
func HandleCreateProject(logger *slog.Logger, s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[api.CreateProjectRequest](r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.Name) == "" {
			httputil.WriteError(w, http.StatusBadRequest, "name is required")
			return
		}

		var desc string
		if req.Description != nil {
			desc = *req.Description
		}

		project, err := s.CreateProject(r.Context(), req.Name, desc)
		if err != nil {
			if err == store.ErrInvalidProjectName {
				httputil.WriteError(w, http.StatusBadRequest, "invalid project name")
				return
			}
			logger.Error("create project", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		_ = httputil.WriteJSON(w, http.StatusCreated, project)
	})
}

// HandleListProjects returns a handler for GET /.
func HandleListProjects(logger *slog.Logger, s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects, err := s.ListProjects(r.Context())
		if err != nil {
			logger.Error("list projects", slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = httputil.WriteJSON(w, http.StatusOK, projects)
	})
}

// HandleGetProject returns a handler for GET /{slug}.
func HandleGetProject(logger *slog.Logger, s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			httputil.WriteError(w, http.StatusBadRequest, "missing project slug")
			return
		}

		project, err := s.GetProjectBySlug(r.Context(), slug)
		if err != nil {
			if err == store.ErrProjectNotFound {
				httputil.WriteError(w, http.StatusNotFound, "project not found")
				return
			}
			logger.Error("get project", slog.String("slug", slug), slog.String("err", err.Error()))
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = httputil.WriteJSON(w, http.StatusOK, project)
	})
}
