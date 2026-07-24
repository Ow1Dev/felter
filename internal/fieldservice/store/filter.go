package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
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
		return fmt.Errorf("%w: unknown filter operator: %s", ErrInvalidFilter, filter.Op)
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
			case OpEq, OpGt, OpGte, OpLt, OpLte, OpLike:
				return false, nil
			case OpNe:
				return true, nil
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

// filterToSQL converts a validated filter AST into a Postgres SQL subquery
// that returns matching record_ids, plus its bound arguments.
// The returned SQL uses $1 for schema_id; caller arguments must start with schemaID.
func filterToSQL(filter *FilterNode, fieldsMap map[string]schemaFieldMeta) (sql string, args []any, err error) {
	if filter == nil {
		return "", nil, nil
	}
	if err := validateFilter(filter); err != nil {
		return "", nil, err
	}
	b := &sqlBuilder{
		fieldsMap: fieldsMap,
		nextArg:   2, // $1 is reserved for schema_id
	}
	sql, err = b.build(filter)
	if err != nil {
		return "", nil, err
	}
	return sql, b.args, nil
}

type sqlBuilder struct {
	fieldsMap map[string]schemaFieldMeta
	args      []any
	nextArg   int
}

func (b *sqlBuilder) addArg(v any) string {
	arg := fmt.Sprintf("$%d", b.nextArg)
	b.args = append(b.args, v)
	b.nextArg++
	return arg
}

func (b *sqlBuilder) build(filter *FilterNode) (string, error) {
	switch filter.Op {
	case OpAnd:
		if len(filter.Conditions) == 0 {
			return `(SELECT record_id FROM field_values WHERE schema_id = $1 GROUP BY record_id)`, nil
		}
		parts := make([]string, len(filter.Conditions))
		for i, c := range filter.Conditions {
			p, err := b.build(&c)
			if err != nil {
				return "", err
			}
			parts[i] = "(" + p + ")"
		}
		return strings.Join(parts, "\nINTERSECT\n"), nil
	case OpOr:
		if len(filter.Conditions) == 0 {
			return `(SELECT record_id FROM field_values WHERE schema_id = $1 AND 1=0 GROUP BY record_id)`, nil
		}
		parts := make([]string, len(filter.Conditions))
		for i, c := range filter.Conditions {
			p, err := b.build(&c)
			if err != nil {
				return "", err
			}
			parts[i] = "(" + p + ")"
		}
		return strings.Join(parts, "\nUNION\n"), nil
	default:
		return b.buildLeaf(filter)
	}
}

