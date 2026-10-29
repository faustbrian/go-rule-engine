package ruleengine_test

import (
	"context"
	"math"
	"testing"
	"time"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

func TestSharedListStorageStaysImmutable(t *testing.T) {
	t.Parallel()

	input := []ruleengine.Value{ruleengine.String("a"), ruleengine.List(ruleengine.String("inner"))}
	outer := ruleengine.List(input...)
	literal := ruleengine.Literal(outer)
	path := ruleengine.MustPath("list")
	facts, err := ruleengine.NewContext(ruleengine.Fact{Path: path, Value: outer})
	if err != nil {
		t.Fatal(err)
	}

	// Every route by which a caller can reach list elements returns a copy.
	input[0] = ruleengine.String("mutated input")
	if items, _ := outer.ListValue(); len(items) == 2 {
		items[0] = ruleengine.String("mutated ListValue")
		nested, _ := items[1].ListValue()
		nested[0] = ruleengine.String("mutated nested")
	}
	if items, ok := outer.Interface().([]ruleengine.Value); ok {
		items[0] = ruleengine.String("mutated Interface")
	}
	if items, _ := facts.Lookup(path).ListValue(); len(items) == 2 {
		items[0] = ruleengine.String("mutated Lookup")
	}

	want := ruleengine.List(ruleengine.String("a"), ruleengine.List(ruleengine.String("inner")))
	for name, got := range map[string]ruleengine.Value{"value": outer, "lookup": facts.Lookup(path)} {
		if !got.Equal(want) {
			t.Fatalf("%s list changed: %#v", name, got.Interface())
		}
	}
	matched, err := ruleengine.Compare(ruleengine.OpEqual, literal, ruleengine.Literal(want)).Evaluate(context.Background(), facts)
	if err != nil || !matched {
		t.Fatalf("literal list changed: %v, %v", matched, err)
	}
	inner, _ := want.ListValue()
	if contains, err := ruleengine.Compare(ruleengine.OpContains, literal, ruleengine.Literal(inner[1])).Evaluate(context.Background(), facts); err != nil || !contains {
		t.Fatalf("nested list membership = %v, %v", contains, err)
	}
}

func TestListMembershipDoesNotAllocate(t *testing.T) {
	list, facts := benchmarkList(t)
	ctx := context.Background()
	code := ruleengine.Variable(ruleengine.MustPath("code"))
	last := ruleengine.Variable(ruleengine.MustPath("last"))
	codes := ruleengine.Variable(ruleengine.MustPath("codes"))
	// Each case scans the whole 256-entry list; positive controls prove the
	// facts resolved, since a Missing operand would also report no match.
	tests := []struct {
		name      string
		predicate ruleengine.Predicate
		want      bool
	}{
		{"in literal absent", ruleengine.Compare(ruleengine.OpIn, code, ruleengine.Literal(list)), false},
		{"in literal last", ruleengine.Compare(ruleengine.OpIn, last, ruleengine.Literal(list)), true},
		{"not_in literal", ruleengine.Compare(ruleengine.OpNotIn, code, ruleengine.Literal(list)), true},
		{"contains fact absent", ruleengine.Compare(ruleengine.OpContains, codes, ruleengine.Literal(ruleengine.String("absent"))), false},
		{"contains fact last", ruleengine.Compare(ruleengine.OpContains, codes, last), true},
		{"list equality", ruleengine.Compare(ruleengine.OpEqual, codes, ruleengine.Literal(list)), true},
	}
	for _, test := range tests {
		allocations := testing.AllocsPerRun(100, func() {
			matched, err := test.predicate.Evaluate(ctx, facts)
			if err != nil || matched != test.want {
				t.Fatalf("%s = %v, %v; want %v", test.name, matched, err, test.want)
			}
		})
		if allocations != 0 {
			t.Errorf("%s allocations = %v, want 0", test.name, allocations)
		}
	}
	copied := ruleengine.List(mustListValue(t, list)...)
	if allocations := testing.AllocsPerRun(100, func() {
		if !list.Equal(copied) {
			t.Fatal("Equal(copied list) = false")
		}
	}); allocations != 0 {
		t.Errorf("Equal allocations = %v, want 0", allocations)
	}
}

func TestValueEqual(t *testing.T) {
	t.Parallel()

	decimal := func(text string) ruleengine.Value {
		value, err := ruleengine.Decimal(text)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	instant := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	tests := []struct {
		name        string
		left, right ruleengine.Value
		want        bool
	}{
		{"missing", ruleengine.Missing(), ruleengine.Missing(), true},
		{"null", ruleengine.Null(), ruleengine.Null(), true},
		{"missing is not null", ruleengine.Missing(), ruleengine.Null(), false},
		{"bool", ruleengine.Bool(true), ruleengine.Bool(true), true},
		{"bool differs", ruleengine.Bool(true), ruleengine.Bool(false), false},
		{"int", ruleengine.Int(3), ruleengine.Int(3), true},
		{"int is not float", ruleengine.Int(1), ruleengine.Float(1), false},
		{"float", ruleengine.Float(1.5), ruleengine.Float(1.5), true},
		{"NaN", ruleengine.Float(math.NaN()), ruleengine.Float(math.NaN()), false},
		{"string", ruleengine.String("x"), ruleengine.String("x"), true},
		{"string is not decimal", ruleengine.String("1"), decimal("1"), false},
		{"time instant across zones", ruleengine.Time(instant), ruleengine.Time(instant.In(time.FixedZone("x", 7200))), true},
		{"duration", ruleengine.Duration(time.Second), ruleengine.Duration(time.Second), true},
		{"decimal value not spelling", decimal("1.50"), decimal("1.5e0"), true},
		{"decimal differs", decimal("1.5"), decimal("1.51"), false},
		{"list", ruleengine.List(ruleengine.Int(1), ruleengine.List(decimal("2"))), ruleengine.List(ruleengine.Int(1), ruleengine.List(decimal("2.0"))), true},
		{"list order", ruleengine.List(ruleengine.Int(1), ruleengine.Int(2)), ruleengine.List(ruleengine.Int(2), ruleengine.Int(1)), false},
		{"list length", ruleengine.List(ruleengine.Int(1)), ruleengine.List(ruleengine.Int(1), ruleengine.Int(1)), false},
		{"empty lists", ruleengine.List(), ruleengine.List(), true},
	}
	for _, test := range tests {
		if got := test.left.Equal(test.right); got != test.want {
			t.Errorf("%s: Equal = %v, want %v", test.name, got, test.want)
		}
		if got := test.right.Equal(test.left); got != test.want {
			t.Errorf("%s: reversed Equal = %v, want %v", test.name, got, test.want)
		}
		// Equal agrees with OpEqual wherever the operator accepts the operands.
		if test.left.Kind() == test.right.Kind() && test.left.Kind() != ruleengine.KindMissing {
			matched, err := ruleengine.Compare(ruleengine.OpEqual, ruleengine.Literal(test.left), ruleengine.Literal(test.right)).Evaluate(context.Background(), ruleengine.Context{})
			if err != nil || matched != test.want {
				t.Errorf("%s: OpEqual = %v, %v; want %v", test.name, matched, err, test.want)
			}
		}
	}
}

func mustListValue(t *testing.T, value ruleengine.Value) []ruleengine.Value {
	t.Helper()
	items, ok := value.ListValue()
	if !ok {
		t.Fatal("ListValue() did not return a list")
	}
	return items
}

func TestListAccessors(t *testing.T) {
	t.Parallel()

	nested := ruleengine.List(ruleengine.Int(2))
	list := ruleengine.List(ruleengine.Int(1), nested)
	if length, ok := list.ListLen(); !ok || length != 2 {
		t.Fatalf("ListLen() = %d, %v", length, ok)
	}
	if first, ok := list.ListAt(0); !ok || !first.Equal(ruleengine.Int(1)) {
		t.Fatalf("ListAt(0) = %#v, %v", first.Interface(), ok)
	}
	if second, ok := list.ListAt(1); !ok || !second.Equal(nested) {
		t.Fatalf("ListAt(1) = %#v, %v", second.Interface(), ok)
	}
	for _, index := range []int{-1, 2, int(^uint(0) >> 1)} {
		if value, ok := list.ListAt(index); ok || value.Kind() != ruleengine.KindMissing {
			t.Errorf("ListAt(%d) = %v, %v; want Missing, false", index, value.Kind(), ok)
		}
	}
	if length, ok := ruleengine.List().ListLen(); !ok || length != 0 {
		t.Fatalf("empty ListLen() = %d, %v", length, ok)
	}
	if _, ok := ruleengine.List().ListAt(0); ok {
		t.Fatal("empty ListAt(0) reported an element")
	}
	for _, other := range []ruleengine.Value{ruleengine.Missing(), ruleengine.Null(), ruleengine.String("ab"), ruleengine.Int(1)} {
		if length, ok := other.ListLen(); ok || length != 0 {
			t.Errorf("%v ListLen() = %d, %v", other.Kind(), length, ok)
		}
		if _, ok := other.ListAt(0); ok {
			t.Errorf("%v ListAt(0) reported an element", other.Kind())
		}
	}
	// Mutating a ListValue copy is not visible through the accessors.
	items := mustListValue(t, list)
	items[0] = ruleengine.String("mutated")
	if first, _ := list.ListAt(0); !first.Equal(ruleengine.Int(1)) {
		t.Fatalf("ListAt(0) after copy mutation = %#v", first.Interface())
	}
}

// listMemberOperator is a custom membership operator that reads its list
// operand through the non-allocating accessors, as Checkout's operators do.
type listMemberOperator struct{}

func (listMemberOperator) Name() ruleengine.OperatorName { return "list_member" }
func (listMemberOperator) Signatures() []ruleengine.Signature {
	return []ruleengine.Signature{{Left: ruleengine.KindString, Right: ruleengine.KindList}}
}
func (listMemberOperator) Evaluate(_ context.Context, left, right ruleengine.Value) (bool, error) {
	length, _ := right.ListLen()
	for index := range length {
		if item, _ := right.ListAt(index); item.Equal(left) {
			return true, nil
		}
	}
	return false, nil
}

func listMemberPlan(tb testing.TB, list ruleengine.Value, probe string) ruleengine.Plan {
	tb.Helper()
	compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), listMemberOperator{})
	if err != nil {
		tb.Fatal(err)
	}
	set := ruleengine.RuleSet{ID: "custom", Rules: []ruleengine.Rule{{ID: "member", When: ruleengine.Compare("list_member", ruleengine.Variable(ruleengine.MustPath(probe)), ruleengine.Literal(list))}}}
	plan, _, err := compiler.Compile(context.Background(), set)
	if err != nil {
		tb.Fatal(err)
	}
	return plan
}

