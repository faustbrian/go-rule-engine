package ruleengine_test

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

func mustDecimal(t testing.TB, text string) ruleengine.Value {
	t.Helper()
	value, err := ruleengine.Decimal(text)
	if err != nil {
		t.Fatalf("Decimal(%q) error = %v", text, err)
	}
	return value
}

func TestDecimalGrammar(t *testing.T) {
	t.Parallel()

	accepted := []string{
		"0", "-0", "0.0", "-0.000", "1", "-1", "10", "1.50", "0.10", "123456789012345678901234567890",
		"1e5", "1E5", "1e+5", "1e-5", "1.5e-0", "0e999999999999999999", "1e000000000000000000000000000005",
	}
	for _, text := range accepted {
		value, err := ruleengine.Decimal(text)
		if err != nil {
			t.Errorf("Decimal(%q) error = %v", text, err)
			continue
		}
		if value.Kind() != ruleengine.KindDecimal {
			t.Errorf("Decimal(%q).Kind() = %v", text, value.Kind())
		}
	}
	rejected := []string{
		"", "-", "+1", "01", "-01", "00", "00.5", ".5", "1.", "1.e5", "1e", "1e+", "1e-", "e5",
		" 1", "1 ", "1_000", "0x10", "NaN", "Infinity", "-Infinity", "1.5.5", "1e5e5", "--1", "1,5", "١",
		"1e1000000000000000000",
	}
	for _, text := range rejected {
		if _, err := ruleengine.Decimal(text); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
			t.Errorf("Decimal(%q) error = %v, want invalid_fact", text, err)
		}
	}
}

func TestDecimalTextRoundTripsLexically(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"1.50", "-0", "1E+05", "0.000", "123.4500e-2"} {
		value := mustDecimal(t, text)
		got, ok := value.DecimalText()
		if !ok || got != text {
			t.Fatalf("DecimalText() = %q, %v; want %q", got, ok, text)
		}
		if value.Interface() != text {
			t.Fatalf("Interface() = %#v", value.Interface())
		}
	}
	if _, ok := ruleengine.String("1.50").DecimalText(); ok {
		t.Fatal("String.DecimalText() reported a decimal")
	}
	if _, ok := mustDecimal(t, "1").StringValue(); ok {
		t.Fatal("Decimal.StringValue() reported a string")
	}
}

func evaluateBuiltin(t *testing.T, operator ruleengine.OperatorName, left, right ruleengine.Value) (bool, error) {
	t.Helper()
	return ruleengine.Compare(operator, ruleengine.Literal(left), ruleengine.Literal(right)).Evaluate(context.Background(), ruleengine.Context{})
}