func (b *sqlBuilder) buildLeaf(filter *FilterNode) (string, error) {
	if filter.Field == "record_id" {
		valStr := valueToString(filter.Value)
		if filter.Op != OpLike {
			if _, err := uuid.Parse(valStr); err != nil {
				return "", fmt.Errorf("%w: invalid record_id %q", ErrInvalidFilter, valStr)
			}
		}
		valArg := b.addArg(valStr)
		switch filter.Op {
		case OpEq:
			return fmt.Sprintf(`SELECT record_id FROM field_values WHERE schema_id = $1 AND record_id = %s GROUP BY record_id`, valArg), nil
		case OpNe:
			return fmt.Sprintf(`SELECT record_id FROM field_values WHERE schema_id = $1 AND record_id <> %s GROUP BY record_id`, valArg), nil
		case OpGt, OpGte, OpLt, OpLte:
			opSQL, err := leafOpSQL(filter.Op)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(`SELECT record_id FROM field_values WHERE schema_id = $1 AND record_id %s %s GROUP BY record_id`, opSQL, valArg), nil
		case OpLike:
			return fmt.Sprintf(`SELECT record_id FROM field_values WHERE schema_id = $1 AND POSITION(LOWER(%s) IN LOWER(record_id::text)) > 0 GROUP BY record_id`, valArg), nil
		default:
			return "", fmt.Errorf("unsupported leaf operator for SQL: %s", filter.Op)
		}
	}

	meta, ok := b.fieldsMap[filter.Field]
	if !ok {
		// Unknown field behaves as missing on every record (matches old client-side behaviour).
		if filter.Op == OpNe {
			return `(SELECT record_id FROM field_values WHERE schema_id = $1 GROUP BY record_id)`, nil
		}
		return `(SELECT record_id FROM field_values WHERE schema_id = $1 AND 1=0 GROUP BY record_id)`, nil
	}

	if filter.Op == OpLike && meta.Type != api.String {
		return "", fmt.Errorf("like operator requires string field")
	}
	if (filter.Op == OpGt || filter.Op == OpGte || filter.Op == OpLt || filter.Op == OpLte) && meta.Type == api.Boolean {
		return "", fmt.Errorf("cannot compare boolean with >, >=, <, <=")
	}

	fieldArg := b.addArg(filter.Field)
	valStr := valueToString(filter.Value)

	// For Ne we use an EXCEPT subquery that matches equality on the right side.
	// Like is handled separately in the type switch below.
	cmpOp := "="
	if filter.Op != OpNe && filter.Op != OpLike {
		var err error
		cmpOp, err = leafOpSQL(filter.Op)
		if err != nil {
			return "", err
		}
	}

	var cond string
	switch meta.Type {
	case api.Int:
		valArg := b.addArg(valStr)
		cond = fmt.Sprintf("CAST(value AS BIGINT) %s CAST(%s AS BIGINT)", cmpOp, valArg)
	case api.Float:
		valArg := b.addArg(valStr)
		cond = fmt.Sprintf("CAST(value AS DOUBLE PRECISION) %s CAST(%s AS DOUBLE PRECISION)", cmpOp, valArg)
	case api.Boolean:
		valArg := b.addArg(valStr)
		cond = fmt.Sprintf("CAST(value AS BOOLEAN) %s CAST(%s AS BOOLEAN)", cmpOp, valArg)
	case api.String:
		if filter.Op == OpLike {
			valArg := b.addArg(valStr)
			cond = fmt.Sprintf("POSITION(LOWER(%s) IN LOWER(value)) > 0", valArg)
		} else {
			valArg := b.addArg(valStr)
			cond = fmt.Sprintf("value %s %s", cmpOp, valArg)
		}
	case api.Date, api.Datetime:
		if filter.Op == OpLike {
			valArg := b.addArg(valStr)
			cond = fmt.Sprintf("POSITION(LOWER(%s) IN LOWER(value)) > 0", valArg)
		} else {
			valArg := b.addArg(valStr)
			cond = fmt.Sprintf("CAST(value AS TIMESTAMPTZ) %s CAST(%s AS TIMESTAMPTZ)", cmpOp, valArg)
		}
	default:
		return "", fmt.Errorf("unsupported field type for filter: %s", meta.Type)
	}

	if filter.Op == OpNe {
		return fmt.Sprintf(`(SELECT record_id FROM field_values WHERE schema_id = $1 GROUP BY record_id) EXCEPT (SELECT record_id FROM field_values WHERE schema_id = $1 AND field_key = %s AND %s GROUP BY record_id)`, fieldArg, cond), nil
	}

	return fmt.Sprintf(`SELECT record_id FROM field_values WHERE schema_id = $1 AND field_key = %s AND %s GROUP BY record_id`, fieldArg, cond), nil
}

func leafOpSQL(op FilterOp) (string, error) {
	switch op {
	case OpEq:
		return "=", nil
	case OpNe:
		return "<>", nil
	case OpGt:
		return ">", nil
	case OpGte:
		return ">=", nil
	case OpLt:
		return "<", nil
	case OpLte:
		return "<=", nil
	default:
		return "", fmt.Errorf("unsupported leaf operator for SQL: %s", op)
	}
}
