// Package handlers contains HTTP handler constructors for API endpoints.
package handlers

import (
	"net/http"
	"strings"

	"github.com/Ow1Dev/felter/internal/httputil"
	"github.com/Ow1Dev/felter/internal/projectservice/api"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
)

// HandleProjects returns a handler for POST /projects (create) and GET /projects (list).
func HandleProjects(s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateProject(w, r, s)
		case http.MethodGet:
			handleListProjects(w, r, s)
		default:
			httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}

func handleCreateProject(w http.ResponseWriter, r *http.Request, s store.Store) {
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
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = httputil.WriteJSON(w, http.StatusCreated, project)
}

func handleListProjects(w http.ResponseWriter, r *http.Request, s store.Store) {
	projects, err := s.ListProjects(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = httputil.WriteJSON(w, http.StatusOK, projects)
}

// HandleProject returns a handler for GET /projects/{slug}.
func HandleProject(s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		path := r.URL.Path
		slug := strings.TrimPrefix(path, "/projects/")
		if slug == path {
			httputil.WriteError(w, http.StatusBadRequest, "missing project slug")
			return
		}

		project, err := s.GetProjectBySlug(r.Context(), slug)
		if err != nil {
			if err == store.ErrProjectNotFound {
				httputil.WriteError(w, http.StatusNotFound, "project not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_ = httputil.WriteJSON(w, http.StatusOK, project)
	})
}
