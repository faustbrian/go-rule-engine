package ruleengine

import "context"

// RuleID is a stable identifier within a rule set.
type RuleID string

// ConflictStrategy controls how ordered matches are selected.
type ConflictStrategy uint8

const (
	// FirstMatch selects only the first deterministically ordered match.
	FirstMatch ConflictStrategy = iota
	// CollectAll selects every unique matching rule.
	CollectAll
	// ErrorOnMultiple rejects more than one unique match.
	ErrorOnMultiple
)

// RuleOrder selects how Compile orders rules for evaluation.
type RuleOrder uint8

const (
	// OrderPriority orders by descending Priority, then ascending ID. It is the
	// zero value and the only order before this option existed; canonical JSON
	// and hashes omit it.
	OrderPriority RuleOrder = iota
	// OrderDeclared evaluates rules in input slice order. Priority is retained
	// and serialized as metadata but does not affect evaluation.
	OrderDeclared
)

// MissingMode selects how comparisons treat an operand whose fact path is
// absent. It applies to compiled plans; direct Predicate.Evaluate calls always
// use MissingFalse. Exists reports presence in every mode.
type MissingMode uint8

const (
	// MissingFalse makes a comparison with an absent operand false without
	// invoking the operator, including negated operators and custom operators.
	// It is the zero value; canonical JSON and hashes omit it.
	MissingFalse MissingMode = iota
	// MissingNull resolves an absent operand to Null before operator dispatch,
	// exactly as if null had been supplied. Built-in and custom operators then
	// apply their ordinary kind checks, so a Null/Int comparison is a
	// type_mismatch evaluation error.
	MissingNull
	// MissingUnknown makes a comparison with an absent operand Unknown under
	// three-valued Kleene logic: Not keeps Unknown, All is false if any child is
	// false and otherwise Unknown if any child is Unknown, and Any is true if any
	// child is true and otherwise Unknown if any child is Unknown. A rule matches
	// only when its proposition is true. Constants, Exists and custom predicates
	// are always known.
	MissingUnknown
)

// Rule is a proposition with stable metadata and optional derived facts.
type Rule struct {
	// ID is nonempty and unique within the set; Compile validates its byte limit.
	ID RuleID
	// Namespace is optional metadata, not a path or authorization scope.
	Namespace string
	// Priority orders rules descending, with ascending ID breaking ties, under
	// OrderPriority. Under OrderDeclared it is metadata only.
	Priority int
	// Tags are unique bounded metadata strings; Compile copies the slice.
	Tags []string
	// When must be non-nil. Custom predicates remain borrowed by the plan.
	When Predicate
	// Derive supplies literal facts on a match; duplicate derived paths are rejected.
	Derive []Fact
}

// RuleSet is an independently compiled collection of rules.
type RuleSet struct {
	// ID is the nonempty rule-set identifier.
	ID string
	// Namespace is optional rule-set metadata and does not filter evaluation.
	Namespace string
	// Strategy selects matches; its zero value is FirstMatch.
	Strategy ConflictStrategy
	// Rules are validated and ordered by Compile; custom predicate state is not copied.
	Rules []Rule
	// Order selects evaluation order; its zero value is OrderPriority.
	Order RuleOrder
	// Missing selects absent-operand semantics; its zero value is MissingFalse.
	Missing MissingMode
}

// Predicate evaluates a snapshot. Implementations must supply bounded,
// deterministic, side-effect-free behavior and honor cancellation. A plan
// retains the implementation; concurrent evaluations require concurrent-safe
// implementations. The engine neither recovers panics nor interrupts callbacks.
type Predicate interface {
	// Evaluate returns the match and failure; a non-nil error invalidates the match.
	Evaluate(context.Context, Context) (bool, error)
}

// PredicateFunc adapts a trusted callback without copying captured state.
// A nil function panics when invoked and is not rejected by compilation.
// Its dependencies are opaque and it cannot be canonically serialized.
type PredicateFunc func(context.Context, Context) (bool, error)

// Evaluate invokes the adapted predicate function.
func (predicate PredicateFunc) Evaluate(ctx context.Context, facts Context) (bool, error) {
	return predicate(ctx, facts)
}

