package ruleengine_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

type policyContextKey struct{}

func TestOwnedCacheRejectsZeroCompilerBeforeCacheAccess(t *testing.T) {
	var compiler ruleengine.Compiler
	cache := &recordingCache{}
	set := ruleengine.RuleSet{ID: "ordinary", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.True()}}}
	plan, diagnostics, err := compiler.CompileCached(context.Background(), set, cache)
	if !ruleengine.IsCode(err, ruleengine.CodeInvalidLimit) {
		t.Fatalf("zero compiler error = %v, want invalid limit", err)
	}
	if !reflect.DeepEqual(plan, ruleengine.Plan{}) || diagnostics != nil {
		t.Fatalf("rejected compile returned plan or diagnostics: %#v, %#v", plan, diagnostics)
	}
	if cache.gets != 0 || cache.puts != 0 {
		t.Fatalf("invalid compiler reached cache: gets=%d puts=%d", cache.gets, cache.puts)
	}
}

func TestOwnedResolutionRejectsCanceledCallerBeforeResolver(t *testing.T) {
	missing := ruleengine.MustPath("facts", "missing")
	retained := ruleengine.MustPath("facts", "retained")
	set := ruleengine.RuleSet{ID: "ordinary", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.Exists(missing)}}}
	plan, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).Compile(context.Background(), set)
	if err != nil {
		t.Fatal(err)
	}
	base, err := ruleengine.NewContext(ruleengine.Fact{Path: retained, Value: ruleengine.Int(7), Owner: ruleengine.OwnerResource})
	if err != nil {
		t.Fatal(err)
	}
	before := base.Lookup(retained)
	resolver := &mapResolver{values: map[string]ruleengine.Value{missing.String(): ruleengine.Int(1)}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := plan.EvaluateResolved(ctx, base, resolver)
	if result.Decision != ruleengine.Indeterminate || len(result.Errors) != 1 || !ruleengine.IsCode(result.Errors[0], ruleengine.CodeEvaluation) {
		t.Fatalf("canceled resolution result = %#v", result)
	}
	if len(resolver.calls) != 0 || len(result.MatchedRules) != 0 || len(result.Explanation) != 0 {
		t.Fatalf("canceled resolution performed work: calls=%v result=%#v", resolver.calls, result)
	}
	owner, found := base.Owner(retained)
	if !reflect.DeepEqual(base.Lookup(retained), before) || !found || owner != ruleengine.OwnerResource || base.Lookup(missing).Kind() != ruleengine.KindMissing {
		t.Fatal("canceled resolution changed the caller's base snapshot")
	}
}

type observingPlanCache struct {
	recordingCache
	get, put func(context.Context)
}

func (cache *observingPlanCache) Get(ctx context.Context, key string) (ruleengine.Plan, bool, error) {
	if cache.get != nil {
		cache.get(ctx)
	}
	return cache.recordingCache.Get(ctx, key)
}

func (cache *observingPlanCache) Put(ctx context.Context, key string, plan ruleengine.Plan) error {
	if cache.put != nil {
		cache.put(ctx)
	}
	return cache.recordingCache.Put(ctx, key, plan)
}

func TestOwnedCacheReplacesDifferentAcceptedExplanationPolicy(t *testing.T) {
	set := ruleengine.RuleSet{ID: "policy", Strategy: ruleengine.CollectAll, Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.True()}, {ID: "two", When: ruleengine.True()}}}
	limits := ruleengine.DefaultLimits()
	limits.MaxExplanation = 2
	cache := &recordingCache{}
	first, _, err := ruleengine.NewCompiler(limits).CompileCached(context.Background(), set, cache)
	if err != nil {
		t.Fatal(err)
	}
	facts, _ := ruleengine.NewContext()
	if result := first.Evaluate(context.Background(), facts); result.Decision != ruleengine.Matched || len(result.Explanation) != 2 {
		t.Fatalf("initial result = %#v", result)
	}
	limits.MaxExplanation = 1
	compiler := ruleengine.NewCompiler(limits)
	second, _, err := compiler.CompileCached(context.Background(), set, cache)
	if err != nil {
		t.Fatal(err)
	}
	if result := second.Evaluate(context.Background(), facts); result.Decision != ruleengine.Matched || len(result.Explanation) != 1 {
		t.Fatalf("changed policy result = %#v", result)
	}
	if cache.puts != 2 || first.Hash() != second.Hash() {
		t.Fatal("accepted policy did not replace identical-definition plan")
	}
	third, _, err := compiler.CompileCached(context.Background(), set, cache)
	if err != nil || cache.puts != 2 || third.Hash() != second.Hash() {
		t.Fatalf("matching policy reuse = %v; puts=%d", err, cache.puts)
	}
}

func TestOwnedCacheSharesDeadlineAndPreservesEarlierCaller(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		ctx := context.WithValue(context.Background(), policyContextKey{}, "ordinary")
		var callerDeadline time.Time
		if earlier {
			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Minute))
			defer cancel()
			callerDeadline, _ = ctx.Deadline()
		}
		limits := ruleengine.DefaultLimits()
		limits.EvaluationTimeout = 2 * time.Minute
		var getContext context.Context
		cache := &observingPlanCache{get: func(current context.Context) { getContext = current }, put: func(current context.Context) {
			deadline, ok := current.Deadline()
			getDeadline, getOK := getContext.Deadline()
			if !ok || !getOK || !deadline.Equal(getDeadline) || current != getContext || current.Value(policyContextKey{}) != "ordinary" {
				t.Fatal("cache operation context not shared/bounded")
			}
			if earlier && !deadline.Equal(callerDeadline) {
				t.Fatal("cache widened caller deadline")
			}
		}}
		set := ruleengine.RuleSet{ID: "context", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.True()}}}
		if _, _, err := ruleengine.NewCompiler(limits).CompileCached(ctx, set, cache); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOwnedCacheCancellationAfterLookupRejectsHit(t *testing.T) {
	compiler := ruleengine.NewCompiler(ruleengine.DefaultLimits())
	set := ruleengine.RuleSet{ID: "cancel-hit", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.True()}}}
	cache := &observingPlanCache{}
	if _, _, err := compiler.CompileCached(context.Background(), set, cache); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache.get = func(context.Context) { cancel() }
	if _, _, err := compiler.CompileCached(ctx, set, cache); err == nil {
		t.Fatal("canceled lookup returned successful cached plan")
	}
	if cache.puts != 1 {
		t.Fatal("canceled lookup reached cache write")
	}
}

