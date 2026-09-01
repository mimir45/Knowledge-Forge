package dataset

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// TestPackagedCaptureListGates is a regression guard against config/code drift, and it
// exists because that exact mismatch once shipped green: the tag tests above check hand-written lists and
// pkg/config's tests check that the packaged layer parses, so nothing ever asserted that
// the two agree. D2Tag read "d2_advisor" while the packaged dataset.capture said "d2", so
// the gate returned false forever — silently, since a capture write is a side channel
// that deliberately never fails a command.
//
// It lives here rather than in pkg/config because pkg/config cannot import pkg/dataset:
// dataset -> vault -> config is a real edge (vault/validate.go), so the reverse is a
// cycle. Loading with no optional layers gives the packaged base alone, which is exactly
// what a fresh install runs under.
//
// It now covers all five tiers. It used to be scoped to just D2 and D4, from a time when
// d1/d3/d5 were read by nothing in the capture path; every tag in the packaged list now
// gates a real write path, so every tag
// is worth pinning. Assert against the packaged layer, never a hand-written list — the
// whole point is that config and code are checked against each other.
func TestPackagedCaptureListGates(t *testing.T) {
	t.Setenv(config.EnvVar, "") // a developer's own $FORGE_CONFIG must not reach this
	c, err := config.Load(config.Options{ProjectDir: t.TempDir(), HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if !c.Dataset.Enabled {
		t.Fatal("packaged dataset.enabled = false; the gates below are moot")
	}
	for _, tier := range Tiers() {
		// D6 has no write path and is deliberately absent from
		// dataset.capture — Enabled would correctly report false for it forever, which is
		// not the mismatch this test guards against.
		if tier.Derived {
			continue
		}
		if !tier.Enabled(c.Dataset) {
			t.Errorf("packaged dataset.capture %v does not enable %s (want the %q entry)",
				c.Dataset.Capture, tier.Kind, tier.Tag)
		}
	}
}

// TestConsentResolvesWithNoHomeLayer is the git-hook environment, and it exists because
// this phase moved a config read onto that path: cmd/forge/capture.go now calls loadConfig
// before harvesting, and skips capture when it fails.
//
// That fail-closed choice is right for a consent check but it has a bad failure mode if
// mere absence can trip it — a hook runs with a minimal environment, prints nothing, and
// never fails a commit, so D3 capture would stop silently and only .forge/capture.log
// would know. pkg/config/load.go:69 discards os.UserHomeDir()'s error deliberately ("an
// unresolvable home skips that layer, not fails") and the base layer is embedded in the
// binary, so an unreachable home resolves to the packaged config rather than an error.
// This pins that: absence must never be mistaken for withheld consent.
func TestConsentResolvesWithNoHomeLayer(t *testing.T) {
	t.Setenv(config.EnvVar, "")
	c, err := config.Load(config.Options{ProjectDir: t.TempDir(), HomeDir: "/nonexistent"})
	if err != nil {
		t.Fatalf("config.Load with an unreachable home: %v", err)
	}
	if !D3.Enabled(c.Dataset) {
		t.Error("D3 capture is off under a hook-like environment; absence read as a no")
	}
}
