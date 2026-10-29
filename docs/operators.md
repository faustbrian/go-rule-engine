# Operators

[Documentation index](README.md) · [Repository entry](../README.md)

All operators require exact kinds. Missing operands return false unless the
rule set selects another [missing mode](rule-sets.md). Null is only comparable
with null through equality. There is no coercion.

| Operator | Left | Right | Meaning |
| --- | --- | --- | --- |
| `equal`, `not_equal` | any same kind | same kind | structural equality |
| `less_than`, `less_or_equal` | int, float, string, time, duration, decimal | same kind | total ordering |
| `greater_than`, `greater_or_equal` | int, float, string, time, duration, decimal | same kind | total ordering |
| `in`, `not_in` | any | list | exact element membership |
| `contains` | string or list | string or any | substring or exact element |
| `starts_with`, `ends_with` | string | string | byte-exact UTF-8 affix |
| `matches` | string | string literal | Go RE2 regular expression |

Regex patterns are validated during compilation, limited by `MaxRegexBytes`,
and use Go's linear-time regular-expression implementation. Dynamic patterns
are rejected.

Decimal equality, membership and ordering compare exact numeric values, so
`1.0` equals `1.00` and `0.10` is less than `0.2`. A decimal never equals an
int, float or string.

`All`, `Any`, and `Not` are logical combinators. Children execute left to
right. `All` stops on false; `Any` stops on true. Under `MissingUnknown` an
unknown child does not stop either combinator. Operator truth tables are
executable in [operator_truth_table_test.go](../operator_truth_table_test.go).

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
