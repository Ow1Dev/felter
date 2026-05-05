package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ow1Dev/felter/internal/fieldservice/config"
)

func TestServerRoutes(t *testing.T) {
	cfg := config.Config{
		Address: ":0",
	}
	h := New(cfg)

	t.Run("hello route", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("not found json", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/nope", nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
	})
}
