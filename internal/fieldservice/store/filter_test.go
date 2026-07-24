package store

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
)

func makeRecord(values map[string]any) fieldvalue.FieldRecord {
	return fieldvalue.FieldRecord{
		RecordId:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		SchemaKey: "test",
		Values:    values,
	}
}

func TestValidateFilter_Nil(t *testing.T) {
	if err := validateFilter(nil); err != nil {
		t.Fatalf("expected nil filter to be valid, got %v", err)
	}
}

func TestValidateFilter_LeafOps(t *testing.T) {
	ops := []FilterOp{OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpLike}
	for _, op := range ops {
		f := &FilterNode{Op: op, Field: "x", Value: "y"}
		if err := validateFilter(f); err != nil {
			t.Fatalf("op %q should be valid: %v", op, err)
		}
	}
}

func TestValidateFilter_InvalidOp(t *testing.T) {
	f := &FilterNode{Op: "invalid", Field: "x", Value: "y"}
	if err := validateFilter(f); err == nil {
		t.Fatal("expected error for invalid op")
	}
}

func TestValidateFilter_NestedAnd(t *testing.T) {
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "a", Value: 1},
			{Op: OpGt, Field: "b", Value: 2},
		},
	}
	if err := validateFilter(f); err != nil {
		t.Fatalf("nested and should be valid: %v", err)
	}
}

func TestValidateFilter_NestedInvalid(t *testing.T) {
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "a", Value: 1},
			{Op: "bad", Field: "b", Value: 2},
		},
	}
	if err := validateFilter(f); err == nil {
		t.Fatal("expected error for nested invalid op")
	}
}

func TestEvaluateFilter_Eq(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{Op: OpEq, Field: "name", Value: "alice"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_Eq_Mismatch(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{Op: OpEq, Field: "name", Value: "bob"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatal("expected no match")
	}
}

func TestEvaluateFilter_Ne(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{Op: OpNe, Field: "name", Value: "bob"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_Gt(t *testing.T) {
	rec := makeRecord(map[string]any{"score": int64(10)})
	f := &FilterNode{Op: OpGt, Field: "score", Value: float64(5)}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_Gte(t *testing.T) {
	rec := makeRecord(map[string]any{"score": int64(10)})
	f := &FilterNode{Op: OpGte, Field: "score", Value: int64(10)}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for equal value with gte")
	}
}

func TestEvaluateFilter_Lt(t *testing.T) {
	rec := makeRecord(map[string]any{"score": int64(3)})
	f := &FilterNode{Op: OpLt, Field: "score", Value: float64(5)}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_Lte(t *testing.T) {
	rec := makeRecord(map[string]any{"score": int64(5)})
	f := &FilterNode{Op: OpLte, Field: "score", Value: int64(5)}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for equal value with lte")
	}
}

func TestEvaluateFilter_Like(t *testing.T) {
	rec := makeRecord(map[string]any{"title": "Hello World"})
	f := &FilterNode{Op: OpLike, Field: "title", Value: "world"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for case-insensitive like")
	}
}

func TestEvaluateFilter_Like_NonString(t *testing.T) {
	rec := makeRecord(map[string]any{"score": int64(5)})
	f := &FilterNode{Op: OpLike, Field: "score", Value: "5"}
	_, err := evaluateFilter(f, rec)
	if err == nil {
		t.Fatal("expected error for like on non-string")
	}
}

func TestEvaluateFilter_And(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice", "score": int64(10)})
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "name", Value: "alice"},
			{Op: OpGt, Field: "score", Value: float64(5)},
		},
	}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_And_ShortCircuit(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "name", Value: "bob"},
			{Op: OpGt, Field: "score", Value: float64(5)},
		},
	}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatal("expected no match")
	}
}

func TestEvaluateFilter_Or(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{
		Op: OpOr,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "name", Value: "bob"},
			{Op: OpEq, Field: "name", Value: "alice"},
		},
	}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
}

func TestEvaluateFilter_Or_ShortCircuit(t *testing.T) {
	rec := makeRecord(map[string]any{"name": "alice"})
	f := &FilterNode{
		Op: OpOr,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "name", Value: "bob"},
			{Op: OpEq, Field: "name", Value: "charlie"},
		},
	}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatal("expected no match")
	}
}

func TestEvaluateFilter_RecordID(t *testing.T) {
	rec := makeRecord(map[string]any{})
	f := &FilterNode{Op: OpEq, Field: "record_id", Value: rec.RecordId.String()}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for record_id")
	}
}

func TestEvaluateFilter_MissingField_Eq(t *testing.T) {
	rec := makeRecord(map[string]any{})
	f := &FilterNode{Op: OpEq, Field: "missing", Value: "x"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatal("expected no match for missing field with eq")
	}
}

