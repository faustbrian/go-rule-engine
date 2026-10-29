package ruleengine_test

import (
	"context"
	"strconv"
	"testing"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

// BenchmarkCompile includes validation, copying and ordering of a three-rule
// fixture; compiler and input construction are outside timing.
func BenchmarkCompile(b *testing.B) {
	set, _ := benchmarkFixture()
	compiler := ruleengine.NewCompiler(ruleengine.DefaultLimits())
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, _, err := compiler.Compile(ctx, set); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEvaluate measures a precompiled three-rule decision over a prepared
// fact snapshot; compilation and fact construction are outside timing.
func BenchmarkEvaluate(b *testing.B) {
	set, facts := benchmarkFixture()
	plan, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).Compile(context.Background(), set)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if result := plan.Evaluate(ctx, facts); result.Decision != ruleengine.Matched {
			b.Fatalf("Evaluate() = %#v", result)
		}
	}
}

func benchmarkFixture() (ruleengine.RuleSet, ruleengine.Context) {
	country := ruleengine.MustPath("shipment", "country")
	weight := ruleengine.MustPath("shipment", "weight_grams")
	tags := ruleengine.MustPath("shipment", "tags")
	set := ruleengine.RuleSet{ID: "benchmark", Strategy: ruleengine.CollectAll, Rules: []ruleengine.Rule{
		{ID: "country", Priority: 30, When: ruleengine.Compare(ruleengine.OpEqual, ruleengine.Variable(country), ruleengine.Literal(ruleengine.String("FI")))},
		{ID: "weight", Priority: 20, When: ruleengine.Compare(ruleengine.OpGreaterOrEqual, ruleengine.Variable(weight), ruleengine.Literal(ruleengine.Int(1_000)))},
		{ID: "tag", Priority: 10, When: ruleengine.Compare(ruleengine.OpContains, ruleengine.Variable(tags), ruleengine.Literal(ruleengine.String("express")))},
	}}
	facts, _ := ruleengine.NewContext(
		ruleengine.Fact{Path: country, Value: ruleengine.String("FI")},
		ruleengine.Fact{Path: weight, Value: ruleengine.Int(1_500)},
		ruleengine.Fact{Path: tags, Value: ruleengine.List(ruleengine.String("express"), ruleengine.String("fragile"))},
	)
	return set, facts
}

// benchmarkList returns a 256-entry string list, the size of Rate's largest
// membership sets, and a fact snapshot holding it and an absent probe value.
func benchmarkList(tb testing.TB) (ruleengine.Value, ruleengine.Context) {
	tb.Helper()
	items := make([]ruleengine.Value, 256)
	for index := range items {
		items[index] = ruleengine.String("code-" + strconv.Itoa(index))
	}
	list := ruleengine.List(items...)
	facts, err := ruleengine.NewContext(
		ruleengine.Fact{Path: ruleengine.MustPath("codes"), Value: list},
		ruleengine.Fact{Path: ruleengine.MustPath("code"), Value: ruleengine.String("absent")},
		ruleengine.Fact{Path: ruleengine.MustPath("last"), Value: ruleengine.String("code-255")},
	)
	if err != nil {
		tb.Fatal(err)
	}
	return list, facts
}

// BenchmarkInLiteralList256 scans a 256-entry literal list for an absent fact
// value, so the operator visits every element and resolves the literal.
func BenchmarkInLiteralList256(b *testing.B) {
	list, facts := benchmarkList(b)
	predicate := ruleengine.Compare(ruleengine.OpIn, ruleengine.Variable(ruleengine.MustPath("code")), ruleengine.Literal(list))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if matched, err := predicate.Evaluate(ctx, facts); err != nil || matched {
			b.Fatalf("Evaluate() = %v, %v", matched, err)
		}
	}
}

// BenchmarkContainsFactList256 scans a 256-entry list fact, which is looked up
// from the context on every evaluation.
func BenchmarkContainsFactList256(b *testing.B) {
	_, facts := benchmarkList(b)
	predicate := ruleengine.Compare(ruleengine.OpContains, ruleengine.Variable(ruleengine.MustPath("codes")), ruleengine.Literal(ruleengine.String("absent")))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if matched, err := predicate.Evaluate(ctx, facts); err != nil || matched {
			b.Fatalf("Evaluate() = %v, %v", matched, err)
		}
	}
}

// BenchmarkLiteralListResolve256 resolves a 256-entry literal through an
// equality against itself, isolating literal resolution and list equality.
func BenchmarkLiteralListResolve256(b *testing.B) {
	list, facts := benchmarkList(b)
	predicate := ruleengine.Compare(ruleengine.OpEqual, ruleengine.Literal(list), ruleengine.Literal(list))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if matched, err := predicate.Evaluate(ctx, facts); err != nil || !matched {
			b.Fatalf("Evaluate() = %v, %v", matched, err)
		}
	}
}

// literalSink keeps constructed operands live so the compiler cannot elide
// the construction being measured.
var literalSink ruleengine.Operand

// BenchmarkLiteralList256 measures constructing a literal operand from an
// existing 256-entry list value.
func BenchmarkLiteralList256(b *testing.B) {
	list, _ := benchmarkList(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		literalSink = ruleengine.Literal(list)
	}
}

// BenchmarkPlanInList256 evaluates a compiled one-rule plan whose proposition
// tests membership in a 256-entry literal list.
func BenchmarkPlanInList256(b *testing.B) {
	list, facts := benchmarkList(b)
	set := ruleengine.RuleSet{ID: "membership", Rules: []ruleengine.Rule{{ID: "in", When: ruleengine.Compare(ruleengine.OpIn, ruleengine.Variable(ruleengine.MustPath("code")), ruleengine.Literal(list))}}}
	plan, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).Compile(context.Background(), set)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if result := plan.Evaluate(ctx, facts); result.Decision != ruleengine.Unmatched {
			b.Fatalf("Evaluate() = %#v", result)
		}
	}
}

// BenchmarkCustomOperatorList256 evaluates a compiled plan whose custom
// operator scans a 256-entry literal list with ListLen and ListAt.
func BenchmarkCustomOperatorList256(b *testing.B) {
	list, facts := benchmarkList(b)
	plan := listMemberPlan(b, list, "code")
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if result := plan.Evaluate(ctx, facts); result.Decision != ruleengine.Unmatched {
			b.Fatalf("Evaluate() = %#v", result)
		}
	}
}

// listValueMemberOperator is the copying ListValue variant kept for comparison.
type listValueMemberOperator struct{ listMemberOperator }

func (listValueMemberOperator) Evaluate(_ context.Context, left, right ruleengine.Value) (bool, error) {
	items, _ := right.ListValue()
	for _, item := range items {
		if item.Equal(left) {
			return true, nil
		}
	}
	return false, nil
}

// BenchmarkCustomOperatorListValue256 is the same scan through ListValue,
// which copies the list on every call.
func BenchmarkCustomOperatorListValue256(b *testing.B) {
	list, facts := benchmarkList(b)
	compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), listValueMemberOperator{})
	if err != nil {
		b.Fatal(err)
	}
	set := ruleengine.RuleSet{ID: "custom", Rules: []ruleengine.Rule{{ID: "member", When: ruleengine.Compare("list_member", ruleengine.Variable(ruleengine.MustPath("code")), ruleengine.Literal(list))}}}
	plan, _, err := compiler.Compile(context.Background(), set)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if result := plan.Evaluate(ctx, facts); result.Decision != ruleengine.Unmatched {
			b.Fatalf("Evaluate() = %#v", result)
		}
	}
}
