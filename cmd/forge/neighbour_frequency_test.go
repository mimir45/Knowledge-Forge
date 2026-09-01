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
// only says ten notes overlapped at all), and, since B-036 widened the neighbour band's
// pool from TopN=10 to NeighbourWindow=20, how many queries exhaust that wider pool itself
// (poolSaturated — recall.NeighbourPool(pool) hits all 20 slots, meaning even the widened
// window ran out of room, not just the floor). calibration.golden's JPA row is the proof
// rankCapped and neighbour admission diverge: it has a nonzero top score (0.119) and 0
// neighbours in the same run.

const freqGoldenPath = "testdata/neighbour-frequency.golden"
const ecosystemsPath = "testdata/query-ecosystems.txt"

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

// minOf mirrors neighbour_floor_test.go's maxOf/median — no shared minimum existed there
// because the sweep never needed one before this file added a per-query summary line.
func minOf(counts []int) int {
	if len(counts) == 0 {
		return 0
	}
	m := counts[0]
	for _, c := range counts[1:] {
		if c < m {
			m = c
		}
	}
	return m
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
	eco := loadQueryEcosystems(t)
	freq := map[string]*freqRow{}
	rankCapped, poolSaturated := 0, 0
	counts := make([]int, 0, len(queries))
	for _, q := range queries {
		pool := recall.RankPool(recall.Query{Question: q.question}, docs, calibrationNow)
		// Equivalent to len(recall.Rank(...)) == recall.TopN: Rank is truncate(pool, TopN),
		// and pool is already nonZero-filtered, so truncation hits TopN iff pool has at
		// least that many entries.
		if len(pool) >= recall.TopN {
			rankCapped++
		}
		nPool := recall.NeighbourPool(pool)
		if len(nPool) == recall.NeighbourWindow {
			poolSaturated++
		}
		ns := recall.DefaultThresholds.Neighbours(nPool)
		counts = append(counts, len(ns))
		for _, n := range ns {
			r, ok := freq[n.Slug]
			if !ok {
				r = &freqRow{slug: n.Slug}
				freq[n.Slug] = r
			}
			r.queries = append(r.queries, q.question)
		}
	}
	return renderFrequencyTable(queries, eco, freq, rankCapped, poolSaturated, counts)
}

// loadQueryEcosystems parses testdata/query-ecosystems.txt — B-036's own named
// prerequisite: which ecosystem cluster each of neighbour-labels.txt's fifteen queries
// belongs to, written and committed before any scoring change. Format mirrors
// loadNeighbourLabels: a "Q:" line followed by an "Ecosystem:" line, "#" and blank lines
// ignored. Keyed by question text so it joins against labelledQuery without depending on
// row order matching between the two files.
func loadQueryEcosystems(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(ecosystemsPath)
	if err != nil {
		t.Fatalf("%v — B-036's ecosystem rollup needs this file; see its header for the labelling rule", err)
	}
	defer func() { _ = f.Close() }()
	out, pending, sc := map[string]string{}, "", bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
			continue
		case strings.HasPrefix(line, "Q:"):
			pending = strings.TrimSpace(strings.TrimPrefix(line, "Q:"))
		case strings.HasPrefix(line, "Ecosystem:"):
			if pending == "" {
				t.Fatalf("%s: Ecosystem: line before any Q: line", ecosystemsPath)
			}
			out[pending] = strings.TrimSpace(strings.TrimPrefix(line, "Ecosystem:"))
			pending = ""
		default:
			t.Fatalf("%s: unparsable line %q", ecosystemsPath, line)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading %s: %v", ecosystemsPath, err)
	}
	return out
}

func renderFrequencyTable(queries []labelledQuery, eco map[string]string, freq map[string]*freqRow, rankCapped, poolSaturated int, neighbourCounts []int) string {
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
	fmt.Fprintf(&b, "Neighbour document frequency over %s — %d queries, floor %.3f,\n"+
		"NeighbourWindow=%d.\n\n", labelsPath, len(queries), recall.DefaultThresholds.Neighbour, recall.NeighbourWindow)
	fmt.Fprintf(&b, "%d/%d queries have recall.Rank return a full TopN=%d nonzero-scored\n"+
		"candidates (before the floor filters anything).\n",
		rankCapped, len(queries), recall.TopN)
	fmt.Fprintf(&b, "%d/%d queries exhaust the widened NeighbourWindow=%d pool itself (even\n"+
		"the window, not just the floor, ran out of room).\n",
		poolSaturated, len(queries), recall.NeighbourWindow)
	fmt.Fprintf(&b, "Neighbours per query: min %d, median %d, max %d.\n\n",
		minOf(neighbourCounts), median(neighbourCounts), maxOf(neighbourCounts))
	b.WriteString("| Note | N of M (all) | Queries (only listed when N > 1) |\n|---|---|---|\n")
	for _, r := range rows {
		list := ""
		if len(r.queries) > 1 {
			list = strings.Join(r.queries, "; ")
		}
		fmt.Fprintf(&b, "| %s | %d/%d | %s |\n", r.slug, len(r.queries), len(queries), list)
	}
	b.WriteString(springClusterRollup(queries, eco, rows))
	return b.String()
}

// springClusterRollup answers B-036's actual hypothesis — "a note scoring on every query
// in an ecosystem" — which the full-fifteen N/M column above cannot: a note scattered
// across unrelated ecosystems and a note universal within one both produce the same raw
// count. `spring` is testdata/query-ecosystems.txt's only cluster with M >= 2 in this
// label set (6 of 15); the others are singletons a ratio can't say anything about.
func springClusterRollup(queries []labelledQuery, eco map[string]string, rows []*freqRow) string {
	total := 0
	for _, q := range queries {
		if eco[q.question] == "spring" {
			total++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\nWithin the `spring` cluster (%d of %d queries; the only cluster in\n"+
		"this label set with M >= 2 — see testdata/query-ecosystems.txt):\n\n", total, len(queries))
	b.WriteString("| Note | N of M (spring) |\n|---|---|\n")
	for _, r := range rows {
		n := 0
		for _, q := range r.queries {
			if eco[q] == "spring" {
				n++
			}
		}
		if n == 0 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d/%d |\n", r.slug, n, total)
	}
	return b.String()
}
