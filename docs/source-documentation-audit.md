# Source documentation audit

[Documentation index](README.md) · [Repository entry](../README.md)

## Boundary and reproducibility

Tier A source-only pass from `d65ef8ab3844a1021a542d44eabea776f4e257dd`.
Runtime, dependencies, API identities, specifications and dated evidence remain
unchanged. The original unrelated checkout is not part of this change.
Reproduce with `git ls-files '*.go'`, `git ls-files '*/go.mod' 'go.mod'`,
`golib source-docs inventory --json` and `golib source-docs check --json`.
Catalog producer: Go 1.27.2; module minima: Go 1.27.0.

## Complete inventory

Five modules own 56 tracked Go files. Root owns ruleengine and jsonast;
three independently versioned adapters own math, measurement and temporal.
The competitor module has a test-only package, absent from the public-package
catalog, and is included below. No vendor, generated Go, platform-specific
file, build constraint or generator directive occurs in this inventory.

| Module | Package | Kind | Files |
| --- | --- | --- | ---: |
| `.` | `jsonast` | source | 1 |
| `.` | `jsonast_test` | test | 1 |
| `.` | `ruleengine` | source | 12 |
| `.` | `ruleengine` | test | 5 |
| `.` | `ruleengine_test` | test | 12 |
| `adapters/math` | `ruleenginemath` | source | 1 |
| `adapters/math` | `ruleenginemath` | test | 1 |
| `adapters/math` | `ruleenginemath_test` | test | 5 |
| `adapters/measurement` | `ruleenginemeasurement` | source | 1 |
| `adapters/measurement` | `ruleenginemeasurement` | test | 1 |
| `adapters/measurement` | `ruleenginemeasurement_test` | test | 5 |
| `adapters/temporal` | `ruleenginetemporal` | source | 1 |
| `adapters/temporal` | `ruleenginetemporal` | test | 1 |
| `adapters/temporal` | `ruleenginetemporal_test` | test | 8 |
| `benchmarks/competitors` | `competitors_test` | test | 1 |

Every tracked Go file and its declared members, tests and helpers was inspected.
Public declarations and interface/field obligations are checked structurally;
uppercase members of private DTOs/test fixtures are not public API gaps.

```text
adapters/math/benchmark_test.go
adapters/math/coverage_internal_test.go
adapters/math/decimal.go
adapters/math/decimal_test.go
adapters/math/example_test.go
adapters/math/fuzz_test.go
adapters/math/property_test.go
adapters/measurement/benchmark_test.go
adapters/measurement/coverage_internal_test.go
adapters/measurement/example_test.go
adapters/measurement/fuzz_test.go
adapters/measurement/measurement.go
adapters/measurement/measurement_test.go
adapters/measurement/property_test.go
adapters/temporal/benchmark_test.go
adapters/temporal/compatibility_test.go
adapters/temporal/concurrency_test.go
adapters/temporal/coverage_internal_test.go
adapters/temporal/example_test.go
adapters/temporal/fuzz_test.go
adapters/temporal/relations_test.go
adapters/temporal/temporal.go
adapters/temporal/temporal_test.go
adapters/temporal/validation_test.go
benchmark_test.go
benchmarks/competitors/competitors_test.go
cache.go
cache_definition_identity_test.go
cache_policy_context_test.go
cache_test.go
canonical_cancellation_internal_test.go
chaining_test.go
compiler.go
compiler_bounds_test.go
context.go
context_test.go
coverage_internal_test.go
doc.go
engine_test.go
errors.go
example_test.go
extensions_test.go
fuzz_test.go
hardening_test.go
jsonast/jsonast.go
jsonast/jsonast_test.go
limits.go
model.go
mutation_boundaries_internal_test.go
operator.go
operator_truth_table_test.go
path.go
plan.go
serialization.go
serialization_test.go
value.go
```

## Missing documentation versus technical questions

Initial objective check identifies 52 missing public member comments in
limits, model, plan, operator, compiler, context and cache. Existing canonical
package comments need ownership/entry-point detail. Existing blanket immutable
plan and side-effect-free predicate wording needs the borrowed-extension caveat.
Private seams needing explanation: static dependency visibility, shallow rule
copy versus immutable built-in values, copy-on-write facts, sticky errors and
partial outputs, canonical identity versus callback identity, LRU locking,
JSON variant validation and ordered time comparison. Fuzz boundaries,
benchmark timing and test-double boundaries are included; existing
cancellation-count rationale is already useful and retained.

Observed source limitations, not invented intent:

- Callbacks execute synchronously, are not recovered or forcibly preempted,
  and built-in structural validation cannot inspect arbitrary predicates.
- Evaluation checks cancellation before owned steps, not after every callback;
  a final successful callback can yield Matched after canceling its context.
- MaxFacts applies at context construction, not to accumulated resolved/derived
  facts. Path construction limits are not rechecked against compiler limits.
