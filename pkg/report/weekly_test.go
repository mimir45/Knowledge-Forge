package report

import (
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/linkcheck"
	"github.com/mimir45/Knowledge-Forge/pkg/similarity"
)

func weeklyFixture() WeeklyInput {
	return WeeklyInput{
		Week: 32, Year: 2026,
		Broken: []drift.Finding{
			{Note: "notes/concept/a.md", Ref: "x.go:12", Verdict: drift.Broken, Reason: "file gone"},
		},
		Uncovered: []Uncovered{
			{Symbol: "Widget.Render", Path: "src/widget.go", LOC: 40, Commits: 6},
		},
		UncoveredDays: 90,
		DuplicatePairs: []similarity.Pair{
			{A: "notes/concept/dup1.md", B: "notes/concept/dup2.md", Score: 0.90},
			{A: "notes/concept/low1.md", B: "notes/concept/low2.md", Score: 0.50},
		},
		DeadCitations: []Citation{
			{Status: linkcheck.Status{URL: "https://gone.example/", Verdict: linkcheck.Dead},
				Notes: []string{"notes/howto/b.md"}},
		},
		StaleCount: 2, MergeCandidates: 1, OrphanCount: 3,
		Stats: VaultStats{Notes: 100, HitRate: 42.5, Orphans: 3, Drift: 1},
		Asks: []Ask{
			{Topic: "known-topic", Count: 1, Written: true},
			{Topic: "unwritten-topic", Count: 2, Written: false},
		},
		Slugs: map[string]string{
			"notes/concept/a.md": "a", "notes/howto/b.md": "b",
			"notes/concept/dup1.md": "dup1", "notes/concept/dup2.md": "dup2",
		},
		Now: at,
	}
}

// The title line is a literal format per the original spec, not the generic "Title — date"
// every other renderer uses — a week is keyed by ISOWeek's own year, not the header helper.
func TestWeeklyTitleLineIsTheSpecFormat(t *testing.T) {
	got := string(RenderWeekly(weeklyFixture()))
	if !strings.HasPrefix(got, "# Week 32, 2026\n") {
		t.Fatalf("title line = %q, want the literal ADDENDUM format", firstLines(got, 1))
	}
}

// Act now must rank BROKEN drift first, then undocumented churn, then near-duplicates at
// the strict spec threshold, then dead links — and the 0.50 pair must be filtered out.
func TestWeeklyActNowRanksByKindAndFiltersLooseDuplicates(t *testing.T) {
	got := string(RenderWeekly(weeklyFixture()))
	broken := strings.Index(got, "BROKEN")
	churn := strings.Index(got, "Widget.Render")
	dup := strings.Index(got, "0.90 near-duplicate")
	dead := strings.Index(got, "dead")
	if broken < 0 || churn < 0 || dup < 0 || dead < 0 {
		t.Fatalf("missing an Act now line:\n%s", got)
	}
	if !(broken < churn && churn < dup && dup < dead) {
		t.Errorf("Act now out of order: broken=%d churn=%d dup=%d dead=%d", broken, churn, dup, dead)
	}
	if strings.Contains(got, "0.50 near-duplicate") {
		t.Error("a pair below the 0.85 spec threshold must not appear in Act now")
	}
}

// An empty Act now must explain itself, not print a bare "none" that reads
// like a bug.
func TestWeeklyActNowExplainsWhenEmpty(t *testing.T) {
	in := weeklyFixture()
	in.Broken, in.Uncovered, in.DuplicatePairs, in.DeadCitations = nil, nil, nil, nil
	got := string(RenderWeekly(in))
	for _, want := range []string{"churn window", "0.85 spec threshold"} {
		if !strings.Contains(got, want) {
			t.Errorf("empty Act now should explain itself with %q, got:\n%s", want, got)
		}
	}
}

// A first-ever run has no prior snapshot, and the Vault section must say so rather than
// reporting the vault's entire history as one week's delta.
func TestWeeklyVaultOmitsDeltaOnFirstRun(t *testing.T) {
	got := string(RenderWeekly(weeklyFixture()))
	if strings.Contains(got, "(+100)") {
		t.Error("no Prev snapshot should not produce a delta")
	}
	if !strings.Contains(got, "no prior snapshot") {
		t.Errorf("missing the first-run caveat:\n%s", got)
	}
}

// With a Prev snapshot, deltas must be signed and computed field-by-field, not just repeat
// the current totals.
func TestWeeklyVaultShowsSignedDeltaAgainstPrev(t *testing.T) {
	in := weeklyFixture()
	prev := VaultStats{Notes: 90, HitRate: 40, Orphans: 5, Drift: 3}
	in.Prev = &prev
	got := string(RenderWeekly(in))
	if !strings.Contains(got, "(+10)") {
		t.Errorf("notes delta should be +10, got:\n%s", got)
	}
	if !strings.Contains(got, "(-2)") {
		t.Errorf("orphans delta should be -2, got:\n%s", got)
	}
}

// Gaps must list only unwritten, repeated asks — the same rule gaps.md itself applies.
func TestWeeklyGapsListsOnlyUnwrittenAsks(t *testing.T) {
	got := string(RenderWeekly(weeklyFixture()))
	if !strings.Contains(got, "unwritten-topic") {
		t.Error("an unwritten, repeated ask should appear in Gaps")
	}
	if strings.Contains(got, "known-topic") {
		t.Error("an already-written topic must not appear in Gaps")
	}
}

func TestWeeklyIsByteIdenticalOnRerun(t *testing.T) {
	in := weeklyFixture()
	first := RenderWeekly(in)
	for i := 0; i < 5; i++ {
		if string(RenderWeekly(in)) != string(first) {
			t.Fatal("weekly.md changed between runs on identical input")
		}
	}
}
