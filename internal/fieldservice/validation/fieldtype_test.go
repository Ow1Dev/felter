package validation

import (
	"strconv"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
)

func TestValidateValue(t *testing.T) {
	cases := []struct {
		name      string
		fieldType api.FieldType
		value     any
		wantErr   bool
		wantMsg   string
	}{
		{"string valid", api.String, "hello", false, ""},
		{"string invalid", api.String, 42, true, "expected string"},
		{"int valid", api.Int, float64(42), false, ""},
		{"int from float invalid", api.Int, float64(42.5), true, "expected int"},
		{"int wrong type", api.Int, "42", true, "expected int"},
		{"float valid", api.Float, float64(3.14), false, ""},
		{"float wrong type", api.Float, "3.14", true, "expected float"},
		{"boolean valid", api.Boolean, true, false, ""},
		{"boolean invalid", api.Boolean, "true", true, "expected boolean"},
		{"date valid", api.Date, "2026-07-15T00:00:00Z", false, ""},
		{"date invalid", api.Date, "not-a-date", true, "expected date/datetime"},
		{"datetime valid", api.Datetime, "2026-07-15T12:30:00Z", false, ""},
		{"datetime invalid type", api.Datetime, 123, true, "expected date/datetime"},
		{"unknown type", api.FieldType("unknown"), "x", true, "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateValue("testField", tc.value, tc.fieldType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if fe, ok := err.(FieldError); ok {
					if fe.Field != "testField" {
						t.Fatalf("field = %q, want %q", fe.Field, "testField")
					}
				} else {
					t.Fatalf("expected FieldError, got %T", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCoerceValue(t *testing.T) {
	cases := []struct {
		name      string
		fieldType api.FieldType
		raw       string
		want      any
		wantErr   bool
	}{
		{"string", api.String, "hello", "hello", false},
		{"int", api.Int, "42", int64(42), false},
		{"int invalid", api.Int, "abc", nil, true},
		{"float", api.Float, "3.14", float64(3.14), false},
		{"float invalid", api.Float, "abc", nil, true},
		{"boolean true", api.Boolean, "true", true, false},
		{"boolean false", api.Boolean, "false", false, false},
		{"boolean invalid", api.Boolean, "maybe", nil, true},
		{"date", api.Date, "2026-07-15T00:00:00Z", time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), false},
		{"datetime", api.Datetime, "2026-07-15T12:30:00Z", time.Date(2026, 7, 15, 12, 30, 0, 0, time.UTC), false},
		{"datetime invalid", api.Datetime, "bad", nil, true},
		{"unknown type", api.FieldType("unknown"), "x", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CoerceValue(tc.raw, tc.fieldType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch want := tc.want.(type) {
			case time.Time:
				gotT, ok := got.(time.Time)
				if !ok || !gotT.Equal(want) {
					t.Fatalf("got %v, want %v", got, want)
				}
			default:
				if got != want {
					t.Fatalf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		fieldType api.FieldType
		value     any
	}{
		{"string", api.String, "hello"},
		{"int", api.Int, float64(42)},
		{"float", api.Float, float64(3.14)},
		{"boolean", api.Boolean, true},
		{"date", api.Date, "2026-07-15T00:00:00Z"},
		{"datetime", api.Datetime, "2026-07-15T12:30:00Z"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate the original value
			if err := ValidateValue("f", tc.value, tc.fieldType); err != nil {
				t.Fatalf("validate: %v", err)
			}

			// Simulate storage: convert to string
			var raw string
			switch v := tc.value.(type) {
			case string:
				raw = v
			case float64:
				if tc.fieldType == api.Int {
					raw = strconv.FormatInt(int64(v), 10)
				} else {
					raw = strconv.FormatFloat(v, 'f', -1, 64)
				}
			case bool:
				raw = strconv.FormatBool(v)
			default:
				raw = v.(string)
			}

			// Coerce back
			coerced, err := CoerceValue(raw, tc.fieldType)
			if err != nil {
				t.Fatalf("coerce: %v", err)
			}

			// Basic sanity check on coerced type
			switch tc.fieldType {
			case api.String:
				if _, ok := coerced.(string); !ok {
					t.Fatalf("expected string, got %T", coerced)
				}
			case api.Int:
				if _, ok := coerced.(int64); !ok {
					t.Fatalf("expected int64, got %T", coerced)
				}
			case api.Float:
				if _, ok := coerced.(float64); !ok {
					t.Fatalf("expected float64, got %T", coerced)
				}
			case api.Boolean:
				if _, ok := coerced.(bool); !ok {
					t.Fatalf("expected bool, got %T", coerced)
				}
			case api.Date, api.Datetime:
				if _, ok := coerced.(time.Time); !ok {
					t.Fatalf("expected time.Time, got %T", coerced)
				}
			}
		})
	}
}
