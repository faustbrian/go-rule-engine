# Evaluation

[Documentation index](README.md) · [Repository entry](../README.md)

Root v2 `EvaluateResolved` uses one `EvaluationTimeout` deadline for fact
resolution and subsequent evaluation. Earlier caller deadlines and context
values are preserved. Trusted resolvers and operators must honor cancellation;
the engine does not detach or forcibly preempt callbacks.

Evaluation applies the compiled order to one immutable context. The result
contains `Decision`, ordered matched rule IDs, a bounded explanation, bounded
redacted errors, duration, and derived facts.

`Matched` means at least one selected rule matched and no evaluation error
occurred. `Unmatched` means no selected rule matched. `Indeterminate` means
cancellation, timeout, operator failure, conflict, or a resource bound stopped
a reliable decision. Integrations must not treat `Indeterminate` as matched.

Forward chaining is literal and monotonic: a match may add an absent derived
fact. Re-deriving the same value is stable; changing an existing value is a
conflict. The compiler rejects static dependency cycles and evaluation stops
at `MaxIterations` even when a custom predicate obscures dependencies.

`EvaluateResolved` requests only required missing paths, in lexical order,
through the explicit resolver passed by the caller. Ordinary `Evaluate` never
performs I/O through built-ins. Custom predicates and operators can perform
caller-owned I/O and are not sandboxed. Errors preserve partial matches and
derivations; explanation truncation is silent, and a terminal error can append
one entry beyond the predicate-error budget. Cancellation is checked before
owned steps, not after every callback; a final callback can return a match
after canceling its context. Do not infer a hard execution-duration bound.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
