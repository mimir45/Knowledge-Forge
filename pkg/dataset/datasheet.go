package dataset

import (
	"fmt"
	"sort"
	"strings"
)

// renderDatasheet writes the datasheet document beside every export: counts.
func renderDatasheet(rep ExportReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Datasheet — %s (%s)\n\n", rep.Kind, rep.Format)
	writeProvenance(&b, rep)
	writeDistribution(&b, "Stack distribution", rep.Stacks, rep.Records)
	writeDistribution(&b, "Engine trail", rep.EngineTrail, rep.Records)
	writeLimitations(&b, rep)
	return b.String()
}

func writeProvenance(b *strings.Builder, rep ExportReport) {
	fmt.Fprintf(b, "| field | value |\n|---|---|\n")
	fmt.Fprintf(b, "| set | `%s` (%s) |\n", rep.Set, rep.Kind)
	fmt.Fprintf(b, "| format | `%s` |\n", rep.Format)
	fmt.Fprintf(b, "| records exported | %d |\n", rep.Records)
	fmt.Fprintf(b, "| records available | %d |\n", rep.Available)
	fmt.Fprintf(b, "| date range | %s |\n", dateRange(rep))
	fmt.Fprintf(b, "| anonymized | %s |\n", anonLabel(rep))
	if rep.DroppedBy != "" {
		fmt.Fprintf(b, "| filtered | %s |\n", rep.DroppedBy)
	}
	b.WriteString("\nThis file was produced locally by `forge export-dataset`. Nothing in " +
		"Knowledge Forge transmits it anywhere; the data is yours and moving it is a " +
		"decision you make by hand.\n")
}

func dateRange(rep ExportReport) string {
	if rep.From.IsZero() {
		return "empty"
	}
	return rep.From.Format("2006-01-02") + " → " + rep.To.Format("2006-01-02")
}

// anonLabel's d6 branch says "raw text", not "raw captured text" — d6 is derived.
func anonLabel(rep ExportReport) string {
	if rep.Anonymized {
		return fmt.Sprintf("yes, %d redactions", rep.Redactions)
	}
	if rep.Set == D6Tag {
		return "**no** — this export is raw text (a derived set cannot be anonymized)"
	}
	return "**no** — this export contains raw captured text"
}

// writeDistribution prints a share table, or says plainly that the tier does not record
// the field. An empty table would read as "none were used", which is a different claim.
func writeDistribution(b *strings.Builder, title string, dist map[string]int, total int) {
	fmt.Fprintf(b, "\n## %s\n\n", title)
	if len(dist) == 0 {
		b.WriteString("Not recorded for this tier.\n")
		return
	}
	for _, k := range rankedKeys(dist) {
		fmt.Fprintf(b, "- `%s` — %d (%s)\n", k, dist[k], share(dist[k], total))
	}
	if top := rankedKeys(dist)[0]; total > 0 && dist[top]*100/total >= 60 {
		fmt.Fprintf(b, "\n**Bias:** %s of records are `%s`. Do not expect this corpus to "+
			"generalize outside it.\n", share(dist[top], total), top)
	}
}

func share(n, total int) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%d%%", n*100/total)
}

// writeLimitations is the section that must not be trimmed.
func writeLimitations(b *strings.Builder, rep ExportReport) {
	b.WriteString("\n## Limitations\n\n")
	for _, l := range append(commonLimits(rep), tierLimits(rep)...) {
		fmt.Fprintf(b, "- %s\n", l)
	}
}

func commonLimits(rep ExportReport) []string {
	out := []string{
		"Single-author corpus. Every pair comes from one developer's vault, so what looks " +
			"like a preference is at least partly one person's habit.",
		accumulationLimit(rep.Set),
	}
	if rep.Anonymized {
		out = append(out, "Anonymization redacts token-, address- and path-shaped content. "+
			"**Topic slugs and profile values are kept**, because they are the semantic and "+
			"conditioning features the routing and style tiers carry and hashing them makes "+
			"those corpora untrainable. Anything spelled as a plain kebab-case name — a topic "+
			"named after a product, a framework named after an in-house SDK — survives. Read "+
			"the topic and profile fields before sharing.")
	}
	return out
}

