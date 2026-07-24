// Package validation provides field type validation and coercion for schema values.
package validation

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
)

// FieldError provides structured detail for a validation failure.
type FieldError struct {
	Field        string `json:"field"`
	ExpectedType string `json:"expected_type"`
	Got          string `json:"got"`
}

func (e FieldError) Error() string {
	return fmt.Sprintf("field %q: expected %s, got %s", e.Field, e.ExpectedType, e.Got)
}

// ValidateValue checks that a JSON-decoded value matches the declared schema type.
func ValidateValue(fieldKey string, value any, fieldType api.FieldType) error {
	switch fieldType {
	case api.String:
		if _, ok := value.(string); !ok {
			return FieldError{Field: fieldKey, ExpectedType: "string", Got: fmt.Sprintf("%T", value)}
		}
	case api.Int:
		f, ok := value.(float64)
		if !ok {
			return FieldError{Field: fieldKey, ExpectedType: "int", Got: fmt.Sprintf("%T", value)}
		}
		if math.Mod(f, 1.0) != 0 {
			return FieldError{Field: fieldKey, ExpectedType: "int", Got: fmt.Sprintf("%v", value)}
		}
	case api.Float:
		if _, ok := value.(float64); !ok {
			return FieldError{Field: fieldKey, ExpectedType: "float", Got: fmt.Sprintf("%T", value)}
		}
	case api.Boolean:
		if _, ok := value.(bool); !ok {
			return FieldError{Field: fieldKey, ExpectedType: "boolean", Got: fmt.Sprintf("%T", value)}
		}
	case api.Date, api.Datetime:
		s, ok := value.(string)
		if !ok {
			return FieldError{Field: fieldKey, ExpectedType: "date/datetime", Got: fmt.Sprintf("%T", value)}
		}
		if _, err := time.Parse(time.RFC3339, s); err != nil {
			return FieldError{Field: fieldKey, ExpectedType: "date/datetime", Got: s}
		}
	default:
		return FieldError{Field: fieldKey, ExpectedType: string(fieldType), Got: fmt.Sprintf("%T", value)}
	}
	return nil
}

// CoerceValue converts a stored TEXT value back to its native Go type.
func CoerceValue(raw string, fieldType api.FieldType) (any, error) {
	switch fieldType {
	case api.String:
		return raw, nil
	case api.Int:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("coerce int: %w", err)
		}
		return v, nil
	case api.Float:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("coerce float: %w", err)
		}
		return v, nil
	case api.Boolean:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("coerce bool: %w", err)
		}
		return v, nil
	case api.Date, api.Datetime:
		v, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, fmt.Errorf("coerce datetime: %w", err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown field type: %s", fieldType)
	}
}
