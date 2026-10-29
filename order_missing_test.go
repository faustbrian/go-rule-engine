package ruleengine_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

// recordingOperator reports true and records the operand kinds it received.
type recordingOperator struct {
	name       ruleengine.OperatorName
	signatures []ruleengine.Signature
	calls      *[][2]ruleengine.Kind
	mu         *sync.Mutex
}

func (operator recordingOperator) Name() ruleengine.OperatorName { return operator.name }
func (operator recordingOperator) Signatures() []ruleengine.Signature {
	return operator.signatures
}
func (operator recordingOperator) Evaluate(_ context.Context, left, right ruleengine.Value) (bool, error) {
	if operator.calls != nil {
		operator.mu.Lock()
		*operator.calls = append(*operator.calls, [2]ruleengine.Kind{left.Kind(), right.Kind()})
		operator.mu.Unlock()
	}
	return true, nil
}

func compileSet(t *testing.T, compiler ruleengine.Compiler, set ruleengine.RuleSet) ruleengine.Plan {
	t.Helper()
	plan, diagnostics, err := compiler.Compile(context.Background(), set)
	if err != nil {
		t.Fatalf("Compile(%s) = %#v, %v", set.ID, diagnostics, err)
	}
	return plan
}

func defaultCompiler() ruleengine.Compiler { return ruleengine.NewCompiler(ruleengine.DefaultLimits()) }

