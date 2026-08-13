// Package report renders analyses to markdown. Nothing here reads a model; every
// number it prints comes from the static core.
package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Entry is one note as the index needs it.
type Entry struct {
	Rel, Slug, Title, Type string
	Stack                  []string
	Updated, Verified      time.Time
	FreshnessDays          int
	Valid                  bool
	Orphan                 bool
}

// IndexInput is everything `forge index` renders from.
type IndexInput struct {
	Entries []Entry
	Gaps    []string // topics asked but never written; empty until the ask log exists
	Now     time.Time
	MaxSize int // byte budget; 0 means the DESIGN §7.1 default of 4096
}

const defaultBudget = 4096

// RenderIndex produces _index.md. It is deterministic for a given (entries, day): the
// header carries a date, not a timestamp, so running it twice in one day is a no-op on
// disk. That is what "idempotent" has to mean for a file with a rebuild stamp in it.
func RenderIndex(in IndexInput) []byte {
	if in.MaxSize == 0 {
		in.MaxSize = defaultBudget
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Vault index — %d notes — rebuilt %s\n",
		len(in.Entries), in.Now.Format("2006-01-02"))
	b.WriteString(summaryLine(in))
	writeByStack(&b, in)
	writeRecent(&b, in)
	writeStale(&b, in)
	writeGaps(&b, in)
	return []byte(Trim(b.String(), in.MaxSize))
}

func summaryLine(in IndexInput) string {
	invalid, orphans := 0, 0
	for _, e := range in.Entries {
		if !e.Valid {
			invalid++
		}
		if e.Orphan {
			orphans++
		}
	}
	return fmt.Sprintf("\n%d contract-valid · %d failing · %d orphaned\n",
		len(in.Entries)-invalid, invalid, orphans)
}

func writeByStack(b *strings.Builder, in IndexInput) {
	counts := map[string]int{}
	for _, e := range in.Entries {
		for _, s := range e.Stack {
			counts[s]++
		}
	}
	b.WriteString("\n## By stack\n\n")
	if len(counts) == 0 {
		b.WriteString("_none recorded_\n")
		return
	}
	for _, kv := range topN(counts, 20) {
		fmt.Fprintf(b, "- **%s** — %d\n", kv.k, kv.v)
	}
}

type kv struct {
	k string
	v int
}

// topN sorts by count descending, then key ascending, so ties never reorder between runs.
func topN(counts map[string]int, n int) []kv {
	out := make([]kv, 0, len(counts))
	for k, v := range counts {
		out = append(out, kv{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].v != out[j].v {
			return out[i].v > out[j].v
		}
		return out[i].k < out[j].k
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func writeRecent(b *strings.Builder, in IndexInput) {
	es := append([]Entry(nil), in.Entries...)
	sort.Slice(es, func(i, j int) bool {
		if !es[i].Updated.Equal(es[j].Updated) {
			return es[i].Updated.After(es[j].Updated)
		}
		return es[i].Slug < es[j].Slug
	})
	b.WriteString("\n## Recently updated\n\n")
	for _, e := range head(es, 15) {
		fmt.Fprintf(b, "- [[%s]] — %s\n", e.Slug, e.Updated.Format("2006-01-02"))
	}
}

// writeStale is the section the truncation budget must never drop: it is the only
// actionable output of the index, so it is rendered before Gaps and always keeps counts.
func writeStale(b *strings.Builder, in IndexInput) {
	stale := staleEntries(in)
	fmt.Fprintf(b, "\n## Stale — %d past their freshness window\n\n", len(stale))
	for _, e := range head(stale, 15) {
		fmt.Fprintf(b, "- [[%s]] — verified %s (%dd budget)\n",
			e.Slug, e.Verified.Format("2006-01-02"), e.FreshnessDays)
	}
	if len(stale) > 15 {
		fmt.Fprintf(b, "- _… and %d more_\n", len(stale)-15)
	}
}

func staleEntries(in IndexInput) []Entry {
	var out []Entry
	for _, e := range in.Entries {
		if e.FreshnessDays <= 0 || e.Verified.IsZero() {
			continue
		}
		if in.Now.Sub(e.Verified).Hours()/24 > float64(e.FreshnessDays) {
			out = append(out, e)
		}
	}
	sortByVerified(out)
	return out
}

// sortByVerified needs the slug behind the date. `verified` is a date, not a timestamp, so
// notes verified on the same day are the common case rather than a rare tie — and this list
// is truncated to 15, so an unbroken tie does not merely reorder the report, it changes
// which notes appear in it at all. The same defect in pkg/drift's name table made drift.md
// oscillate between runs on an unchanged tree.
func sortByVerified(out []Entry) {
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Verified.Equal(out[j].Verified) {
			return out[i].Verified.Before(out[j].Verified)
		}
		return out[i].Slug < out[j].Slug
	})
}

func writeGaps(b *strings.Builder, in IndexInput) {
	b.WriteString("\n## Gaps — asked but never written\n\n")
	if len(in.Gaps) == 0 {
		b.WriteString("_no ask log yet_\n")
		return
	}
	for _, g := range head(in.Gaps, 10) {
		fmt.Fprintf(b, "- %s\n", g)
	}
}

func head[T any](s []T, n int) []T {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// Trim cuts at a line boundary and says so, rather than emitting a half-written
// bullet that a SessionStart hook would feed to a model as if it were complete.
// Exported for cmd/forge's session-context hook, which shares this exact 4KB-budget
// contract for the profile it appends after the index.
func Trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max-64]
	if i := strings.LastIndexByte(cut, '\n'); i > 0 {
		cut = cut[:i+1]
	}
	return cut + "\n_[index truncated to fit the 4KB budget]_\n"
}
