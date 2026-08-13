package report

import (
	"testing"
	"time"
)

// A week that straddles a year boundary must key off ISOWeek's own returned year, not
// t.Year() — 2025-12-29 is a Monday in ISO week 1 of 2026.
func TestWeekKeyUsesISOYearNotCalendarYear(t *testing.T) {
	got := WeekKey(time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC))
	if got != "2026-W01" {
		t.Errorf("WeekKey = %q, want 2026-W01", got)
	}
}

// A store that has never been saved must behave like an empty one, not panic or error —
// every vault before its first weekly run has no weekly-stats.json at all.
func TestOpenWeeklyStoreMissingFileIsEmpty(t *testing.T) {
	s := OpenWeeklyStore(t.TempDir())
	if len(s.Weeks) != 0 {
		t.Fatalf("got %v, want empty", s.Weeks)
	}
	if s.Prev("2026-W32") != nil {
		t.Error("Prev on an empty store must be nil")
	}
}

// Prev must find the most recent week strictly before the key, skipping over gaps and
// never returning the current week's own entry.
func TestWeeklyStorePrevSkipsGapsAndSelf(t *testing.T) {
	s := OpenWeeklyStore(t.TempDir())
	s.Record("2026-W10", VaultStats{Notes: 10})
	s.Record("2026-W20", VaultStats{Notes: 20})
	s.Record("2026-W32", VaultStats{Notes: 32})

	got := s.Prev("2026-W32")
	if got == nil || got.Notes != 20 {
		t.Fatalf("Prev(2026-W32) = %v, want the W20 snapshot", got)
	}
}

// A second run in the same week must update the existing snapshot, not stack a duplicate
// that would make Prev see two entries for one week.
func TestWeeklyStoreRecordOverwritesSameWeek(t *testing.T) {
	s := OpenWeeklyStore(t.TempDir())
	s.Record("2026-W32", VaultStats{Notes: 10})
	s.Record("2026-W32", VaultStats{Notes: 15})
	if len(s.Weeks) != 1 || s.Weeks["2026-W32"].Notes != 15 {
		t.Fatalf("got %v, want one entry with Notes=15", s.Weeks)
	}
}

func TestWeeklyStorePruneKeepsOnlyMostRecentN(t *testing.T) {
	s := OpenWeeklyStore(t.TempDir())
	for _, k := range []string{"2026-W01", "2026-W02", "2026-W03", "2026-W04"} {
		s.Record(k, VaultStats{})
	}
	s.Prune(2)
	if len(s.Weeks) != 2 {
		t.Fatalf("got %d weeks, want 2", len(s.Weeks))
	}
	if _, ok := s.Weeks["2026-W03"]; !ok {
		t.Error("W03 should have survived pruning")
	}
	if _, ok := s.Weeks["2026-W04"]; !ok {
		t.Error("W04 should have survived pruning")
	}
}

// Save/reopen round trip: what Record wrote must be exactly what a fresh OpenWeeklyStore
// reads back, since this is the only thing weekly.md's delta depends on across runs.
func TestWeeklyStoreSaveAndReopenRoundTrips(t *testing.T) {
	dir := t.TempDir()
	s := OpenWeeklyStore(dir)
	s.Record("2026-W32", VaultStats{Notes: 100, HitRate: 42.5, Orphans: 3, Drift: 1})
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reopened := OpenWeeklyStore(dir)
	got, ok := reopened.Weeks["2026-W32"]
	if !ok || got.Notes != 100 || got.HitRate != 42.5 {
		t.Fatalf("reopened = %+v, ok=%v, want the saved snapshot", got, ok)
	}
}

// Save on a clean store (nothing recorded) must not write a file — mirrors
// pkg/drift/demotions.go's dirty-flag contract.
func TestWeeklyStoreSaveNoopWhenClean(t *testing.T) {
	dir := t.TempDir()
	s := OpenWeeklyStore(dir)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	reopened := OpenWeeklyStore(dir)
	if len(reopened.Weeks) != 0 {
		t.Error("Save on a clean store should not have written a file")
	}
}
