# rule-engine/adapters/temporal

The `rule-engine/adapters/temporal` module is the optional bridge between exact
`temporal/instant` periods and the deterministic custom-operator boundary in
`rule-engine`. It encodes
instants and periods as tagged strings so the core engine does not acquire a
temporal dependency.

The adapter does not read a clock, resolve named time zones, perform calendar
arithmetic, register global state, start goroutines, schedule work, or perform
I/O.

## Install

```sh
go get github.com/faustbrian/go-rule-engine/adapters/temporal@v1
```

## Quick start

```go
start := time.Date(2026, time.July, 19, 10, 0, 0, 0, time.UTC)
window, err := instant.New(start, start.Add(time.Hour), temporal.ClosedOpen)
if err != nil {
    return err
}
windowValue, err := ruleenginetemporal.Period(window)
if err != nil {
    return err
}
pointValue, err := ruleenginetemporal.Instant(start.Add(30 * time.Minute))
if err != nil {
    return err
}

compiler, err := ruleengine.NewCompilerWithOperators(
    ruleengine.DefaultLimits(),
    ruleenginetemporal.Operators()...,
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
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/temporal)
- [Parent package documentation](../../docs/README.md)

## Compatibility and support

This module follows Semantic Versioning. Report vulnerabilities through the
[parent security policy](../../SECURITY.md).

Shared package-selection and ownership guidance is in the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.3.0/docs/ecosystem/README.md).

## License

MIT. See [LICENSE](LICENSE).
