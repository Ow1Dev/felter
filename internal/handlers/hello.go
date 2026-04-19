package handlers

import (
	"net/http"

	"github.com/Ow1Dev/felter/internal/httputil"
)

// HandleHello returns a handler that replies with a hello world message.
func HandleHello() http.Handler {
	type response struct {
		Message string `json:"message"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = httputil.WriteJSON(w, http.StatusOK, response{Message: "hello world"})
	})
}
