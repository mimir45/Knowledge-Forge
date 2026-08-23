package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"knowledge-forge/pkg/recall"
)

// This is B-033's derivation harness. It exists as a test rather than as numbers in a
// commit message for one reason: B-032 is next in the queue and moves blend's
// denominator, so every score shifts and the floor has to be re-derived against a scale
// that does not exist yet. A sweep that lives in prose cannot be re-run; this one can,
// with -update, and the diff is the review surface.
//
// It deliberately does not route through calibration.golden. The floor is a property of
// Thresholds.Neighbours, so the sweep constructs Thresholds directly per candidate value
// and never touches Rank or Decide — the two the floor must not influence.

const (
	labelsPath      = "testdata/neighbour-labels.txt"
	sweepGoldenPath = "testdata/neighbour-sweep.golden"
)

// labelledQuery is one row of neighbour-labels.txt: an adjacent-topic question and the
// notes a new note on that topic should link to, decided before any score was measured.
type labelledQuery struct {
	question string
	want     map[string]bool
}

// TestNeighbourFloorSweep reports precision, recall and link volume at each candidate
// floor. It asserts nothing about which floor is best — that argument belongs in
// BACKLOG.md, where a human wrote it. What it asserts is that the argument stays
// reproducible: a scorer change that moves the sweep fails here until it is re-recorded.
func TestNeighbourFloorSweep(t *testing.T) {
	got := sweepTable(t)
	if *updateGolden {
		if err := os.WriteFile(sweepGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", sweepGoldenPath)
		return
	}
	want, err := os.ReadFile(sweepGoldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestNeighbourFloorSweep -update", err)
	}
	if got != string(want) {
		t.Errorf("neighbour sweep changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// sweepFloors are the candidate values. The band is >= Neighbour && < Update, so 0.55 is
// excluded: at the update threshold itself the band is empty and emits nothing, which
// validate.go:126 permits but no derivation would choose.
func sweepFloors() []float64 {
	var out []float64
	for v := 0.100; v < 0.5001; v += 0.025 {
		out = append(out, v)
	}
	return out
}

// sweepTable scores the fifteen labelled queries once, then re-filters the same candidate
// lists at each floor. Scoring once is not just speed: it makes every row a function of
// one ranking, so a difference between rows is the floor and nothing else.
func sweepTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	queries := loadNeighbourLabels(t)
	ranked := make([][]recall.Candidate, len(queries))
	for i, q := range queries {
		ranked[i] = recall.Rank(recall.Query{Question: q.question}, docs, calibrationNow)
	}
	var b strings.Builder
	b.WriteString(sweepHeader(queries))
	for _, floor := range sweepFloors() {
		b.WriteString(sweepRow(queries, ranked, floor))
	}
	return b.String()
}

func sweepHeader(queries []labelledQuery) string {
	return fmt.Sprintf(
		"Neighbour floor sweep over %s — %d queries, %d labelled neighbours.\n"+
			"Precision and recall are micro-averaged: summed over queries, not averaged\n"+
			"per query, so a query with five expected neighbours weighs five times one\n"+
			"with a single expected neighbour.\n\n"+
			"| Floor | Emitted | TP | Precision | Recall | F1 | Median/query | Max/query | Empty queries |\n"+
			"|---|---|---|---|---|---|---|---|---|\n",
		labelsPath, len(queries), wantTotal(queries))
}

// sweepRow measures one floor across every query.
func sweepRow(queries []labelledQuery, ranked [][]recall.Candidate, floor float64) string {
	th := recall.Thresholds{Answer: 0.85, Update: 0.55, Neighbour: floor}
	var emitted, tp, empty int
	counts := make([]int, 0, len(queries))
	for i, q := range queries {
		ns := th.Neighbours(ranked[i])
		emitted += len(ns)
		tp += hits(ns, q.want)
		counts = append(counts, len(ns))
		if len(ns) == 0 {
			empty++
		}
	}
	p, r := ratio(tp, emitted), ratio(tp, wantTotal(queries))
	return fmt.Sprintf("| %.3f | %d | %d | %.3f | %.3f | %.3f | %d | %d | %d |\n",
		floor, emitted, tp, p, r, f1(p, r), median(counts), maxOf(counts), empty)
}

// hits counts emitted neighbours that the label file expected.
func hits(ns []recall.Candidate, want map[string]bool) int {
	n := 0
	for _, c := range ns {
		if want[c.Slug] {
			n++
		}
	}
	return n
}

func ratio(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}

func f1(p, r float64) float64 {
	if p+r == 0 {
		return 0
	}
	return 2 * p * r / (p + r)
}

func median(counts []int) int {
	if len(counts) == 0 {
		return 0
	}
	s := append([]int(nil), counts...)
	sort.Ints(s)
	return s[len(s)/2]
}

func maxOf(counts []int) int {
	m := 0
	for _, c := range counts {
		if c > m {
			m = c
		}
	}
	return m
}

func wantTotal(queries []labelledQuery) int {
	n := 0
	for _, q := range queries {
		n += len(q.want)
	}
	return n
}

// loadNeighbourLabels parses testdata/neighbour-labels.txt. The format is deliberately
// not YAML or JSON: the file is read by humans deciding whether a label is honest, and a
// "Q:" line followed by "- slug" lines is the shape that survives review.
func loadNeighbourLabels(t *testing.T) []labelledQuery {
	t.Helper()
	f, err := os.Open(labelsPath)
	if err != nil {
		t.Fatalf("%v — the label file is the derivation; without it there is no sweep", err)
	}
	defer func() { _ = f.Close() }()
	out, sc := []labelledQuery{}, bufio.NewScanner(f)
	for sc.Scan() {
		out = appendLabelLine(t, out, strings.TrimSpace(sc.Text()))
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading %s: %v", labelsPath, err)
	}
	return out
}

// appendLabelLine folds one line into the query list. A "- slug" line before any "Q:"
// line is a malformed file, not an empty label set, so it fails loudly.
func appendLabelLine(t *testing.T, out []labelledQuery, line string) []labelledQuery {
	switch {
	case line == "" || strings.HasPrefix(line, "#"):
		return out
	case strings.HasPrefix(line, "Q:"):
		q := strings.TrimSpace(strings.TrimPrefix(line, "Q:"))
		return append(out, labelledQuery{question: q, want: map[string]bool{}})
	case strings.HasPrefix(line, "- "):
		if len(out) == 0 {
			t.Fatalf("%s: neighbour %q appears before any Q: line", labelsPath, line)
		}
		out[len(out)-1].want[strings.TrimSpace(line[2:])] = true
		return out
	}
	t.Fatalf("%s: unparsable line %q", labelsPath, line)
	return out
}
