package ruleengine

import (
	"context"
	"time"
)

// Decision is the rule-set evaluation state.
type Decision uint8

const (
	// Unmatched means no selected rule matched and no error occurred.
	Unmatched Decision = iota
	// Matched means at least one selected rule matched without an error.
	Matched
	// Indeterminate means an error or bound prevented a reliable decision.
	Indeterminate
)

// Explanation records one bounded rule evaluation without fact values.
type Explanation struct {
	// RuleID identifies an attempted evaluation and may recur across passes.
	RuleID RuleID
	// Matched is the predicate boolean, even when that predicate also returns an error.
	Matched bool
	// Unknown reports a MissingUnknown proposition that was neither true nor
	// false; Matched is then false. It is always false in other missing modes.
	Unknown bool
}

// Result contains a decision and possibly partial evaluation evidence. Returned
// slices belong to the caller; mutating them does not change the plan. They are
// not synchronized for concurrent mutation by callers.
type Result struct {
	// Decision is authoritative; partial matches do not override Indeterminate.
	Decision Decision
	// MatchedRules contains unique IDs in first-observed evaluation order.
	MatchedRules []string
	// Explanation records attempts up to MaxExplanation; truncation is silent.
	Explanation []Explanation
	// Errors contains redacted engine errors. Predicate failures are capped by
	// MaxDiagnostics; a terminal error can add one more. Underlying callback errors
	// are not wrapped in these entries; a predicate failure keeps code
	// evaluation_error and exposes an underlying engine code through Error.Cause.
	Errors []error
	// Duration measures Evaluate only, excluding prior resolver work.
	Duration time.Duration
	// DerivedFacts contains only newly inserted derivations, not input/resolved facts;
	// partial derivations remain inspectable on failure.
	DerivedFacts Context
}

// Plan stores immutable built-in state and can be evaluated concurrently if
// borrowed predicates/operators are concurrent-safe. Successful compilation
// is required; the zero Plan has no useful rule-set contract. No goroutines,
// shutdown or callback isolation are provided.
type Plan struct {
	setID         string
	namespace     string
	strategy      ConflictStrategy
	missing       MissingMode
	rules         []Rule
	limits        Limits
	operators     map[OperatorName]registeredOperator
	requiredPaths []Path
	hash          string
}

// Hash returns the canonical definition hash when the plan was compiled
// through CompileCached.
func (plan Plan) Hash() string { return plan.hash }

// FactResolver supplies explicitly requested missing facts. Implementations
// must be deterministic; EvaluateResolved invokes paths in lexical order.
type FactResolver interface {
	// Resolve returns a supplied value and provenance when found is true. False
	// means absence; a non-nil error stops resolution. It must honor cancellation
	// and not panic. Implementations own any I/O and synchronization.
	Resolve(context.Context, Path) (Value, Owner, bool, error)
}

// EvaluateResolved resolves only required paths absent from the base context,
// validates every found value, then evaluates the completed snapshot. Paths
// hidden inside custom predicates cannot be discovered. The base is unchanged;
// resolution is sequential but concurrent calls can share the resolver.
// One cooperative deadline covers resolution and evaluation. Resolution errors
// return Indeterminate without partial resolved facts or a measured Duration.
// Aggregate resolved facts are not checked against MaxFacts. A nil resolver
// returns Indeterminate; ctx must be non-nil.
func (plan Plan) EvaluateResolved(ctx context.Context, base Context, resolver FactResolver) Result {
	if resolver == nil {
		return Result{Decision: Indeterminate, Errors: []error{newError(CodeEvaluation, "fact resolver is nil")}}
	}
	evaluationContext, cancel := context.WithTimeout(ctx, plan.limits.EvaluationTimeout)
	defer cancel()
	working := base
	for _, path := range plan.requiredPaths {
		if err := evaluationContext.Err(); err != nil {
			return Result{Decision: Indeterminate, Errors: []error{newError(CodeEvaluation, "fact resolution canceled or timed out")}}
		}
		if working.Lookup(path).kind != KindMissing {
			continue
		}
		value, owner, found, err := resolver.Resolve(evaluationContext, path)
		if err != nil {
			return Result{Decision: Indeterminate, Errors: []error{newError(CodeEvaluation, "fact resolution failed")}}
		}
		if err := evaluationContext.Err(); err != nil {
			return Result{Decision: Indeterminate, Errors: []error{newError(CodeEvaluation, "fact resolution canceled or timed out")}}
		}
		if !found {
			continue
		}
		fact := Fact{Path: path, Value: value, Owner: owner}
		if _, err := NewContextWithLimits(plan.limits, fact); err != nil {
			return Result{Decision: Indeterminate, Errors: []error{newError(CodeInvalidFact, "resolved fact is invalid")}}
		}
		working = working.withFact(fact)
	}
	return plan.Evaluate(evaluationContext, working)
}

