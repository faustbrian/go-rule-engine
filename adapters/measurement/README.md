# Exact measurement rule operators

`rule-engine/adapters/measurement` is an optional bridge between immutable
`measurement.Quantity` values and the rule engine's explicitly registered
custom operators. It compares compatible dimensions with
`measurement.ExactConversion()` and never supplies a unit or rounding policy.

This stable, independently released module requires Go 1.26.6 or newer. Its
releases use `adapters/measurement/v*` tags.

## Install

```console
go get github.com/faustbrian/go-rule-engine/adapters/measurement@v1
```

Import the canonical module path directly:

```go
import ruleenginemeasurement "github.com/faustbrian/go-rule-engine/adapters/measurement"
```

## Quick start

```go
operators := ruleenginemeasurement.Operators()
compiler, err := ruleengine.NewCompilerWithOperators(
    ruleengine.DefaultLimits(),
    operators...,
)

limit := measurement.MustNew(decimal.New(1), measurement.Kilogram)
operand := ruleenginemeasurement.Quantity(limit)
```

Register the returned operators on each compiler that needs them. Registration
is caller-owned; importing this module changes no global registry and starts no
goroutine or I/O.

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-rule-engine/adapters/measurement` | Encode exact quantities and register five compatible-dimension comparison operators. |

This module has no public subpackages. `Quantity` returns an immutable rule
value, and `Operators` returns fresh caller-owned operator and signature
slices. The module owns no resource, callback, cache, clock, or background work
and therefore exposes no shutdown operation.

## When to use it

Use this adapter when rule facts already carry validated exact quantities and
explicit units. Resolve aliases, localization, missing units, and rounding
policy before encoding. Do not use it for display conversion or compare its
tagged strings with the rule engine's built-in string operators.

## Guarantees

- quantities use the canonical `quantity:v1|<amount>|<unit>` string tag;
- the decimal amount and canonical measurement unit symbol are both explicit;
- all five relations use exact compatible-unit conversion;
- malformed, noncanonical, unknown, oversized, incompatible, and
  unrepresentable inputs fail without coercion;
- operator and signature slices are fresh values safe for concurrent reuse;
- cancellation is checked before and between bounded evaluation stages.

## Limitations

The adapter intentionally rejects exact conversions whose decimal result does
not terminate, even when a rounded result might be operationally useful.
Quantities remain strings to preserve the core rule engine's closed value-kind
model, so callers must use `Quantity` instead of constructing tags manually.

Invalid and incompatible quantities have separate stable error categories;
owned decimal, measurement, and context causes remain inspectable. Rejected
amounts and units are not copied into adapter diagnostics. The complete limits
and security guidance applies to persisted or otherwise hostile values.

## Documentation

- [Documentation index](docs/README.md)
- [Encoding](docs/encoding.md)
- [Conversion and dimensions](docs/conversion.md)
- [Operators and API](docs/api.md)
- [Limits](docs/limits.md)
- [Security](docs/security.md)
- [Examples](docs/examples.md)
- [Adoption](docs/adoption.md)
- [Compatibility](docs/compatibility.md)
- [Migration](docs/migration.md)
- [FAQ](docs/faq.md)
- [Troubleshooting](docs/faq.md)
- [Executable example](example_test.go)
- [Performance benchmark](benchmark_test.go)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/measurement)
- [Support](../../SUPPORT.md)
- [Contributing](../../CONTRIBUTING.md)
- [Security reporting](../../SECURITY.md)
- [License](LICENSE)
- [Parent package documentation](../../docs/README.md)

Release history is in the [Changelog](CHANGELOG.md).

Shared package-selection and ownership guidance is in the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md).

This module follows Semantic Versioning. Its v1 tag, operator names, exact
conversion policy, and error categories are compatibility contracts. The
module exports no dedicated test helper; tests can construct fresh operators
and use ordinary rule-engine compiler APIs without global setup.

## Development

```console
make check MODULES=adapters/measurement
```

## License

MIT.
See [LICENSE](LICENSE).
