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

// VaultStats is one week's headline numbers, snapshotted so the next run can show a
// delta.
type VaultStats struct {
	Notes   int
	HitRate float64
	Orphans int
	Drift   int // notes carrying a BROKEN or SUSPECT finding
}

// WeeklyInput is what moc/weekly/YYYY-WW.md renders from.
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

// RenderWeekly produces moc/weekly/YYYY-WW.md — a ranked rollup per
// docs/ARCHITECTURE.md §10 (Flow C — Weekly check).
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

// actNowLines ranks by kind — BROKEN, then undocumented churn, then near-duplicates.
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
		"regardless of vault health: the churn window can leave " +
		"\"undocumented and moving\" empty on repos with low recent churn, and " +
		"near-duplicate pairs almost never clear the 0.85 spec threshold this " +
		"section requires._\n")
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