func TestEvaluateFilter_MissingField_Ne(t *testing.T) {
	rec := makeRecord(map[string]any{})
	f := &FilterNode{Op: OpNe, Field: "missing", Value: "x"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for missing field with ne")
	}
}

func TestEvaluateFilter_MissingField_Gt(t *testing.T) {
	rec := makeRecord(map[string]any{})
	f := &FilterNode{Op: OpGt, Field: "missing", Value: float64(1)}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatal("expected no match for missing field with gt")
	}
}

func TestEvaluateFilter_TimeCoercion(t *testing.T) {
	ts := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	rec := makeRecord(map[string]any{"created": ts})
	f := &FilterNode{Op: OpEq, Field: "created", Value: "2026-07-15T00:00:00Z"}
	match, err := evaluateFilter(f, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatal("expected match for time coercion")
	}
}

func TestEvaluateFilter_TimeCoercion_Invalid(t *testing.T) {
	ts := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	rec := makeRecord(map[string]any{"created": ts})
	f := &FilterNode{Op: OpEq, Field: "created", Value: "not-a-date"}
	_, err := evaluateFilter(f, rec)
	if err == nil {
		t.Fatal("expected error for invalid time string")
	}
}

func TestCompareEqual_String(t *testing.T) {
	if !compareEqual("a", "a") {
		t.Fatal("expected equal strings")
	}
	if compareEqual("a", "b") {
		t.Fatal("expected unequal strings")
	}
}

func TestCompareEqual_Int64(t *testing.T) {
	if !compareEqual(int64(5), int64(5)) {
		t.Fatal("expected equal int64")
	}
	if compareEqual(int64(5), int64(6)) {
		t.Fatal("expected unequal int64")
	}
}

func TestCompareEqual_Int64_Float64(t *testing.T) {
	if !compareEqual(int64(5), float64(5.0)) {
		t.Fatal("expected int64 == float64")
	}
	if !compareEqual(float64(5.0), int64(5)) {
		t.Fatal("expected float64 == int64")
	}
}

func TestCompareEqual_Bool(t *testing.T) {
	if !compareEqual(true, true) {
		t.Fatal("expected true == true")
	}
	if compareEqual(true, false) {
		t.Fatal("expected true != false")
	}
}

func TestCompareEqual_Time(t *testing.T) {
	a := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if !compareEqual(a, b) {
		t.Fatal("expected equal times")
	}
}

func TestCompareGreater_String(t *testing.T) {
	gt, err := compareGreater("b", "a")
	if err != nil || !gt {
		t.Fatal("expected b > a")
	}
}

func TestCompareGreater_Int64(t *testing.T) {
	gt, err := compareGreater(int64(10), int64(5))
	if err != nil || !gt {
		t.Fatal("expected 10 > 5")
	}
}

func TestCompareGreater_Int64_Float64(t *testing.T) {
	gt, err := compareGreater(int64(10), float64(5.0))
	if err != nil || !gt {
		t.Fatal("expected int64 10 > float64 5")
	}
}

func TestCompareGreater_Time(t *testing.T) {
	a := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	gt, err := compareGreater(a, b)
	if err != nil || !gt {
		t.Fatal("expected later time > earlier time")
	}
}

