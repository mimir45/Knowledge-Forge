package main

import "testing"

// isLockedStage guards a third path.
func TestIsLockedStageMatchesEngineLockedStages(t *testing.T) {
	for _, s := range []string{"recall", "write", "index"} {
		if !isLockedStage(s) {
			t.Errorf("isLockedStage(%q) = false, want true", s)
		}
	}
	if isLockedStage("research") {
		t.Error("isLockedStage(\"research\") = true, research is not locked")
	}
}

// runEngineRecord must refuse to stamp write=host even though TrailEntry itself has no
// opinion about locked stages — the guard has to sit in the CLI, not the pure package.
func TestRunEngineRecordRefusesLockedStageWithNonNoneTier(t *testing.T) {
	root := fixtureCopy(t)
	code := runEngineRecord(root, "concepts/hibernate.md", "write", "host")
	if code == 0 {
		t.Fatal("runEngineRecord(write, host) = 0, want a nonzero refusal")
	}
}

// write=none must stay legal — the lock is on the tier, not the stage name itself.
func TestRunEngineRecordAllowsLockedStageWithNoneTier(t *testing.T) {
	root := fixtureCopy(t)
	code := runEngineRecord(root, "concepts/hibernate.md", "write", "none")
	if code != 0 {
		t.Fatalf("runEngineRecord(write, none) = %d, want 0", code)
	}
}