// Evaluate applies the compiled order, adding absent literal derivations until
// stable or bounded by MaxIterations. Predicate errors are sticky Indeterminate
// but other rules may still match and derive facts. Conflict or cancellation
// returns the accumulated evidence; it does not roll back external side effects.
// The supplied snapshot is not revalidated against plan limits.
// ctx must be non-nil. Callbacks run synchronously without panic recovery or
// forced preemption; cancellation is checked before owned steps, not after
// every callback. A final callback can therefore return a match after canceling
// its context. A result is not an authorization or action-execution decision.
func (plan Plan) Evaluate(ctx context.Context, facts Context) (result Result) {
	started := time.Now()
	derived, _ := NewContext()
	result = Result{Decision: Unmatched, DerivedFacts: derived}
	defer func() { result.Duration = time.Since(started) }()

	evaluationContext, cancel := context.WithTimeout(ctx, plan.limits.EvaluationTimeout)
	defer cancel()
	working := facts
	matchedRules := make(map[RuleID]bool)
	// A later match cannot erase an earlier predicate failure.
	hadError := false
	for iteration := 0; iteration < plan.limits.MaxIterations; iteration++ {
		changed := false
		stop := false
		for _, rule := range plan.rules {
			if err := evaluationContext.Err(); err != nil {
				result.Decision = Indeterminate
				result.Errors = append(result.Errors, newError(CodeEvaluation, "evaluation canceled or timed out"))
				return result
			}
			value, err := plan.evaluatePredicate(evaluationContext, rule.When, working)
			matched := value == truthTrue
			if len(result.Explanation) < plan.limits.MaxExplanation {
				result.Explanation = append(result.Explanation, Explanation{RuleID: rule.ID, Matched: matched, Unknown: value == truthUnknown})
			}
			if err != nil {
				hadError = true
				result.Decision = Indeterminate
				if len(result.Errors) < plan.limits.MaxDiagnostics {
					failure := newError(CodeEvaluation, "rule evaluation failed")
					failure.cause = causeCode(err)
					result.Errors = append(result.Errors, failure)
				}
				continue
			}
			if !matched {
				continue
			}
			if !matchedRules[rule.ID] {
				matchedRules[rule.ID] = true
				result.MatchedRules = append(result.MatchedRules, string(rule.ID))
			}
			if !hadError {
				result.Decision = Matched
			}
			for _, fact := range rule.Derive {
				existing := working.Lookup(fact.Path)
				if existing.kind == KindMissing {
					working = working.withFact(fact)
					result.DerivedFacts = result.DerivedFacts.withFact(fact)
					changed = true
					continue
				}
				if !valuesEqual(existing, fact.Value) {
					result.Decision = Indeterminate
					result.Errors = append(result.Errors, newError(CodeConflict, "derived fact conflicts with existing fact"))
					return result
				}
			}
			if plan.strategy == FirstMatch {
				stop = true
				break
			}
			if plan.strategy == ErrorOnMultiple && len(result.MatchedRules) > 1 {
				result.Decision = Indeterminate
				result.Errors = append(result.Errors, newError(CodeConflict, "multiple rules matched"))
				return result
			}
		}
		if !changed || stop {
			return result
		}
	}
	result.Decision = Indeterminate
	result.Errors = append(result.Errors, newError(CodeLimitExceeded, "forward chaining iteration limit exceeded"))
	return result
}

// truth is a Kleene three-valued proposition result. Only MissingUnknown
// comparisons originate truthUnknown; other plans see true and false only.
type truth uint8

const (
	truthFalse truth = iota
	truthTrue
	truthUnknown
)

func knownTruth(matched bool) truth {
	if matched {
		return truthTrue
	}
	return truthFalse
}

func (value truth) negate() truth {
	switch value {
	case truthTrue:
		return truthFalse
	case truthFalse:
		return truthTrue
	case truthUnknown:
		return truthUnknown
	default:
		return value
	}
}

// evaluatePredicate returns the child result alongside an error, matching the
// historical boolean contract where Explanation.Matched records that value.
func (plan Plan) evaluatePredicate(ctx context.Context, predicate Predicate, facts Context) (truth, error) {
	if err := ctx.Err(); err != nil {
		return truthFalse, err
	}
	switch typed := predicate.(type) {
	case comparisonPredicate:
		left := typed.left.resolve(facts)
		right := typed.right.resolve(facts)
		if left.kind == KindMissing || right.kind == KindMissing {
			switch plan.missing {
			case MissingNull:
				left, right = missingAsNull(left), missingAsNull(right)
			case MissingUnknown:
				return truthUnknown, nil
			case MissingFalse:
				return truthFalse, nil
			default:
				return truthFalse, nil
			}
		}
		if custom, exists := plan.operators[typed.operator]; exists {
			if !supportsSignature(custom.signatures, left.kind, right.kind) {
				return truthFalse, newError(CodeTypeMismatch, "custom operator operands are incompatible")
			}
			matched, err := custom.implementation.Evaluate(ctx, left, right)
			return knownTruth(matched), err
		}
		matched, err := evaluateBuiltin(typed.operator, left, right)
		return knownTruth(matched), err
	case allPredicate:
		result := truthTrue
		for _, child := range typed.children {
			value, err := plan.evaluatePredicate(ctx, child, facts)
			if err != nil || value == truthFalse {
				return value, err
			}
			if value == truthUnknown {
				result = truthUnknown
			}
		}
		return result, nil
	case anyPredicate:
		result := truthFalse
		for _, child := range typed.children {
			value, err := plan.evaluatePredicate(ctx, child, facts)
			if err != nil || value == truthTrue {
				return value, err
			}
			if value == truthUnknown {
				result = truthUnknown
			}
		}
		return result, nil
	case notPredicate:
		value, err := plan.evaluatePredicate(ctx, typed.child, facts)
		return value.negate(), err
	default:
		matched, err := predicate.Evaluate(ctx, facts)
		return knownTruth(matched), err
	}
}

func missingAsNull(value Value) Value {
	if value.kind == KindMissing {
		return Null()
	}
	return value
}
