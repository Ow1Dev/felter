package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ow1Dev/felter/internal/config"
)

func TestServerRoutes(t *testing.T) {
	cfg := config.Config{
		Address:        ":0",
		AllowedOrigins: []string{"http://localhost:4200"},
	}
	h := New(cfg)

	t.Run("hello route", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d", rr.Code)
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body.Message != "hello world" {
			t.Fatalf("unexpected: %q", body.Message)
		}
	})

	t.Run("not found json", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/nope", nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct == "" {
			t.Fatalf("content-type not set")
		}
	})
}
