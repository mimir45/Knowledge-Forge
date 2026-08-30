package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"knowledge-forge/pkg/recall"
)

// This is B-038 question (b)'s second Phase 1 harness: does an alternative tie-break
// change how well the neighbour band (0.150, unmoved) recovers testdata/
// neighbour-labels.txt's ground-truth wanted slugs? tiebreak_comparison_test.go answers
// "does the output change"; this file answers "is the change better or worse," reusing
// neighbour_floor_test.go's own precision/recall/F1 machinery (hits, ratio, f1, median,
// maxOf, loadNeighbourLabels) rather than re-deriving it — one strategy per row instead
// of one floor per row, same 15 labelled queries, same fixed floor (0.150).
//
// Phase 2's decision rule treats Recency and Verified as a noise floor: both are
// unmotivated reorderings of the same window, so their ΔF1 vs Path brackets the movement
// attributable to "any reordering" alone. DocFreq only counts as a real signal if its
// ΔF1 sits clearly outside that bracket — see docs/BACKLOG.md's B-038 entry for the
// verdict, not this file, which only measures.
const tiebreakPrecisionGoldenPath = "testdata/tiebreak-neighbour-precision.golden"

func TestTieBreakNeighbourPrecision(t *testing.T) {
	got := tieBreakPrecisionTable(t)
	if *updateGolden {
		if err := os.WriteFile(tiebreakPrecisionGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", tiebreakPrecisionGoldenPath)
		return
	}
	want, err := os.ReadFile(tiebreakPrecisionGoldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestTieBreakNeighbourPrecision -update", err)
	}
	if got != string(want) {
		t.Errorf("tiebreak neighbour precision changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// tieBreakPrecisionStrategies includes Path as an explicit row — unlike
// tiebreak_comparison_test.go, which only ever measures alternatives *against* it, this
// table's whole point is comparing all four numbers side by side.
func tieBreakPrecisionStrategies(docs []recall.Doc) []struct {
	name string
	tb   recall.TieBreak
} {
	return []struct {
		name string
		tb   recall.TieBreak
	}{
		{"Path", recall.PathTieBreak},
		{"Recency", recall.RecencyTieBreak},
		{"Verified", recall.VerifiedTieBreak},
		{"DocFreq", recall.DocFreqTieBreak(docs)},
	}
}

func tieBreakPrecisionTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	queries := loadNeighbourLabels(t)
	var b strings.Builder
	b.WriteString(tieBreakPrecisionHeader(queries))
	for _, s := range tieBreakPrecisionStrategies(docs) {
		b.WriteString(tieBreakPrecisionRow(docs, queries, s.name, s.tb))
	}
	return b.String()
}

func tieBreakPrecisionHeader(queries []labelledQuery) string {
	return fmt.Sprintf(
		"Neighbour precision/recall by tie-break strategy, floor fixed at %.3f "+
			"(DefaultThresholds.Neighbour, unmoved) — %s, %d queries, %d labelled "+
			"neighbours. Path is today's shipped row; the other three are B-038 question "+
			"(b)'s candidates. Precision/recall are micro-averaged, same convention as "+
			"neighbour-sweep.golden.\n\n"+
			"| Strategy | Emitted | TP | Precision | Recall | F1 | Median/query | Max/query | Empty queries |\n"+
			"|---|---|---|---|---|---|---|---|---|\n",
		recall.DefaultThresholds.Neighbour, labelsPath, len(queries), wantTotal(queries))
}

// tieBreakPrecisionRow mirrors sweepRow (neighbour_floor_test.go) with strategy in place
// of floor: same Thresholds.Neighbours(NeighbourPool(...)) call, same metrics.
func tieBreakPrecisionRow(docs []recall.Doc, queries []labelledQuery, strategy string, tb recall.TieBreak) string {
	th := recall.DefaultThresholds
	var emitted, tp, empty int
	counts := make([]int, 0, len(queries))
	for _, q := range queries {
		pool := recall.RankPoolWithTieBreak(recall.Query{Question: q.question}, docs, calibrationNow,
			recall.BodyPassSize, tb)
		ns := th.Neighbours(recall.NeighbourPool(pool))
		emitted += len(ns)
		tp += hits(ns, q.want)
		counts = append(counts, len(ns))
		if len(ns) == 0 {
			empty++
		}
	}
	p, r := ratio(tp, emitted), ratio(tp, wantTotal(queries))
	return fmt.Sprintf("| %s | %d | %d | %.3f | %.3f | %.3f | %d | %d | %d |\n",
		strategy, emitted, tp, p, r, f1(p, r), median(counts), maxOf(counts), empty)
}