func TestDefaultCanonicalBytesAndHashesAreUnchanged(t *testing.T) {
	t.Parallel()

	// Hash recorded from v2.0.2 source b0e2fb5 for the maintained fixture.
	const fixtureHash = "7dfe533d1ef3d7971bdbce891c2738b6f87e7650603294f8fed92ee5aa2e365f"
	data, err := os.ReadFile("jsonast/testdata/location-routing.json")
	if err != nil {
		t.Fatal(err)
	}
	set, _, err := ruleengine.ParseJSON(data, ruleengine.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if set.Order != ruleengine.OrderPriority || set.Missing != ruleengine.MissingFalse {
		t.Fatalf("fixture options = %v, %v", set.Order, set.Missing)
	}
	if hash, hashErr := ruleengine.CanonicalHash(set); hashErr != nil || hash != fixtureHash {
		t.Fatalf("CanonicalHash(fixture) = %q, %v; want %q", hash, hashErr, fixtureHash)
	}
	canonical, err := ruleengine.MarshalCanonical(set)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"order"`, `"missing"`, `"decimal"`} {
		if strings.Contains(string(canonical), field) {
			t.Fatalf("default canonical JSON contains %s: %s", field, canonical)
		}
	}
}

func TestRuleOrder(t *testing.T) {
	t.Parallel()

	rules := []ruleengine.Rule{
		{ID: "c", Priority: 1, When: ruleengine.True()},
		{ID: "a", Priority: 5, When: ruleengine.True()},
		{ID: "b", Priority: 5, When: ruleengine.True()},
		{ID: "d", Priority: 1, When: ruleengine.False()},
		{ID: "e", Priority: 9, When: ruleengine.True()},
	}
	facts := ruleengine.Context{}
	tests := []struct {
		order    ruleengine.RuleOrder
		strategy ruleengine.ConflictStrategy
		matched  []string
		explain  []ruleengine.RuleID
	}{
		{ruleengine.OrderPriority, ruleengine.FirstMatch, []string{"e"}, []ruleengine.RuleID{"e"}},
		{ruleengine.OrderPriority, ruleengine.CollectAll, []string{"e", "a", "b", "c"}, []ruleengine.RuleID{"e", "a", "b", "c", "d"}},
		{ruleengine.OrderDeclared, ruleengine.FirstMatch, []string{"c"}, []ruleengine.RuleID{"c"}},
		{ruleengine.OrderDeclared, ruleengine.CollectAll, []string{"c", "a", "b", "e"}, []ruleengine.RuleID{"c", "a", "b", "d", "e"}},
	}
	for _, test := range tests {
		set := ruleengine.RuleSet{ID: "ordered", Strategy: test.strategy, Order: test.order, Rules: rules}
		result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), facts)
		if !reflect.DeepEqual(result.MatchedRules, test.matched) {
			t.Errorf("order %v strategy %v MatchedRules = %v, want %v", test.order, test.strategy, result.MatchedRules, test.matched)
		}
		explained := make([]ruleengine.RuleID, len(result.Explanation))
		for index, entry := range result.Explanation {
			explained[index] = entry.RuleID
		}
		if !reflect.DeepEqual(explained, test.explain) {
			t.Errorf("order %v strategy %v Explanation = %v, want %v", test.order, test.strategy, explained, test.explain)
		}
	}
	// Declared order follows the slice, so reversing it changes the first match.
	reversed := make([]ruleengine.Rule, len(rules))
	for index, rule := range rules {
		reversed[len(rules)-1-index] = rule
	}
	set := ruleengine.RuleSet{ID: "ordered", Order: ruleengine.OrderDeclared, Rules: reversed}
	if result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), facts); !reflect.DeepEqual(result.MatchedRules, []string{"e"}) {
		t.Fatalf("reversed declared FirstMatch = %v", result.MatchedRules)
	}
	invalid := ruleengine.RuleSet{ID: "ordered", Order: ruleengine.OrderDeclared + 1, Rules: rules}
	if _, _, err := defaultCompiler().Compile(context.Background(), invalid); !ruleengine.IsCode(err, ruleengine.CodeInvalidRule) {
		t.Fatalf("Compile(unknown order) error = %v", err)
	}
}

func TestRuleOrderJSON(t *testing.T) {
	t.Parallel()

	rules := []ruleengine.Rule{
		{ID: "z", Priority: 1, When: ruleengine.True()},
		{ID: "a", Priority: 2, When: ruleengine.True()},
	}
	declared := ruleengine.RuleSet{ID: "ordered", Order: ruleengine.OrderDeclared, Rules: rules}
	canonical, err := ruleengine.MarshalCanonical(declared)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(canonical), `"strategy":"first_match","order":"declared","rules":[{"id":"z","priority":1`) {
		t.Fatalf("declared canonical JSON = %s", canonical)
	}
	parsed, _, err := ruleengine.ParseJSON(canonical, ruleengine.DefaultLimits())
	if err != nil || parsed.Order != ruleengine.OrderDeclared || parsed.Rules[0].ID != "z" || parsed.Rules[0].Priority != 1 {
		t.Fatalf("ParseJSON(declared) = %#v, %v", parsed, err)
	}
	again, err := ruleengine.MarshalCanonical(parsed)
	if err != nil || string(again) != string(canonical) {
		t.Fatalf("declared round trip = %s, %v", again, err)
	}
	declaredHash, _ := ruleengine.CanonicalHash(declared)
	priorityHash, _ := ruleengine.CanonicalHash(ruleengine.RuleSet{ID: "ordered", Rules: rules})
	if declaredHash == priorityHash {
		t.Fatal("declared and priority order share a hash")
	}

	base := `{"version":"2","id":"o","strategy":"first_match",%s"rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"true"},"derive":[]}]}`
	for _, option := range []string{`"order":"Declared",`, `"order":"random",`, `"order":1,`, `"order":null,"order":"x",`} {
		if _, _, parseErr := ruleengine.ParseJSON([]byte(strings.Replace(base, "%s", option, 1)), ruleengine.DefaultLimits()); !ruleengine.IsCode(parseErr, ruleengine.CodeInvalidJSON) {
			t.Errorf("ParseJSON(%s) error = %v, want invalid_json", option, parseErr)
		}
	}
	explicit, _, err := ruleengine.ParseJSON([]byte(strings.Replace(base, "%s", `"order":"priority",`, 1)), ruleengine.DefaultLimits())
	if err != nil || explicit.Order != ruleengine.OrderPriority {
		t.Fatalf("ParseJSON(explicit priority) = %v, %v", explicit.Order, err)
	}
	if canonical, _ := ruleengine.MarshalCanonical(explicit); strings.Contains(string(canonical), `"order"`) {
		t.Fatalf("explicit default order was serialized: %s", canonical)
	}
}

// Truth values for the Kleene tables: T and F are constants and U is an
// equality on an absent path under MissingUnknown.
var kleene = map[string]func() ruleengine.Predicate{
	"T": ruleengine.True,
	"F": ruleengine.False,
	"U": func() ruleengine.Predicate {
		return ruleengine.Compare(ruleengine.OpEqual, ruleengine.Variable(ruleengine.MustPath("absent")), ruleengine.Literal(ruleengine.Int(1)))
	},
}

func evaluateTruth(t *testing.T, mode ruleengine.MissingMode, predicate ruleengine.Predicate) string {
	t.Helper()
	set := ruleengine.RuleSet{ID: "kleene", Missing: mode, Rules: []ruleengine.Rule{{ID: "r", When: predicate}}}
	result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), ruleengine.Context{})
	if len(result.Errors) != 0 || len(result.Explanation) != 1 {
		t.Fatalf("Evaluate() = %#v", result)
	}
	entry := result.Explanation[0]
	switch {
	case entry.Matched && !entry.Unknown && result.Decision == ruleengine.Matched:
		return "T"
	case !entry.Matched && !entry.Unknown && result.Decision == ruleengine.Unmatched:
		return "F"
	case !entry.Matched && entry.Unknown && result.Decision == ruleengine.Unmatched:
		return "U"
	default:
		t.Fatalf("inconsistent result %#v", result)
		return ""
	}
}

func TestMissingUnknownKleeneTruthTables(t *testing.T) {
	t.Parallel()

	and := map[[2]string]string{
		{"T", "T"}: "T", {"T", "F"}: "F", {"T", "U"}: "U",
		{"F", "T"}: "F", {"F", "F"}: "F", {"F", "U"}: "F",
		{"U", "T"}: "U", {"U", "F"}: "F", {"U", "U"}: "U",
	}
	or := map[[2]string]string{
		{"T", "T"}: "T", {"T", "F"}: "T", {"T", "U"}: "T",
		{"F", "T"}: "T", {"F", "F"}: "F", {"F", "U"}: "U",
		{"U", "T"}: "T", {"U", "F"}: "U", {"U", "U"}: "U",
	}
	for pair, want := range and {
		if got := evaluateTruth(t, ruleengine.MissingUnknown, ruleengine.All(kleene[pair[0]](), kleene[pair[1]]())); got != want {
			t.Errorf("All(%s, %s) = %s, want %s", pair[0], pair[1], got, want)
		}
	}
	for pair, want := range or {
		if got := evaluateTruth(t, ruleengine.MissingUnknown, ruleengine.Any(kleene[pair[0]](), kleene[pair[1]]())); got != want {
			t.Errorf("Any(%s, %s) = %s, want %s", pair[0], pair[1], got, want)
		}
	}
	for operand, want := range map[string]string{"T": "F", "F": "T", "U": "U"} {
		if got := evaluateTruth(t, ruleengine.MissingUnknown, ruleengine.Not(kleene[operand]())); got != want {
			t.Errorf("Not(%s) = %s, want %s", operand, got, want)
		}
	}
	// Unknown is not double-negated into true, and nesting keeps Kleene rules.
	if got := evaluateTruth(t, ruleengine.MissingUnknown, ruleengine.Not(ruleengine.Not(kleene["U"]()))); got != "U" {
		t.Errorf("Not(Not(U)) = %s", got)
	}
	if got := evaluateTruth(t, ruleengine.MissingUnknown, ruleengine.Any(ruleengine.All(kleene["U"](), kleene["T"]()), ruleengine.Not(kleene["T"]()))); got != "U" {
		t.Errorf("Any(All(U, T), Not(T)) = %s", got)
	}
	// Every comparison operator, including negated and set tests, is Unknown on a missing operand.
	absent := ruleengine.Variable(ruleengine.MustPath("absent"))
	for _, comparison := range []ruleengine.Predicate{
		ruleengine.Compare(ruleengine.OpNotEqual, absent, ruleengine.Literal(ruleengine.Int(1))),
		ruleengine.Compare(ruleengine.OpLessThan, ruleengine.Literal(ruleengine.Int(1)), absent),
		ruleengine.Compare(ruleengine.OpIn, absent, ruleengine.Literal(ruleengine.List(ruleengine.Int(1)))),
		ruleengine.Compare(ruleengine.OpNotIn, absent, ruleengine.Literal(ruleengine.List(ruleengine.Int(1)))),
		ruleengine.Compare(ruleengine.OpContains, absent, ruleengine.Literal(ruleengine.String("x"))),
		ruleengine.Compare(ruleengine.OpMatches, absent, ruleengine.Literal(ruleengine.String("x"))),
	} {
		if got := evaluateTruth(t, ruleengine.MissingUnknown, comparison); got != "U" {
			t.Errorf("%#v under MissingUnknown = %s", comparison, got)
		}
		if got := evaluateTruth(t, ruleengine.MissingFalse, comparison); got != "F" {
			t.Errorf("%#v under MissingFalse = %s", comparison, got)
		}
		if got := evaluateTruth(t, ruleengine.MissingFalse, ruleengine.Not(comparison)); got != "T" {
			t.Errorf("Not(%#v) under MissingFalse = %s", comparison, got)
		}
	}
}

func TestMissingUnknownKeepsFirstMatchSearching(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "first", Missing: ruleengine.MissingUnknown, Order: ruleengine.OrderDeclared, Rules: []ruleengine.Rule{
		{ID: "unknown", When: kleene["U"](), Derive: []ruleengine.Fact{{Path: ruleengine.MustPath("u"), Value: ruleengine.Bool(true)}}},
		{ID: "known", When: ruleengine.True()},
	}}
	result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), ruleengine.Context{})
	want := []ruleengine.Explanation{{RuleID: "unknown", Unknown: true}, {RuleID: "known", Matched: true}}
	if result.Decision != ruleengine.Matched || !reflect.DeepEqual(result.MatchedRules, []string{"known"}) ||
		!reflect.DeepEqual(result.Explanation, want) || result.DerivedFacts.Lookup(ruleengine.MustPath("u")).Kind() != ruleengine.KindMissing {
		t.Fatalf("Evaluate() = %#v", result)
	}
}

