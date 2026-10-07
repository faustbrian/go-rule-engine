# Migration from Shipit and Cline Ruler

## Root v2 and pending adapter majors

Root v2.0.0 uses `github.com/faustbrian/go-rule-engine/v2`. Math adapter
v2.0.0, measurement adapter v3.0.0 and temporal adapter v2.0.0 remain pending
publication. Select only released tags available through the public proxy.
Applications using adapters must keep their root-v1-compatible module set
until every required successor is published; then update exposed root and
adapter types together.

Delivery is ordered root first. This source keeps the delivered math v1.0.1,
measurement v2.0.1 and temporal v1 adapters, and the competitor module,
on their existing public root v1 dependency. Their complete nominal migration
is retained in commit `6668aa1721dd38988e42e1106a89a9e25bd184d4` and resumes
after actual root v2 publication. No local replacement or unpublished module
proxy substitutes for that public dependency. Successor major versions remain
required before adopting root v2 types in adapter signatures.

| Module | Baseline before migration | Successor import path |
| --- | --- | --- |
| Root | v1.0.0 | `github.com/faustbrian/go-rule-engine/v2` |
| Math adapter | v1.0.1 | `github.com/faustbrian/go-rule-engine/adapters/math/v2` |
| Measurement adapter | v2.0.1 | `github.com/faustbrian/go-rule-engine/adapters/measurement/v3` |
| Temporal adapter | v1.0.0 | `github.com/faustbrian/go-rule-engine/adapters/temporal/v2` |

The adapters retain Math v1.1.2, Measurement v2.0.1 and Temporal v1.1.0;
only the Rule Engine nominal types change. Historical API snapshots and
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