func TestCustomOperatorListAccessDoesNotAllocate(t *testing.T) {
	list, facts := benchmarkList(t)
	ctx := context.Background()
	probe := facts.Lookup(ruleengine.MustPath("last"))
	// The operator body over a 256-entry list allocates nothing.
	if allocations := testing.AllocsPerRun(100, func() {
		if matched, err := (listMemberOperator{}).Evaluate(ctx, probe, list); err != nil || !matched {
			t.Fatalf("Evaluate(last) = %v, %v", matched, err)
		}
	}); allocations != 0 {
		t.Errorf("custom operator allocations = %v, want 0", allocations)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		length, _ := list.ListLen()
		for index := range length {
			_, _ = list.ListAt(index)
		}
	}); allocations != 0 {
		t.Errorf("ListLen/ListAt allocations = %v, want 0", allocations)
	}
	// Through a compiled plan, list size adds nothing beyond the fixed cost of
	// evaluating a one-rule plan.
	for _, test := range []struct {
		probe   string
		outcome ruleengine.Decision
	}{{"last", ruleengine.Matched}, {"code", ruleengine.Unmatched}} {
		plan := listMemberPlan(t, list, test.probe)
		small := listMemberPlan(t, ruleengine.List(ruleengine.String("code-255")), test.probe)
		measure := func(plan ruleengine.Plan) float64 {
			return testing.AllocsPerRun(100, func() {
				if result := plan.Evaluate(ctx, facts); result.Decision != test.outcome {
					t.Fatalf("Evaluate(%s) = %#v", test.probe, result)
				}
			})
		}
		if large, oneEntry := measure(plan), measure(small); large != oneEntry {
			t.Errorf("%s: 256-entry plan allocations = %v, one-entry = %v", test.probe, large, oneEntry)
		}
	}
}
