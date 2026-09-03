package qualitygate

import (
	"strings"
	"testing"
	"time"
)

// TestFreshnessGateAtBoundary pins recall.IsStale's exact boundary via the injected
// clock, instead of racing time.Now() — one day inside freshness_days passes.
func TestFreshnessGateAtBoundary(t *testing.T) {
	s := testSchema(t)
	verified := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	src := strings.Replace(goodNote, "verified: 2026-08-07", "verified: 2026-01-01", 1)
	src = strings.Replace(src, "freshness_days: 365", "freshness_days: 30", 1)

	fresh := noteFrom(t, src, "notes/concept/x.md")
	justInside := verified.AddDate(0, 0, 30)
	o := freshnessGateAt(fresh, s, ModeCreate, justInside)
	if o.Verdict != Pass {
		t.Errorf("29 days after verified with freshness_days=30 = %+v, want Pass", o)
	}

	justOutside := verified.AddDate(0, 0, 31)
	o = freshnessGateAt(fresh, s, ModeCreate, justOutside)
	if o.Verdict != Fail || o.Remedy != DropConfidence {
		t.Errorf("31 days after verified with freshness_days=30 = %+v, want Fail/DropConfidence", o)
	}
}

// TestFreshnessGateAtUndatedCreateIsSkipped confirms CREATE's one divergence from
// UPDATE: an undated brand-new draft is Skipped here.
func TestFreshnessGateAtUndatedCreateIsSkipped(t *testing.T) {
	s := testSchema(t)
	src := strings.Replace(goodNote, "updated: 2026-08-07", "updated: \"\"", 1)
	src = strings.Replace(src, "verified: 2026-08-07", "verified: \"\"", 1)
	draft := noteFrom(t, src, "notes/concept/x.md")
	o := freshnessGateAt(draft, s, ModeCreate, time.Now())
	if o.Verdict != Skipped {
		t.Errorf("undated CREATE draft = %+v, want Skipped", o)
	}
}
