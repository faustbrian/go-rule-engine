package ruleengine

import "time"

// Limits bounds compilation and evaluation work. Zero values are invalid
// except for the decimal budgets, where zero rejects every decimal value so
// limits written before KindDecimal existed remain valid. Callers should start
// with DefaultLimits and reduce values as needed.
type Limits struct {
	// MaxRules caps the number of rules in one compiled set.
	MaxRules int
	// MaxFacts caps input facts to NewContextWithLimits, not aggregate resolution or chaining.
	MaxFacts int
	// MaxASTDepth caps predicate depth (root 1) and value nesting (root 0).
	MaxASTDepth int
	// MaxOperands caps predicate nodes per rule, including logical nodes.
	MaxOperands int
	// MaxCollection caps each list independently, not the sum across nested lists.
	MaxCollection int
	// MaxStringBytes caps each UTF-8 string value in bytes.
	MaxStringBytes int
	// MaxDecimalBytes caps each decimal's lexical text in bytes. Zero rejects
	// decimals; negative values are invalid.
	MaxDecimalBytes int
	// MaxDecimalExponent caps the absolute power of ten of a nonzero decimal's
	// most significant digit, so 1e6144 and 1e-6144 both need 6144. It guards
	// against hostile exponents; negative values are invalid.
	MaxDecimalExponent int
	// MaxDefinitionBytes caps the raw ParseJSON input before decoding.
	MaxDefinitionBytes int
	// MaxRegexBytes caps a literal matches pattern at compilation.
	MaxRegexBytes int
	// MaxIdentifierBytes caps set/rule IDs and optional namespaces in bytes.
	MaxIdentifierBytes int
	// MaxTags caps the number of unique tags on each rule.
	MaxTags int
	// MaxTagBytes caps each tag in bytes.
	MaxTagBytes int
	// MaxPathBytes caps dotted path bytes at NewPath construction.
	MaxPathBytes int
	// MaxPathSegments caps the segment count at NewPath construction.
	MaxPathSegments int
	// MaxIterations caps complete forward-chaining passes before Indeterminate.
	MaxIterations int
	// MaxDerivedFacts caps declared derivations per rule and across the compiled set.
	MaxDerivedFacts int
	// MaxDiagnostics caps recorded predicate failures; a terminal error may add one entry.
	MaxDiagnostics int
	// MaxExplanation caps recorded rule evaluations; later entries are omitted.
	MaxExplanation int
	// EvaluationTimeout supplies a cooperative deadline to Evaluate, EvaluateResolved and
	// CompileCached. Synchronous work and callbacks are not forcibly preempted.
	EvaluationTimeout time.Duration
}

// DefaultLimits returns positive process-local construction, compilation and
// evaluation budgets. Applications must choose budgets suited to their workload.
func DefaultLimits() Limits {
	return Limits{
		MaxRules:       1_000,
		MaxFacts:       10_000,
		MaxASTDepth:    64,
		MaxOperands:    10_000,
		MaxCollection:  10_000,
		MaxStringBytes: 1 << 20,
		// 256 significant characters plus a sign and a decimal point.
		MaxDecimalBytes: 258,
		// The IEEE 754 decimal128 maximum exponent, well beyond monetary use.
		MaxDecimalExponent: 6_144,
		MaxDefinitionBytes: 4 << 20,
		MaxRegexBytes:      4 << 10,
		MaxIdentifierBytes: 256,
		MaxTags:            64,
		MaxTagBytes:        128,
		MaxPathBytes:       1_024,
		MaxPathSegments:    64,
		MaxIterations:      32,
		MaxDerivedFacts:    1_000,
		MaxDiagnostics:     100,
		MaxExplanation:     1_000,
		EvaluationTimeout:  5 * time.Second,
	}
}

func (l Limits) validate() error {
	if l.MaxRules <= 0 || l.MaxFacts <= 0 || l.MaxASTDepth <= 0 || l.MaxOperands <= 0 ||
		l.MaxCollection <= 0 || l.MaxStringBytes <= 0 || l.MaxRegexBytes <= 0 || l.MaxPathBytes <= 0 ||
		l.MaxDefinitionBytes <= 0 || l.MaxIdentifierBytes <= 0 || l.MaxTags <= 0 || l.MaxTagBytes <= 0 ||
		l.MaxPathSegments <= 0 || l.MaxIterations <= 0 || l.MaxDerivedFacts <= 0 ||
		l.MaxDiagnostics <= 0 || l.MaxExplanation <= 0 || l.EvaluationTimeout <= 0 ||
		l.MaxDecimalBytes < 0 || l.MaxDecimalExponent < 0 {
		return newError(CodeInvalidLimit, "limits must be positive; decimal limits must be non-negative")
	}
	return nil
}
