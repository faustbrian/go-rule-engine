# Exact decimal operators for rule-engine

The `rule-engine/adapters/math` module is the optional bridge between
[`math/decimal`](https://pkg.go.dev/github.com/faustbrian/go-math/decimal) and
[`rule-engine`](../..). It encodes
finite decimals as tagged string values and supplies deterministic equality and
ordering operators. The core rule engine does not depend on the math module.

## Install

```sh
go get github.com/faustbrian/go-rule-engine/adapters/math@v1
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

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/math)
- [Parent package documentation](../../docs/README.md)

## Compatibility and support

This module follows Semantic Versioning. Report vulnerabilities through the
[parent security policy](../../SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
