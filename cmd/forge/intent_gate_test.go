package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
)

// This is the derivation behind intent.go's gate.

const (
	intentLabelsPath = "testdata/intent-gate-labels.txt"
	intentGolden     = "testdata/intent-gate.golden"
)

// intentPrompt is one labelled prompt: fire means the vault already answers it.
type intentPrompt struct {
	text string
	fire bool
}

// minFireAdmitted is the recall floor.
const minFireAdmitted = 16

// TestIntentGateSeparation asserts the two properties the gate promises, in the
// asymmetric shape intent.go argues for: no QUIET prompt is ever admitted.
func TestIntentGateSeparation(t *testing.T) {
	scored := scoreIntentPrompts(t)
	fired := 0
	for _, p := range scored {
		admitted := p.score >= intentGate
		if admitted && !p.prompt.fire {
			t.Errorf("QUIET prompt admitted: %q scored %.3f, gate %.2f",
				p.prompt.text, p.score, intentGate)
		}
		if admitted && p.prompt.fire {
			fired++
		}
	}
	if fired < minFireAdmitted {
		t.Errorf("gate %.2f admits %d FIRE prompts, want >= %d — see intent.go's derivation",
			intentGate, fired, minFireAdmitted)
	}
	compareIntentGolden(t, intentTable(scored))
}

// scoredPrompt pairs a labelled prompt with what the current scorer gives it.
type scoredPrompt struct {
	prompt intentPrompt
	score  float64
}

// scoreIntentPrompts ranks every labelled prompt against the calibration corpus.
func scoreIntentPrompts(t *testing.T) []scoredPrompt {
	docs := calibrationCorpus(t)
	prompts := loadIntentPrompts(t)
	out := make([]scoredPrompt, 0, len(prompts))
	for _, p := range prompts {
		var score float64
		if cands := recall.Rank(recall.Query{Question: p.text}, docs, calibrationNow); len(cands) > 0 {
			score = cands[0].Score
		}
		out = append(out, scoredPrompt{prompt: p, score: score})
	}
	return out
}

// intentTable renders the two classes sorted by score.
func intentTable(scored []scoredPrompt) string {
	sorted := append([]scoredPrompt(nil), scored...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].score > sorted[j].score })
	var b strings.Builder
	fmt.Fprintf(&b, "Intent gate: %.2f. Corpus %s, %d prompts.\n\n",
		intentGate, calibrationVault, len(sorted))
	b.WriteString("| Score | Class | Admitted | Prompt |\n|---|---|---|---|\n")
	for _, s := range sorted {
		fmt.Fprintf(&b, "| %.3f | %s | %v | %s |\n",
			s.score, intentClass(s.prompt.fire), s.score >= intentGate, s.prompt.text)
	}
	b.WriteString(intentMargin(sorted))
	return b.String()
}

func intentClass(fire bool) string {
	if fire {
		return "FIRE"
	}
	return "QUIET"
}

// intentMargin states where the two classes actually part.
func intentMargin(sorted []scoredPrompt) string {
	lowFire, highQuiet := 1.0, 0.0
	for _, s := range sorted {
		if s.prompt.fire && s.score < lowFire {
			lowFire = s.score
		}
		if !s.prompt.fire && s.score > highQuiet {
			highQuiet = s.score
		}
	}
	return fmt.Sprintf(
		"\nLowest FIRE %.3f, highest QUIET %.3f, margin %.3f — every value from the top of\n"+
			"QUIET to the old 0.7 is false-positive-free here, so this measurement rules out\n"+
			"0.7 but does not pick its replacement. See intent.go for what does.\n",
		lowFire, highQuiet, lowFire-highQuiet)
}

func compareIntentGolden(t *testing.T, got string) {
	t.Helper()
	if *updateGolden {
		if err := os.WriteFile(intentGolden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", intentGolden)
		return
	}
	want, err := os.ReadFile(intentGolden)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestIntentGate -update", err)
	}
	if got != string(want) {
		t.Errorf("intent gate table changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// loadIntentPrompts parses testdata/intent-gate-labels.txt: one "FIRE:" or "QUIET:" line
// per prompt.
func loadIntentPrompts(t *testing.T) []intentPrompt {
	t.Helper()
	f, err := os.Open(intentLabelsPath)
	if err != nil {
		t.Fatalf("%v — the label file is the derivation; without it the gate is a guess", err)
	}
	defer func() { _ = f.Close() }()
	out, sc := []intentPrompt{}, bufio.NewScanner(f)
	for sc.Scan() {
		out = appendIntentLine(t, out, strings.TrimSpace(sc.Text()))
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading %s: %v", intentLabelsPath, err)
	}
	return out
}

func appendIntentLine(t *testing.T, out []intentPrompt, line string) []intentPrompt {
	switch {
	case line == "" || strings.HasPrefix(line, "#"):
		return out
	case strings.HasPrefix(line, "FIRE:"):
		return append(out, intentPrompt{text: trimLabel(line, "FIRE:"), fire: true})
	case strings.HasPrefix(line, "QUIET:"):
		return append(out, intentPrompt{text: trimLabel(line, "QUIET:")})
	}
	t.Fatalf("%s: unparsable line %q", intentLabelsPath, line)
	return out
}

func trimLabel(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}