func TestCompareGreater_Unsupported(t *testing.T) {
	_, err := compareGreater([]int{1}, []int{2})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestFilterToSQL_Nil(t *testing.T) {
	sql, args, err := filterToSQL(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "" {
		t.Fatalf("expected empty sql, got %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestFilterToSQL_EqString(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "name", Value: "alice"}
	fields := map[string]schemaFieldMeta{"name": {Key: "name", Type: api.String}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "field_key = $2") || !strings.Contains(sql, "value = $3") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "name" || args[1] != "alice" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_EqInt(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "score", Value: float64(42)}
	fields := map[string]schemaFieldMeta{"score": {Key: "score", Type: api.Int}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS BIGINT) = CAST($3 AS BIGINT)") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "score" || args[1] != "42" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_EqFloat(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "rating", Value: float64(3.14)}
	fields := map[string]schemaFieldMeta{"rating": {Key: "rating", Type: api.Float}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS DOUBLE PRECISION) = CAST($3 AS DOUBLE PRECISION)") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "rating" || args[1] != "3.14" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_EqBoolean(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "active", Value: true}
	fields := map[string]schemaFieldMeta{"active": {Key: "active", Type: api.Boolean}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS BOOLEAN) = CAST($3 AS BOOLEAN)") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "active" || args[1] != "true" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_GtInt(t *testing.T) {
	f := &FilterNode{Op: OpGt, Field: "score", Value: float64(10)}
	fields := map[string]schemaFieldMeta{"score": {Key: "score", Type: api.Int}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS BIGINT) > CAST($3 AS BIGINT)") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "score" || args[1] != "10" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_GtBooleanError(t *testing.T) {
	f := &FilterNode{Op: OpGt, Field: "active", Value: true}
	fields := map[string]schemaFieldMeta{"active": {Key: "active", Type: api.Boolean}}
	_, _, err := filterToSQL(f, fields)
	if err == nil {
		t.Fatal("expected error for gt on boolean")
	}
}

func TestFilterToSQL_LikeString(t *testing.T) {
	f := &FilterNode{Op: OpLike, Field: "title", Value: "world"}
	fields := map[string]schemaFieldMeta{"title": {Key: "title", Type: api.String}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "POSITION(LOWER($3) IN LOWER(value)) > 0") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "title" || args[1] != "world" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_LikeIntError(t *testing.T) {
	f := &FilterNode{Op: OpLike, Field: "score", Value: "42"}
	fields := map[string]schemaFieldMeta{"score": {Key: "score", Type: api.Int}}
	_, _, err := filterToSQL(f, fields)
	if err == nil {
		t.Fatal("expected error for like on int field")
	}
}

func TestFilterToSQL_NeString(t *testing.T) {
	f := &FilterNode{Op: OpNe, Field: "name", Value: "alice"}
	fields := map[string]schemaFieldMeta{"name": {Key: "name", Type: api.String}}
	sql, args, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "EXCEPT") || !strings.Contains(sql, "value = $3") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != "name" || args[1] != "alice" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_RecordID(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "record_id", Value: "11111111-1111-1111-1111-111111111111"}
	sql, args, err := filterToSQL(f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "record_id = $2") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_RecordIDNe(t *testing.T) {
	f := &FilterNode{Op: OpNe, Field: "record_id", Value: "11111111-1111-1111-1111-111111111111"}
	sql, args, err := filterToSQL(f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "record_id <> $2") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_RecordIDLike(t *testing.T) {
	f := &FilterNode{Op: OpLike, Field: "record_id", Value: "1111"}
	sql, args, err := filterToSQL(f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "POSITION(LOWER($2) IN LOWER(record_id::text)) > 0") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "1111" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestFilterToSQL_UnknownFieldEq(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "missing", Value: "x"}
	sql, _, err := filterToSQL(f, map[string]schemaFieldMeta{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "1=0") {
		t.Fatalf("expected empty result sql, got: %s", sql)
	}
}

func TestFilterToSQL_UnknownFieldNe(t *testing.T) {
	f := &FilterNode{Op: OpNe, Field: "missing", Value: "x"}
	sql, _, err := filterToSQL(f, map[string]schemaFieldMeta{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(sql, "1=0") {
		t.Fatalf("expected all-records sql, got: %s", sql)
	}
	if !strings.Contains(sql, "GROUP BY record_id") {
		t.Fatalf("expected record_id selection, got: %s", sql)
	}
}

func TestFilterToSQL_And(t *testing.T) {
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "a", Value: "1"},
			{Op: OpGt, Field: "b", Value: float64(2)},
		},
	}
	fields := map[string]schemaFieldMeta{
		"a": {Key: "a", Type: api.String},
		"b": {Key: "b", Type: api.Int},
	}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "INTERSECT") {
		t.Fatalf("expected INTERSECT, got: %s", sql)
	}
}

func TestFilterToSQL_Or(t *testing.T) {
	f := &FilterNode{
		Op: OpOr,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "a", Value: "1"},
			{Op: OpEq, Field: "b", Value: "2"},
		},
	}
	fields := map[string]schemaFieldMeta{
		"a": {Key: "a", Type: api.String},
		"b": {Key: "b", Type: api.String},
	}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "UNION") {
		t.Fatalf("expected UNION, got: %s", sql)
	}
}

func TestFilterToSQL_EmptyAnd(t *testing.T) {
	f := &FilterNode{Op: OpAnd, Conditions: []FilterNode{}}
	sql, _, err := filterToSQL(f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "GROUP BY record_id") || strings.Contains(sql, "1=0") {
		t.Fatalf("expected all-records sql, got: %s", sql)
	}
}

func TestFilterToSQL_EmptyOr(t *testing.T) {
	f := &FilterNode{Op: OpOr, Conditions: []FilterNode{}}
	sql, _, err := filterToSQL(f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "1=0") {
		t.Fatalf("expected empty-result sql, got: %s", sql)
	}
}

func TestFilterToSQL_NestedAndOr(t *testing.T) {
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{
				Op: OpOr,
				Conditions: []FilterNode{
					{Op: OpEq, Field: "a", Value: "1"},
					{Op: OpEq, Field: "b", Value: "2"},
				},
			},
			{Op: OpEq, Field: "c", Value: "3"},
		},
	}
	fields := map[string]schemaFieldMeta{
		"a": {Key: "a", Type: api.String},
		"b": {Key: "b", Type: api.String},
		"c": {Key: "c", Type: api.String},
	}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "UNION") {
		t.Fatalf("expected UNION in nested OR branch, got: %s", sql)
	}
	if !strings.Contains(sql, "INTERSECT") {
		t.Fatalf("expected INTERSECT for top-level AND, got: %s", sql)
	}
	// The nested OR branch must be parenthesized before INTERSECT.
	unionIdx := strings.Index(sql, "UNION")
	intersectIdx := strings.Index(sql, "INTERSECT")
	if unionIdx == -1 || intersectIdx == -1 {
		t.Fatalf("expected both UNION and INTERSECT, got: %s", sql)
	}
	if unionIdx >= intersectIdx {
		t.Fatalf("expected UNION before INTERSECT, got: %s", sql)
	}
	// UNION appears before INTERSECT in the joined output, so we expect
	// the OR block (with UNION) to be wrapped in parentheses.
	orBlock := sql[:intersectIdx]
	if !strings.HasPrefix(strings.TrimSpace(orBlock), "(") || !strings.HasSuffix(strings.TrimSpace(orBlock), ")") {
		t.Fatalf("expected nested OR branch to be parenthesized, got: %s", orBlock)
	}
}