type observingResolver struct{ observe func(context.Context) }

func (resolver observingResolver) Resolve(ctx context.Context, _ ruleengine.Path) (ruleengine.Value, ruleengine.Owner, bool, error) {
	resolver.observe(ctx)
	return ruleengine.String("FI"), ruleengine.OwnerResource, true, nil
}

type observingOperator struct{ observe func(context.Context) }

func (observingOperator) Name() ruleengine.OperatorName { return "observed" }
func (observingOperator) Signatures() []ruleengine.Signature {
	return []ruleengine.Signature{{Left: ruleengine.KindString, Right: ruleengine.KindString}}
}
func (operator observingOperator) Evaluate(ctx context.Context, _, _ ruleengine.Value) (bool, error) {
	operator.observe(ctx)
	return true, nil
}

func TestOwnedResolutionAndEvaluationShareDeadlinePolicy(t *testing.T) {
	for _, test := range []struct{ cancelDuringResolve, earlier bool }{{false, false}, {false, true}, {true, true}} {
		limits := ruleengine.DefaultLimits()
		limits.EvaluationTimeout = 2 * time.Minute
		base := context.WithValue(context.Background(), policyContextKey{}, "ordinary")
		ctx, cancel := context.WithCancel(base)
		if test.earlier {
			cancel()
			ctx, cancel = context.WithDeadline(base, time.Now().Add(time.Minute))
		}
		defer cancel()
		callerDeadline, _ := ctx.Deadline()
		var resolvedDeadline time.Time
		operatorCalls := 0
		operator := observingOperator{observe: func(current context.Context) {
			operatorCalls++
			deadline, ok := current.Deadline()
			if !ok || !deadline.Equal(resolvedDeadline) || current.Value(policyContextKey{}) != "ordinary" {
				t.Fatal("evaluation lost shared deadline/value policy")
			}
		}}
		compiler, err := ruleengine.NewCompilerWithOperators(limits, operator)
		if err != nil {
			t.Fatal(err)
		}
		set := ruleengine.RuleSet{ID: "resolve", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.Compare("observed", ruleengine.Variable(ruleengine.MustPath("facts", "code")), ruleengine.Literal(ruleengine.String("SE")))}}}
		plan, _, err := compiler.Compile(context.Background(), set)
		if err != nil {
			t.Fatal(err)
		}
		facts, _ := ruleengine.NewContext()
		resolver := observingResolver{observe: func(current context.Context) {
			var ok bool
			resolvedDeadline, ok = current.Deadline()
			if !ok || test.earlier && !resolvedDeadline.Equal(callerDeadline) || current.Value(policyContextKey{}) != "ordinary" {
				t.Fatal("resolver lost caller deadline/value")
			}
			if test.cancelDuringResolve {
				cancel()
			}
		}}
		result := plan.EvaluateResolved(ctx, facts, resolver)
		if test.cancelDuringResolve {
			if result.Decision != ruleengine.Indeterminate || len(result.Errors) == 0 || operatorCalls != 0 {
				t.Fatalf("canceled resolution = %#v, operators=%d", result, operatorCalls)
			}
		} else if result.Decision != ruleengine.Matched || operatorCalls != 1 {
			t.Fatalf("resolution result = %#v", result)
		}
	}
}

func TestOwnedCacheCancellationAfterPutDoesNotClaimRollback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := &observingPlanCache{put: func(context.Context) { cancel() }}
	set := ruleengine.RuleSet{ID: "cancel-put", Rules: []ruleengine.Rule{{ID: "one", When: ruleengine.True()}}}
	if _, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).CompileCached(ctx, set, cache); err == nil {
		t.Fatal("canceled cache write returned successful plan")
	}
	if cache.puts != 1 || cache.key == "" || cache.plan.Hash() != cache.key {
		t.Fatal("test lost the completed cache write side effect")
	}
}

func TestOwnedResolverCancellationPrecedesReturnedFactValidation(t *testing.T) {
	limits := ruleengine.DefaultLimits()
	limits.MaxStringBytes = 1
	set := ruleengine.RuleSet{ID: "cancel-validation", Rules: []ruleengine.Rule{{
		ID: "one", When: ruleengine.Compare(ruleengine.OpEqual,
			ruleengine.Variable(ruleengine.MustPath("facts", "code")),
			ruleengine.Literal(ruleengine.String("F"))),
	}}}
	plan, _, err := ruleengine.NewCompiler(limits).Compile(context.Background(), set)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	facts, _ := ruleengine.NewContext()
	resolver := observingResolver{observe: func(context.Context) { cancel() }}
	result := plan.EvaluateResolved(ctx, facts, resolver)
	if result.Decision != ruleengine.Indeterminate || len(result.Errors) != 1 ||
		!ruleengine.IsCode(result.Errors[0], ruleengine.CodeEvaluation) ||
		ruleengine.IsCode(result.Errors[0], ruleengine.CodeInvalidFact) {
		t.Fatalf("canceled returned fact validation = %#v; errors=%v", result, result.Errors)
	}
}