func TestDecimalExactComparison(t *testing.T) {
	t.Parallel()

	tests := []struct {
		operator    ruleengine.OperatorName
		left, right string
		want        bool
	}{
		{ruleengine.OpEqual, "1.0", "1.00", true},
		{ruleengine.OpEqual, "1", "1e0", true},
		{ruleengine.OpEqual, "100", "1e2", true},
		{ruleengine.OpEqual, "0.001", "1e-3", true},
		{ruleengine.OpEqual, "-0", "0", true},
		{ruleengine.OpEqual, "-0.0", "0e10", true},
		{ruleengine.OpEqual, "1.5", "1.50001", false},
		{ruleengine.OpNotEqual, "1.0", "1.00", false},
		{ruleengine.OpNotEqual, "-1", "1", true},
		{ruleengine.OpLessThan, "0.10", "0.2", true},
		{ruleengine.OpLessThan, "0.2", "0.10", false},
		{ruleengine.OpLessThan, "-2", "-1.5", true},
		{ruleengine.OpLessThan, "-1", "0", true},
		{ruleengine.OpLessThan, "0", "1e-6144", true},
		{ruleengine.OpLessThan, "9.99", "10", true},
		{ruleengine.OpLessThan, "1.5", "1.51", true},
		{ruleengine.OpLessThan, "-1.51", "-1.5", true},
		{ruleengine.OpLessOrEqual, "1.50", "1.5", true},
		{ruleengine.OpGreaterThan, "1e3", "999.999", true},
		{ruleengine.OpGreaterOrEqual, "-0", "0", true},
		// 0.1+0.2 differs from 0.3 in binary floating point but not here.
		{ruleengine.OpEqual, "0.30000000000000000001", "0.3", false},
		{ruleengine.OpGreaterThan, "0.30000000000000000001", "0.3", true},
	}
	for _, test := range tests {
		got, err := evaluateBuiltin(t, test.operator, mustDecimal(t, test.left), mustDecimal(t, test.right))
		if err != nil || got != test.want {
			t.Errorf("%s(%s, %s) = %v, %v; want %v", test.operator, test.left, test.right, got, err, test.want)
		}
	}
	list := ruleengine.List(mustDecimal(t, "2.0"), mustDecimal(t, "3"))
	if got, err := evaluateBuiltin(t, ruleengine.OpIn, mustDecimal(t, "2"), list); err != nil || !got {
		t.Fatalf("in(2, [2.0, 3]) = %v, %v", got, err)
	}
	if got, err := evaluateBuiltin(t, ruleengine.OpNotIn, mustDecimal(t, "3.00"), list); err != nil || got {
		t.Fatalf("not_in(3.00, [2.0, 3]) = %v, %v", got, err)
	}
	if got, err := evaluateBuiltin(t, ruleengine.OpContains, list, mustDecimal(t, "30e-1")); err != nil || !got {
		t.Fatalf("contains([2.0, 3], 30e-1) = %v, %v", got, err)
	}
	if got, err := evaluateBuiltin(t, ruleengine.OpEqual, list, ruleengine.List(mustDecimal(t, "2"), mustDecimal(t, "3.0"))); err != nil || !got {
		t.Fatalf("equal(list, list) = %v, %v", got, err)
	}
}

func TestDecimalCrossKindIsTypeMismatch(t *testing.T) {
	t.Parallel()

	decimal := mustDecimal(t, "1")
	for _, other := range []ruleengine.Value{ruleengine.Int(1), ruleengine.Float(1), ruleengine.String("1"), ruleengine.Null()} {
		for _, operator := range []ruleengine.OperatorName{ruleengine.OpEqual, ruleengine.OpNotEqual, ruleengine.OpLessThan, ruleengine.OpGreaterOrEqual} {
			if _, err := evaluateBuiltin(t, operator, decimal, other); !ruleengine.IsCode(err, ruleengine.CodeTypeMismatch) {
				t.Errorf("%s(decimal, %v) error = %v, want type_mismatch", operator, other.Kind(), err)
			}
		}
	}
	// Membership compares element kinds without error, so cross-kind elements never match.
	if got, err := evaluateBuiltin(t, ruleengine.OpIn, decimal, ruleengine.List(ruleengine.Int(1), ruleengine.Float(1), ruleengine.String("1"))); err != nil || got {
		t.Fatalf("in(decimal 1, [int 1, float 1, string 1]) = %v, %v", got, err)
	}
	set := ruleengine.RuleSet{ID: "mismatch", Rules: []ruleengine.Rule{{ID: "r", When: ruleengine.Compare(ruleengine.OpEqual, ruleengine.Literal(decimal), ruleengine.Literal(ruleengine.Int(1)))}}}
	if _, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).Compile(context.Background(), set); !ruleengine.IsCode(err, ruleengine.CodeTypeMismatch) {
		t.Fatalf("Compile(decimal = int) error = %v", err)
	}
}

