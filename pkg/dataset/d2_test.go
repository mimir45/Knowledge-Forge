package dataset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

func TestEnabledRequiresTheD2Tag(t *testing.T) {
	on := func(tags ...string) config.Dataset {
		return config.Dataset{Enabled: true, Capture: tags}
	}
	if D2.Enabled(on("d1", "d3")) {
		t.Error("D2.Enabled() = true without d2 in the list")
	}
	if !D2.Enabled(on("d1", "d2")) {
		t.Error("D2.Enabled() = false with d2 present")
	}
	// The old "d2_advisor" spelling must not keep working, or the config/code mismatch
	// it once caused could silently return.
	if D2.Enabled(on("d2_advisor")) {
		t.Error("D2.Enabled() = true for the old d2_advisor spelling")
	}
}

// TestEnabledHonoursTheMasterSwitch pins the gate this phase added.
func TestEnabledHonoursTheMasterSwitch(t *testing.T) {
	off := config.Dataset{Enabled: false, Capture: []string{"d1", "d2", "d3", "d4", "d5"}}
	for _, tier := range Tiers() {
		if tier.Enabled(off) {
			t.Errorf("%s.Enabled() = true with dataset.enabled false", tier.Tag)
		}
	}
}

func TestAppendD2WritesAJSONLLine(t *testing.T) {
	root := t.TempDir()
	p := D2Pair{Kind: D2Kind, Stage: "verify", Draft: "draft text",
		Critique: `{"confidence":"medium"}`, CapturedAt: time.Now()}
	if err := AppendD2(root, p); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, D2Path))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"draft":"draft text"`) {
		t.Errorf("d2.jsonl missing the draft field:\n%s", out)
	}
}

// TestAppendD2NeverDeduplicates: two real advisor calls on the same draft are two real
// data points, not a re-fired hook — both lines must survive.
func TestAppendD2NeverDeduplicates(t *testing.T) {
	root := t.TempDir()
	p := D2Pair{Kind: D2Kind, Stage: "verify", Draft: "same", Critique: "same"}
	if err := AppendD2(root, p); err != nil {
		t.Fatal(err)
	}
	if err := AppendD2(root, p); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(root, D2Path))
	if n := strings.Count(string(out), "\n"); n != 2 {
		t.Errorf("got %d lines, want 2", n)
	}
}