func TestExistsIsKnownInEveryMissingMode(t *testing.T) {
	t.Parallel()

	facts, err := ruleengine.NewContext(ruleengine.Fact{Path: ruleengine.MustPath("null"), Value: ruleengine.Null()})
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []ruleengine.MissingMode{ruleengine.MissingFalse, ruleengine.MissingNull, ruleengine.MissingUnknown} {
		set := ruleengine.RuleSet{ID: "exists", Missing: mode, Strategy: ruleengine.CollectAll, Rules: []ruleengine.Rule{
			{ID: "absent", When: ruleengine.Exists(ruleengine.MustPath("absent"))},
			{ID: "not-absent", When: ruleengine.Not(ruleengine.Exists(ruleengine.MustPath("absent")))},
			{ID: "null", When: ruleengine.Exists(ruleengine.MustPath("null"))},
		}}
		result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), facts)
		want := []ruleengine.Explanation{{RuleID: "absent"}, {RuleID: "not-absent", Matched: true}, {RuleID: "null", Matched: true}}
		if !reflect.DeepEqual(result.Explanation, want) {
			t.Errorf("mode %v Explanation = %#v", mode, result.Explanation)
		}
	}
}

func TestMissingNullResolvesAbsentOperandsToNull(t *testing.T) {
	t.Parallel()

	absent := ruleengine.Variable(ruleengine.MustPath("absent"))
	set := ruleengine.RuleSet{ID: "null", Missing: ruleengine.MissingNull, Strategy: ruleengine.CollectAll, Rules: []ruleengine.Rule{
		{ID: "equals-null", When: ruleengine.Compare(ruleengine.OpEqual, absent, ruleengine.Literal(ruleengine.Null()))},
		{ID: "in-list", When: ruleengine.Compare(ruleengine.OpIn, absent, ruleengine.Literal(ruleengine.List(ruleengine.Null())))},
		{ID: "two-absent", When: ruleengine.Compare(ruleengine.OpEqual, absent, ruleengine.Variable(ruleengine.MustPath("other")))},
	}}
	result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), ruleengine.Context{})
	if result.Decision != ruleengine.Matched || !reflect.DeepEqual(result.MatchedRules, []string{"equals-null", "in-list", "two-absent"}) {
		t.Fatalf("Evaluate() = %#v", result)
	}

	// Null is then strictly typed, as an explicitly supplied null would be.
	mismatch := ruleengine.RuleSet{ID: "null", Missing: ruleengine.MissingNull, Rules: []ruleengine.Rule{
		{ID: "int", When: ruleengine.Compare(ruleengine.OpEqual, absent, ruleengine.Literal(ruleengine.Int(1)))},
	}}
	result = compileSet(t, defaultCompiler(), mismatch).Evaluate(context.Background(), ruleengine.Context{})
	if result.Decision != ruleengine.Indeterminate || len(result.Errors) != 1 || !ruleengine.IsCode(result.Errors[0], ruleengine.CodeEvaluation) {
		t.Fatalf("Evaluate(null = int) = %#v", result)
	}
	var engineError *ruleengine.Error
	if !errorAs(result.Errors[0], &engineError) {
		t.Fatalf("error type = %T", result.Errors[0])
	}
	if cause, ok := engineError.Cause(); !ok || cause != ruleengine.CodeTypeMismatch {
		t.Fatalf("Cause() = %q, %v", cause, ok)
	}
}

