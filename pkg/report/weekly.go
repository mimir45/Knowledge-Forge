package report

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/linkcheck"
	"github.com/mimir45/Knowledge-Forge/pkg/similarity"
)

// VaultStats is one week's headline numbers, snapshotted so the next run can show a delta.
// HitRate is asks resolved by ANSWER_FROM_VAULT as a share of all asks recorded — cumulative
// to date, not filtered to the week, because DESIGN §14 events carry a timestamp but nothing
// downstream of loadAskLog windows by it yet. That is a known simplification, not a bug.
type VaultStats struct {
	Notes   int
	HitRate float64
	Orphans int
	Drift   int // notes carrying a BROKEN or SUSPECT finding
}

// WeeklyInput is what moc/weekly/YYYY-WW.md renders from. Week and Year must come from
// time.Time.ISOWeek(), not Now.Year() — a week can straddle a calendar year boundary and
// ISOWeek's own returned year is the one that keys it correctly.
//
// Broken, Uncovered, DuplicatePairs and DeadCitations are Act now's four raw signal kinds,
// passed through as the same typed values their own reports (drift.md, codebase.md,
// duplicates.md, deadlinks.md) already compute — this renderer re-derives nothing, it only
// re-ranks and re-labels for a different audience. DuplicatePairs here is filtered at the
// spec's 0.85 (specThreshold), not duplicates.md's lower operating threshold: an "act now"
// merge is a different claim than a "candidate worth a look".
type WeeklyInput struct {
	Week, Year int

	Broken         []drift.Finding
	Uncovered      []Uncovered
	UncoveredDays  int
	DuplicatePairs []similarity.Pair
	DeadCitations  []Citation

	StaleCount      int
	MergeCandidates int
	OrphanCount     int

	Stats VaultStats
	Prev  *VaultStats

	Asks []Ask

	Slugs map[string]string
	Now   time.Time
}

// RenderWeekly produces moc/weekly/YYYY-WW.md — ADDENDUM §C's ranked rollup. Its four
// sections and their literal emoji headers are the spec; the sentences under them are this
// renderer's own, not copied from the example, because the example's numbers are fiction.
func RenderWeekly(in WeeklyInput) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# Week %d, %d\n", in.Week, in.Year)
	writeActNow(&b, in)
	writeReview(&b, in)
	writeVaultSection(&b, in)
	writeWeeklyGaps(&b, in)
	return []byte(b.String())
}

func writeActNow(b *strings.Builder, in WeeklyInput) {
	lines := actNowLines(in)
	fmt.Fprintf(b, "\n## \U0001F534 Act now (%d)\n\n", len(lines))
	if len(lines) == 0 {
		writeActNowCaveat(b)
		return
	}
	for _, l := range lines {
		fmt.Fprintf(b, "- %s\n", l)
	}
}

// actNowLines ranks by kind — BROKEN, then undocumented churn, then near-duplicates, then
// dead links — and each kind carries its own deterministic order in from the report it was
// computed by, so no further tiebreak is needed here.
func actNowLines(in WeeklyInput) []string {
	var lines []string
	lines = append(lines, brokenLines(in)...)
	lines = append(lines, uncoveredLines(in)...)
	lines = append(lines, duplicateLines(in)...)
	lines = append(lines, deadLinkLines(in)...)
	return lines
}

func brokenLines(in WeeklyInput) []string {
	byNote := groupByNote(in.Broken, drift.Broken)
	var out []string
	for _, rel := range slices.Sorted(maps.Keys(byNote)) {
		var reasons []string
		for _, f := range byNote[rel] {
			reasons = append(reasons, f.Reason)
		}
		out = append(out, fmt.Sprintf("%s — BROKEN: %s",
			note(in.Slugs[rel], rel), strings.Join(reasons, "; ")))
	}
	return out
}

func uncoveredLines(in WeeklyInput) []string {
	u := append([]Uncovered(nil), in.Uncovered...)
	sortUncovered(u)
	var out []string
	for _, s := range head(u, 5) {
		out = append(out, fmt.Sprintf("`%s` — %d LOC, %d changes/%dd, 0 notes",
			s.Symbol, s.LOC, s.Commits, in.UncoveredDays))
	}
	return out
}

func duplicateLines(in WeeklyInput) []string {
	var out []string
	for _, p := range in.DuplicatePairs {
		if p.Score < specThreshold {
			continue
		}
		out = append(out, fmt.Sprintf("%.2f near-duplicate — %s · %s",
			p.Score, note(in.Slugs[p.A], p.A), note(in.Slugs[p.B], p.B)))
	}
	return out
}