type constantPredicate bool

func (predicate constantPredicate) Evaluate(ctx context.Context, _ Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return bool(predicate), nil
}

// True returns an always-true predicate.
func True() Predicate { return constantPredicate(true) }

// False returns an always-false predicate.
func False() Predicate { return constantPredicate(false) }

type allPredicate struct{ children []Predicate }

func (predicate allPredicate) Evaluate(ctx context.Context, facts Context) (bool, error) {
	for _, child := range predicate.children {
		matched, err := child.Evaluate(ctx, facts)
		if err != nil || !matched {
			return matched, err
		}
	}
	return true, nil
}

// All copies the child slice and evaluates left-to-right, stopping at false or
// an error. Compile rejects an empty list or nil child; direct evaluation of
// an empty All returns true. Child implementations are borrowed, not cloned.
func All(children ...Predicate) Predicate {
	return allPredicate{children: append([]Predicate(nil), children...)}
}

type anyPredicate struct{ children []Predicate }

func (predicate anyPredicate) Evaluate(ctx context.Context, facts Context) (bool, error) {
	for _, child := range predicate.children {
		matched, err := child.Evaluate(ctx, facts)
		if err != nil || matched {
			return matched, err
		}
	}
	return false, nil
}

// Any copies the child slice and evaluates left-to-right, stopping at true or
// an error. Compile rejects an empty list or nil child; direct evaluation of
// an empty Any returns false. Child implementations are borrowed, not cloned.
func Any(children ...Predicate) Predicate {
	return anyPredicate{children: append([]Predicate(nil), children...)}
}

type notPredicate struct{ child Predicate }

func (predicate notPredicate) Evaluate(ctx context.Context, facts Context) (bool, error) {
	matched, err := predicate.child.Evaluate(ctx, facts)
	return !matched, err
}

// Not negates the child match while preserving its error. Compile rejects a
// nil child; direct evaluation with nil panics.
func Not(child Predicate) Predicate { return notPredicate{child: child} }

type existsPredicate struct{ path Path }

func (predicate existsPredicate) Evaluate(ctx context.Context, facts Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return facts.Lookup(predicate.path).kind != KindMissing, nil
}

// Exists reports whether a path was supplied, including an explicit null.
func Exists(path Path) Predicate { return existsPredicate{path: path} }

// Operand resolves a typed value from Literal or Variable. Its private methods
// prevent implementations outside this package; nil operands are not valid.
type Operand interface {
	resolve(Context) Value
	staticKind() (Kind, bool)
}

type literalOperand struct{ value Value }

func (operand literalOperand) resolve(Context) Value    { return operand.value.clone() }
func (operand literalOperand) staticKind() (Kind, bool) { return operand.value.kind, true }

// Literal returns an immutable literal operand.
func Literal(value Value) Operand { return literalOperand{value: value.clone()} }

type variableOperand struct{ path Path }

func (operand variableOperand) resolve(facts Context) Value { return facts.Lookup(operand.path) }
func (operand variableOperand) staticKind() (Kind, bool)    { return KindMissing, false }

// Variable resolves an explicit fact path during evaluation.
func Variable(path Path) Operand { return variableOperand{path: path} }

type comparisonPredicate struct {
	operator OperatorName
	left     Operand
	right    Operand
}

// Compare creates a typed binary comparison without validation. Compile checks
// operators and known literal kinds; variable kinds are checked at evaluation.
// Custom operators require a compiled plan; direct Evaluate uses only built-ins.
// Missing values produce false, including for negated comparison operators;
// a compiled plan applies its RuleSet.Missing mode instead.
// Time ordering uses UnixNano; unlike equality, its valid range is restricted.
// left and right must be non-nil; nil can panic during Compile or evaluation.
func Compare(operator OperatorName, left, right Operand) Predicate {
	return comparisonPredicate{operator: operator, left: left, right: right}
}

func (predicate comparisonPredicate) Evaluate(ctx context.Context, facts Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return evaluateBuiltin(predicate.operator, predicate.left.resolve(facts), predicate.right.resolve(facts))
}
