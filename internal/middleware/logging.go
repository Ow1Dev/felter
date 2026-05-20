// Package middleware provides HTTP middleware for correlation ID propagation and structured logging.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Ow1Dev/felter/internal/log"
)

const correlationHeader = "X-Correlation-ID"

// CorrelationID generates a new correlation ID and stores it in the request context.
// Any client-provided X-Correlation-ID header is ignored; the proxy always generates its own.
// The header is set on the request so downstream services receive it.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		r.Header.Set(correlationHeader, id)

		ctx := log.SetCorrelationID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLogger logs each request at Info level with method, path, status, duration, and correlation ID.
func RequestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		dur := time.Since(start)
		corr := r.Header.Get(correlationHeader)

		logger.Info("request",
			slog.String("corr", corr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.status),
			slog.Int64("dur_ms", dur.Milliseconds()),
		)
	})
}

// Recoverer converts panics to JSON 500 responses and logs at Error level.
func Recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				corr := r.Header.Get(correlationHeader)
				logger.Error("panic recovered",
					slog.String("corr", corr),
					slog.Any("panic", rec),
				)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