func TestCustomOperatorsReceiveNullOnlyUnderMissingNull(t *testing.T) {
	t.Parallel()

	for mode, wantCalls := range map[ruleengine.MissingMode][][2]ruleengine.Kind{
		ruleengine.MissingFalse:   nil,
		ruleengine.MissingNull:    {{ruleengine.KindNull, ruleengine.KindInt}},
		ruleengine.MissingUnknown: nil,
	} {
		var calls [][2]ruleengine.Kind
		operator := recordingOperator{
			name:       "seen",
			signatures: []ruleengine.Signature{{Left: ruleengine.KindNull, Right: ruleengine.KindInt}},
			calls:      &calls, mu: &sync.Mutex{},
		}
		compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), operator)
		if err != nil {
			t.Fatal(err)
		}
		set := ruleengine.RuleSet{ID: "custom", Missing: mode, Rules: []ruleengine.Rule{
			{ID: "r", When: ruleengine.Compare("seen", ruleengine.Variable(ruleengine.MustPath("absent")), ruleengine.Literal(ruleengine.Int(1)))},
		}}
		result := compileSet(t, compiler, set).Evaluate(context.Background(), ruleengine.Context{})
		if !reflect.DeepEqual(calls, wantCalls) {
			t.Errorf("mode %v calls = %v, want %v", mode, calls, wantCalls)
		}
		if matched := result.Decision == ruleengine.Matched; matched != (mode == ruleengine.MissingNull) {
			t.Errorf("mode %v result = %#v", mode, result)
		}
	}

	// Without a Null signature, MissingNull is a redacted type mismatch.
	strict := recordingOperator{name: "strict", signatures: []ruleengine.Signature{{Left: ruleengine.KindInt, Right: ruleengine.KindInt}}}
	compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), strict)
	if err != nil {
		t.Fatal(err)
	}
	set := ruleengine.RuleSet{ID: "strict", Missing: ruleengine.MissingNull, Rules: []ruleengine.Rule{
		{ID: "r", When: ruleengine.Compare("strict", ruleengine.Variable(ruleengine.MustPath("absent")), ruleengine.Literal(ruleengine.Int(1)))},
	}}
	result := compileSet(t, compiler, set).Evaluate(context.Background(), ruleengine.Context{})
	var engineError *ruleengine.Error
	if result.Decision != ruleengine.Indeterminate || len(result.Errors) != 1 || !errorAs(result.Errors[0], &engineError) {
		t.Fatalf("Evaluate(strict) = %#v", result)
	}
	if cause, ok := engineError.Cause(); !ok || cause != ruleengine.CodeTypeMismatch {
		t.Fatalf("Cause() = %q, %v", cause, ok)
	}
}

