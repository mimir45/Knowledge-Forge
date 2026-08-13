package report

import (
	"fmt"
	"strings"
	"time"

	"knowledge-forge/pkg/similarity"
)

// DuplicatesInput is what duplicates.md renders from.
type DuplicatesInput struct {
	Pairs     []similarity.Pair // already filtered at Threshold, best first
	Threshold float64
	Compared  int               // candidate pairs banding nominated
	Slugs     map[string]string // note path -> slug
	Types     map[string]string // note path -> type, to show what was scoped together
	Now       time.Time
}

// specThreshold is ADDENDUM §B.4's ">0.85 similar". It is printed, not applied — see
// references/duplicate-spec.md, which is what the code follows where the two disagree.
const specThreshold = 0.85

// TopDuplicatePair returns the highest-scoring pair that clears the 0.85 spec threshold —
// the same bar duplicates.md's header and weekly.go's Act now section use, not the lower
// operating threshold pairs is otherwise filtered at. For check.ai_pass's merge-proposal
// sub-task: pairs must already be sorted best-first, as similarity.Index.Pairs returns it.
func TopDuplicatePair(pairs []similarity.Pair) (similarity.Pair, bool) {
	for _, p := range pairs {
		if p.Score >= specThreshold {
			return p, true
		}
	}
	return similarity.Pair{}, false
}

// RenderDuplicates produces duplicates.md.
//
// The header states outright that nothing in the vault crosses §B.4's 0.85. That is the
// honest headline rather than a footnote: a reader who knows the spec would otherwise see
// a list of 0.4-scoring pairs and assume the threshold in the doc was met. It was not, and
// no shingle width makes it met — 0.85 is a copy-paste detector and these notes are not
// copy-paste, they are the same behaviour written up twice, months apart, in other words.
func RenderDuplicates(in DuplicatesInput) []byte {
	var b strings.Builder
	header(&b, "Duplicates", duplicatesSummary(in), in.Now)
	if len(in.Pairs) == 0 {
		empty(&b, "no pair scores at or above the threshold")
		return []byte(b.String())
	}
	b.WriteString("\n## Merge candidates\n\n")
	for _, p := range in.Pairs {
		writePair(&b, in, p)
	}
	writeMethod(&b, in)
	return []byte(b.String())
}

func duplicatesSummary(in DuplicatesInput) string {
	return fmt.Sprintf(
		"**%d %s** at or above %.2f, from %d candidate pairs.\n\n"+
			"No pair in this vault reaches ADDENDUM §B.4's 0.85 — at that threshold this "+
			"report is empty and stays empty. See `references/duplicate-spec.md` §1.\n",
		len(in.Pairs), plural(len(in.Pairs), "pair", "pairs"), in.Threshold, in.Compared)
}

func writePair(b *strings.Builder, in DuplicatesInput, p similarity.Pair) {
	fmt.Fprintf(b, "- **%.2f** — %s · %s",
		p.Score, note(in.Slugs[p.A], p.A), note(in.Slugs[p.B], p.B))
	if t := in.Types[p.A]; t != "" {
		fmt.Fprintf(b, "  _(%s)_", t)
	}
	b.WriteString("\n")
}

// writeMethod records what the number means, in the report rather than only in the spec.
// A similarity score with no stated method is a number a reader cannot argue with.
func writeMethod(b *strings.Builder, in DuplicatesInput) {
	fmt.Fprintf(b, "\n## Method\n\n"+
		"%d-hash MinHash over single-word shingles of the note body, frontmatter excluded, "+
		"banded %dx%d. Only notes of the same type are ever compared: the vault's "+
		"highest-scoring pairs are cross-type (a decision and the pitfall that caused it) "+
		"and they are not duplicates, they are the taxonomy working.\n",
		similarity.Hashes, similarity.Bands, similarity.Rows)
}
