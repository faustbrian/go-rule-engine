// Package jsonast exposes the repository-owned JSON rule-definition DSL,
// grammar versions 1 and 2.
// The grammar is the canonical representation implemented by ruleengine;
// unknown or variant-incompatible fields are rejected. Parse validates with
// built-in operators; Marshal uses default compiler limits. Custom registries
// need the core Compiler methods. Byte storage is not retained. This facade
// introduces no dependencies beyond the core and owns no evaluation lifecycle.
package jsonast

import ruleengine "github.com/faustbrian/go-rule-engine/v2"

// Parse decodes and validates a JSON AST definition.
func Parse(data []byte, limits ruleengine.Limits) (ruleengine.RuleSet, []ruleengine.Diagnostic, error) {
	return ruleengine.ParseJSON(data, limits)
}

// Marshal returns the canonical JSON AST representation.
func Marshal(set ruleengine.RuleSet) ([]byte, error) {
	return ruleengine.MarshalCanonical(set)
}
