# FAQ

## Is this an authorization engine?

No. It has no permit, deny, role, permission, principal, or combining defaults.

## Is missing equal to null?

No. Missing is returned only for absent paths. Null is an explicit supplied
value.

## Are numeric strings coerced?

No. Normalize them before constructing facts or reject the input.

## Can a rule call my model or database?

No. Supply facts directly or use the explicit resolver boundary before
evaluation. Arbitrary method calls and reflection discovery are unsupported.

## Can I register operators globally?

No. Operators belong to one compiler, making ownership and concurrency clear.

## Why can custom predicates not be serialized?

Function behavior has no canonical portable representation or inspectable
grammar. Use built-in AST nodes or a typed registered operator.

## Troubleshooting

### Why does compilation reject an operator?

The compiler only accepts built-in operators and the custom operators supplied
to that compiler. Register the adapter or custom operator explicitly, verify
its versioned name, and check that both operand kinds match its signature.

### Why does evaluation report a missing fact?

Missing is distinct from null. Confirm that the supplied context contains the
exact path used by the compiled rule and that a resolver returned a value
before its deadline. Do not substitute null unless null is the domain value.

### Why is a benchmark result different on another machine?

Benchmark results depend on CPU, operating system, Go version, corpus, and
sample duration. Compare equivalent work with the same environment and use
`benchstat`; see the [performance guide](performance.md).
