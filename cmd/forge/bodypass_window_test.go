package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
)

// This is B-038's measurement harness: its own "shape when picked up" text asks whether
// raising or removing BodyPassSize ever moves a verdict, "measured at scale," not just on
// the one row B-031's closure spot-checked. It changes nothing shipped — RankPool still
// calls RankPoolWithBodyPass(..., BodyPassSize) and pkg/recall's
// TestRankPoolWithBodyPassMatchesRankPoolAtShippedSize pins that delegation — this file
// only compares today's capped window against an uncapped one and records what differs.
// B-038's own "do not": this does not raise BodyPassSize, does not touch sortByScore's
// path tie-break, and does not pick a tie-break replacement. Question (b) — what the
// tie-break should favor instead of path, if this measurement shows the cap ever binds —
// stays fully open.
//
// Query set: calibration_test.go's 9 calibrationQueries plus the 15 question texts from
// testdata/neighbour-labels.txt (loadNeighbourLabels, .question only — the file's
// ground-truth "- slug" labels are never read here, since this is a diff measurement, not
// a relevance judgment, and needs no new human labelling). No overlap between the two
// sets. Corpus and clock are calibrationCorpus(t) / calibrationNow, reused as-is.
//
// Two caveats that must travel with any reading of this golden, stated once here and
// again in the golden's own header so a reader of just the table doesn't miss them:
//   - the body channel is always active (bodyChannel sets Active: true unconditionally,
//     pkg/recall/score.go), so opening a previously-unopened doc with no body-term hits
//     dilutes its blended score rather than leaving it unchanged — widening the window is
//     not one-directional, and "no change" is not the only outcome the mechanism predicts.
//   - Thresholds.ResultFrom only populates Neighbours on a CREATE verdict, so a
//     capped-CREATE -> uncapped-UPDATE flip shows up as a neighbours change for a verdict
//     reason, not a neighbour-band reason. Both verdict columns are in the row so the two
//     are distinguishable, but a "neighbours changed" count alone conflates them.
const bodyPassGoldenPath = "testdata/bodypass-window.golden"