// accumulationLimit is D6's one structural difference from every other common limit.
func accumulationLimit(set string) string {
	if set == D6Tag {
		return "This is a point-in-time derivation over `forge logback`'s map, not a " +
			"forward-accumulating capture. Re-running the export later can " +
			"emit *fewer* pairs than before if a citation stops resolving or a repo's cached " +
			"code index goes stale — unlike D1-D5, an old record here is not guaranteed to " +
			"still be present."
	}
	return "Capture is forward-only. Nothing before the day the relevant tier was switched " +
		"on appears here, and the date range above is the honest start of the record."
}

func tierLimits(rep ExportReport) []string {
	switch rep.Set {
	case D1Tag:
		return []string{
			fmt.Sprintf("**Outcome label is partial: %s (%d of %d records).** A joined pair "+
				"carries whether the note it led to was actually published or quarantined "+
				"when the join succeeded; the rest are (question features → routing "+
				"decision) pairs only — supervision on the router's own output, not evidence "+
				"the router is correct. A pair joins only when the caller threaded recall's "+
				"run_id back through `forge gate --run-id`; forgetting it degrades to the "+
				"unjoined case silently, it does not fail, so do not read this share as a "+
				"census of how often the routing decision was right.",
				share(rep.D1Joined, rep.Records), rep.D1Joined, rep.Records),
			"**A joined outcome is the last one reported, not the first.** A quarantine " +
				"followed by a `--previous-draft` repair that re-passes `--run-id` (SKILL.md's " +
				"Stage 4) appends a second D1Outcome for the same run_id; export keeps the " +
				"later record. A repair that forgets `--run-id` leaves the pair labelled " +
				"`quarantined` even though the note went on to publish — that pair is " +
				"indistinguishable from a real quarantine in this export.",
			"**Recall calls only, not every ranking.** The original spec (since removed) said " +
				"\"every run\", but " +
				"`forge intent` also ranks the vault on every prompt submission and is " +
				"deliberately excluded: it has a 50ms budget and a passive hint is not a " +
				"question anyone asked.",
		}
	case D2Tag:
		return []string{"Advisor runs only, so the corpus is skewed toward notes that were " +
			"worth spending a T3 call on — the routine ones are absent by construction."}
	case D3Tag:
		return []string{"A pair exists only where a human edited a generated note within the " +
			"capture window. A note accepted unchanged produces nothing, so the corpus is " +
			"biased toward the notes the generator got wrong."}
	case D4Tag:
		return []string{"Only retries that were explicitly joined with `--previous-draft` " +
			"appear. A draft fixed without passing that flag back is invisible here."}
	case D6Tag:
		return []string{
			"**Derived, not captured.** This tier has no write path: every export re-resolves " +
				"(repo, path, symbol) → note pairs from the vault's citations against whatever " +
				"`.forge/code-index-<repo>.json` caches this machine currently holds. An empty " +
				"export usually means no repo has been indexed here yet (`forge logback`), not " +
				"that nothing is documented.",
			"**--anonymize is always refused for this set.** Repo, path and symbol strings " +
				"are the feature D6 exists to carry, and also the most employer-identifying " +
				"strings in the system; `anonymize.go`'s note-path answer (hash the slug, keep " +
				"the type) has no equivalent that leaves this tier useful. Every export is raw " +
				"text — read the repo names before sharing.",
			"**An unresolved or ambiguous citation contributes no pair.** Symbol lookup is " +
				"exact-name-then-trailing-member (`codeindex.File.Lookup`, the same rule " +
				"`forge check`'s coverage numbers use); a citation with a typo or a renamed " +
				"symbol drops out silently rather than appearing with a guessed target.",
			"**Records ≠ notes.** The original spec (since removed) gave volume as " +
				"\"= note count\", but a note " +
				"citing five symbols yields five records here — read this tier's count as " +
				"citation count, not note count.",
		}
	case D5Tag:
		return []string{"**Gate is not enforced in Go.** `forge gate` running before every " +
			"write is an invariant of `skills/forge/SKILL.md`, not of the binary, so a note " +
			"written around the skill never reaches this tier. Treat the count as a subset " +
			"of accepted notes, not a census.",
			"Profile fields are captured only where `profiles/me.md` exists. On a vault " +
				"where `forge init` has not run, the conditioning half of the pair is empty."}
	}
	return nil
}

func rankedKeys(m map[string]int) []string {
	out := sortedKeys(m)
	sort.SliceStable(out, func(i, j int) bool { return m[out[i]] > m[out[j]] })
	return out
}

// sortedKeys keeps every rendered table deterministic.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
