package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// StalenessInput is what staleness.md renders from.
type StalenessInput struct {
	Entries []Entry
	Asks    map[string]int // slug -> times asked; empty until the capture log has data
	Now     time.Time
}

// Overdue is one stale note with its rank inputs kept visible.
type Overdue struct {
	Entry
	DaysOverdue int
	Asks        int
	Score       int // the ranking key actually used
}

// RenderStaleness produces staleness.md, ranked so the top line is the one note to fix today.
//
// ADDENDUM §B.4's ranking is ask frequency x days overdue, and that product is what this
// implements. It is not what this vault can currently rank by: every note carries
// `origin: import`, so no note has ever been asked for, every ask count is 0, and the
// product is 0 for all of them. Ranking then falls back to days overdue alone and the
// header says so.
//
// The fallback is conditional on the ask counts actually being empty rather than hardcoded.
// Phase 4 starts recording asks; if this were wired to days-overdue permanently, that data
// would land and the ranking would silently stay degenerate.
func RenderStaleness(in StalenessInput) []byte {
	overdue := rankOverdue(in)
	var b strings.Builder
	header(&b, "Staleness", stalenessSummary(in, overdue), in.Now)
	if len(overdue) == 0 {
		empty(&b, "no note is past its freshness window")
		return []byte(b.String())
	}
	b.WriteString("\n")
	for _, o := range head(overdue, 40) {
		writeOverdue(&b, o)
	}
	if len(overdue) > 40 {
		fmt.Fprintf(&b, "- _… and %d more_\n", len(overdue)-40)
	}
	return []byte(b.String())
}

// haveAsks reports whether the ask log carries any signal at all.
func haveAsks(asks map[string]int) bool {
	for _, n := range asks {
		if n > 0 {
			return true
		}
	}
	return false
}

func rankOverdue(in StalenessInput) []Overdue {
	weighted := haveAsks(in.Asks)
	var out []Overdue
	for _, e := range in.Entries {
		od, ok := overdueBy(e, in.Now)
		if !ok {
			continue
		}
		asks := in.Asks[e.Slug]
		out = append(out, Overdue{Entry: e, DaysOverdue: od, Asks: asks,
			Score: rankScore(od, asks, weighted)})
	}
	sortOverdue(out)
	return out
}

func rankScore(daysOverdue, asks int, weighted bool) int {
	if weighted {
		return asks * daysOverdue
	}
	return daysOverdue
}

// overdueBy returns how many days past its freshness window a note is. A note with no
// freshness budget or no verification date is not stale, it is unmeasured — a different
// problem, and coverage.md's, not this report's.
func overdueBy(e Entry, now time.Time) (int, bool) {
	if e.FreshnessDays <= 0 || e.Verified.IsZero() {
		return 0, false
	}
	over := days(e.Verified, now) - e.FreshnessDays
	if over <= 0 {
		return 0, false
	}
	return over, true
}

func sortOverdue(out []Overdue) {
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Slug < out[j].Slug
	})
}

func stalenessSummary(in StalenessInput, overdue []Overdue) string {
	if !haveAsks(in.Asks) {
		return fmt.Sprintf("**%d of %d notes are past their freshness window.**\n\n"+
			"Ranked by days overdue. §B.4 ranks by _ask frequency x days overdue_, which is "+
			"0 for every note in this vault: all of them are `origin: import` and none has "+
			"been asked for yet. The product returns as soon as the capture log has data.\n",
			len(overdue), len(in.Entries))
	}
	return fmt.Sprintf("**%d of %d notes are past their freshness window.**\n\n"+
		"Ranked by ask frequency x days overdue — the top line is the note to fix today.\n",
		len(overdue), len(in.Entries))
}

func writeOverdue(b *strings.Builder, o Overdue) {
	fmt.Fprintf(b, "- **%dd overdue** — %s", o.DaysOverdue, note(o.Slug, o.Rel))
	if o.Asks > 0 {
		fmt.Fprintf(b, " · asked %dx", o.Asks)
	}
	fmt.Fprintf(b, " · verified %s (%dd budget)\n",
		o.Verified.Format("2006-01-02"), o.FreshnessDays)
}