// TestBodyPassWindowEffect asserts nothing about which window size is correct — that
// argument, if this measurement shows one is needed, belongs in BACKLOG.md, where a human
// writes it. What it asserts is that the measurement stays reproducible: a scorer change
// that moves it fails here until it is re-recorded with -update.
func TestBodyPassWindowEffect(t *testing.T) {
	got := bodyPassTable(t)
	if *updateGolden {
		if err := os.WriteFile(bodyPassGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", bodyPassGoldenPath)
		return
	}
	want, err := os.ReadFile(bodyPassGoldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestBodyPassWindowEffect -update", err)
	}
	if got != string(want) {
		t.Errorf("bodypass window table changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// bodyPassQueries returns the 9 calibration queries plus neighbour-labels.txt's 15
// question texts — 24 distinct queries, wider than calibration.golden alone, per B-038's
// own unblock condition. Ground-truth neighbour labels are discarded; only .question is
// read.
func bodyPassQueries(t *testing.T) []string {
	qs := append([]string{}, calibrationQueries...)
	for _, lq := range loadNeighbourLabels(t) {
		qs = append(qs, lq.question)
	}
	return qs
}

// bodyPassRow is one query's capped (BodyPassSize=20, today's shipped behavior) vs
// uncapped (every doc gets a body pass) comparison.
type bodyPassRow struct {
	question                                     string
	tieAtZero, shutOut                           int
	cappedSlug, cappedScore, cappedVerdict       string
	fullSlug, fullScore, fullVerdict             string
	top1Changed, candidatesChanged, neighChanged bool
}

func bodyPassTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	queries := bodyPassQueries(t)
	rows := make([]bodyPassRow, len(queries))
	for i, q := range queries {
		rows[i] = computeBodyPassRow(docs, q)
	}

	shutOutN, top1N, candN, neighN := 0, 0, 0, 0
	for _, r := range rows {
		if r.shutOut > 0 {
			shutOutN++
		}
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

	var b strings.Builder
	fmt.Fprintf(&b, "BodyPass window effect over %d queries (9 from calibration_test.go's "+
		"calibrationQueries, 15 from testdata/neighbour-labels.txt's Q: lines only — "+
		"ground-truth labels not used), capped at BodyPassSize=%d vs uncapped (every doc "+
		"body-scored).\n", len(rows), recall.BodyPassSize)
	b.WriteString("Caveats: the body channel is always active, so opening more docs is " +
		"not one-directionally helpful — a previously-unopened doc with no body-term hits " +
		"is diluted, not left unchanged. Neighbours is populated on a CREATE verdict only, " +
		"so a capped-CREATE -> uncapped-UPDATE flip reads as a neighbours change for a " +
		"verdict reason, not a neighbour-band reason; compare the Verdict columns before " +
		"attributing a neighbours change to the band itself.\n")
	fmt.Fprintf(&b, "%d/%d rows have ShutOut>0 (the tie-break actually decided who got a "+
		"body pass). %d/%d Top-1 changed. %d/%d Candidates changed. %d/%d Neighbours "+
		"changed.\n\n", shutOutN, len(rows), top1N, len(rows), candN, len(rows), neighN, len(rows))
	b.WriteString("| Query | TieAtZero | ShutOut | Top-1 capped | Top-1 uncapped | Top1Δ | CandidatesΔ | NeighboursΔ |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		b.WriteString(renderBodyPassRow(r))
	}
	return b.String()
}

func computeBodyPassRow(docs []recall.Doc, question string) bodyPassRow {
	q := recall.Query{Question: question}
	nonzero := len(recall.RankPoolWithBodyPass(q, docs, calibrationNow, 0))
	tieAtZero := len(docs) - nonzero
	slotsForZeroGroup := recall.BodyPassSize - nonzero
	if slotsForZeroGroup < 0 {
		slotsForZeroGroup = 0
	}
	shutOut := tieAtZero - slotsForZeroGroup
	if shutOut < 0 {
		shutOut = 0
	}

	pool20 := recall.RankPoolWithBodyPass(q, docs, calibrationNow, recall.BodyPassSize)
	poolFull := recall.RankPoolWithBodyPass(q, docs, calibrationNow, len(docs))
	res20 := recall.DefaultThresholds.ResultFrom(q, pool20)
	resFull := recall.DefaultThresholds.ResultFrom(q, poolFull)

	r := bodyPassRow{question: question, tieAtZero: tieAtZero, shutOut: shutOut}
	r.cappedSlug, r.cappedScore, r.cappedVerdict = top1Of(res20)
	r.fullSlug, r.fullScore, r.fullVerdict = top1Of(resFull)
	r.top1Changed = r.cappedSlug != r.fullSlug || r.cappedScore != r.fullScore
	r.candidatesChanged = !equalSlugs(res20.Candidates, resFull.Candidates)
	r.neighChanged = !equalSlugs(res20.Neighbours, resFull.Neighbours)
	return r
}

func top1Of(res recall.Result) (slug, score, verdict string) {
	verdict = string(res.Verdict)
	if len(res.Candidates) == 0 {
		return "(none)", "—", verdict
	}
	return res.Candidates[0].Slug, fmt.Sprintf("%.3f", res.TopScore), verdict
}

func equalSlugs(a, b []recall.Candidate) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Slug != b[i].Slug {
			return false
		}
	}
	return true
}

func renderBodyPassRow(r bodyPassRow) string {
	return fmt.Sprintf("| %s | %d | %d | %s (%s, %s) | %s (%s, %s) | %v | %v | %v |\n",
		r.question, r.tieAtZero, r.shutOut,
		r.cappedSlug, r.cappedScore, r.cappedVerdict,
		r.fullSlug, r.fullScore, r.fullVerdict,
		r.top1Changed, r.candidatesChanged, r.neighChanged)
}