// missingModeRules covers every built-in node kind, custom operators and
// negations so that mode equivalence is checked beyond a single comparison.
func missingModeRules() []ruleengine.Rule {
	country := ruleengine.Variable(ruleengine.MustPath("country"))
	weight := ruleengine.Variable(ruleengine.MustPath("weight"))
	nothing := ruleengine.Variable(ruleengine.MustPath("nothing"))
	return []ruleengine.Rule{
		{ID: "eq", When: ruleengine.Compare(ruleengine.OpEqual, country, ruleengine.Literal(ruleengine.String("FI")))},
		{ID: "ne", When: ruleengine.Not(ruleengine.Compare(ruleengine.OpNotEqual, country, ruleengine.Literal(ruleengine.String("SE"))))},
		{ID: "all", When: ruleengine.All(ruleengine.Compare(ruleengine.OpGreaterThan, weight, ruleengine.Literal(ruleengine.Int(5))), ruleengine.Exists(ruleengine.MustPath("country")))},
		{ID: "any", When: ruleengine.Any(ruleengine.False(), ruleengine.Compare(ruleengine.OpIn, country, ruleengine.Literal(ruleengine.List(ruleengine.String("FI")))))},
		{ID: "null", When: ruleengine.Compare(ruleengine.OpEqual, nothing, ruleengine.Literal(ruleengine.Null()))},
		{ID: "custom", When: ruleengine.Compare("seen", weight, ruleengine.Literal(ruleengine.Int(1)))},
		{ID: "mismatch", When: ruleengine.Compare(ruleengine.OpEqual, weight, ruleengine.Variable(ruleengine.MustPath("country")))},
		{ID: "func", When: ruleengine.PredicateFunc(func(context.Context, ruleengine.Context) (bool, error) { return true, nil })},
	}
}

