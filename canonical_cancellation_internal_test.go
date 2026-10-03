package ruleengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

// canonicalCheckpointContext schedules real cancellation at an Err checkpoint.
// It is used synchronously: Done and Err retain the embedded context's contract.
type canonicalCheckpointContext struct {
	context.Context
	cancel   context.CancelFunc
	cancelAt int
	calls    int
}

func (ctx *canonicalCheckpointContext) Err() error {
	ctx.calls++
	if ctx.calls == ctx.cancelAt {
		ctx.cancel()
	}
	return ctx.Context.Err()
}

func TestCanonicalMarshalHonorsLateCancellation(t *testing.T) {
	for _, test := range []struct {
		name       string
		checkpoint int
	}{{"before rule encoding", 3}, {"after encoding", 4}} {
		t.Run(test.name, func(t *testing.T) {
			base, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx := &canonicalCheckpointContext{Context: base, cancel: cancel, cancelAt: test.checkpoint}
			set := RuleSet{ID: "ordinary", Rules: []Rule{{ID: "one", When: True()}}}
			// This one-rule fixture has two Compile checks, then the pre-encoding
			// and post-encoding checks. The checkpoint is a scheduling seam.
			encoded, err := NewCompiler(DefaultLimits()).marshalCanonical(ctx, set)
			if encoded != nil || !errors.Is(err, context.Canceled) {
				t.Fatalf("marshalCanonical() = %q, %v; want nil bytes and cancellation", encoded, err)
			}
			assertCanonicalContextCanceled(ctx, t)
		})
	}
}

func TestCanonicalHashHonorsLateCancellation(t *testing.T) {
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &canonicalCheckpointContext{Context: base, cancel: cancel, cancelAt: 5}
	set := RuleSet{ID: "ordinary", Rules: []Rule{{ID: "one", When: True()}}}
	// The fifth checkpoint follows successful canonical encoding and hashing.
	hash, err := NewCompiler(DefaultLimits()).canonicalHash(ctx, set)
	if hash != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("canonicalHash() = %q, %v; want empty hash and cancellation", hash, err)
	}
	assertCanonicalContextCanceled(ctx, t)
}

func TestCanonicalSerializationUncanceledControl(t *testing.T) {
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &canonicalCheckpointContext{Context: base, cancel: cancel}
	set := RuleSet{ID: "ordinary", Rules: []Rule{{ID: "one", When: True()}}}
	compiler := NewCompiler(DefaultLimits())
	want := []byte(`{"version":"1","id":"ordinary","strategy":"first_match","rules":[{"id":"one","priority":0,"tags":null,"when":{"kind":"true"},"derive":[]}]}`)
	encoded, err := compiler.marshalCanonical(ctx, set)
	if err != nil || !bytes.Equal(encoded, want) {
		t.Fatalf("marshalCanonical() = %q, %v; want %q", encoded, err, want)
	}
	digest := sha256.Sum256(want)
	wantHash := hex.EncodeToString(digest[:])
	hash, err := compiler.canonicalHash(ctx, set)
	if err != nil || hash != wantHash {
		t.Fatalf("canonicalHash() = %q, %v; want %q", hash, err, wantHash)
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("uncanceled context error = %v", err)
	}
	select {
	case <-ctx.Done():
		t.Fatal("uncanceled context Done is closed")
	default:
	}
}

func assertCanonicalContextCanceled(ctx context.Context, t *testing.T) {
	t.Helper()
	select {
	case <-ctx.Done():
	default:
		t.Fatal("canceled context Done is not closed")
	}
	for range 2 {
		if err := ctx.Err(); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled context error = %v, want stable cancellation", err)
		}
	}
}
