package ruleengine_test

import (
	"context"
	"reflect"
	"testing"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

func TestCompileCachedReplacesForeignDefinitionWithEqualLimits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	compiler := ruleengine.NewCompiler(ruleengine.DefaultLimits())
	cache, err := ruleengine.NewMemoryPlanCache(2)
	if err != nil {
		t.Fatal(err)
	}
	foreignSet := ruleengine.RuleSet{ID: "cache-foreign", Strategy: ruleengine.FirstMatch, Rules: []ruleengine.Rule{{ID: "foreign", When: ruleengine.False()}}}
	foreign, diagnostics, err := compiler.CompileCached(ctx, foreignSet, cache)
	if err != nil || diagnostics != nil {
		t.Fatalf("foreign compile = %#v, %v", diagnostics, err)
	}
	wantedSet := ruleengine.RuleSet{ID: "cache-target", Strategy: ruleengine.FirstMatch, Rules: []ruleengine.Rule{{ID: "wanted", When: ruleengine.True()}}}
	const wantedHash = "4108247d891d7f5083fdd68c8f4f231ec804781479baee0bd1e596ba6f6d99ee"
	if foreign.Hash() == "" || foreign.Hash() == wantedHash {
		t.Fatalf("foreign hash = %q", foreign.Hash())
	}
	if err = cache.Put(ctx, wantedHash, foreign); err != nil {
		t.Fatal(err)
	}
	facts, err := ruleengine.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	assertWanted := func(plan ruleengine.Plan) {
		t.Helper()
		if plan.Hash() != wantedHash {
			t.Errorf("plan hash = %q, want %q", plan.Hash(), wantedHash)
		}
		result := plan.Evaluate(ctx, facts)
		if result.Decision != ruleengine.Matched || !reflect.DeepEqual(result.MatchedRules, []string{"wanted"}) || !reflect.DeepEqual(result.Explanation, []ruleengine.Explanation{{RuleID: "wanted", Matched: true}}) || len(result.Errors) != 0 {
			t.Errorf("wanted evaluation = %#v", result)
		}
	}
	for range 2 {
		plan, diagnostics, err := compiler.CompileCached(ctx, wantedSet, cache)
		if err != nil || diagnostics != nil {
			t.Fatalf("wanted compile = %#v, %v", diagnostics, err)
		}
		assertWanted(plan)
		stored, found, err := cache.Get(ctx, wantedHash)
		if err != nil || !found {
			t.Fatalf("stored replacement found = %t, error = %v", found, err)
		}
		assertWanted(stored)
	}
	result := foreign.Evaluate(ctx, facts)
	if result.Decision != ruleengine.Unmatched || len(result.MatchedRules) != 0 || !reflect.DeepEqual(result.Explanation, []ruleengine.Explanation{{RuleID: "foreign", Matched: false}}) || len(result.Errors) != 0 {
		t.Fatalf("held foreign evaluation = %#v", result)
	}
}
