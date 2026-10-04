package ruleengine_test

import (
	"context"
	"testing"
	"time"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

type recordingCache struct {
	plan       ruleengine.Plan
	key        string
	gets       int
	puts       int
	getBounded bool
	putBounded bool
}

func (cache *recordingCache) Get(ctx context.Context, key string) (ruleengine.Plan, bool, error) {
	_, cache.getBounded = ctx.Deadline()
	cache.gets++
	return cache.plan, cache.key == key, nil
}

func (cache *recordingCache) Put(ctx context.Context, key string, plan ruleengine.Plan) error {
	_, cache.putBounded = ctx.Deadline()
	cache.puts++
	cache.key = key
	cache.plan = plan
	return nil
}

func TestCompileCachedReusesOnlyMatchingImmutablePlans(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "cached", Rules: []ruleengine.Rule{{ID: "match", When: ruleengine.True()}}}
	compiler := ruleengine.NewCompiler(ruleengine.DefaultLimits())
	cache := &recordingCache{}

	first, _, err := compiler.CompileCached(context.Background(), set, cache)
	if err != nil {
		t.Fatalf("CompileCached(first) error = %v", err)
	}
	second, _, err := compiler.CompileCached(context.Background(), set, cache)
	if err != nil {
		t.Fatalf("CompileCached(second) error = %v", err)
	}
	if cache.gets != 2 || cache.puts != 1 {
		t.Fatalf("cache gets = %d, puts = %d", cache.gets, cache.puts)
	}
	if !cache.getBounded || !cache.putBounded {
		t.Fatalf("cache contexts bounded: get=%t put=%t", cache.getBounded, cache.putBounded)
	}
	if first.Hash() == "" || second.Hash() != first.Hash() {
		t.Fatalf("plan hashes = %q and %q", first.Hash(), second.Hash())
	}
	facts, _ := ruleengine.NewContext()
	if result := second.Evaluate(context.Background(), facts); result.Decision != ruleengine.Matched {
		t.Fatalf("cached Evaluate() = %#v", result)
	}
}

func TestCompileCachedDoesNotReuseAPlanWithDifferentLimits(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "cached-limits", Rules: []ruleengine.Rule{
		{ID: "one", When: ruleengine.True()},
		{ID: "two", When: ruleengine.True()},
	}}
	cache := &recordingCache{}
	if _, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).CompileCached(context.Background(), set, cache); err != nil {
		t.Fatalf("CompileCached(default) error = %v", err)
	}
	strict := ruleengine.DefaultLimits()
	strict.MaxRules = 1
	if _, _, err := ruleengine.NewCompiler(strict).CompileCached(context.Background(), set, cache); !ruleengine.IsCode(err, ruleengine.CodeLimitExceeded) {
		t.Fatalf("CompileCached(strict) error = %v, want limit exceeded", err)
	}
}

func TestCompileCachedCanonicalizesWithCompilerPolicy(t *testing.T) {
	t.Parallel()

	customSet := ruleengine.RuleSet{ID: "custom-cache", Rules: []ruleengine.Rule{{
		ID: "custom",
		When: ruleengine.Compare("same_length",
			ruleengine.Literal(ruleengine.String("FI")), ruleengine.Literal(ruleengine.String("SE"))),
	}}}
	customCompiler, err := ruleengine.NewCompilerWithOperators(ruleengine.DefaultLimits(), sameLengthOperator{})
	if err != nil {
		t.Fatalf("NewCompilerWithOperators() error = %v", err)
	}
	if _, _, err := customCompiler.CompileCached(context.Background(), customSet, &recordingCache{}); err != nil {
		t.Fatalf("CompileCached(custom operator) error = %v", err)
	}

	limits := ruleengine.DefaultLimits()
	limits.MaxRules = 2
	rules := []ruleengine.Rule{{ID: "one", When: ruleengine.True()}, {ID: "two", When: ruleengine.True()}}
	if _, _, err := ruleengine.NewCompiler(limits).CompileCached(context.Background(), ruleengine.RuleSet{ID: "permissive", Rules: rules}, &recordingCache{}); err != nil {
		t.Fatalf("CompileCached(non-default limits) error = %v", err)
	}
}

func TestCompileCachedStopsCanonicalizationBeforeCacheAccess(t *testing.T) {
	t.Parallel()

	set := ruleengine.RuleSet{ID: "canceled-cache", Rules: []ruleengine.Rule{{ID: "rule", When: ruleengine.True()}}}
	for name, ctx := range map[string]context.Context{
		"canceled":         canceledContext(),
		"expired deadline": expiredContext(),
	} {
		t.Run(name, func(t *testing.T) {
			cache := &recordingCache{}
			if _, _, err := ruleengine.NewCompiler(ruleengine.DefaultLimits()).CompileCached(ctx, set, cache); err == nil {
				t.Fatal("CompileCached() error = nil")
			}
			if cache.gets != 0 || cache.puts != 0 {
				t.Fatalf("cache gets = %d, puts = %d", cache.gets, cache.puts)
			}
		})
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func expiredContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	cancel()
	return ctx
}

func TestMemoryPlanCacheIsBoundedAndLRU(t *testing.T) {
	t.Parallel()

	cache, err := ruleengine.NewMemoryPlanCache(2)
	if err != nil {
		t.Fatalf("NewMemoryPlanCache() error = %v", err)
	}
	ctx := context.Background()
	if err := cache.Put(ctx, "a", ruleengine.Plan{}); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(ctx, "b", ruleengine.Plan{}); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := cache.Get(ctx, "a"); !ok {
		t.Fatal("cache lost a before capacity was reached")
	}
	if err := cache.Put(ctx, "c", ruleengine.Plan{}); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := cache.Get(ctx, "b"); ok {
		t.Fatal("cache did not evict least recently used b")
	}
	if cache.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", cache.Len())
	}
}

func TestMemoryPlanCacheHonorsCancellationAndInputValidation(t *testing.T) {
	t.Parallel()

	if _, err := ruleengine.NewMemoryPlanCache(0); err == nil {
		t.Fatal("NewMemoryPlanCache(0) error = nil")
	}
	cache, _ := ruleengine.NewMemoryPlanCache(1)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := cache.Put(canceled, "key", ruleengine.Plan{}); err == nil {
		t.Fatal("Put() error = nil for canceled context")
	}
	if _, _, err := cache.Get(canceled, "key"); err == nil {
		t.Fatal("Get() error = nil for canceled context")
	}
}
