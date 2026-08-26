package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"knowledge-forge/pkg/recall"
)

// This is B-036's own unblock-condition harness — docs/TODO.md names it explicitly: "a
// per-note 'appears in N of M query results' column added to TestNeighbourFloorSweep's
// harness." It lives in its own file rather than inside neighbour_floor_test.go because
// it isn't a floor sweep: the floor is fixed at recall.DefaultThresholds.Neighbour and
// nothing varies. It answers a different question — does a note's neighbour eligibility
// depend on how many unrelated queries it turns up for, a document-frequency property
// recall-spec.md §2.3.1 computes for terms and never for notes.
//
// B-036's hypothesis was "a note scoring on every query in an ecosystem" — not on every
// query, full stop. A raw N/M count over a query set spanning several unrelated
// ecosystems (Spring, Keycloak, React, Docker, ...) cannot tell those apart: a note at
// 5/15 could be scattered noise, or it could be 5/5 on exactly the Spring-flavored
// subset, which is the hypothesis confirmed, not refuted. So for every slug that appears
// in more than one query, this records *which* queries — the golden is read alongside
// them, not just at its counts.
//
// It also counts two different things queries can saturate, which are not the same
// number: how many queries return a full TopN=10 *nonzero-scored* candidates from
// recall.Rank (rankCapped — recall.Rank returns nonZero(cands) truncated to TopN, so this
// only says ten notes overlapped at all), and how many queries emit a full TopN=10
// *neighbours* after Thresholds.Neighbours filters by the 0.150 floor (neighbourCapped —
// this is the number that actually bears on "did the floor run out of room before the
// window did"). calibration.golden's JPA row is the proof they diverge: it has a nonzero
// top score (0.119) and 0 neighbours in the same run.

const freqGoldenPath = "testdata/neighbour-frequency.golden"

// TestNeighbourDocumentFrequency reports, for every note admitted as a neighbour in at
// least one of the fifteen testdata/neighbour-labels.txt queries, how many of those
// queries admit it and which ones. It asserts nothing about which count is too high —
// that judgment belongs in BACKLOG.md — only that the count and the query list stay
// reproducible.
func TestNeighbourDocumentFrequency(t *testing.T) {
	got := frequencyTable(t)
	if *updateGolden {
		if err := os.WriteFile(freqGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", freqGoldenPath)
		return
	}
	want, err := os.ReadFile(freqGoldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestNeighbourDocumentFrequency -update", err)
	}
	if got != string(want) {
		t.Errorf("neighbour document-frequency changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// freqRow is one note's frequency plus which queries admitted it — the query list is
// what lets a reader tell "scattered" from "ecosystem-universal" apart; the count alone
// cannot.
type freqRow struct {
	slug    string
	queries []string
}

// frequencyTable scores the same fifteen labelled queries TestNeighbourFloorSweep uses,
// at the shipped floor, and counts both saturation shapes plus per-slug membership.
func frequencyTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	queries := loadNeighbourLabels(t)
	freq := map[string]*freqRow{}
	rankCapped, neighbourCapped := 0, 0
	for _, q := range queries {
		ranked := recall.Rank(recall.Query{Question: q.question}, docs, calibrationNow)
		if len(ranked) == recall.TopN {
			rankCapped++
		}
		ns := recall.DefaultThresholds.Neighbours(ranked)
		if len(ns) == recall.TopN {
			neighbourCapped++
		}
		for _, n := range ns {
			r, ok := freq[n.Slug]
			if !ok {
				r = &freqRow{slug: n.Slug}
				freq[n.Slug] = r
			}
			r.queries = append(r.queries, q.question)
		}
	}
	return renderFrequencyTable(queries, freq, rankCapped, neighbourCapped)
}

func renderFrequencyTable(queries []labelledQuery, freq map[string]*freqRow, rankCapped, neighbourCapped int) string {
	rows := make([]*freqRow, 0, len(freq))
	for _, r := range freq {
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if len(rows[i].queries) != len(rows[j].queries) {
			return len(rows[i].queries) > len(rows[j].queries)
		}
		return rows[i].slug < rows[j].slug
	})
	var b strings.Builder
	fmt.Fprintf(&b, "Neighbour document frequency over %s — %d queries, floor %.3f.\n\n",
		labelsPath, len(queries), recall.DefaultThresholds.Neighbour)
	fmt.Fprintf(&b, "%d/%d queries have recall.Rank return a full TopN=%d nonzero-scored\n"+
		"candidates (before the floor filters anything).\n",
		rankCapped, len(queries), recall.TopN)
	fmt.Fprintf(&b, "%d/%d queries emit a full TopN=%d neighbours after the floor (the\n"+
		"number that says the window, not just the floor, ran out of room).\n\n",
		neighbourCapped, len(queries), recall.TopN)
	b.WriteString("| Note | N of M | Queries (only listed when N > 1) |\n|---|---|---|\n")
	for _, r := range rows {
		list := ""
		if len(r.queries) > 1 {
			list = strings.Join(r.queries, "; ")
		}
		fmt.Fprintf(&b, "| %s | %d/%d | %s |\n", r.slug, len(r.queries), len(queries), list)
	}
	return b.String()
}
