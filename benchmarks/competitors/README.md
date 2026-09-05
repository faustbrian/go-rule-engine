# Competitor benchmark harness

This internal, non-releasable module compares one equivalent country-and-weight
decision across `go-rule-engine`, Expr, and Grule. It isolates competitor
dependencies from every public rule-engine module and provides engineering
evidence; it is not a package-selection guide or a claim that one engine is
universally faster.

The harness requires Go 1.26.6 or newer. It has no supported installation path,
public package, semantic-version release, or runtime dependency relationship
with consumers.

## Run

From the repository root:

```console
go test ./benchmarks/competitors/... -run '^TestEquivalentDecisionBaselines$'
go test ./benchmarks/competitors/... -run '^$' \
  -bench '^BenchmarkEquivalentDecision$' -benchmem -count 10
```

The repository gate runs the nested module through its own module directory.
For publishable comparisons, run the same commit and Go toolchain on one idle
machine, retain CPU and operating-system details, use enough samples for
`benchstat`, and report latency and allocations together.

## Comparison boundary

All three cases compile an equivalent predicate before the timer starts and
evaluate matching facts while timed. The Grule case resets its reusable
knowledge instance and mutable result field for each iteration. Differences in
language, compilation model, data representation, rule-set breadth, caching,
concurrency, diagnostics, and application integration remain outside this
narrow comparison.

Do not use these results to select a production engine without measuring the
actual rule corpus, fact shape, lifecycle, and correctness requirements. Do not
add a competitor dependency to the root or adapter modules.

## Navigation

- [Benchmark source](competitors_test.go)
- [Rule-engine performance guidance](../../docs/performance.md)
- [Rule-engine documentation index](../../docs/README.md)
- [Contribution guide](../../CONTRIBUTING.md)
- [Security policy](../../SECURITY.md)
- [Support policy](../../SUPPORT.md)
- [License](../../LICENSE)
- [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)

This harness belongs only in the Golib engineering inventory. Consumers should
start with the public [`go-rule-engine` module](../../README.md).
