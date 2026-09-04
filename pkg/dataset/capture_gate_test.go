package dataset

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// TestPackagedCaptureListGates is a regression guard against config/code drift, and it
// exists because that exact mismatch once shipped green.
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
// this phase moved a config read onto that path.
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
