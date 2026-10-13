# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-rule-engine/v2

`rule-engine` is a deterministic, typed, inspectable engine for evaluating facts and propositions. It compiles immutable execution plans, evaluates them concurrently without hidden I/O, supports bounded forward chaining, and emits redacted diagnostics and explanations.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-rule-engine/v2`](../doc.go) | Use to construct typed facts and rules, compile bounded immutable execution plans, evaluate them deterministically, and encode canonical definitions. |
| [`github.com/faustbrian/go-rule-engine/v2/jsonast`](../jsonast/jsonast.go) | Use for the compatibility facade that parses and marshals the core versioned JSON AST without importing adapter-specific operators. |

## github.com/faustbrian/go-rule-engine/adapters/math/v2

The `rule-engine/adapters/math` module is the optional bridge between `math/decimal` and `rule-engine`. It encodes finite decimals as tagged string values and supplies deterministic equality and ordering operators. The core rule engine does not depend on the math module.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-rule-engine/adapters/math/v2`](../adapters/math/decimal.go) | Use when rule definitions need exact go-math decimal values and explicitly registered comparison operators. |

## github.com/faustbrian/go-rule-engine/adapters/measurement/v3

`rule-engine/adapters/measurement` is an optional bridge between immutable `measurement.Quantity` values and the rule engine's explicitly registered custom operators. It compares compatible dimensions with `measurement.ExactConversion()` and never supplies a unit or rounding policy.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-rule-engine/adapters/measurement/v3`](../adapters/measurement/measurement.go) | Use when rule definitions need exact go-measurement quantities and explicitly registered dimension-safe comparison operators. |

## github.com/faustbrian/go-rule-engine/adapters/temporal/v2

The `rule-engine/adapters/temporal` module is the optional bridge between exact `temporal/instant` periods and the deterministic custom-operator boundary in `rule-engine`. It encodes instants and periods as tagged strings so the core engine does not acquire a temporal dependency.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-rule-engine/adapters/temporal/v2`](../adapters/temporal/temporal.go) | Use when rule definitions need exact go-temporal instants and periods with explicitly registered relation operators. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
