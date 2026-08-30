package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"knowledge-forge/pkg/recall"
)

// This is B-038 question (b)'s Phase 1 measurement harness: does an alternative
// tie-break (recency, verified date, document-frequency — pkg/recall/tiebreak.go)
// change what the path tie-break decides, and by how much? It changes nothing shipped:
// sortByScore still uses PathTieBreak alone; this file only compares it against the
// three alternatives via RankPoolWithTieBreak/BodyPassWindow, both new measurement-only
// seams (see rank.go's "review flag, not convenience" comments on both).
//
// Query set: the same 24 queries as bodypass_window_test.go's bodyPassQueries() — 9
// calibration + 15 from neighbour-labels.txt, text only.
//
// A predicted invariant this file's first draft got wrong, corrected the same run rather
// than forced to pass: the draft assumed an alternative tie-break can only ever change
// WHICH candidates fill the window, a strict subset of B-038 question (a)'s uncapped
// measurement (0/24 Top-1 changes). That's true of window-membership effects, but it
// missed a second, more direct one — the tie-break also decides the FINAL order among
// candidates that land on the exact same score after the body pass (e.g. two docs both
// scoring 0.170), which can move Top-1's identity with no window-membership change at
// all and no verdict change either. Measured: 2/72 rows (both under Recency and
// Verified, same query, same score, CREATE→CREATE both sides — see Top1Changed below).
// This file now measures Top-1 stability instead of asserting it.
//
// DocFreq saturation rate is the fraction of tied-at-zero pairs DocFreqTieBreak cannot
// separate (returns 0 — either untagged, or an exact specificity tie). BACKLOG B-036's
// closure note found a related idf-based signal saturated heavily for a different
// purpose; this is the same risk, measured directly rather than assumed not to apply.
const tiebreakGoldenPath = "testdata/tiebreak-comparison.golden"

