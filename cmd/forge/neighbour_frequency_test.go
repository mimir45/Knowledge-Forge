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
// It also counts how many queries hit recall.Rank's TopN=10 truncation before
// Neighbours ever sees an 11th candidate. A document-frequent note excluded from
// Neighbours' output can only be replaced by a note Rank never computed in the first
// place when a query is TopN-capped — so this number decides whether "filter inside
// Neighbours" can even work, independent of whether filtering is a good idea.

const freqGoldenPath = "testdata/neighbour-frequency.golden"

// TestNeighbourDocumentFrequency reports, for every note admitted as a neighbour in at
// least one of the fifteen testdata/neighbour-labels.txt queries, how many of those
// queries admit it. It asserts nothing about which count is too high — that judgment
// belongs in BACKLOG.md — only that the count stays reproducible.
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

// frequencyTable scores the same fifteen labelled queries TestNeighbourFloorSweep uses,
// at the shipped floor, and counts how many queries admit each slug as a neighbour and
// how many queries hit Rank's TopN cap.
func frequencyTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	queries := loadNeighbourLabels(t)
	freq := map[string]int{}
	capped := 0
	for _, q := range queries {
		ranked := recall.Rank(recall.Query{Question: q.question}, docs, calibrationNow)
		if len(ranked) == recall.TopN {
			capped++
		}
		for _, n := range recall.DefaultThresholds.Neighbours(ranked) {
			freq[n.Slug]++
		}
	}
	return renderFrequencyTable(queries, freq, capped)
}

func renderFrequencyTable(queries []labelledQuery, freq map[string]int, capped int) string {
	type row struct {
		slug string
		n    int
	}
	rows := make([]row, 0, len(freq))
	for slug, n := range freq {
		rows = append(rows, row{slug, n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].n != rows[j].n {
			return rows[i].n > rows[j].n
		}
		return rows[i].slug < rows[j].slug
	})
	var b strings.Builder
	fmt.Fprintf(&b, "Neighbour document frequency over %s — %d queries, floor %.3f.\n",
		labelsPath, len(queries), recall.DefaultThresholds.Neighbour)
	fmt.Fprintf(&b, "%d/%d queries hit Rank's TopN=%d cap (emitted the maximum candidates\n"+
		"Neighbours could ever see, before the floor filters anything).\n\n",
		capped, len(queries), recall.TopN)
	b.WriteString("| Note | Appears in N of M queries |\n|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %d/%d |\n", r.slug, r.n, len(queries))
	}
	return b.String()
}
