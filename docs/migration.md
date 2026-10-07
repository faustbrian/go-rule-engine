# Migration from Shipit and Cline Ruler

## Root v2 and published adapter majors

Root v2.0.0 uses `github.com/faustbrian/go-rule-engine/v2`. Math adapter
v2.0.0, measurement adapter v3.0.0 and temporal adapter v2.0.0 are published
through the public proxy. Applications using adapters must update exposed
root and adapter types together.

Delivery was ordered root first, followed by the three adapter major releases.
The competitor module also consumes the public root v2 dependency. Historical
math v1.0.1, measurement v2.0.1 and temporal v1.0.1 remain unchanged and expose
root v1 types. No local replacement or unpublished module proxy substitutes
for the public dependency. Adopt the published successor majors when updating
applications to root v2.

| Module | Baseline before migration | Successor import path |
| --- | --- | --- |
| Root | v1.0.0 | `github.com/faustbrian/go-rule-engine/v2` |
| Math adapter | v1.0.1 | `github.com/faustbrian/go-rule-engine/adapters/math/v2` |
| Measurement adapter | v2.0.1 | `github.com/faustbrian/go-rule-engine/adapters/measurement/v3` |
| Temporal adapter | v1.0.1 | `github.com/faustbrian/go-rule-engine/adapters/temporal/v2` |

The initial adapter major releases retained Math v1.1.2, Measurement
v2.0.1 and Temporal v1.1.0; only the Rule Engine nominal types changed.
The compatible math v2.0.1 and measurement v3.0.1 patches select Math
v1.1.3 without changing arithmetic behavior or the other domain dependencies. Historical API snapshots and
released tags remain unchanged. Rebuild caller-scoped caches when compiler
limits or operator implementations change. Verify all trusted callbacks
cooperate with the bounded context. Public `MarshalCanonical` and
`CanonicalHash` still use a background context; only internal cached
canonicalization receives the operation context. JSON grammar remains version 1.

Inventory every existing rule's identifier, input fields, missing/null
behavior, priority, conflict behavior, output, and failure policy. Do not begin
by translating syntax.

Map source fields to explicit paths and typed values. Replace implicit
truthiness and coercion with exact propositions. Preserve source priority and
choose a rule-set conflict strategy deliberately. Represent computed outputs
as bounded derived facts only when they are side-effect free.

For each source rule, keep fixtures for match, non-match, missing, null,
boundary, Unicode, and invalid input. Run the old and new evaluators against
the same normalized snapshot and compare decisions and selected rule IDs.
Record deliberate differences before activation.

Migrate in shadow mode by canonical hash. Observe mismatches without logging
sensitive facts. Cut over only after every production rule has differential
evidence and `Indeterminate` is wired to the owning service's safe state.
