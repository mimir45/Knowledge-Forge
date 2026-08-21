package dataset

import (
	"testing"

	"knowledge-forge/pkg/config"
)

// TestPackagedCaptureListGates is B-024's regression guard, and it exists because the
// mismatch it pins shipped green: the tag tests above check hand-written lists and
// pkg/config's tests check that the packaged layer parses, so nothing ever asserted that
// the two agree. D2Tag read "d2_advisor" while the packaged dataset.capture said "d2", so
// Enabled returned false forever — silently, since a capture write is a side channel that
// deliberately never fails a command.
//
// It lives here rather than in pkg/config because pkg/config cannot import pkg/dataset:
// dataset -> vault -> config is a real edge (vault/validate.go), so the reverse is a
// cycle. Loading with no optional layers gives the packaged base alone, which is exactly
// what a fresh install runs under.
//
// Scoped to D2 and D4 deliberately. Those are the only two tags in the tree and the only
// two readers of cfg.Dataset.Capture (cmd/forge/engine_run.go, cmd/forge/gate.go). The
// list's d1/d3/d5 entries are read by nothing — see BACKLOG B-030 — so asserting on them
// would pin a gate that does not exist.
func TestPackagedCaptureListGates(t *testing.T) {
	t.Setenv(config.EnvVar, "") // a developer's own $FORGE_CONFIG must not reach this
	c, err := config.Load(config.Options{ProjectDir: t.TempDir(), HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if !c.Dataset.Enabled {
		t.Fatal("packaged dataset.enabled = false; the gates below are moot")
	}
	if !Enabled(c.Dataset.Capture) {
		t.Errorf("packaged dataset.capture %v does not enable D2 (want the %q entry)",
			c.Dataset.Capture, D2Tag)
	}
	if !D4Enabled(c.Dataset.Capture) {
		t.Errorf("packaged dataset.capture %v does not enable D4 (want the %q entry)",
			c.Dataset.Capture, D4Tag)
	}
}
