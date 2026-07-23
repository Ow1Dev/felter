package store

import (
	"testing"
	"time"

	"github.com/google/uuid"

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
	_, err := evaluateFilter(f, rec)
	if err == nil {
		t.Fatal("expected error for missing field with gt")
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