func TestEveryMissingModeMatchesDefaultWhenNoFactIsMissing(t *testing.T) {
	t.Parallel()

	facts, err := ruleengine.NewContext(
		ruleengine.Fact{Path: ruleengine.MustPath("country"), Value: ruleengine.String("FI")},
		ruleengine.Fact{Path: ruleengine.MustPath("weight"), Value: ruleengine.Int(10)},
		ruleengine.Fact{Path: ruleengine.MustPath("nothing"), Value: ruleengine.Null()},
	)
	if err != nil {
		t.Fatal(err)
	}
	operator := recordingOperator{name: "seen", signatures: []ruleengine.Signature{{Left: ruleengine.KindInt, Right: ruleengine.KindInt}}}
	compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), operator)
	if err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		decision    ruleengine.Decision
		matched     []string
		explanation []ruleengine.Explanation
		errors      []string
	}
	evaluate := func(mode ruleengine.MissingMode, strategy ruleengine.ConflictStrategy) outcome {
		set := ruleengine.RuleSet{ID: "equivalent", Missing: mode, Strategy: strategy, Rules: missingModeRules()}
		result := compileSet(t, compiler, set).Evaluate(context.Background(), facts)
		errorsText := make([]string, len(result.Errors))
		for index, err := range result.Errors {
			errorsText[index] = err.Error()
		}
		return outcome{result.Decision, result.MatchedRules, result.Explanation, errorsText}
	}
	for _, strategy := range []ruleengine.ConflictStrategy{ruleengine.FirstMatch, ruleengine.CollectAll, ruleengine.ErrorOnMultiple} {
		want := evaluate(ruleengine.MissingFalse, strategy)
		if strategy == ruleengine.CollectAll && len(want.matched) < 5 {
			t.Fatalf("baseline matched too little to compare: %#v", want)
		}
		for _, mode := range []ruleengine.MissingMode{ruleengine.MissingNull, ruleengine.MissingUnknown} {
			if got := evaluate(mode, strategy); !reflect.DeepEqual(got, want) {
				t.Errorf("strategy %v mode %v = %#v, want %#v", strategy, mode, got, want)
			}
		}
	}
}

type absentResolver struct{}

func (absentResolver) Resolve(context.Context, ruleengine.Path) (ruleengine.Value, ruleengine.Owner, bool, error) {
	return ruleengine.Value{}, ruleengine.OwnerUnspecified, false, nil
}

func TestResolverNotFoundFollowsMissingMode(t *testing.T) {
	t.Parallel()

	absent := ruleengine.Variable(ruleengine.MustPath("absent"))
	rules := []ruleengine.Rule{
		{ID: "not-equal", When: ruleengine.Compare(ruleengine.OpNotEqual, absent, ruleengine.Literal(ruleengine.Int(1)))},
		{ID: "is-null", When: ruleengine.Compare(ruleengine.OpEqual, absent, ruleengine.Literal(ruleengine.Null()))},
		{ID: "negated", When: ruleengine.Not(ruleengine.Compare(ruleengine.OpEqual, absent, ruleengine.Literal(ruleengine.Null())))},
	}
	tests := map[ruleengine.MissingMode][]ruleengine.Explanation{
		ruleengine.MissingFalse: {{RuleID: "is-null"}, {RuleID: "negated", Matched: true}, {RuleID: "not-equal"}},
		// not_equal(Null, Int) is a type mismatch, as for an explicit null.
		ruleengine.MissingNull:    {{RuleID: "is-null", Matched: true}, {RuleID: "negated"}, {RuleID: "not-equal"}},
		ruleengine.MissingUnknown: {{RuleID: "is-null", Unknown: true}, {RuleID: "negated", Unknown: true}, {RuleID: "not-equal", Unknown: true}},
	}
	for mode, want := range tests {
		set := ruleengine.RuleSet{ID: "resolved", Missing: mode, Strategy: ruleengine.CollectAll, Rules: rules}
		result := compileSet(t, defaultCompiler(), set).EvaluateResolved(context.Background(), ruleengine.Context{}, absentResolver{})
		if !reflect.DeepEqual(result.Explanation, want) {
			t.Errorf("mode %v Explanation = %#v, want %#v", mode, result.Explanation, want)
		}
		wantDecision := map[ruleengine.MissingMode]ruleengine.Decision{
			ruleengine.MissingFalse: ruleengine.Matched, ruleengine.MissingNull: ruleengine.Indeterminate, ruleengine.MissingUnknown: ruleengine.Unmatched,
		}[mode]
		if result.Decision != wantDecision {
			t.Errorf("mode %v Decision = %v, want %v", mode, result.Decision, wantDecision)
		}
	}
}

