# rule-engine

[![CI](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-rule-engine/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-root_evidence-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-rule-engine.svg)](https://pkg.go.dev/github.com/faustbrian/go-rule-engine)
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

Root v2.0.0 is published on Go 1.27.0. Published root v1.0.0 remains available
under its original import path. This source prepares the three optional
adapter successor majors against the public root v2 dependency because their
public signatures expose root types. Their releases remain pending.
See [migration](docs/migration.md) before adopting root v2 or its pending
adapter successors. Source on main alone does not establish publication.

## Install

Select the published root-v2 release through the public Go proxy. The three
adapter successors remain separate pending releases:

```sh
go get github.com/faustbrian/go-rule-engine/v2@v2.0.0
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
| `github.com/faustbrian/go-rule-engine/adapters/math/v2` | Pending successor, not currently available: exact-decimal comparison operators for root v2. |
| `github.com/faustbrian/go-rule-engine/adapters/measurement/v3` | Pending successor, not currently available: exact compatible-unit comparison operators for root v2. |
| `github.com/faustbrian/go-rule-engine/adapters/temporal/v2` | Pending successor, not currently available: exact instant and period relation operators for root v2. |

The currently published adapter paths are `adapters/math`,
`adapters/measurement/v2`, and `adapters/temporal` beneath
`github.com/faustbrian/go-rule-engine`. They expose root-v1 types and cannot be
composed directly with root v2. The current nested modules use the successor
paths above; source on main does not make their pending tags available.

The compiler and compiled plans are caller-owned values. The root module
performs no hidden I/O, starts no background work, and owns no runtime resource
that requires shutdown. Applications own any resolver I/O and its lifecycle.

## Guarantees

- Missing and null are distinct typed values; there is no truthiness or
  implicit coercion.
- Priorities sort descending and equal priorities sort by rule ID ascending.
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

`make check` runs formatting, module hygiene, vet, static analysis, lint,
tests, production coverage under the selected module policy, race tests,
fuzz smoke tests,
mutation tests, benchmarks, documentation checks, API compatibility,
security policy checks, vulnerability scanning, and workflow validation.

The root collects real coverage profiles in evidence mode; optional adapters
retain exact coverage. Counts do not certify behavioral adequacy or release
readiness. CI currently uses pinned development tooling source, not a published
v2 binary; see the [tooling route](CONTRIBUTING.md#repository-tooling).

The module requires Go 1.27.0 and has no runtime dependencies.
Exact decimal, temporal-period, and measurement adapters live in isolated
nested modules described in the [extension guide](docs/extensions.md), so core
consumers do not inherit their dependency graphs.

For shared construction, ownership, lifecycle, and composition guidance, see
the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Domain utilities family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

## License

MIT. See [LICENSE](LICENSE).
