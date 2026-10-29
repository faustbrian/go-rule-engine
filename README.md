# rule-engine

[![CI](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-risk_selected-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-optional_local%2Fselected_scheduled-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-rule-engine/v2.svg)](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/v2)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-rule-engine?sort=semver)](https://github.com/faustbrian/go-rule-engine/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`rule-engine` is a deterministic, typed, inspectable engine for evaluating
facts and propositions. It compiles immutable execution plans, evaluates them
concurrently without hidden I/O, supports bounded forward chaining, and emits
redacted diagnostics and explanations.

It is intentionally not an authorization system, feature-flag service,
validator, workflow engine, database query layer, or action runner. Those
products may adapt its decisions while retaining their own fail-closed and
domain semantics.

Root v2.0.0 and the optional math v2.0.0, measurement v3.0.0 and temporal
v2.0.0 adapters are published on Go 1.27.0. Their public signatures expose
root v2 types. Published root v1.0.0 and its compatible adapter releases remain
available under their original import paths.
See [migration](docs/migration.md) before updating root and adapter types together.

## Install

Select the published root-v2 release through the public Go proxy. Install only
the optional adapters your application uses:

```sh
go get github.com/faustbrian/go-rule-engine/v2@v2.0.0
go get github.com/faustbrian/go-rule-engine/adapters/math/v2@v2.0.0
go get github.com/faustbrian/go-rule-engine/adapters/measurement/v3@v3.0.0
go get github.com/faustbrian/go-rule-engine/adapters/temporal/v2@v2.0.0
```

Import the canonical root-v2 module path directly:

```go
import ruleengine "github.com/faustbrian/go-rule-engine/v2"
```

## Quick start

```go
country := ruleengine.MustPath("shipment", "country")
set := ruleengine.RuleSet{ID: "routing", Rules: []ruleengine.Rule{{
    ID: "finland",
    When: ruleengine.Compare(ruleengine.OpEqual,
        ruleengine.Variable(country),
        ruleengine.Literal(ruleengine.String("FI"))),
}}}
plan, diagnostics, err := ruleengine.NewCompiler(
    ruleengine.DefaultLimits(),
).Compile(context.Background(), set)
```

See the executable [package example](example_test.go), the
[quick start](docs/quickstart.md), and the [JSON AST fixture](jsonast/testdata/location-routing.json).

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-rule-engine/v2` | Construct typed facts and rules, compile immutable bounded plans, and evaluate them deterministically. |
| `github.com/faustbrian/go-rule-engine/v2/jsonast` | Parse and marshal the versioned core JSON AST without adapter-specific operators. |
| `github.com/faustbrian/go-rule-engine/adapters/math/v2` | Exact-decimal comparison operators for root v2. |
| `github.com/faustbrian/go-rule-engine/adapters/measurement/v3` | Exact compatible-unit comparison operators for root v2. |
| `github.com/faustbrian/go-rule-engine/adapters/temporal/v2` | Exact instant and period relation operators for root v2. |

The historical root-v1-compatible adapter paths are `adapters/math`,
`adapters/measurement/v2`, and `adapters/temporal` beneath
`github.com/faustbrian/go-rule-engine`. They expose root-v1 types and cannot be
composed directly with root v2. Use the published successor paths above for
root v2 applications.

The compiler and compiled plans are caller-owned values. The root module
performs no hidden I/O, starts no background work, and owns no runtime resource
that requires shutdown. Applications own any resolver I/O and its lifecycle.

## Guarantees

- Missing and null are distinct typed values; there is no truthiness or
  implicit coercion. A rule set may opt into treating absent operands as
  null or as Kleene unknown.
- Priorities sort descending and equal priorities sort by rule ID ascending,
  unless a rule set opts into declared order.
- Decimal values compare exactly by numeric value, never through floats.
- Logical operands evaluate left to right and short-circuit deterministically.
- Compilation rejects duplicate IDs, unknown operators, incompatible literal
  types, dependency cycles, non-finite floats, unsafe regexes, and every
  configured bound violation.
- Evaluation bounds time, iterations, derived facts, explanations, and errors.
- Canonical JSON and SHA-256 hashes are stable for equivalent definitions.
- Built-in plans and contexts are immutable and safe for concurrent reuse.

## Documentation

- [Documentation index](docs/README.md)
- [Model](docs/model.md)
- [Operators](docs/operators.md)
- [Types and coercion](docs/types-and-coercion.md)
- [Compilation](docs/compilation.md)
- [Evaluation](docs/evaluation.md)
- [Rule sets](docs/rule-sets.md)
- [Extensions](docs/extensions.md)
- [JSON AST](docs/json-ast.md)
- [Limits](docs/limits.md)
- [Security](docs/security.md)
- [Performance](docs/performance.md)
- [Migration](docs/migration.md)
- [Integration](docs/integration.md)
- [Cookbook](docs/cookbook.md)
- [FAQ](docs/faq.md)
- [Compatibility](docs/compatibility.md)
- [Support](SUPPORT.md)
- [Security reporting](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)

## Verification

`GOWORK=off go test ./...` runs standalone deterministic tests and fuzz seeds
in the selected module. `make check` and `make ci` use the public Tools v2.1.1
verifier's bounded local route, including the selected context/cache/operator
race; changed-source CI adds causally selected fuzz smoke. `make check-full`
retains the original full coverage, race, mutation, nine native fuzz targets,
integration, API, security, docs and benchmark campaigns. See
[contributor routes](CONTRIBUTING.md#verification) for focused active fuzzing
and optional local mutation.

Full root coverage collects real profiles in evidence mode; optional adapters
retain exact coverage. Counts do not certify behavioral adequacy or release
readiness. See the [tooling route](CONTRIBUTING.md#repository-tooling).

The module requires Go 1.27.0 and has no runtime dependencies.
Exact decimal, temporal-period, and measurement adapters live in isolated
nested modules described in the [extension guide](docs/extensions.md), so core
consumers do not inherit their dependency graphs.

For shared construction, ownership, lifecycle, and composition guidance, see
the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Domain utilities family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

## License

MIT. See [LICENSE](LICENSE).

## Documentation navigation

[Complete documentation map](docs/README.md#complete-documentation-map) · [Current source package map](docs/package-map.md)

Use the [versioned ecosystem entry](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md) for package
selection. The [frozen compatibility cohort](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/compatibility-sets.md)
is separate from current package releases and unpublished source changes.
Adopting them does not silently update that cohort.
