package main

import "testing"

// A sha no repository can resolve must not read as "nothing changed".
func TestGateFailsClosedOnUnresolvableAnchor(t *testing.T) {
	cfg := driftCfg{
		since: "0000000000000000000000000000000000000000",
		repos: repoList{{Name: "nowhere", Root: t.TempDir()}},
	}
	gate, err := gateOf(cfg)
	if err == nil {
		t.Fatalf("gate = %v, want an error", gate)
	}
	if gate != nil {
		t.Errorf("gate = %v, want nil so no caller mistakes it for a clean run", gate)
	}
}

// Without --since-commit there is no gate at all: every citation is evaluated, which is
// what the weekly check does.
func TestNoAnchorMeansNoGate(t *testing.T) {
	gate, err := gateOf(driftCfg{repos: repoList{{Name: "nowhere", Root: t.TempDir()}}})
	if err != nil || gate != nil {
		t.Fatalf("gate = %v, err = %v, want nil, nil", gate, err)
	}
}
