package httputil

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONAndError(t *testing.T) {
	t.Run("write json ok", func(t *testing.T) {
		rr := httptest.NewRecorder()
		type resp struct {
			X string `json:"x"`
		}
		if err := WriteJSON(rr, http.StatusCreated, resp{X: "y"}); err != nil {
			t.Fatalf("WriteJSON error: %v", err)
		}
		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct == "" {
			t.Fatalf("content-type not set")
		}
		if got := rr.Body.String(); got == "" {
			t.Fatalf("empty body")
		}
	})

	t.Run("write error helper", func(t *testing.T) {
		rr := httptest.NewRecorder()
		WriteError(rr, http.StatusTeapot, "oops")
		if rr.Code != http.StatusTeapot {
			t.Fatalf("status = %d", rr.Code)
		}
		if got := rr.Body.String(); got == "" {
			t.Fatalf("empty body")
		}
	})
}

func TestDecode(t *testing.T) {
	type payload struct {
		A int `json:"a"`
	}

	t.Run("ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"a": 3}`))
		r.Header.Set("Content-Type", "application/json")
		got, err := Decode[payload](r)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if got.A != 3 {
			t.Fatalf("unexpected value: %+v", got)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{`))
		_, err := Decode[payload](r)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !errors.Is(err, io.EOF) && err.Error() == "" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
