# JSON AST

[Documentation index](README.md) · [Repository entry](../README.md)

The `jsonast` package implements grammar versions `1` and `2`. The root object
contains `version`, `id`, optional `namespace`, `strategy`, and `rules`.
Version `2` adds optional root `order` (`priority` or `declared`), optional
root `missing` (`false`, `null` or `unknown`) and the `decimal` value type.
Version `1` documents that use any of these are rejected. Canonical output
omits default order and missing mode and writes version `2` only when a
non-default order, a non-default missing mode or a decimal value is present,
so every definition expressible in version `1` keeps its bytes and hash. Every rule has
`id`, optional `namespace`, `priority`, `tags`, `when`, and `derive`.

Predicate `kind` is one of `true`, `false`, `exists`, `compare`, `all`, `any`,
or `not`. Operands are `variable` with a segment array or `literal` with one
typed value. Values use `type` plus exactly one matching field. Null and
missing have no value field. Times use RFC 3339 with nanoseconds; durations
use signed nanoseconds. Decimals use a JSON string holding the original
number text, for example `{"type":"decimal","decimal":"1.50"}`.

Unknown fields, fields belonging to another variant, trailing JSON values,
unknown operators, invalid kinds, invalid paths, and every limit violation are
rejected. Unknown-field handling retains encoding/json duplicate-name and
case-folding behavior; it is not complete required-member schema validation.
The byte budget precedes decoding; variant/depth checks run on decoded DTOs.
Source values are not copied into diagnostics.

Canonical output fixes field order, sorts rules by evaluation order (input
order under `declared`), sorts and
deduplicates tags, sorts derived facts by path, normalizes times to UTC, and
uses stable typed representations. Logical child order and list order remain
unchanged because they are semantic.

The maintained grammar fixture is
[location-routing.json](../jsonast/testdata/location-routing.json). JMESPath,
SQL-like, and GraphQL-like DSLs are not implemented; accepting those strings
would imply undocumented grammar and security behavior.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