func TestDecimalLimits(t *testing.T) {
	t.Parallel()

	path := ruleengine.MustPath("amount")
	defaults := ruleengine.DefaultLimits()
	if defaults.MaxDecimalBytes < 258 || defaults.MaxDecimalExponent <= 0 {
		t.Fatalf("DefaultLimits decimal budgets = %d, %d", defaults.MaxDecimalBytes, defaults.MaxDecimalExponent)
	}
	limits := defaults
	limits.MaxDecimalBytes = 4
	limits.MaxDecimalExponent = 3
	for text, wantOK := range map[string]bool{
		"1.25": true, "12.5": true, "-1.5": true, "1.255": false,
		"1e3": true, "9999": true, "1e4": false, "1e-3": true, "1e-4": false,
		"0e99": true,
	} {
		_, err := ruleengine.NewContextWithLimits(limits, ruleengine.Fact{Path: path, Value: mustDecimal(t, text)})
		if (err == nil) != wantOK {
			t.Errorf("NewContextWithLimits(%q) error = %v, want ok=%v", text, err, wantOK)
		}
		if err != nil && !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
			t.Errorf("NewContextWithLimits(%q) code = %v", text, err)
		}
	}
	// The bound applies to the value's leading power, not the written exponent:
	// 10e2 is 1e3, while 10e3 is 1e4.
	for _, text := range []string{"10e2", "1e+3"} {
		if _, err := ruleengine.NewContextWithLimits(limits, ruleengine.Fact{Path: path, Value: mustDecimal(t, text)}); err != nil {
			t.Fatalf("NewContextWithLimits(%s) error = %v", text, err)
		}
	}
	if _, err := ruleengine.NewContextWithLimits(limits, ruleengine.Fact{Path: path, Value: mustDecimal(t, "10e3")}); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
		t.Fatalf("NewContextWithLimits(10e3) error = %v", err)
	}
	set := ruleengine.RuleSet{ID: "limits", Rules: []ruleengine.Rule{{ID: "r", When: ruleengine.Compare(ruleengine.OpEqual, ruleengine.Variable(path), ruleengine.Literal(mustDecimal(t, "1.255")))}}}
	if _, _, err := ruleengine.NewCompiler(limits).Compile(context.Background(), set); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
		t.Fatalf("Compile(oversized literal) error = %v", err)
	}
	negative := defaults
	negative.MaxDecimalExponent = -1
	if _, err := ruleengine.NewContextWithLimits(negative); !ruleengine.IsCode(err, ruleengine.CodeInvalidLimit) {
		t.Fatalf("negative MaxDecimalExponent error = %v", err)
	}
	negative = defaults
	negative.MaxDecimalBytes = -1
	if _, err := ruleengine.NewContextWithLimits(negative); !ruleengine.IsCode(err, ruleengine.CodeInvalidLimit) {
		t.Fatalf("negative MaxDecimalBytes error = %v", err)
	}
}

func TestDecimalLimitsZeroDisable(t *testing.T) {
	t.Parallel()

	limits := ruleengine.DefaultLimits()
	limits.MaxDecimalBytes = 0
	limits.MaxDecimalExponent = 0
	if _, err := ruleengine.NewContextWithLimits(limits, ruleengine.Fact{Path: ruleengine.MustPath("ordinary"), Value: ruleengine.Int(1)}); err != nil {
		t.Fatalf("zero decimal limits rejected ordinary facts: %v", err)
	}
	if _, err := ruleengine.NewContextWithLimits(limits, ruleengine.Fact{Path: ruleengine.MustPath("amount"), Value: mustDecimal(t, "1")}); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
		t.Fatalf("zero MaxDecimalBytes accepted a decimal: %v", err)
	}
}

func TestDecimalHostileExponentGuard(t *testing.T) {
	t.Parallel()

	// Exponent digits beyond int64 range are refused at construction, before
	// any budget, and within-range hostile exponents fail MaxDecimalExponent
	// without materializing 10^exponent.
	if _, err := ruleengine.Decimal("1e" + strings.Repeat("9", 19)); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
		t.Fatalf("Decimal(1e<19 digits>) error = %v", err)
	}
	for _, text := range []string{"1e999999999999999999", "1e-999999999999999999", "-9.9e999999999999999999"} {
		value := mustDecimal(t, text)
		if _, err := ruleengine.NewContext(ruleengine.Fact{Path: ruleengine.MustPath("x"), Value: value}); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
			t.Fatalf("NewContext(%s) error = %v", text, err)
		}
	}
	definition := `{"version":"2","id":"hostile","strategy":"first_match","rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"compare","operator":"equal","left":{"kind":"literal","value":{"type":"decimal","decimal":"1e999999999"}},"right":{"kind":"literal","value":{"type":"decimal","decimal":"1"}}},"derive":[]}]}`
	if _, _, err := ruleengine.ParseJSON([]byte(definition), ruleengine.DefaultLimits()); !ruleengine.IsCode(err, ruleengine.CodeInvalidFact) {
		t.Fatalf("ParseJSON(hostile exponent) error = %v", err)
	}
}