func TestMissingModeJSON(t *testing.T) {
	t.Parallel()

	rules := []ruleengine.Rule{{ID: "r", When: kleene["U"]()}}
	hashes := map[string]bool{}
	for mode, spelling := range map[ruleengine.MissingMode]string{ruleengine.MissingFalse: "", ruleengine.MissingNull: `"missing":"null",`, ruleengine.MissingUnknown: `"missing":"unknown",`} {
		set := ruleengine.RuleSet{ID: "m", Missing: mode, Rules: rules}
		canonical, err := ruleengine.MarshalCanonical(set)
		if err != nil {
			t.Fatal(err)
		}
		if spelling == "" && strings.Contains(string(canonical), `"missing"`) || !strings.Contains(string(canonical), `"strategy":"first_match",`+spelling+`"rules"`) {
			t.Fatalf("mode %v canonical = %s", mode, canonical)
		}
		parsed, _, err := ruleengine.ParseJSON(canonical, ruleengine.DefaultLimits())
		if err != nil || parsed.Missing != mode {
			t.Fatalf("mode %v ParseJSON = %v, %v", mode, parsed.Missing, err)
		}
		hash, _ := ruleengine.CanonicalHash(set)
		hashes[hash] = true
	}
	if len(hashes) != 3 {
		t.Fatalf("missing modes produced %d distinct hashes", len(hashes))
	}
	combined := ruleengine.RuleSet{ID: "m", Order: ruleengine.OrderDeclared, Missing: ruleengine.MissingUnknown, Rules: rules}
	canonical, err := ruleengine.MarshalCanonical(combined)
	if err != nil || !strings.Contains(string(canonical), `"order":"declared","missing":"unknown","rules"`) {
		t.Fatalf("combined canonical = %s, %v", canonical, err)
	}
	base := `{"version":"2","id":"m","strategy":"first_match",%s"rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"true"},"derive":[]}]}`
	for _, option := range []string{`"missing":"Unknown",`, `"missing":"absent",`, `"missing":true,`} {
		if _, _, parseErr := ruleengine.ParseJSON([]byte(strings.Replace(base, "%s", option, 1)), ruleengine.DefaultLimits()); !ruleengine.IsCode(parseErr, ruleengine.CodeInvalidJSON) {
			t.Errorf("ParseJSON(%s) error = %v, want invalid_json", option, parseErr)
		}
	}
	explicit, _, err := ruleengine.ParseJSON([]byte(strings.Replace(base, "%s", `"missing":"false",`, 1)), ruleengine.DefaultLimits())
	if err != nil || explicit.Missing != ruleengine.MissingFalse {
		t.Fatalf("ParseJSON(explicit false) = %v, %v", explicit.Missing, err)
	}
	invalid := ruleengine.RuleSet{ID: "m", Missing: ruleengine.MissingUnknown + 1, Rules: rules}
	if _, _, err := defaultCompiler().Compile(context.Background(), invalid); !ruleengine.IsCode(err, ruleengine.CodeInvalidRule) {
		t.Fatalf("Compile(unknown missing mode) error = %v", err)
	}
}

