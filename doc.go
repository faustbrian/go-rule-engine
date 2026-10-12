// Package ruleengine compiles typed propositions and evaluates them over
// immutable fact snapshots. Start with DefaultLimits, NewPath, NewContext and
// NewCompiler; compile before reusing a Plan. The repository-owned JSON AST
// grammar is version 1, exposed through ParseJSON and the jsonast facade.
//
// Built-in state is immutable; custom predicates/operators and cache/resolver
// callbacks remain caller-owned trust boundaries. They must supply concurrency
// safety, bounded work and cooperative cancellation. Calls are synchronous;
// there are no background goroutines, callback sandbox or shutdown lifecycle.
// The core has no external Go dependencies or cgo requirements.
//
// Results may contain partial matches or derivations with Indeterminate. This
// package owns no authorization, feature rollout, workflow, persistence or
// action execution; consumers own their interpretation and failure policy.
package ruleengine
