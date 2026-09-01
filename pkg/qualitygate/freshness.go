package qualitygate

import (
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// freshnessGate reuses recall.IsStale rather than reimplementing the Verified-then-
// Updated fallback: that predicate is already the one governing ANSWER_FROM_VAULT vs
// UPDATE(refresh) at recall time, so a draft that would recall as stale the moment it
// lands must not publish as fresh. A draft that is already stale on arrival gets
// DropConfidence, not RetryOnce — no amount of retrying fixes a stale Verified date;
// only backdating it would, and that is not this gate's job.
//
// mode is the one place CREATE and UPDATE genuinely diverge here: recall.IsStale treats
// an undatable doc as stale, which is right for an UPDATE (the published note really has
// no verified/updated to trust) but wrong for a brand-new CREATE draft that simply hasn't
// been dated yet — schema.go's RetryOnce already covers that missing-required-field case,
// so freshness.go skips rather than piling on a second, misleading "stale" verdict.
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

// freshnessDaysOr mirrors cmd/forge/index.go's atoiOr(freshness_days, schema default) —
// duplicated rather than imported because cmd/forge depends on pkg/qualitygate, not the
// other way around.
func freshnessDaysOr(draft *vault.Note, s *vault.Schema) int {
	var n int
	if _, err := fmt.Sscanf(draft.FM.Str("freshness_days"), "%d", &n); err == nil && n > 0 {
		return n
	}
	return s.FreshnessDefault(draft.FM.Str("type"))
}