func TestDecimalCanonicalJSON(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "money", Rules: []ruleengine.Rule{{
		ID:     "threshold",
		When:   ruleengine.Compare(ruleengine.OpGreaterOrEqual, ruleengine.Variable(ruleengine.MustPath("total")), ruleengine.Literal(mustDecimal(t, "1.50"))),
		Derive: []ruleengine.Fact{{Path: ruleengine.MustPath("fee"), Value: mustDecimal(t, "0.10")}},
	}}}
	canonical, err := ruleengine.MarshalCanonical(set)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`{"type":"decimal","decimal":"1.50"}`, `{"type":"decimal","decimal":"0.10"}`} {
		if !strings.Contains(string(canonical), fragment) {
			t.Fatalf("canonical JSON %s lacks %s", canonical, fragment)
		}
	}
	parsed, _, err := ruleengine.ParseJSON(canonical, ruleengine.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	again, err := ruleengine.MarshalCanonical(parsed)
	if err != nil || string(again) != string(canonical) {
		t.Fatalf("round trip = %s, %v", again, err)
	}
	// Spelling is identity: equal values with different text hash differently.
	spelled := set
	spelled.Rules = []ruleengine.Rule{set.Rules[0]}
	spelled.Rules[0].When = ruleengine.Compare(ruleengine.OpGreaterOrEqual, ruleengine.Variable(ruleengine.MustPath("total")), ruleengine.Literal(mustDecimal(t, "1.5")))
	first, _ := ruleengine.CanonicalHash(set)
	second, _ := ruleengine.CanonicalHash(spelled)
	if first == second {
		t.Fatal("different decimal spellings produced one hash")
	}
	facts, err := ruleengine.NewContext(ruleengine.Fact{Path: ruleengine.MustPath("total"), Value: mustDecimal(t, "1.5")})
	if err != nil {
		t.Fatal(err)
	}
	plan, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).Compile(context.Background(), parsed)
	if err != nil {
		t.Fatal(err)
	}
	result := plan.Evaluate(context.Background(), facts)
	if result.Decision != ruleengine.Matched {
		t.Fatalf("Evaluate(total 1.5 >= 1.50) = %#v", result)
	}
	fee, _ := result.DerivedFacts.Lookup(ruleengine.MustPath("fee")).DecimalText()
	if fee != "0.10" {
		t.Fatalf("derived fee = %q", fee)
	}

	for _, invalid := range []string{
		`{"type":"decimal","decimal":"01"}`,
		`{"type":"decimal"}`,
		`{"type":"decimal","decimal":"1","string":"1"}`,
		`{"type":"string","decimal":"1"}`,
		`{"type":"decimal","decimal":1}`,
	} {
		definition := `{"version":"2","id":"bad","strategy":"first_match","rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"compare","operator":"equal","left":{"kind":"variable","path":["x"]},"right":{"kind":"literal","value":` + invalid + `}},"derive":[]}]}`
		if _, _, err := ruleengine.ParseJSON([]byte(definition), ruleengine.DefaultLimits()); !ruleengine.IsCode(err, ruleengine.CodeInvalidJSON) {
			t.Errorf("ParseJSON(%s) error = %v, want invalid_json", invalid, err)
		}
	}
}

func TestDecimalCustomOperatorSignature(t *testing.T) {
	t.Parallel()

	operator := recordingOperator{name: "decimal_seen", signatures: []ruleengine.Signature{{Left: ruleengine.KindDecimal, Right: ruleengine.KindDecimal}}}
	compiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), operator)
	if err != nil {
		t.Fatal(err)
	}
	set := ruleengine.RuleSet{ID: "custom", Rules: []ruleengine.Rule{{ID: "r", When: ruleengine.Compare("decimal_seen", ruleengine.Literal(mustDecimal(t, "1")), ruleengine.Literal(mustDecimal(t, "2")))}}}
	if _, _, err := compiler.Compile(context.Background(), set); err != nil {
		t.Fatalf("Compile(decimal signature) error = %v", err)
	}
}

