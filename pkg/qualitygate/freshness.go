package qualitygate

import (
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// freshnessGate reuses recall.IsStale rather than reimplementing the Verified-then-
// Updated fallback.
func freshnessGate(draft *vault.Note, s *vault.Schema, mode Mode) Outcome {
	return freshnessGateAt(draft, s, mode, time.Now())
}

// freshnessGateAt is freshnessGate with the clock injected, so a test can pin the exact
// freshness_days boundary instead of racing time.Now().
func freshnessGateAt(draft *vault.Note, s *vault.Schema, mode Mode, now time.Time) Outcome {
	if draft.FM == nil {
		return Outcome{Gate: "freshness", Verdict: Skipped, Detail: "no frontmatter"}
	}
	updated, verified := draft.FM.Str("updated"), draft.FM.Str("verified")
	if mode == ModeCreate && updated == "" && verified == "" {
		return Outcome{Gate: "freshness", Verdict: Skipped, Detail: "undated CREATE draft: schema gate covers this"}
	}
	d := recall.Doc{
		Rel:           draft.Rel,
		Updated:       updated,
		Verified:      verified,
		FreshnessDays: freshnessDaysOr(draft, s),
	}
	if d.FreshnessDays <= 0 {
		return Outcome{Gate: "freshness", Verdict: Skipped, Detail: "freshness_days<=0: never stale"}
	}
	if recall.IsStale(d, now) {
		return Outcome{Gate: "freshness", Verdict: Fail, Remedy: DropConfidence,
			Detail: fmt.Sprintf("stale: verified=%q updated=%q freshness_days=%d",
				d.Verified, d.Updated, d.FreshnessDays)}
	}
	return Outcome{Gate: "freshness", Verdict: Pass}
}

// freshnessDaysOr mirrors cmd/forge/index.go's atoiOr(freshness_days, schema default).
func freshnessDaysOr(draft *vault.Note, s *vault.Schema) int {
	var n int
	if _, err := fmt.Sscanf(draft.FM.Str("freshness_days"), "%d", &n); err == nil && n > 0 {
		return n
	}
	return s.FreshnessDefault(draft.FM.Str("type"))
}
