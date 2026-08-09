package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"knowledge-forge/pkg/recall"
)

// printExplain writes the score breakdown to stderr so stdout stays parseable JSON.
// Its job is to make a surprising verdict debuggable without a rebuild: which channels
// were active, what each contributed, and what the renormalizing denominator was.
func printExplain(w io.Writer, q recall.Query, res recall.Result) {
	fmt.Fprintf(w, "query terms: %s\n", strings.Join(recall.Terms(q.Question), ", "))
	if len(q.Stack) > 0 {
		fmt.Fprintf(w, "stack hints: %s\n", strings.Join(q.Stack, ", "))
	}
	if len(res.Candidates) == 0 {
		fmt.Fprint(w, "\nno candidate scored above zero on any channel\n")
	}
	for _, c := range res.Candidates {
		explainOne(w, c)
	}
	// Printed on every path, including the empty one: a caller reading stderr for the
	// verdict must not get silence on the exact case the verdict matters most, CREATE.
	fmt.Fprintf(w, "\nverdict: %s\n", res.Verdict)
	for _, n := range res.Neighbours {
		fmt.Fprintf(w, "  link to: %-48s %.3f\n", n.Slug, n.Score)
	}
}

func explainOne(w io.Writer, c recall.Candidate) {
	fmt.Fprintf(w, "\n%-52s %.3f%s\n", c.Slug, c.Score, staleMark(c.Stale))
	num, den := 0.0, 0.0
	for _, ch := range c.Channels {
		if !ch.Active {
			fmt.Fprintf(w, "  %-6s   inactive — the query supplied no %s input\n", ch.Name, ch.Name)
			continue
		}
		num += ch.Weight * ch.Value
		den += ch.Weight
		fmt.Fprintf(w, "  %-6s %.3f x %.1f = %.3f%s\n",
			ch.Name, ch.Value, ch.Weight, ch.Weight*ch.Value, hitList(ch.Hits))
		// Since B-008 the terms in a hit list no longer count equally, so the list alone
		// no longer explains the value. Print the weights that produced it.
		if len(ch.Terms) > 0 {
			fmt.Fprintf(w, "         idf %s\n", weightList(ch.Terms))
		}
	}
	fmt.Fprintf(w, "  sum   %.3f / %.3f = %.3f\n", num, den, c.Score)
}

func staleMark(stale bool) string {
	if stale {
		return "  [stale]"
	}
	return ""
}

// weightList renders a channel's per-term IDF, sorted so the line is byte-stable.
func weightList(terms map[string]float64) string {
	keys := make([]string, 0, len(terms))
	for t := range terms {
		keys = append(keys, t)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, t := range keys {
		parts[i] = fmt.Sprintf("%s %.2f", t, terms[t])
	}
	return strings.Join(parts, ", ")
}

func hitList(hits []string) string {
	if len(hits) == 0 {
		return ""
	}
	return "   " + strings.Join(hits, ", ")
}
