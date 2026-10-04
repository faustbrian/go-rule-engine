# Security

Treat rule definitions and facts as untrusted input. Parse with service-specific
limits, compile before storage or activation, retain the canonical hash, and
evaluate only immutable plans. Never interpolate diagnostics into logs with
raw request facts.

The engine has no eval, JavaScript runtime, dynamic Go loading, reflection-based
model discovery, arbitrary method calls, database access, action execution,
global registry, or background goroutines. Regexes use Go RE2 and literal
patterns. All ordinary evaluation is process-local and side-effect free.

Custom operators, predicates, resolvers, and caches are explicit trust
boundaries. Review them for deterministic output, bounded work, cancellation,
concurrency safety, I/O, and value disclosure. Do not register a custom
predicate in authorization or feature-flag paths unless its failure maps to
the owning product's fail-closed state.

The versioned [threat model](threat-model.md) distinguishes current controls,
caller-owned risks and unresolved cache/context assurance boundaries.

Report vulnerabilities according to [SECURITY.md](../SECURITY.md).
