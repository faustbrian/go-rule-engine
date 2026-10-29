package ruleengine_test

import (
	"testing"

	ruleengine "github.com/faustbrian/go-rule-engine/v2"
)

// FuzzParseJSON admits bounded grammar-v1/v2 bytes and checks canonical reparse
// and hash stability for accepted definitions, not external DSL conformance.
func FuzzParseJSON(f *testing.F) {
	f.Add([]byte(`{"version":"1","id":"empty","strategy":"first_match","rules":[]}`))
	f.Add([]byte(`{"version":"999","id":"invalid","strategy":"first_match","rules":[]}`))
	f.Add([]byte(`{"version":"2","id":"v2","strategy":"collect_all","order":"declared","missing":"unknown","rules":[{"id":"r","priority":0,"tags":[],"when":{"kind":"compare","operator":"less_than","left":{"kind":"variable","path":["x"]},"right":{"kind":"literal","value":{"type":"decimal","decimal":"1.50"}}},"derive":[]}]}`))
	f.Add([]byte("not-json"))
	f.Fuzz(func(t *testing.T, input []byte) {
		limits := ruleengine.DefaultLimits()
		limits.MaxDefinitionBytes = 64 << 10
		set, _, err := ruleengine.ParseJSON(input, limits)
		if err != nil {
			return
		}
		canonical, err := ruleengine.MarshalCanonical(set)
		if err != nil {
			t.Fatalf("MarshalCanonical() error = %v", err)
		}
		roundTrip, _, err := ruleengine.ParseJSON(canonical, limits)
		if err != nil {
			t.Fatalf("ParseJSON(canonical) error = %v", err)
		}
		first, _ := ruleengine.CanonicalHash(set)
		second, _ := ruleengine.CanonicalHash(roundTrip)
		if first != second {
			t.Fatalf("round-trip hashes = %q and %q", first, second)
		}
	})
}

// FuzzPath checks that accepted hostile segment pairs preserve exact segment
// identity rather than being split or normalized into another fact key.
func FuzzPath(f *testing.F) {
	f.Add("shipment", "country")
	f.Add("..", "value")
	f.Add("shipment.recipient", "name")
	f.Fuzz(func(t *testing.T, first, second string) {
		path, err := ruleengine.NewPath(ruleengine.DefaultLimits(), first, second)
		if err != nil {
			return
		}
		segments := path.Segments()
		if len(segments) != 2 || segments[0] != first || segments[1] != second {
			t.Fatalf("Segments() = %#v", segments)
		}
	})
}
