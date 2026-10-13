# Quick start

[Documentation index](README.md) · [Repository entry](../README.md)

Create paths and values explicitly, build an immutable `Context`, compile a
`RuleSet`, then reuse the `Plan` for any number of evaluations. Always inspect
the compile error and diagnostics before retaining a plan. Treat
`Indeterminate` as an error state rather than a non-match.

Use `CollectAll` when every matching rule matters, `FirstMatch` when only the
highest ordered match matters, and `ErrorOnMultiple` when simultaneous matches
indicate an invalid definition. The zero-value strategy is `FirstMatch`.

The complete runnable flow is in [example_test.go](../example_test.go).

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