func TestRuleErrorCausePreservesEngineCode(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "cause", Strategy: ruleengine.CollectAll, Rules: []ruleengine.Rule{
		{ID: "mismatch", When: ruleengine.Compare(ruleengine.OpEqual, ruleengine.Variable(ruleengine.MustPath("x")), ruleengine.Literal(ruleengine.String("1")))},
		{ID: "callback", When: ruleengine.PredicateFunc(func(context.Context, ruleengine.Context) (bool, error) { return false, context.DeadlineExceeded })},
	}}
	facts, err := ruleengine.NewContext(ruleengine.Fact{Path: ruleengine.MustPath("x"), Value: ruleengine.Int(1)})
	if err != nil {
		t.Fatal(err)
	}
	result := compileSet(t, defaultCompiler(), set).Evaluate(context.Background(), facts)
	if len(result.Errors) != 2 {
		t.Fatalf("Errors = %#v", result.Errors)
	}
	// Priority order sorts "callback" before "mismatch"; context errors carry no engine code.
	wantCauses := map[int]ruleengine.Code{1: ruleengine.CodeTypeMismatch}
	for index, failure := range result.Errors {
		var engineError *ruleengine.Error
		if !errorAs(failure, &engineError) || engineError.Code() != ruleengine.CodeEvaluation || failure.Error() != "evaluation_error: rule evaluation failed" {
			t.Fatalf("Errors[%d] = %v", index, failure)
		}
		cause, ok := engineError.Cause()
		if want, known := wantCauses[index]; cause != want || ok != known {
			t.Fatalf("Errors[%d].Cause() = %q, %v", index, cause, ok)
		}
	}
}

func errorAs(err error, target **ruleengine.Error) bool { return errors.As(err, target) }

func TestGrammarVersionGatesNewFeatures(t *testing.T) {
	t.Parallel()

	rule := func(value string) string {
		return `"rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"compare","operator":"equal","left":{"kind":"variable","path":["x"]},"right":{"kind":"literal","value":` + value + `}},"derive":[]}]}`
	}
	decimalList := `{"type":"list","list":[{"type":"decimal","decimal":"1"}]}`
	features := []string{
		`"order":"declared",` + rule(`{"type":"int","int":1}`),
		`"order":"priority",` + rule(`{"type":"int","int":1}`),
		`"missing":"unknown",` + rule(`{"type":"int","int":1}`),
		`"missing":"false",` + rule(`{"type":"int","int":1}`),
		rule(`{"type":"decimal","decimal":"1"}`),
		rule(decimalList),
		`"rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"true"},"derive":[{"path":["fee"],"owner":"unspecified","value":{"type":"decimal","decimal":"0.10"}}]}]}`,
		`"rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"not","child":{"kind":"any","children":[{"kind":"compare","operator":"in","left":{"kind":"variable","path":["x"]},"right":{"kind":"literal","value":` + decimalList + `}}]}},"derive":[]}]}`,
	}
	for _, body := range features {
		version1 := `{"version":"1","id":"g","strategy":"first_match",` + body
		if _, _, err := ruleengine.ParseJSON([]byte(version1), ruleengine.DefaultLimits()); !ruleengine.IsCode(err, ruleengine.CodeInvalidJSON) {
			t.Errorf("ParseJSON(v1 %s) error = %v, want invalid_json", body, err)
		}
		version2 := `{"version":"2","id":"g","strategy":"first_match",` + body
		set, _, err := ruleengine.ParseJSON([]byte(version2), ruleengine.DefaultLimits())
		if err != nil {
			t.Errorf("ParseJSON(v2 %s) error = %v", body, err)
			continue
		}
		canonical, err := ruleengine.MarshalCanonical(set)
		if err != nil {
			t.Fatal(err)
		}
		explicitDefault := strings.Contains(body, `"priority",`) || strings.Contains(body, `"false",`)
		if wantVersion := map[bool]string{true: `{"version":"1",`, false: `{"version":"2",`}[explicitDefault]; !strings.HasPrefix(string(canonical), wantVersion) {
			t.Errorf("MarshalCanonical(%s) = %s, want prefix %s", body, canonical, wantVersion)
		}
	}
	// A version 2 document without features is accepted and canonicalizes to version 1.
	plain := `{"version":"2","id":"g","strategy":"first_match",` + rule(`{"type":"int","int":1}`)
	set, _, err := ruleengine.ParseJSON([]byte(plain), ruleengine.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if canonical, _ := ruleengine.MarshalCanonical(set); !strings.HasPrefix(string(canonical), `{"version":"1",`) {
		t.Fatalf("featureless canonical = %s", canonical)
	}
}
