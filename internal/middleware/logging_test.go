package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ow1Dev/felter/internal/log"
)

func TestCorrelationID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corr := log.CorrelationID(r.Context())
		if corr == "" {
			t.Error("correlation ID not found in context")
		}
		w.Header().Set("X-Correlation-ID", corr)
	})

	t.Run("generates ID when header missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		CorrelationID(handler).ServeHTTP(rr, req)

		if rr.Header().Get("X-Correlation-ID") == "" {
			t.Error("expected X-Correlation-ID header to be set")
		}
	})

	t.Run("always generates new ID even if header provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Correlation-ID", "client-provided-id")
		rr := httptest.NewRecorder()

		CorrelationID(handler).ServeHTTP(rr, req)

		got := rr.Header().Get("X-Correlation-ID")
		if got == "client-provided-id" {
			t.Error("expected proxy-generated ID, not client-provided")
		}
		if got == "" {
			t.Error("expected X-Correlation-ID header to be set")
		}
	})

	t.Run("sets header on request", func(t *testing.T) {
		var capturedID string
		h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			capturedID = r.Header.Get("X-Correlation-ID")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		CorrelationID(h).ServeHTTP(httptest.NewRecorder(), req)

		if capturedID == "" {
			t.Error("expected X-Correlation-ID header on request")
		}
	})
}

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		handler    http.HandlerFunc
		wantStatus int
	}{
		{
			name:       "200 OK",
			method:     http.MethodGet,
			path:       "/hello",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "404 not found",
			method:     http.MethodGet,
			path:       "/missing",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) },
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "500 internal error",
			method:     http.MethodPost,
			path:       "/error",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set(correlationHeader, "test-corr")
			rr := httptest.NewRecorder()

			RequestLogger(logger, tt.handler).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("failed to parse log entry: %v", err)
			}

			if got, want := entry["msg"], "request"; got != want {
				t.Errorf("msg = %v, want %v", got, want)
			}
			if got, want := entry["corr"], "test-corr"; got != want {
				t.Errorf("corr = %v, want %v", got, want)
			}
			if got, want := entry["method"], tt.method; got != want {
				t.Errorf("method = %v, want %v", got, want)
			}
			if got, want := entry["path"], tt.path; got != want {
				t.Errorf("path = %v, want %v", got, want)
			}
			if got, want := int(entry["status"].(float64)), tt.wantStatus; got != want {
				t.Errorf("status = %d, want %d", got, want)
			}
			if _, ok := entry["dur_ms"]; !ok {
				t.Error("missing dur_ms field")
			}
		})
	}
}

func TestRecoverer(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
		wantLog    bool
	}{
		{
			name: "no panic passes through",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
			wantLog:    false,
		},
		{
			name: "panic returns 500 JSON",
			handler: func(_ http.ResponseWriter, _ *http.Request) {
				panic("something went wrong")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"internal server error"}`,
			wantLog:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(correlationHeader, "test-corr")
			rr := httptest.NewRecorder()

			Recoverer(logger, tt.handler).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				body := rr.Body.String()
				if len(body) >= len(tt.wantBody) {
					body = body[:len(tt.wantBody)]
				}
				if body != tt.wantBody {
					t.Errorf("body = %q, want %q", body, tt.wantBody)
				}
			}

			if tt.wantLog {
				var entry map[string]any
				if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
					t.Fatalf("failed to parse log entry: %v", err)
				}
				if got, want := entry["msg"], "panic recovered"; got != want {
					t.Errorf("msg = %v, want %v", got, want)
				}
				if got, want := entry["corr"], "test-corr"; got != want {
					t.Errorf("corr = %v, want %v", got, want)
				}
				if got := entry["level"]; got != "ERROR" {
					t.Errorf("level = %v, want ERROR", got)
				}
			}
		})
	}
}
