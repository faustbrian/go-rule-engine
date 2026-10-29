# Types and coercion

[Documentation index](README.md) · [Repository entry](../README.md)

The engine never converts strings to numbers, integers to floats, timestamps
to strings, empty values to false, or null to missing. Integrations must
normalize data before constructing facts.

Integers are signed 64-bit values. Floats must be finite; NaN and infinities
are rejected recursively. Times compare as instants after monotonic clock data
is removed. Durations compare as nanoseconds. Strings compare by Go's stable
byte ordering over valid UTF-8. Lists use ordered structural equality.

Decimals are a core kind. `Decimal(text)` accepts exactly the RFC 8259 JSON
number grammar: an optional minus, no leading zeros, an optional fraction and
an optional exponent. `-0` is accepted and equals `0`. No float conversion
takes place; equality and ordering compare the exact value, while
`DecimalText` and canonical JSON keep the original spelling, so `1.5` and
`1.50` are equal but hash differently. Decimals are never equal to ints,
floats or strings and cross-kind comparisons are type mismatches.

Use explicit domain adapters for measurements and for go-math decimal
arithmetic. An adapter should encode the domain value canonically and register
an operator with exact signatures; it must not add implicit core coercion.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
