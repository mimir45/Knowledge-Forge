package main

import (
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// estimatedMinutesSavedPerHit is deliberately rough.
const estimatedMinutesSavedPerHit = 15

// topStatsTopics caps the "most-asked" table — mirrors gaps.md's own head(gaps, 30) cap;
// a long tail of one-off asks isn't the point of this table.
const topStatsTopics = 15

const statsUsage = `usage: forge stats [--vault path]

Reports on the vault's telemetry log (.forge/log.jsonl) and weekly snapshot history
(.forge/weekly-stats.json): hit rate, most-asked topics, gaps (asked 2+ times, never
written), an approximate research-time-saved estimate, and a vault-stats trend. Zero
model calls; a missing or empty log/store just means an emptier report, not an error.
`

func cmdStats(args []string) int {
	fs := flag.NewFlagSet("forge stats", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.Usage = func() { fmt.Fprint(os.Stderr, statsUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	root, code := vaultOrExit("stats", *vaultDir)
	if code != 0 {
		return code
	}
	return runStats(root, os.Stdout)
}

// runStats does the actual work against an io.Writer so it's directly testable without
// a stdout pipe. loadNotes/loadAskLog/OpenWeeklyStore are the exact functions forge
// check already uses for the same data — no parallel reimplementation.
func runStats(root string, w io.Writer) int {
	notes, err := loadNotes(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge stats: %v\n", err)
		return 1
	}
	_, asks := loadAskLog(filepath.Join(root, ".forge", "log.jsonl"), slugMap(notes))
	store := report.OpenWeeklyStore(filepath.Join(root, ".forge"))

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	writeHitRate(tw, asks)
	writeTopTopics(tw, asks)
	writeGaps(tw, asks)
	writeTimeSaved(tw, asks)
	writeTrend(tw, store)
	tw.Flush()
	return 0
}

func writeHitRate(tw *tabwriter.Writer, asks []report.Ask) {
	total := 0
	for _, a := range asks {
		total += a.Count
	}
	fmt.Fprintf(tw, "Hit rate:\t%.1f%%\t(%d asks across %d topics)\n",
		report.HitRate(asks), total, len(asks))
}

func writeTopTopics(tw *tabwriter.Writer, asks []report.Ask) {
	sorted := sortedByCount(asks)
	fmt.Fprintf(tw, "\nMost-asked topics:\n")
	if len(sorted) == 0 {
		fmt.Fprintf(tw, "  none recorded yet\n")
		return
	}
	for _, a := range headAsks(sorted, topStatsTopics) {
		fmt.Fprintf(tw, "  %s\t%dx\t%s\n", a.Topic, a.Count, writtenLabel(a.Written))
	}
}

// writeGaps reuses Step 3's exact-slug "written" resolution.
func writeGaps(tw *tabwriter.Writer, asks []report.Ask) {
	fmt.Fprintf(tw, "\nGaps (asked 2+ times, never written):\n")
	gaps := unwrittenAsks(asks)
	if len(gaps) == 0 {
		fmt.Fprintf(tw, "  none\n")
		return
	}
	for _, a := range gaps {
		fmt.Fprintf(tw, "  %s\t%dx\n", a.Topic, a.Count)
	}
}

func writeTimeSaved(tw *tabwriter.Writer, asks []report.Ask) {
	written := 0
	for _, a := range asks {
		if a.Written {
			written += a.Count
		}
	}
	mins := written * estimatedMinutesSavedPerHit
	fmt.Fprintf(tw, "\nApprox. research time saved: ~%d min (%d hits x %d min/hit, a rough estimate)\n",
		mins, written, estimatedMinutesSavedPerHit)
}

// writeTrend prints Step 5's weekly snapshots.
func writeTrend(tw *tabwriter.Writer, store *report.WeeklyStore) {
	fmt.Fprintf(tw, "\nVault trend (Drift stands in for staleness; no dedicated metric exists):\n")
	if len(store.Weeks) == 0 {
		fmt.Fprintf(tw, "  no weekly snapshots yet — forge check must run at least once\n")
		return
	}
	fmt.Fprintf(tw, "  week\tnotes\thit rate\torphans\tdrift\n")
	for _, k := range slices.Sorted(maps.Keys(store.Weeks)) {
		s := store.Weeks[k]
		fmt.Fprintf(tw, "  %s\t%d\t%.1f%%\t%d\t%d\n", k, s.Notes, s.HitRate, s.Orphans, s.Drift)
	}
}

// sortedByCount and unwrittenAsks share a deterministic tiebreak.
func sortedByCount(asks []report.Ask) []report.Ask {
	out := append([]report.Ask(nil), asks...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Topic < out[j].Topic
	})
	return out
}

func unwrittenAsks(asks []report.Ask) []report.Ask {
	var filtered []report.Ask
	for _, a := range asks {
		if !a.Written && a.Count >= 2 {
			filtered = append(filtered, a)
		}
	}
	return sortedByCount(filtered)
}

func writtenLabel(w bool) string {
	if w {
		return "(written)"
	}
	return "(gap)"
}

func headAsks(asks []report.Ask, n int) []report.Ask {
	if len(asks) <= n {
		return asks
	}
	return asks[:n]
}
