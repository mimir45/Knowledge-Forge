package report

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
)

// DriftInput is what drift.md renders from: one finding per code reference, as
// `forge drift --deep` produced them.
type DriftInput struct {
	Findings []drift.Finding
	Slugs    map[string]string // note path -> slug, for wikilinks
	Now      time.Time
}

// RenderDrift produces drift.md — the report that answers "which of my notes are now
// lying".
func RenderDrift(in DriftInput) []byte {
	var b strings.Builder
	counts := countVerdicts(in.Findings)
	header(&b, "Drift", driftSummary(counts, in.Findings), in.Now)
	writeFindings(&b, in, drift.Broken, "Broken — the code these notes describe is gone")
	writeFindings(&b, in, drift.Suspect, "Suspect — the code changed under these notes")
	writeSkipped(&b, counts)
	return []byte(b.String())
}

func countVerdicts(fs []drift.Finding) map[drift.Verdict]int {
	counts := map[drift.Verdict]int{}
	for _, f := range fs {
		counts[f.Verdict]++
	}
	return counts
}

// driftSummary leads with the number the user asked for: how many notes.
func driftSummary(counts map[drift.Verdict]int, fs []drift.Finding) string {
	broken, suspect := notesWith(fs, drift.Broken), notesWith(fs, drift.Suspect)
	affected := len(notesWith(fs, drift.Broken, drift.Suspect))
	return fmt.Sprintf(
		"**%d %s reference code that has changed** — %d broken, %d suspect.\n\n"+
			"Checked %d citations: %d ok, %d repaired, %d suspect, %d broken, %d skipped.\n",
		affected, plural(affected, "note", "notes"),
		len(broken), len(suspect), len(fs), counts[drift.OK], counts[drift.Repaired],
		counts[drift.Suspect], counts[drift.Broken], counts[drift.Skipped])
}

// notesWith returns the distinct notes carrying at least one finding of any of these
// verdicts. Distinct is the whole job: findings are per reference, the report is per note.
func notesWith(fs []drift.Finding, vs ...drift.Verdict) []string {
	want := map[drift.Verdict]bool{}
	for _, v := range vs {
		want[v] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range fs {
		if want[f.Verdict] && !seen[f.Note] {
			seen[f.Note] = true
			out = append(out, f.Note)
		}
	}
	sort.Strings(out)
	return out
}

// AffectedByDrift is notesWith's count, exported for weekly.md's collector.
func AffectedByDrift(fs []drift.Finding, vs ...drift.Verdict) int {
	return len(notesWith(fs, vs...))
}

func writeFindings(b *strings.Builder, in DriftInput, v drift.Verdict, title string) {
	byNote := groupByNote(in.Findings, v)
	fmt.Fprintf(b, "\n## %s — %d\n", title, len(byNote))
	if len(byNote) == 0 {
		empty(b, "none")
		return
	}
	b.WriteString("\n")
	for _, rel := range slices.Sorted(maps.Keys(byNote)) {
		fmt.Fprintf(b, "- %s\n", note(in.Slugs[rel], rel))
		for _, f := range byNote[rel] {
			fmt.Fprintf(b, "  - `%s` — %s\n", f.Ref, f.Reason)
		}
	}
}

func groupByNote(fs []drift.Finding, v drift.Verdict) map[string][]drift.Finding {
	out := map[string][]drift.Finding{}
	for _, f := range fs {
		if f.Verdict == v {
			out[f.Note] = append(out[f.Note], f)
		}
	}
	for rel := range out {
		sort.Slice(out[rel], func(i, j int) bool {
			if out[rel][i].Ref != out[rel][j].Ref {
				return out[rel][i].Ref < out[rel][j].Ref
			}
			return out[rel][i].Reason < out[rel][j].Reason
		})
	}
	return out
}

// writeSkipped reports the blind spot rather than hiding it.
func writeSkipped(b *strings.Builder, counts map[drift.Verdict]int) {
	if counts[drift.Skipped] == 0 {
		return
	}
	fmt.Fprintf(b, "\n## Not checked — %d\n\n", counts[drift.Skipped])
	b.WriteString("Citations nothing on this machine could adjudicate, usually a repo that " +
		"is not cloned here. They are neither verified nor broken.\n")
}
