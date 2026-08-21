package dataset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnabledRequiresTheD2Tag(t *testing.T) {
	if Enabled([]string{"d1", "d3"}) {
		t.Error("Enabled() = true without d2 in the list")
	}
	if !Enabled([]string{"d1", "d2"}) {
		t.Error("Enabled() = false with d2 present")
	}
	// The pre-B-024 spelling must not keep working, or the mismatch could silently return.
	if Enabled([]string{"d2_advisor"}) {
		t.Error("Enabled() = true for the old d2_advisor spelling")
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
	AppendD2(root, p)
	AppendD2(root, p)
	out, _ := os.ReadFile(filepath.Join(root, D2Path))
	if n := strings.Count(string(out), "\n"); n != 2 {
		t.Errorf("got %d lines, want 2", n)
	}
}
