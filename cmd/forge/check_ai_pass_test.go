package main

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
)

// TestTopBrokenIgnoresSuspectAndTiesBreakOnRef: a SUSPECT finding must never win over a
// BROKEN one regardless of order.
func TestTopBrokenIgnoresSuspectAndTiesBreakOnRef(t *testing.T) {
	fs := []drift.Finding{
		{Note: "notes/concept/b.md", Ref: "z.go#Foo", Verdict: drift.Suspect, Reason: "moved"},
		{Note: "notes/concept/a.md", Ref: "z.go#Foo", Verdict: drift.Broken, Reason: "gone"},
		{Note: "notes/concept/a.md", Ref: "a.go#Bar", Verdict: drift.Broken, Reason: "gone too"},
	}
	got, ok := topBroken(fs)
	if !ok || got.Note != "notes/concept/a.md" || got.Ref != "a.go#Bar" {
		t.Errorf("topBroken = %+v, %v; want notes/concept/a.md a.go#Bar, true", got, ok)
	}
}

func TestTopBrokenFalseWhenNothingBroken(t *testing.T) {
	fs := []drift.Finding{{Note: "n", Ref: "r", Verdict: drift.Suspect}}
	if _, ok := topBroken(fs); ok {
		t.Error("topBroken found a result with no BROKEN findings present")
	}
	if _, ok := topBroken(nil); ok {
		t.Error("topBroken(nil) = _, true; want false")
	}
}
