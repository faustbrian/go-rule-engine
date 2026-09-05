# rule-engine/adapters/temporal

The `rule-engine/adapters/temporal` module is the optional bridge between exact
`temporal/instant` periods and the deterministic custom-operator boundary in
`rule-engine`. It encodes
instants and periods as tagged strings so the core engine does not acquire a
temporal dependency.

The adapter does not read a clock, resolve named time zones, perform calendar
arithmetic, register global state, start goroutines, schedule work, or perform
I/O.

This stable, independently released module requires Go 1.26.6 or newer. Its
releases use `adapters/temporal/v*` tags.

## Install

```sh
go get github.com/faustbrian/go-rule-engine/adapters/temporal@v1
```

Import the canonical module path directly:

```go
import ruleenginetemporal "github.com/faustbrian/go-rule-engine/adapters/temporal"
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

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-rule-engine/adapters/temporal` | Encode exact instants and periods and register deterministic relation operators. |

This module has no public subpackages. `Instant` and `Period` return immutable
tagged rule values, while `Operators` returns fresh caller-owned operator and
signature slices. Importing the module performs no initialization. It owns no
resource or background work and therefore exposes no shutdown operation.

## When to use it

Use this adapter when rules must persist exact instants or bounded periods and
evaluate set relations or instant membership. Resolve civil dates, named time
zones, recurrence, calendars, scheduling, and clock policy before encoding.
Call `go-temporal/instant` directly when no rule-engine serialization boundary
is needed.

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

Malformed, oversized, or unsupported persisted values fail without echoing
their contents, and caller cancellation is preserved. The adapter does not
model leap seconds, unbounded periods, or dates outside the four-digit RFC 3339
year range.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/temporal)
- [Executable example](example_test.go)
- [FAQ and troubleshooting](docs/reference.md#faq)
- [Performance and verification](docs/reference.md#performance-and-verification)
- [Changelog](CHANGELOG.md)
- [Support](../../SUPPORT.md)
- [Contributing](../../CONTRIBUTING.md)
- [Security reporting](../../SECURITY.md)
- [License](LICENSE)
- [Parent package documentation](../../docs/README.md)

## Compatibility and support

This module follows Semantic Versioning. Report vulnerabilities through the
[parent security policy](../../SECURITY.md).

The instant and period tags, separators, bounds, operator names, and relation
semantics form the v1 persistence contract. The module exports no dedicated
test helper; tests can construct fresh operators and exact `time.Time` or
`instant.Period` values without global setup.

Shared package-selection and ownership guidance is in the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md).

## License

MIT. See [LICENSE](LICENSE).
