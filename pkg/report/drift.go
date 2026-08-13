package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"knowledge-forge/pkg/drift"
)

// DriftInput is what drift.md renders from: one finding per code reference, as
// `forge drift --deep` produced them.
type DriftInput struct {
	Findings []drift.Finding
	Slugs    map[string]string // note path -> slug, for wikilinks
	Now      time.Time
}

// RenderDrift produces drift.md — the report that answers "which of my notes are now
// lying". It is the only report whose subject is truth rather than tidiness: an orphan is
// merely hard to find, but a note describing a method that no longer exists is worse than
// no note, because a reader will believe it.
//
// The two verdicts are kept apart on purpose and BROKEN is listed first. BROKEN means the
// file or symbol is gone and the note cannot be right; SUSPECT means the code is still
// there and no longer says the same thing, which is a prompt to re-read rather than
// evidence of an error. Only BROKEN costs a note its confidence — see drift.Finding.Demoting.
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

// driftSummary leads with the number the user asked for: how many notes, not how many
// references. A note with nine broken citations is one note to fix.
//
// The headline is the union of the two verdicts, not their sum. A note can hold both a
// broken and a suspect reference — in this vault one already does — and adding the two
// lists would report it as two notes to fix. The per-verdict counts below it still overlap,
// which is correct: that note does need fixing on both counts.
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

// AffectedByDrift is notesWith's count, exported for weekly.md's collector: the same
// "distinct notes carrying any of these verdicts" question drift.md already answers
// internally, without redefining it a second way outside this package.
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
	for _, rel := range sortedKeys(byNote) {
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
			// One note can cite one ref twice and get two reasons; the ref alone does not
			// separate them, and an unbroken tie is settled by map order rather than by
			// anything about the vault.
			return out[rel][i].Reason < out[rel][j].Reason
		})
	}
	return out
}

// writeSkipped reports the blind spot rather than hiding it. A skipped citation is one
// nothing on this machine could adjudicate — a repo that is not cloned here, most often —
// and a reader who does not know how many there were cannot judge the rest of the numbers.
func writeSkipped(b *strings.Builder, counts map[drift.Verdict]int) {
	if counts[drift.Skipped] == 0 {
		return
	}
	fmt.Fprintf(b, "\n## Not checked — %d\n\n", counts[drift.Skipped])
	b.WriteString("Citations nothing on this machine could adjudicate, usually a repo that " +
		"is not cloned here. They are neither verified nor broken.\n")
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
