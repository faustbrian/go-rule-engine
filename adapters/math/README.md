# Exact decimal operators for rule-engine

The `rule-engine/adapters/math` module is the optional bridge between
[`math/decimal`](https://pkg.go.dev/github.com/faustbrian/go-math/decimal) and
[`rule-engine`](../..). It encodes
finite decimals as tagged string values and supplies deterministic equality and
ordering operators. The core rule engine does not depend on the math module.

This stable, independently released module requires Go 1.26.6 or newer. Its
releases use `adapters/math/v*` tags.

## Install

```sh
go get github.com/faustbrian/go-rule-engine/adapters/math@v1
```

Import the canonical module path directly:

```go
import ruleenginemath "github.com/faustbrian/go-rule-engine/adapters/math"
```

## Quick start

```go
operators := ruleenginemath.Operators()
compiler, err := ruleengine.NewCompilerWithOperators(
	ruleengine.DefaultLimits(),
	operators...,
)
if err != nil {
	return err
}

minimum := ruleenginemath.Decimal(decimal.MustParse("10.00"))
predicate := ruleengine.Compare(
	ruleenginemath.OpDecimalGreaterOrEqual,
	ruleengine.Variable(ruleengine.MustPath("order", "total")),
	ruleengine.Literal(minimum),
)
```

The compiling examples in this module contain complete imports and setup.

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-rule-engine/adapters/math` | Encode exact decimals and register the five versioned decimal comparison operators. |

This module has no public subpackages. The application owns the returned
operator slice and explicitly registers it on each compiler that needs decimal
support. Importing the module performs no I/O, changes no global state, starts
no goroutine, and acquires no resource that requires shutdown.

## When to use it

Use this adapter when rules already model values as exact `go-math/decimal`
values and their tagged representation must survive canonical rule
serialization. Use the rule engine's built-in integer or float operators when
those value kinds express the complete domain contract. Do not use tagged
decimal strings with built-in string ordering.

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

Invalid tags and noncanonical values have stable error categories. Decimal
syntax and limit causes remain available through `errors.Is` and `errors.As`,
and caller cancellation is preserved. Rejected operand contents are not
included in adapter errors. See the guide before selecting limits for hostile
persisted input or publishing benchmark results.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/math)
- [Executable example](example_test.go)
- [Performance benchmark](benchmark_test.go)
- [FAQ and troubleshooting](docs/reference.md#faq)
- [Exactness and limits](docs/reference.md#exactness-and-limits)
- [Changelog](CHANGELOG.md)
- [Support](../../SUPPORT.md)
- [Contributing](../../CONTRIBUTING.md)
- [Security reporting](../../SECURITY.md)
- [License](LICENSE)
- [Parent package documentation](../../docs/README.md)

## Compatibility and support

This module follows Semantic Versioning. Report vulnerabilities through the
[parent security policy](../../SECURITY.md).

The module exports no dedicated test helper. Tests can construct fresh
operators with `Operators` or `OperatorsWithLimits` and use the root rule
engine's ordinary compiler and context APIs without global setup.

Shared package-selection and ownership guidance is in the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md).

## License

MIT. See [LICENSE](LICENSE).