func deadLinkLines(in WeeklyInput) []string {
	counts := deadCountByNote(in.DeadCitations)
	var out []string
	for _, rel := range slices.Sorted(maps.Keys(counts)) {
		n := counts[rel]
		out = append(out, fmt.Sprintf("%d dead %s in %s",
			n, plural(n, "source URL", "source URLs"), note(in.Slugs[rel], rel)))
	}
	return out
}

func deadCountByNote(cs []Citation) map[string]int {
	out := map[string]int{}
	for _, c := range cs {
		if c.Verdict != linkcheck.Dead {
			continue
		}
		for _, rel := range c.Notes {
			out[rel]++
		}
	}
	return out
}

// writeActNowCaveat names the two known, measured reasons this section can be empty on a
// healthy vault rather than printing a bare "none" that reads as a bug report.
func writeActNowCaveat(b *strings.Builder) {
	b.WriteString("_Nothing to act on right now. Two known reasons this can stay thin " +
		"regardless of vault health: BACKLOG B-017 (the churn window can leave " +
		"\"undocumented and moving\" empty on repos with low recent churn) and BACKLOG " +
		"B-019 (near-duplicate pairs almost never clear the 0.85 spec threshold this " +
		"section requires)._\n")
}

func writeReview(b *strings.Builder, in WeeklyInput) {
	total := in.StaleCount + in.MergeCandidates + in.OrphanCount
	fmt.Fprintf(b, "\n## \U0001F7E1 Review (%d)\n\n", total)
	if total == 0 {
		empty(b, "nothing needs a second look")
		return
	}
	fmt.Fprintf(b, "%d %s past their freshness window · %d merge %s · %d %s\n",
		in.StaleCount, plural(in.StaleCount, "note", "notes"),
		in.MergeCandidates, plural(in.MergeCandidates, "candidate", "candidates"),
		in.OrphanCount, plural(in.OrphanCount, "orphan", "orphans"))
}

func writeVaultSection(b *strings.Builder, in WeeklyInput) {
	b.WriteString("\n## \U0001F4CA Vault\n\n")
	fmt.Fprintf(b, "%d %s%s · hit-rate %.0f%%%s · orphans %d%s · drift %d%s\n",
		in.Stats.Notes, plural(in.Stats.Notes, "note", "notes"), delta(in.Prev, in.Stats, notesOf),
		in.Stats.HitRate, deltaPct(in.Prev, in.Stats),
		in.Stats.Orphans, delta(in.Prev, in.Stats, orphansOf),
		in.Stats.Drift, delta(in.Prev, in.Stats, driftOf))
	if in.Prev == nil {
		b.WriteString("\n_First recorded week — no prior snapshot to compare against._\n")
	}
}

func notesOf(s VaultStats) int   { return s.Notes }
func orphansOf(s VaultStats) int { return s.Orphans }
func driftOf(s VaultStats) int   { return s.Drift }

// delta prints "" on the first-ever run rather than "(+312)" against an empty snapshot,
// which would misreport the vault's entire history as one week's work.
func delta(prev *VaultStats, now VaultStats, of func(VaultStats) int) string {
	if prev == nil {
		return ""
	}
	d := of(now) - of(*prev)
	if d == 0 {
		return " (±0)"
	}
	if d > 0 {
		return fmt.Sprintf(" (+%d)", d)
	}
	return fmt.Sprintf(" (%d)", d)
}

func deltaPct(prev *VaultStats, now VaultStats) string {
	if prev == nil {
		return ""
	}
	d := now.HitRate - prev.HitRate
	if d == 0 {
		return " (±0pt)"
	}
	if d > 0 {
		return fmt.Sprintf(" (+%.0fpt)", d)
	}
	return fmt.Sprintf(" (%.0fpt)", d)
}

func writeWeeklyGaps(b *strings.Builder, in WeeklyInput) {
	gaps := unwritten(in.Asks)
	fmt.Fprintf(b, "\n## \U0001F3AF Gaps (asked, never written)\n\n")
	if len(gaps) == 0 {
		b.WriteString("_none — see gaps.md for the full picture, including single asks_\n")
		return
	}
	for i, a := range head(gaps, 10) {
		fmt.Fprintf(b, "%d. %q (%d×)   ", i+1, a.Topic, a.Count)
		if (i+1)%2 == 0 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
}
