# Cookbook

[Documentation index](README.md) · [Repository entry](../README.md)

## Distinguish missing and null

Use `Exists(path)` to test presence. Compare `Variable(path)` with
`Literal(Null())` to test explicit null. A missing comparison returns false
by default; set `RuleSet.Missing` to `MissingNull` to compare absent paths as
null or to `MissingUnknown` for Kleene three-valued logic.

## Evaluate rules in authored order

Set `RuleSet.Order = OrderDeclared` when a first-match table must follow its
authored order rather than priorities.

## Compare money exactly

Construct facts and literals with `Decimal("12.50")` and use the ordinary
equality and ordering operators. Check the error: invalid number text is
rejected rather than rounded.

## Match a typed list

Use `Compare(OpContains, Variable(tags), Literal(String("express")))` for an
exact list member. The variable must resolve to a list at runtime.

## Add a derived fact

Set `Rule.Derive` to literal facts and use `CollectAll` when downstream rules
must observe them. Give each derived path one producer. Compilation rejects
static cycles and duplicate producers.

## Resolve sparse facts

Compile the plan normally, then call `EvaluateResolved` with a resolver. The
plan requests only missing paths used by its propositions, in lexical order.

## Cache compiled plans

Create `NewMemoryPlanCache(capacity)` and call `CompileCached`. The canonical
definition hash is available through `Plan.Hash`.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
