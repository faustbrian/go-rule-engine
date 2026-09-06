# Performance and troubleshooting

## Benchmarking

`BenchmarkQuantityComparison` measures tagged-value parsing, exact compatible-
unit conversion, and comparison through the adapter. Run it from this module:

```sh
go test -run '^$' -bench '^BenchmarkQuantityComparison$' -benchmem -count=10
```

Record the CPU, operating system, Go version, benchmark duration, corpus, and
sample count. Compare equivalent runs with `benchstat`. This benchmark includes
adapter decoding and conversion; do not compare it directly with an already-
typed quantity comparison or publish one machine's result as a portable limit.

Keep canonical quantities at application boundaries and avoid repeatedly
encoding the same validated value on a hot path. Select stricter limits for
hostile persisted input, but measure with the same limits used in production.

## Troubleshooting

### Why does a compatible-unit comparison fail?

Compatibility is necessary but not sufficient. The exact conversion must also
terminate within the configured decimal and measurement limits. Inspect the
stable adapter error and its wrapped cause with `errors.Is` or `errors.As`;
do not fall back to rounding because that changes the comparison contract.

### Why is a quantity tag rejected?

Use `Quantity` to construct the value. Hand-built tags fail when the amount or
unit is malformed, noncanonical, unknown, oversized, or uses an unsupported
tag version. The rejected contents are intentionally absent from errors.

### Why does the built-in string operator give the wrong ordering?

Tagged quantities are persisted as strings, but their byte order is not a
measurement order. Register this adapter on the compiler and use its versioned
quantity operators.
