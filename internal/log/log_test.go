package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestCorrelationID(t *testing.T) {
	t.Run("empty context returns empty string", func(t *testing.T) {
		ctx := context.Background()
		if got := CorrelationID(ctx); got != "" {
			t.Errorf("CorrelationID() = %q, want empty", got)
		}
	})

	t.Run("set and get correlation ID", func(t *testing.T) {
		ctx := context.Background()
		id := "test-corr-123"
		ctx = SetCorrelationID(ctx, id)

		if got := CorrelationID(ctx); got != id {
			t.Errorf("CorrelationID() = %q, want %q", got, id)
		}
	})
}

func TestWithCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	t.Run("no correlation ID omits corr field", func(t *testing.T) {
		buf.Reset()
		ctx := context.Background()
		l := WithCorrelationID(ctx, logger)
		l.Info("test")

		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log entry: %v", err)
		}
		if _, ok := entry["corr"]; ok {
			t.Error("unexpected corr field in log entry")
		}
	})

	t.Run("correlation ID included in log entry", func(t *testing.T) {
		buf.Reset()
		ctx := SetCorrelationID(context.Background(), "abc-123")
		l := WithCorrelationID(ctx, logger)
		l.Info("test")

		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log entry: %v", err)
		}
		if got, want := entry["corr"], "abc-123"; got != want {
			t.Errorf("corr = %v, want %v", got, want)
		}
	})

	t.Run("empty correlation ID omits corr field", func(t *testing.T) {
		buf.Reset()
		ctx := SetCorrelationID(context.Background(), "")
		l := WithCorrelationID(ctx, logger)
		l.Info("test")

		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log entry: %v", err)
		}
		if _, ok := entry["corr"]; ok {
			t.Error("unexpected corr field in log entry")
		}
	})
}