- MaxDiagnostics caps predicate errors; a terminal cancellation, conflict or
  iteration-limit error can add one more entry. Explanations truncate silently.
- Time equality uses time.Time.Equal, while ordering uses UnixNano; the latter
  representation is not meaningful across the full time.Time year range, per
  the [Go 1.27.2 Time.UnixNano contract](https://pkg.go.dev/time@go1.27.2#Time.UnixNano).
- JSON unknown-field rejection is not duplicate-member or case-sensitive-name
  rejection. Grammar v1 is this repository's DSL, not JSON Logic or another DSL.

Unanswered rationale: why these exact default budgets, why time ordering uses
UnixNano rather than time.Time.Compare, and whether broader parser strictness,
aggregate fact budgets or post-callback cancellation should become contracts.
No behavioral repair or conformance/performance assertion is included.

## Tests, variants, generation and maintained surfaces

All root external/internal tests cover context copies, exact bounds, operator
truth tables, static cycles, ordering, cache policy/identity, redaction and
canonical serialization. Cancellation-count contexts protect checkpoints rather
than real-time timing; stepContext intentionally does not model Done
propagation. Adapter tests cover tagged canonical values, exact domain
relations, independent registration and concurrency; temporal compatibility
fixtures protect persisted encodings. Existing examples exercise root and each
adapter. Eight fuzz targets cover parser/path and the six adapter boundaries.
Benchmarks separate setup from evaluation except the intentional Compile
benchmark; competitor Grule resets state inside timing to preserve equivalent
repeated decisions. No performance claim follows from this pass.

The tracked source has no unfinished-debt, deprecation or generated-code
markers. Existing nolint/exhaustive directives remain byte-for-byte unchanged.
Immutable specs, fixture pins, API historical baselines and dated
threat-model evidence are preserved. Current maintained evaluation/limits/
security wording will be reconciled narrowly with source, not rewritten as an
adoption-documentation project.

Native API discovery: .golib.yaml selects apidiff for root api/v2.txt and
adapter math v2, measurement v3, temporal v2. No go-doc snapshot/generator is
selected by verification scripts or tests. Comment-only changes retain all
API baseline bytes. Native docs shell selects local links, root Examples and
vet; proportionate checks select links, rendered docs and Examples, not broad
vet/integration/race/fuzz/mutation/competitor campaigns.

## Selected verification

Passed with Go 1.27.2, GOWORK=off and bounded task-owned GOCACHE,
GOMODCACHE and GOTMPDIR (GOMAXPROCS=2, GOFLAGS=-p=2 -mod=readonly):

- Native gofmt across all 56 tracked Go files: clean.
- Native scanner/parser helper comparing git-show base to current files:
  all 56 executable tokens, literals, implicit semicolons, directives and
  comment-free ASTs identical. This also preserves exported signatures.
- Native `go doc -all .` for root and all three adapters, and
  `go doc -all ./jsonast`: all five rendered and inspected; root repeat stable.
  Test-only competitor package comment parsed/rendered with go/doc/comment.
  After final public-comment refinements, all-file formatting/equivalence and
  affected root/jsonast renders were repeated; root output remained stable.
- Native `go test ./... -run '^Example' -count=1 -timeout=2m` in root and
  each adapter: all four existing Examples passed. jsonast had no Examples;
  its fixture tests were inspected, not re-executed. Example source/runtime and
  dependency inputs were unchanged by final prose refinements; these results
  were reused rather than re-executed.
- Coordinator-owned `golib source-docs check --json`: 56 files, zero findings.
- Local documentation target inspection: 188 file targets, none missing.
- Pinned Gitleaks v8.30.1 policy, added-content/new-file boundary with inline
  allow suppression disabled and an owned empty ignore file: zero findings.
  No full current-file snapshot scan or historical scan was substituted.
- Generation, marker and API discovery: no GoDoc snapshot selected, no Go
  generator, build variants or generated Go; every API baseline unchanged.

All owned Go caches, helper and rendered output were removed immediately after
inspection, including verified absence. Secret snapshots/reports are removed;
shared parent-owned tools remain untouched. Final diff inspection covers all
source comments and maintained-document corrections, with no unresolved
prose findings. No behavioral fix or
new runtime qualification is claimed. Full runtime/vet/race/fuzz/mutation,
competitor campaigns, releases, live services and public consumers were not run.
Publication and exact-source hosted CI remain coordinator-owned.

## Objective Enforcement

The checksum-pinned public Tools v2.4.0 verifier now supplies
`make source-docs-check` as a prerequisite of the existing local and CI Make
routes. Ordinary shared CI selects strict source-documentation checking and
requires its result. Native prerequisites, operation budgets, security and
release controls are retained. Hosted exact-head evidence is recorded by the
coordination ledger; configuration alone does not establish green delivery.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
