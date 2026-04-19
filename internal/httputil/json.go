package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse is a simple JSON error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON writes v as JSON with the provided status code.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// WriteError writes a JSON error response.
func WriteError(w http.ResponseWriter, status int, msg string) {
	// ignore encode error on error path
	_ = WriteJSON(w, status, ErrorResponse{Error: msg})
}

// Decode decodes a JSON request body into T.
func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
