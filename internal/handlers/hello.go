package handlers

import (
	"net/http"
)

// Hello writes a simple hello world JSON response.
func Hello(w http.ResponseWriter, r *http.Request) {
	// response kept simple and explicit here; helpers used in httpserver layer
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"message\":\"hello world\"}\n"))
}