func TestFilterToSQL_NestedAndWithNeEq(t *testing.T) {
	f := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpNe, Field: "a", Value: "1"},
			{Op: OpEq, Field: "b", Value: "2"},
		},
	}
	fields := map[string]schemaFieldMeta{
		"a": {Key: "a", Type: api.String},
		"b": {Key: "b", Type: api.String},
	}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "EXCEPT") {
		t.Fatalf("expected EXCEPT for Ne condition, got: %s", sql)
	}
	if !strings.Contains(sql, "INTERSECT") {
		t.Fatalf("expected INTERSECT for top-level And, got: %s", sql)
	}
	// The Ne branch must be parenthesized before INTERSECT.
	exceptIdx := strings.Index(sql, "EXCEPT")
	intersectIdx := strings.Index(sql, "INTERSECT")
	if exceptIdx == -1 || intersectIdx == -1 {
		t.Fatalf("expected both EXCEPT and INTERSECT, got: %s", sql)
	}
	if exceptIdx >= intersectIdx {
		t.Fatalf("expected EXCEPT before INTERSECT, got: %s", sql)
	}
	// EXCEPT appears before INTERSECT in the joined output, so we expect
	// the Ne block (with EXCEPT) to be wrapped in parentheses.
	neBlock := sql[:intersectIdx]
	if !strings.HasPrefix(strings.TrimSpace(neBlock), "(") || !strings.HasSuffix(strings.TrimSpace(neBlock), ")") {
		t.Fatalf("expected Ne branch to be parenthesized, got: %s", neBlock)
	}
}

func TestFilterToSQL_InvalidRecordID(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "record_id", Value: "not-a-uuid"}
	_, _, err := filterToSQL(f, nil)
	if err == nil {
		t.Fatal("expected error for invalid record_id")
	}
}

func TestFilterToSQL_DateCast(t *testing.T) {
	f := &FilterNode{Op: OpGt, Field: "due", Value: "2026-07-15T00:00:00Z"}
	fields := map[string]schemaFieldMeta{"due": {Key: "due", Type: api.Date}}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS TIMESTAMPTZ)") || !strings.Contains(sql, "CAST($3 AS TIMESTAMPTZ)") {
		t.Fatalf("expected TIMESTAMPTZ casts for date field, got: %s", sql)
	}
}

func TestFilterToSQL_DatetimeCast(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "created", Value: "2026-07-15T12:30:00Z"}
	fields := map[string]schemaFieldMeta{"created": {Key: "created", Type: api.Datetime}}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "CAST(value AS TIMESTAMPTZ)") || !strings.Contains(sql, "CAST($3 AS TIMESTAMPTZ)") {
		t.Fatalf("expected TIMESTAMPTZ casts for datetime field, got: %s", sql)
	}
}

func TestFilterToSQL_StringNoCast(t *testing.T) {
	f := &FilterNode{Op: OpEq, Field: "name", Value: "alice"}
	fields := map[string]schemaFieldMeta{"name": {Key: "name", Type: api.String}}
	sql, _, err := filterToSQL(f, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(sql, "CAST(value AS TIMESTAMPTZ)") {
		t.Fatalf("expected no TIMESTAMPTZ cast for string field, got: %s", sql)
	}
	if !strings.Contains(sql, "value = $3") {
		t.Fatalf("expected plain string comparison, got: %s", sql)
	}
}
