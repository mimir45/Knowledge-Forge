package report

import (
	"strings"
	"testing"
	"time"
)

func day(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

var now = day("2026-08-09")

func sample() IndexInput {
	return IndexInput{
		Now: now,
		Entries: []Entry{
			{Rel: "notes/concept/a.md", Slug: "a", Title: "A", Type: "concept",
				Stack: []string{"java", "spring-boot"}, Updated: day("2026-08-01"),
				Verified: day("2026-08-01"), FreshnessDays: 365, Valid: true},
			{Rel: "notes/howto/b.md", Slug: "b", Title: "B", Type: "howto",
				Stack: []string{"java"}, Updated: day("2026-07-01"),
				Verified: day("2024-01-01"), FreshnessDays: 180, Valid: true},
			{Rel: "notes/pitfall/c.md", Slug: "c", Title: "C", Type: "pitfall",
				Stack: []string{"docker"}, Updated: day("2026-06-01"),
				Verified: day("2026-06-01"), FreshnessDays: 365, Valid: false, Orphan: true},
		},
	}
}

func TestRenderIndexSections(t *testing.T) {
	got := string(RenderIndex(sample()))
	for _, want := range []string{
		"# Vault index — 3 notes — rebuilt 2026-08-09",
		"2 contract-valid · 1 failing · 1 orphaned",
		"## By stack",
		"- **java** — 2",
		"## Recently updated",
		"- [[a]] — 2026-08-01",
		"## Stale — 1 past their freshness window",
		"- [[b]] — verified 2024-01-01 (180d budget)",
		"## Gaps — asked but never written",
		"_no ask log yet_",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// TestRenderIndexIsIdempotent: the header carries a date, not a timestamp, so two runs
// on the same day are byte-identical and `forge index` never bumps the file's mtime.
func TestRenderIndexIsIdempotent(t *testing.T) {
	in := sample()
	first := string(RenderIndex(in))
	for i := 0; i < 20; i++ {
		in.Now = now.Add(time.Duration(i) * time.Minute)
		if got := string(RenderIndex(in)); got != first {
			t.Fatalf("run %d differed:\n%s", i, got)
		}
	}
}

// TestRenderIndexDeterministicTies: stack counts tie constantly in a small vault; ties
// break on key so the output cannot reorder between runs on map iteration.
func TestRenderIndexDeterministicTies(t *testing.T) {
	in := IndexInput{Now: now}
	for _, s := range []string{"zebra", "alpha", "middle"} {
		in.Entries = append(in.Entries, Entry{Slug: s, Stack: []string{s}, Valid: true})
	}
	got := string(RenderIndex(in))
	if strings.Index(got, "**alpha**") > strings.Index(got, "**middle**") ||
		strings.Index(got, "**middle**") > strings.Index(got, "**zebra**") {
		t.Errorf("tied stack counts not sorted by key:\n%s", got)
	}
}

// TestStaleTiesBreakOnSlug: `verified` is a date, so same-day ties are the common case,
// not a corner. The stale list is truncated to 15, so an unbroken tie does not merely
// reorder the section — it decides which notes appear in it. sort.Slice is not stable, and
// the same defect in pkg/drift's name table made drift.md oscillate on an unchanged tree.
func TestStaleTiesBreakOnSlug(t *testing.T) {
	in := IndexInput{Now: now}
	for _, s := range []string{"zebra", "alpha", "middle"} {
		in.Entries = append(in.Entries, Entry{Slug: s, Verified: day("2020-01-01"),
			FreshnessDays: 30, Valid: true})
	}
	first := string(RenderIndex(in))
	for i := 0; i < 30; i++ {
		in.Entries[0], in.Entries[2] = in.Entries[2], in.Entries[0]
		if got := string(RenderIndex(in)); got != first {
			t.Fatalf("run %d reordered same-day stale notes:\n%s", i, got)
		}
	}
	if strings.Index(first, "[[alpha]]") > strings.Index(first, "[[zebra]]") {
		t.Errorf("tied stale notes not sorted by slug:\n%s", first)
	}
}

func TestRenderIndexBudget(t *testing.T) {
	in := IndexInput{Now: now, MaxSize: 800}
	for i := 0; i < 200; i++ {
		in.Entries = append(in.Entries, Entry{
			Slug: strings.Repeat("x", 30), Stack: []string{"java"},
			Updated: day("2026-08-01"), Verified: day("2020-01-01"),
			FreshnessDays: 30, Valid: true,
		})
	}
	got := string(RenderIndex(in))
	if len(got) > 800 {
		t.Errorf("len = %d, over the 800-byte budget", len(got))
	}
	if !strings.Contains(got, "_[index truncated") {
		t.Errorf("truncated without saying so:\n%s", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("truncation did not cut at a line boundary")
	}
}

func TestRenderIndexEmptyVault(t *testing.T) {
	got := string(RenderIndex(IndexInput{Now: now}))
	for _, want := range []string{"# Vault index — 0 notes", "_none recorded_", "## Stale — 0"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// TestStaleIgnoresUnverifiedNotes: a note with no verified date is unknown, not stale.
// Reporting it as stale would bury the notes that genuinely expired.
func TestStaleIgnoresUnverifiedNotes(t *testing.T) {
	in := IndexInput{Now: now, Entries: []Entry{
		{Slug: "novdate", FreshnessDays: 30, Valid: true},
		{Slug: "nobudget", Verified: day("2000-01-01"), Valid: true},
	}}
	if got := string(RenderIndex(in)); !strings.Contains(got, "## Stale — 0") {
		t.Errorf("unverified or budget-less notes counted as stale:\n%s", got)
	}
}

func TestStaleOverflowKeepsTheCount(t *testing.T) {
	in := IndexInput{Now: now}
	for i := 0; i < 18; i++ {
		in.Entries = append(in.Entries, Entry{
			Slug: "s", Verified: day("2020-01-01"), FreshnessDays: 30, Valid: true,
		})
	}
	if got := string(RenderIndex(in)); !strings.Contains(got, "_… and 3 more_") {
		t.Errorf("overflow count missing:\n%s", got)
	}
}
