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

// This is the derivation behind intent.go's gate, and the reason the literal there is
// defended by a test rather than by a config key. Promoting it to config.Recall was the
// other option B-033's plan named; it was rejected because printIntent runs on
// UserPromptSubmit under a 50ms budget and loads no config today, so wiring the chain in
// buys a knob nobody turns at the cost of the one budget in the tree that is tight.
//
// A number in a comment rots. A number pinned to labelled prompts fails when it rots.

const (
	intentLabelsPath = "testdata/intent-gate-labels.txt"
	intentGolden     = "testdata/intent-gate.golden"
)

// intentPrompt is one labelled prompt: fire means the vault already answers it.
type intentPrompt struct {
	text string
	fire bool
}

// minFireAdmitted is the recall floor, and it is set to exactly what intentGate measures
// today rather than to a comfortable value below it. That is deliberate: the failure this
// pins is the one that already happened once, when 0.7 decayed to admitting 3 of 10 as
// B-008 moved the scale under it and nothing failed. A tripwire with slack in it would
// have let that happen again more slowly. Any loss of recall stops the build and gets
// argued about; that is the whole job.
//
// Rescaled 8 -> 16 by B-037's wide-sweep plan (docs/TODO.md), when
// intent-gate-labels.txt widened 10 FIRE prompts -> 20. This is a proportional rescale,
// not a loosened bar: the widened set measures 16/20 admitted, the same 80% the original
// 8/10 measured, so 16 pins exactly what's true today, the same way 8 did. Leaving 8 in
// place would have turned the tripwire into slack — 16 of 20 admitted would have passed
// against a floor meant for 10, silently tolerating a real regression down to 8/20 (40%).
const minFireAdmitted = 16

// TestIntentGateSeparation asserts the two properties the gate promises, in the
// asymmetric shape intent.go argues for: no QUIET prompt is ever admitted — that is the
// never-disturb contract and it is absolute — and at least minFireAdmitted FIRE prompts
// are, which is a floor rather than a target.
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

// scoreIntentPrompts ranks every labelled prompt against the calibration corpus. It uses
// calibrationNow for the same reason the calibration harness does: a wall-clock run would
// let staleness move a score and rot the record on a calendar.
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

// intentTable renders the two classes sorted by score, which is the shape the derivation
// argument reads off: the bottom of FIRE and the top of QUIET are the two numbers the
// gate has to sit between.
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

// intentMargin states where the two classes actually part. It is recorded rather than
// asserted because it is the number that shows the labels cannot pick the gate: at 0.005
// wide, every value from the separating point up to the old 0.7 has zero false positives
// on this set, so intent.go chooses above it on an argument, not on this measurement.
// A margin that goes negative is a different finding — the classes would have overlapped
// and no threshold would work at all.
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
