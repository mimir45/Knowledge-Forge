package store

import (
	"testing"
	"time"
)

func fixedClock(day string) func() time.Time {
	t, _ := time.Parse("2006-01-02", day)
	return func() time.Time { return t }
}

func TestSpendAccumulatesWithinADay(t *testing.T) {
	s := open(t)
	clock := fixedClock("2026-08-10")
	if err := s.Spend("api", 0.10, clock); err != nil {
		t.Fatal(err)
	}
	if err := s.Spend("api", 0.05, clock); err != nil {
		t.Fatal(err)
	}
	remaining, err := s.Remaining("api", 1.00, clock)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := remaining, 0.85; !closeTo(got, want) {
		t.Errorf("Remaining = %v, want %v", got, want)
	}
}

// TestSpendIsPerDay: yesterday's spend must not shrink today's cap, or a day boundary
// that is "just a key that doesn't exist yet" would instead need a reset job.
func TestSpendIsPerDay(t *testing.T) {
	s := open(t)
	if err := s.Spend("advisor", 5.00, fixedClock("2026-08-09")); err != nil {
		t.Fatal(err)
	}
	remaining, err := s.Remaining("advisor", 5.00, fixedClock("2026-08-10"))
	if err != nil {
		t.Fatal(err)
	}
	if !closeTo(remaining, 5.00) {
		t.Errorf("Remaining = %v, want the full cap (5.00) — spend leaked across days", remaining)
	}
}

func TestRemainingIsFullCapWithNoSpend(t *testing.T) {
	s := open(t)
	remaining, err := s.Remaining("api", 2.50, fixedClock("2026-08-10"))
	if err != nil {
		t.Fatal(err)
	}
	if !closeTo(remaining, 2.50) {
		t.Errorf("Remaining = %v, want the untouched cap (2.50)", remaining)
	}
}

// TestResetPreservesBudget is AUDIT §8.4 D-8's whole point: `forge reindex` wipes the
// derived tables and must never touch spend, which markdown does not record.
func TestResetPreservesBudget(t *testing.T) {
	s := open(t)
	clock := fixedClock("2026-08-10")
	put(t, s, sample())
	if err := s.Spend("api", 0.42, clock); err != nil {
		t.Fatal(err)
	}
	if err := s.Reset(); err != nil {
		t.Fatal(err)
	}
	if got := count(t, s, "notes"); got != 0 {
		t.Errorf("notes = %d rows survived Reset, want 0", got)
	}
	remaining, err := s.Remaining("api", 1.00, clock)
	if err != nil {
		t.Fatal(err)
	}
	if !closeTo(remaining, 0.58) {
		t.Errorf("Remaining = %v after Reset, want 0.58 — budget spend was wiped", remaining)
	}
}

func closeTo(a, b float64) bool {
	d := a - b
	return d > -0.0001 && d < 0.0001
}
