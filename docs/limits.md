# Limits

[Documentation index](README.md) · [Repository entry](../README.md)

`Limits` independently bounds rule count, fact count, AST depth, operands,
collection size, string bytes, decimal text bytes and exponent, JSON definition bytes, regex bytes, identifier
bytes, tags, tag bytes, path bytes, path segments, chaining iterations,
derived facts, diagnostics, explanation entries, and evaluation duration.

Start with `DefaultLimits` and reduce budgets for the owning service. Every
field must be positive, except `MaxDecimalBytes` and `MaxDecimalExponent`:
zero rejects every decimal, so `Limits` values written before decimals existed
remain valid. `MaxDecimalExponent` bounds the power of ten of a decimal's
leading digit (default 6144) and `MaxDecimalBytes` its text (default 258).
`Decimal` itself refuses exponents with more than 18 significant digits before
any budget applies. Limits are captured in a compiler and copied into its
plans. They do not mutate globally.

Definition byte limits apply before JSON decoding. Path and value limits apply
during construction. AST and operator limits apply during compilation.
Iteration, explanation and predicate-error limits apply during evaluation.
A terminal error may add one entry beyond MaxDiagnostics. MaxFacts applies at
context construction, not to accumulated resolver/derived facts. Construct
paths with the intended NewPath limits; existing paths are not size-rechecked
by the compiler or context constructor. Timeouts require callback cooperation.
In root v2, `EvaluationTimeout` also bounds the total `CompileCached`
operation and the combined resolver/evaluation operation. The caller's earlier
deadline takes precedence; callbacks remain trusted and must cooperate.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
