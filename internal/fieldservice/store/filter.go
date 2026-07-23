package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
)

// FilterOp represents a filter operation.
type FilterOp string

// Filter operators for record queries.
const (
	OpEq   FilterOp = "eq"   // Equal
	OpNe   FilterOp = "ne"   // Not equal
	OpGt   FilterOp = "gt"   // Greater than
	OpGte  FilterOp = "gte"  // Greater than or equal
	OpLt   FilterOp = "lt"   // Less than
	OpLte  FilterOp = "lte"  // Less than or equal
	OpLike FilterOp = "like" // Substring match
	OpAnd  FilterOp = "and"  // Logical AND
	OpOr   FilterOp = "or"   // Logical OR
)

// FilterNode is a recursive AST node for in-memory record filtering.
type FilterNode struct {
	Op         FilterOp
	Field      string
	Value      any
	Conditions []FilterNode
}

func validateFilter(filter *FilterNode) error {
	if filter == nil {
		return nil
	}
	switch filter.Op {
	case OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpLike:
		return nil
	case OpAnd, OpOr:
		for i := range filter.Conditions {
			if err := validateFilter(&filter.Conditions[i]); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func evaluateFilter(filter *FilterNode, rec fieldvalue.FieldRecord) (bool, error) {
	switch filter.Op {
	case OpAnd:
		for _, c := range filter.Conditions {
			match, err := evaluateFilter(&c, rec)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
		}
		return true, nil
	case OpOr:
		for _, c := range filter.Conditions {
			match, err := evaluateFilter(&c, rec)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	case OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpLike:
		return evaluateLeafFilter(filter, rec)
	default:
		return false, fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func evaluateLeafFilter(filter *FilterNode, rec fieldvalue.FieldRecord) (bool, error) {
	var left any
	if filter.Field == "record_id" {
		left = rec.RecordId.String()
	} else {
		var ok bool
		left, ok = rec.Values[filter.Field]
		if !ok {
			// Field not present on record.
			switch filter.Op {
			case OpEq:
				return false, nil
			case OpNe:
				return true, nil
			default:
				return false, fmt.Errorf("field %q not found in record", filter.Field)
			}
		}
	}

	// Coerce filter value when the record value is time.Time and filter value is a string.
	right := filter.Value
	if lt, ok := left.(time.Time); ok {
		if rs, ok := right.(string); ok {
			if pt, err := time.Parse(time.RFC3339, rs); err == nil {
				right = pt
			} else {
				return false, fmt.Errorf("cannot parse %q as RFC3339 for field %q", rs, filter.Field)
			}
		} else {
			return false, fmt.Errorf("cannot compare time.Time with %T for field %q", right, filter.Field)
		}
		_ = lt
	}

	switch filter.Op {
	case OpEq:
		return compareEqual(left, right), nil
	case OpNe:
		return !compareEqual(left, right), nil
	case OpGt:
		return compareGreater(left, right)
	case OpGte:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return gt || compareEqual(left, right), nil
	case OpLt:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return !gt && !compareEqual(left, right), nil
	case OpLte:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return !gt, nil
	case OpLike:
		ls, lok := left.(string)
		rs, rok := right.(string)
		if !lok || !rok {
			return false, fmt.Errorf("like operator requires string values")
		}
		return strings.Contains(strings.ToLower(ls), strings.ToLower(rs)), nil
	default:
		return false, fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func compareEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case int64:
		switch bv := b.(type) {
		case int64:
			return av == bv
		case float64:
			return float64(av) == bv
		default:
			return false
		}
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int64:
			return av == float64(bv)
		default:
			return false
		}
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case time.Time:
		bv, ok := b.(time.Time)
		return ok && av.Equal(bv)
	default:
		return false
	}
}

func compareGreater(a, b any) (bool, error) {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		if !ok {
			return false, fmt.Errorf("cannot compare string with %T", b)
		}
		return av > bv, nil
	case int64:
		switch bv := b.(type) {
		case int64:
			return av > bv, nil
		case float64:
			return float64(av) > bv, nil
		default:
			return false, fmt.Errorf("cannot compare int64 with %T", b)
		}
	case float64:
		switch bv := b.(type) {
		case float64:
			return av > bv, nil
		case int64:
			return av > float64(bv), nil
		default:
			return false, fmt.Errorf("cannot compare float64 with %T", b)
		}
	case time.Time:
		bv, ok := b.(time.Time)
		if !ok {
			return false, fmt.Errorf("cannot compare time.Time with %T", b)
		}
		return av.After(bv), nil
	default:
		return false, fmt.Errorf("unsupported type for comparison: %T", a)
	}
}
