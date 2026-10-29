# Rule sets

[Documentation index](README.md) · [Repository entry](../README.md)

By default (`OrderPriority`) rules are ordered by priority descending, then ID
ascending, and input slice order has no effect. `RuleSet.Order = OrderDeclared`
evaluates rules in input slice order instead; priority is then metadata that
is still validated and serialized. Tags and namespaces are inspectable
metadata and do not alter evaluation.

`RuleSet.Missing` selects how comparisons treat an operand whose path is
absent. `MissingFalse` (the default) makes the comparison false without calling
the operator. `MissingNull` substitutes `Null()` before built-in or custom
operator dispatch, so the ordinary kind rules apply exactly as for a supplied
null. `MissingUnknown` uses Kleene three-valued logic: the comparison is
unknown, `Not` keeps unknown, `All` is false when any child is false and
otherwise unknown when any child is unknown, and `Any` is true when any child
is true and otherwise unknown when any child is unknown. A rule matches only
on true; `Explanation.Unknown` records an unknown result. `Exists`, constants
and custom predicates are always known. Modes apply to compiled plans only;
direct `Predicate.Evaluate` calls use `MissingFalse`.

Both options are omitted from canonical JSON and hashes at their defaults, so
existing definitions keep their bytes and hashes.

`FirstMatch` selects the first ordered match. `CollectAll` records every unique
match across chaining iterations. `ErrorOnMultiple` becomes indeterminate when
more than one unique rule matches.

Derived paths have one producer per rule set. This removes write races and
ambiguous conflict resolution. If multiple business sources may propose a
value, model those proposals as separate paths and add an explicit resolving
rule.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
