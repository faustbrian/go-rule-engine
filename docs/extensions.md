# Extensions

[Documentation index](README.md) · [Repository entry](../README.md)

`Operator` exposes a name, exact kind signatures, and a context-aware evaluate
method. An absent operand skips the operator unless the rule set uses
`MissingNull`; the operator then receives `Null()` and must declare a matching
null signature or the evaluation fails with a type mismatch. A redacted
evaluation error exposes an underlying engine code through `Error.Cause`.
Operators that read list operands should use `Value.ListLen` and
`Value.ListAt`, which do not allocate; `ListValue` copies the whole list on
every call. Registries belong to a compiler instance. Built-in names and duplicate
custom names cannot be replaced, and no process-global registry exists.
Use that compiler's `MarshalCanonical`, `CanonicalHash`, and `ParseJSON` methods
for stored definitions containing custom operator names; package-level
persistence helpers validate against built-ins only.

`PredicateFunc` is available for bounded logic that cannot be represented by
built-ins. It cannot be serialized, and its fact dependencies are opaque to
static cycle analysis. Avoid it for stored definitions.

`FactResolver` receives one explicit path at a time through
`EvaluateResolved`. Resolver failures are redacted. Implementations must avoid
side effects and return stable snapshots.

`PlanCache` stores immutable plans by canonical hash. `MemoryPlanCache` is a
bounded, locked LRU implementation. Cache failures stop compilation; entries
whose embedded hash or complete compiler limits differ are ignored.
`CompileCached` uses one `EvaluationTimeout` context across compiler-aware
canonicalization, cache reads, compilation and cache writes. An earlier caller
deadline is preserved. Cache implementations must honor cancellation; the
engine cannot preempt trusted callbacks. Cancellation observed after `Put`
returns an error but does not roll back a cache write already performed.

Hash and limits do not identify custom operator implementations. Applications
must scope caches to one compatible compiler/operator registry and rebuild
them when callback semantics change; sharing a cache across different
registries is unsupported.

Authorization and feature-flag adapters must preserve their own fail-closed
behavior. They must map `Indeterminate` to denial or disabled state, never to
success.

Exact optional domain adapters are isolated nested modules:

- [`adapters/math`](../adapters/math) provides go-math decimal ordering over
  tagged strings; the core `Decimal` kind needs no adapter.
- [`adapters/temporal`](../adapters/temporal) provides period relations
  and instant membership.
- [`adapters/measurement`](../adapters/measurement) provides exact
  compatible-unit quantity ordering.

Each module uses tagged canonical string values plus namespaced typed
operators. Their dependencies are not present in the core module graph.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
