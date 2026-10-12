# Compilation

`Compiler.Compile` checks cancellation, limits, metadata, duplicate IDs,
predicate shape, operator existence, literal types, regex syntax, derived fact
validity, duplicate derived paths, and forward-chaining cycles. It then copies
and orders rules by descending priority and ascending ID. Custom predicates
are opaque to shape/dependency analysis, and custom state is not deep-copied.
Existing paths are not rechecked against compiler path-size budgets.

The returned `Plan` has no exported mutable state. Built-in predicates are
immutable. A custom predicate or operator is part of the caller's trust
boundary and must itself be deterministic, immutable, concurrency-safe, and
responsive to context cancellation.

Diagnostics contain identifiers, stable codes, severity, and redacted
messages. They never include literal or fact values. A compile error means no
plan is usable.
