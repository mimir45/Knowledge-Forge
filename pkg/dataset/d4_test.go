package dataset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

func TestD4EnabledRequiresTheD4Tag(t *testing.T) {
	if D4.Enabled(config.Dataset{Enabled: true, Capture: []string{"d2"}}) {
		t.Error("D4.Enabled() = true without d4 in the list")
	}
	if !D4.Enabled(config.Dataset{Enabled: true, Capture: []string{"d2", "d4"}}) {
		t.Error("D4.Enabled() = false with d4 present")
	}
}

// TestTiersAreDistinct guards the registry against the copy-paste it exists to prevent:
// six entries, six tags, five distinct non-empty paths (D6 is derived and has none), no
// two tags sharing either.
func TestTiersAreDistinct(t *testing.T) {
	tags, paths := map[string]bool{}, map[string]bool{}
	for _, tier := range Tiers() {
		if tags[tier.Tag] {
			t.Errorf("duplicate tier tag %q", tier.Tag)
		}
		if tier.Path != "" && paths[tier.Path] {
			t.Errorf("duplicate tier path %q", tier.Path)
		}
		tags[tier.Tag], paths[tier.Path] = true, true
	}
	if len(tags) != 6 {
		t.Errorf("Tiers() has %d distinct tags, want 6", len(tags))
	}
}

func TestAppendD4WritesAJSONLLine(t *testing.T) {
	root := t.TempDir()
	p := D4Pair{Kind: D4Kind, Stage: "gate", FailingDraft: "bad draft",
		GateError: "schema: missing sources", FixedDraft: "good draft", CapturedAt: time.Now()}
	if err := AppendD4(root, p); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, D4Path))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"failing_draft":"bad draft"`) {
		t.Errorf("d4.jsonl missing the failing_draft field:\n%s", out)
	}
}

// TestAppendD4NeverDeduplicates: mirrors D2's rationale — each gate-repair pairing is its
// own event, appended once per call, no Key()-based collapsing.
func TestAppendD4NeverDeduplicates(t *testing.T) {
	root := t.TempDir()
	p := D4Pair{Kind: D4Kind, Stage: "gate", FailingDraft: "x", GateError: "y", FixedDraft: "z"}
	if err := AppendD4(root, p); err != nil {
		t.Fatal(err)
	}
	if err := AppendD4(root, p); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(root, D4Path))
	if n := strings.Count(string(out), "\n"); n != 2 {
		t.Errorf("got %d lines, want 2", n)
	}
}

func TestSaveAndTakePreviousDraftRoundTrips(t *testing.T) {
	root := t.TempDir()
	path, err := SaveFailingDraft(root, "my-slug", []byte("draft body"), []byte("schema: fail"))
	if err != nil {
		t.Fatal(err)
	}
	draft, gateErr, err := TakePreviousDraft(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(draft) != "draft body" || string(gateErr) != "schema: fail" {
		t.Errorf("got draft=%q gateErr=%q", draft, gateErr)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("TakePreviousDraft did not delete the draft file")
	}
	if _, err := os.Stat(path + ".err"); !os.IsNotExist(err) {
		t.Error("TakePreviousDraft did not delete the .err sidecar")
	}
}

// TestTakePreviousDraftConsumesOnce pins the "exactly once" join guarantee: a second
// --previous-draft pointing at an already-consumed path must fail loudly, not silently
// re-pair or return stale content.
func TestTakePreviousDraftConsumesOnce(t *testing.T) {
	root := t.TempDir()
	path, _ := SaveFailingDraft(root, "slug", []byte("d"), []byte("e"))
	if _, _, err := TakePreviousDraft(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := TakePreviousDraft(path); err == nil {
		t.Error("second TakePreviousDraft on the same path succeeded, want an error")
	}
}
