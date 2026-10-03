# Threat model

Version 1 (2026-10-03)

## Scope and assets

This model describes the current root v1 implementation, not unpublished v2
work. Assets are decision integrity, availability, fact confidentiality and
canonical rule identity. Independently released adapters keep their own module
versions; the measurement adapter already has a v2 release.

The engine is not an authorization service, workflow runner or sandbox.
Integrating products own decision effects and fail-closed handling of
`Indeterminate`. Fact `Owner` values describe provenance, not permissions.

## Inputs and trust boundaries

Definitions, identifiers, paths, operands and facts are application input.
Custom predicates, operators, resolvers and caches are explicitly trusted
application collaborators, not isolated untrusted code. The core performs no
implicit network, database or process operations and starts no background work.

| Boundary | Current enforcement | Source |
| --- | --- | --- |
| JSON to definition | Byte admission before decoding, unknown-field rejection, version checks and subsequent typed compilation | [Serialization](../serialization.go) |
| Definition to plan | Positive limits, rule and operand counts, depth, identifier/tag checks, duplicate and cycle rejection | [Compiler](../compiler.go), [limits](../limits.go) |
| Facts to snapshot | Fact count, value type, string/list bounds and owned copies | [Context](../context.go) |
| Plan to decision | Evaluation timeout, cancellation checkpoints, iteration and explanation limits | [Plan](../plan.go) |
| Regular expressions | Literal patterns, pattern-byte checks and Go's standard regexp implementation | [Operators](../operator.go) |
| Diagnostics | Engine-produced error messages omit fact values; explanations contain rule identifiers and results | [Errors](../errors.go), [plan](../plan.go) |
| Cached plans | Canonical definition-hash comparison; bounded entry count in the owned memory cache | [Cache](../cache.go) |

These controls are not a claim that every allocation, callback or traversal is
preemptible or that all ecosystem security acceptance criteria are complete.

## Caller-owned residual risks

| Risk and rationale | Owner and mitigation | Review condition |
| --- | --- | --- |
| Arbitrary callbacks can block, panic, mutate external state or disclose data; an in-process call cannot be forcibly canceled safely. | Application/adapter owner: use trusted bounded implementations, enforced I/O deadlines and an appropriate panic/isolation boundary. | Any new callback or I/O-backed adapter. |
| A regexp match and immutable-value copying are synchronous work between cancellation checkpoints. | Core maintainers and integrators: keep input and collection limits appropriate to the service, and bound concurrent evaluations. | Changes to limits, Go regexp behavior or workload size. |
| A product can interpret `Indeterminate` as success or log raw facts and caller-controlled identifiers. | Product owner: fail closed at security-sensitive consumers and sanitize observability at its own boundary. | Each new decision consumer or logging integration. |

## Unresolved assurance boundaries

`EvaluationTimeout` is created inside `Evaluate`. In current v1,
`EvaluateResolved` calls the resolver with the caller's context before that
timeout begins. `CompileCached` also passes the caller's context to cache
operations. Integrators need an explicit deadline covering those operations;
the evaluation timeout alone does not bound them.

Cache admission checks a definition hash, not complete compiler-limit equality.
Use separate caches for distinct compiler policies; do not reuse cached plans
after tightening a policy. Canonical serialization also compiles with a
background context. Compiler-aware cache admission and broader bounded-context
handling remain pending behavioral work, not accepted as repaired by this
document. Core maintainers own their verification and release disposition.

Review this model when parser, cache identity, limits, callbacks, adapters,
dependencies or security-sensitive consumers change. Scanner results, runtime
regressions and release/consumer evidence remain separate requirements.
Report suspected vulnerabilities through [SECURITY.md](../SECURITY.md).

## Version 2 controls (2026-10-03, unpublished source)

The Version 1 analysis above remains the historical released-source model.
Planned root v2 checks complete `Limits` equality when admitting cached plans,
uses compiler-aware canonicalization, and passes one bounded operation context
through canonicalization/cache compilation and resolver/evaluation. Earlier
caller deadlines and values are preserved. Checkpoints stop subsequent owned
steps after cancellation; completed callback side effects are not rolled back.
Public canonical helpers retain background-context behavior.

Custom registries are not identified by hash plus limits. The application owns
cache isolation per compatible compiler/operator registry, invalidates entries
when callback semantics change, and reviews that isolation for every registry
or cache integration. Trusted callbacks still own prompt cancellation, bounded
I/O and data disclosure; synchronous encoding/hash/value work is not forcibly
preemptible. Core maintainers own exact-source CI and publication. Local tests
and task-local adapter composition are not public release/consumer evidence.

New adapter majors are math v2, measurement v3 and temporal v2 because public
signatures expose root v2 types. Published adapter algorithms and domain
dependency versions are retained. Review this model after changes to callback
registries, cache sharing, limits or adapter dependencies.
