package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CoverageInput is what coverage.md renders from.
//
// Vocabulary is the authoritative `stack` enum out of references/schema.yaml. It is what
// makes this report say anything: counting the stacks that appear in notes tells you what
// you have written about, and only subtracting from the full vocabulary tells you what you
// have not. Without it the report cannot name an absence.
type CoverageInput struct {
	Entries    []Entry
	Vocabulary []string
	Types      []string // the seven note types, likewise from the schema
	Now        time.Time
}

// RenderCoverage produces coverage.md — where the wiki is thin.
//
// "Missing" here means one specific thing: a stack the schema knows about with zero notes
// against it. That is a different absence from gaps.md's (asked about and never written)
// and from codebase.md's (code that churns with no note citing it), and the three are kept
// apart deliberately — they have different denominators and merging them would produce a
// number that answers none of the questions.
func RenderCoverage(in CoverageInput) []byte {
	counts := stackCounts(in.Entries)
	var b strings.Builder
	header(&b, "Coverage", coverageSummary(in, counts), in.Now)
	writeCovered(&b, counts)
	writeUncovered(&b, in, counts)
	writeTypeMix(&b, in)
	return []byte(b.String())
}

func stackCounts(entries []Entry) map[string]int {
	counts := map[string]int{}
	for _, e := range entries {
		for _, s := range e.Stack {
			counts[s]++
		}
	}
	return counts
}

func coverageSummary(in CoverageInput, counts map[string]int) string {
	return fmt.Sprintf("**%d of %d known stacks have at least one note** — %d do not.\n",
		len(counts), len(in.Vocabulary), len(uncovered(in.Vocabulary, counts)))
}

func writeCovered(b *strings.Builder, counts map[string]int) {
	fmt.Fprintf(b, "\n## Covered — %d\n\n", len(counts))
	if len(counts) == 0 {
		b.WriteString("_no note declares a stack_\n")
		return
	}
	for _, kv := range topN(counts, 40) {
		fmt.Fprintf(b, "- **%s** — %d %s\n", kv.k, kv.v, plural(kv.v, "note", "notes"))
	}
}

// writeUncovered is the actionable half. A vocabulary value with no notes is either a real
// hole or a stack you do not actually work in — the report cannot tell the difference and
// does not pretend to.
func writeUncovered(b *strings.Builder, in CoverageInput, counts map[string]int) {
	missing := uncovered(in.Vocabulary, counts)
	fmt.Fprintf(b, "\n## No notes — %d\n\n", len(missing))
	if len(missing) == 0 {
		b.WriteString("_every stack in the vocabulary has a note_\n")
		return
	}
	b.WriteString("Either a hole in the wiki or a stack you do not work in; " +
		"this report cannot tell which.\n\n")
	for _, s := range missing {
		fmt.Fprintf(b, "- %s\n", s)
	}
}

func uncovered(vocabulary []string, counts map[string]int) []string {
	var out []string
	for _, s := range vocabulary {
		if counts[s] == 0 {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// writeTypeMix shows the shape of the vault. A vault that is 90% concept notes records
// what things are and almost nothing about what went wrong or what was decided.
func writeTypeMix(b *strings.Builder, in CoverageInput) {
	counts := map[string]int{}
	for _, t := range in.Types {
		counts[t] += 0
	}
	for _, e := range in.Entries {
		counts[e.Type]++
	}
	b.WriteString("\n## By type\n\n")
	for _, kv := range topN(counts, 20) {
		fmt.Fprintf(b, "- **%s** — %d\n", kv.k, kv.v)
	}
}