// TestTieBreakComparison records the full comparison in a golden table, gated behind the
// shared -update flag. It asserts nothing about which strategy is better — that argument
// belongs in BACKLOG.md, where a human writes it — only that the measurement stays
// reproducible, the same contract every other golden test in this package makes.
func TestTieBreakComparison(t *testing.T) {
	docs := calibrationCorpus(t)
	queries := bodyPassQueries(t)
	rows := allTieBreakRows(docs, queries)
	got := renderTieBreakTable(queries, rows)
	if *updateGolden {
		if err := os.WriteFile(tiebreakGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", tiebreakGoldenPath)
		return
	}
	want, err := os.ReadFile(tiebreakGoldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestTieBreakComparison -update", err)
	}
	if got != string(want) {
		t.Errorf("tiebreak comparison table changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// tieBreakStrategies are the three alternatives to PathTieBreak this item measures.
func tieBreakStrategies(docs []recall.Doc) []struct {
	name string
	tb   recall.TieBreak
} {
	return []struct {
		name string
		tb   recall.TieBreak
	}{
		{"Recency", recall.RecencyTieBreak},
		{"Verified", recall.VerifiedTieBreak},
		{"DocFreq", recall.DocFreqTieBreak(docs)},
	}
}

// tieBreakRow is one (query, strategy) comparison against the path baseline.
type tieBreakRow struct {
	question, strategy              string
	tiedAtZero                      int
	windowAdded, windowRemoved      []string
	top1Changed                     bool
	candidatesChanged, neighChanged bool
	outboxIn, cqrsIn                bool
	saturation                      float64
}

func allTieBreakRows(docs []recall.Doc, queries []string) []tieBreakRow {
	var rows []tieBreakRow
	for _, question := range queries {
		q := recall.Query{Question: question}
		tied := tiedAtZeroDocs(docs, q)
		pathWin := windowPaths(docs, q, recall.PathTieBreak)
		for _, s := range tieBreakStrategies(docs) {
			rows = append(rows, computeTieBreakRow(docs, q, s.name, s.tb, tied, pathWin))
		}
	}
	return rows
}

// computeTieBreakRow compares one strategy's window/output against the path baseline for
// one query. tied and pathWin are computed once per query by the caller (they do not
// depend on strategy) so a 3-strategy sweep does not re-score the corpus 3x per query.
func computeTieBreakRow(docs []recall.Doc, q recall.Query, strategy string, tb recall.TieBreak,
	tied []recall.Doc, pathWin map[string]bool) tieBreakRow {
	altWin := windowPaths(docs, q, tb)
	added, removed := windowDelta(tied, pathWin, altWin)
	path20 := recall.RankPoolWithTieBreak(q, docs, calibrationNow, recall.BodyPassSize, recall.PathTieBreak)
	alt20 := recall.RankPoolWithTieBreak(q, docs, calibrationNow, recall.BodyPassSize, tb)
	resPath := recall.DefaultThresholds.ResultFrom(q, path20)
	resAlt := recall.DefaultThresholds.ResultFrom(q, alt20)
	pSlug, pScore, pVerdict := top1Of(resPath)
	aSlug, aScore, aVerdict := top1Of(resAlt)
	return tieBreakRow{
		question: q.Question, strategy: strategy, tiedAtZero: len(tied),
		windowAdded: added, windowRemoved: removed,
		top1Changed:       pSlug != aSlug || pScore != aScore || pVerdict != aVerdict,
		candidatesChanged: !equalSlugs(resPath.Candidates, resAlt.Candidates),
		neighChanged:      !equalSlugs(resPath.Neighbours, resAlt.Neighbours),
		outboxIn:          inWindow(docs, altWin, "transactional-outbox-pattern"),
		cqrsIn:            inWindow(docs, altWin, "cqrs-and-event-driven-messaging"),
		saturation:        docFreqSaturationRate(docs, tied),
	}
}

// tiedAtZeroDocs is the set of docs whose frontmatter-only score is 0 for this query —
// the population whose body-pass admission the path tie-break decides today.
// RankPoolWithBodyPass(..., 0) never opens any doc (min(len,0)=0), so its output is
// exactly the nonzero-frontmatter-scoring set; everything else ties at zero.
func tiedAtZeroDocs(docs []recall.Doc, q recall.Query) []recall.Doc {
	nonzero := map[string]bool{}
	for _, c := range recall.RankPoolWithBodyPass(q, docs, calibrationNow, 0) {
		nonzero[c.Path] = true
	}
	var out []recall.Doc
	for _, d := range docs {
		if !nonzero[d.Rel] {
			out = append(out, d)
		}
	}
	return out
}

// windowPaths is the set of Paths BodyPassWindow would open under tb, as a lookup set.
func windowPaths(docs []recall.Doc, q recall.Query, tb recall.TieBreak) map[string]bool {
	out := map[string]bool{}
	for _, p := range recall.BodyPassWindow(q, docs, calibrationNow, recall.BodyPassSize, tb) {
		out[p] = true
	}
	return out
}

// windowDelta reports, restricted to the tied-at-zero subset, which slugs the alt
// strategy's window admits that path's didn't (added) and vice versa (removed) — the
// population a path-vs-content tie-break can actually move.
func windowDelta(tied []recall.Doc, pathWin, altWin map[string]bool) (added, removed []string) {
	for _, d := range tied {
		switch {
		case altWin[d.Rel] && !pathWin[d.Rel]:
			added = append(added, d.Slug)
		case pathWin[d.Rel] && !altWin[d.Rel]:
			removed = append(removed, d.Slug)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

// inWindow reports whether the note with this slug is in win. Both of this item's own
// motivating notes are checked this way per query/strategy.
func inWindow(docs []recall.Doc, win map[string]bool, slug string) bool {
	for _, d := range docs {
		if d.Slug == slug {
			return win[d.Rel]
		}
	}
	return false
}

// docFreqSaturationRate is the fraction of pairs within the tied-at-zero group that
// DocFreqTieBreak cannot separate (returns 0 — either side untagged, or an exact
// specificity tie). Independent of which strategy the row is being compared against, so
// it is the same number for all three strategy rows of one query; only the DocFreq row
// prints it (renderTieBreakRow).
func docFreqSaturationRate(docs []recall.Doc, tied []recall.Doc) float64 {
	tb := recall.DocFreqTieBreak(docs)
	tiedPairs, total := 0, 0
	for i := 0; i < len(tied); i++ {
		for j := i + 1; j < len(tied); j++ {
			total++
			if tb(recall.Candidate{Path: tied[i].Rel}, recall.Candidate{Path: tied[j].Rel}) == 0 {
				tiedPairs++
			}
		}
	}
	return ratio(tiedPairs, total)
}

func renderTieBreakTable(queries []string, rows []tieBreakRow) string {
	var b strings.Builder
	b.WriteString(tieBreakHeader(len(queries), rows))
	for _, r := range rows {
		b.WriteString(renderTieBreakRow(r))
	}
	return b.String()
}

func tieBreakHeader(nQueries int, rows []tieBreakRow) string {
	top1N, candN, neighN := 0, 0, 0
	for _, r := range rows {
		if r.top1Changed {
			top1N++
		}
		if r.candidatesChanged {
			candN++
		}
		if r.neighChanged {
			neighN++
		}
	}
	return fmt.Sprintf(
		"Tie-break comparison over %d queries x 3 strategies = %d rows.\n"+
			"%d/%d Top1 changed, %d/%d Candidates changed, %d/%d Neighbours changed.\n"+
			"WindowAdded/WindowRemoved are restricted to the tied-at-zero subset (path can't\n"+
			"move a nonzero-frontmatter candidate). DocFreqSaturation is only printed on the\n"+
			"DocFreq row (it doesn't depend on the strategy being compared).\n\n"+
			"| Query | Strategy | TiedAtZero | WindowAdded | WindowRemoved | Top1Δ | CandidatesΔ | NeighboursΔ | OutboxInWindow | CQRSInWindow | DocFreqSaturation |\n"+
			"|---|---|---|---|---|---|---|---|---|---|---|\n",
		nQueries, len(rows), top1N, len(rows), candN, len(rows), neighN, len(rows))
}

func renderTieBreakRow(r tieBreakRow) string {
	sat := "—"
	if r.strategy == "DocFreq" {
		sat = fmt.Sprintf("%.3f", r.saturation)
	}
	return fmt.Sprintf("| %s | %s | %d | %s | %s | %v | %v | %v | %v | %v | %s |\n",
		r.question, r.strategy, r.tiedAtZero,
		joinOrNone(r.windowAdded), joinOrNone(r.windowRemoved),
		r.top1Changed, r.candidatesChanged, r.neighChanged, r.outboxIn, r.cqrsIn, sat)
}

func joinOrNone(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}
