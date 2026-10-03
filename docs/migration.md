# Migration from Shipit and Cline Ruler

## Pending family major versions

Root v2.0.0, math adapter v2.0.0, measurement adapter v3.0.0 and temporal
adapter v2.0.0 are pending publication. Current users must keep published
versions until the successor modules are available through the public proxy.
After publication update all exposed root and adapter types together:

| Module | Published baseline | Pending import path |
| --- | --- | --- |
| Root | v1.0.0 | `github.com/faustbrian/go-rule-engine/v2` |
| Math adapter | v1.0.1 | `github.com/faustbrian/go-rule-engine/adapters/math/v2` |
| Measurement adapter | v2.0.1 | `github.com/faustbrian/go-rule-engine/adapters/measurement/v3` |
| Temporal adapter | v1.0.0 | `github.com/faustbrian/go-rule-engine/adapters/temporal/v2` |

The adapters retain Math v1.1.2, Measurement v2.0.1 and Temporal v1.0.0;
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