// decimalOracle returns a big.Rat for an accepted decimal, or nil when its
// exponent is too large to materialize cheaply in a test oracle.
func decimalOracle(text string) *big.Rat {
	if _, exponent, found := strings.Cut(strings.ToLower(text), "e"); found {
		digits := strings.TrimLeft(strings.TrimLeft(exponent, "+-"), "0")
		if len(digits) > 4 {
			return nil
		}
	}
	rational, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil
	}
	return rational
}

// FuzzDecimal checks grammar acceptance against encoding/json's RFC 8259
// number syntax and exact ordering against an independent big.Rat oracle.
func FuzzDecimal(f *testing.F) {
	for _, seed := range [][2]string{
		{"1.0", "1.00"}, {"0.10", "0.2"}, {"-0", "0"}, {"1e5", "100000"}, {"01", "1"},
		{"-1.5e-3", "-0.0015"}, {"9.99", "10"}, {"1e999999999999999999", "1"}, {"", "+1"},
	} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, left, right string) {
		leftValue, leftErr := ruleengine.Decimal(left)
		rightValue, rightErr := ruleengine.Decimal(right)
		for _, check := range []struct {
			text string
			err  error
		}{{left, leftErr}, {right, rightErr}} {
			if check.err == nil && !isJSONNumber(check.text) {
				t.Fatalf("Decimal(%q) accepted a non-RFC 8259 number", check.text)
			}
			if check.err != nil && isJSONNumber(check.text) && !hugeExponent(check.text) {
				t.Fatalf("Decimal(%q) rejected an RFC 8259 number: %v", check.text, check.err)
			}
		}
		if leftErr != nil || rightErr != nil {
			return
		}
		leftRat, rightRat := decimalOracle(left), decimalOracle(right)
		if leftRat == nil || rightRat == nil {
			return
		}
		want := leftRat.Cmp(rightRat)
		less, err := evaluateBuiltinFuzz(ruleengine.OpLessThan, leftValue, rightValue)
		if err != nil || less != (want < 0) {
			t.Fatalf("less_than(%q, %q) = %v, %v; oracle %d", left, right, less, err, want)
		}
		equal, err := evaluateBuiltinFuzz(ruleengine.OpEqual, leftValue, rightValue)
		if err != nil || equal != (want == 0) {
			t.Fatalf("equal(%q, %q) = %v, %v; oracle %d", left, right, equal, err, want)
		}
		greater, err := evaluateBuiltinFuzz(ruleengine.OpGreaterThan, leftValue, rightValue)
		if err != nil || greater != (want > 0) {
			t.Fatalf("greater_than(%q, %q) = %v, %v; oracle %d", left, right, greater, err, want)
		}
	})
}

func evaluateBuiltinFuzz(operator ruleengine.OperatorName, left, right ruleengine.Value) (bool, error) {
	return ruleengine.Compare(operator, ruleengine.Literal(left), ruleengine.Literal(right)).Evaluate(context.Background(), ruleengine.Context{})
}

// isJSONNumber uses encoding/json as the independent RFC 8259 grammar oracle.
// The first-byte check restricts it to a number token and the whitespace check
// rejects the insignificant whitespace json.Valid allows around a value.
func isJSONNumber(text string) bool {
	if text == "" || (text[0] != '-' && (text[0] < '0' || text[0] > '9')) {
		return false
	}
	return !strings.ContainsAny(text, " \t\r\n") && json.Valid([]byte(text))
}

// hugeExponent reports exponents beyond the documented 18-significant-digit
// construction guard, the only RFC 8259 numbers Decimal refuses.
func hugeExponent(text string) bool {
	_, exponent, found := strings.Cut(strings.ToLower(text), "e")
	return found && len(strings.TrimLeft(strings.TrimLeft(exponent, "+-"), "0")) > 18
}
