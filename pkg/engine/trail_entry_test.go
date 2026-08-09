package engine

import "testing"

func TestTrailEntryStampable(t *testing.T) {
	entry, ok := TrailEntry("verify", "advisor")
	if !ok || entry != "verify=advisor" {
		t.Errorf("TrailEntry(verify, advisor) = %q, %v", entry, ok)
	}
}

// TestTrailEntryUnstampedStages: intake/plan/synthesize/link are host-orchestration
// bookkeeping per the plan's decision, not audited model-call stages.
func TestTrailEntryUnstampedStages(t *testing.T) {
	for _, stage := range []string{"intake", "plan", "synthesize", "link"} {
		if _, ok := TrailEntry(stage, "host"); ok {
			t.Errorf("TrailEntry(%s, host) reported ok, want false", stage)
		}
	}
}
