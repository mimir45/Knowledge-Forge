package recall

import "sort"

// TieBreak compares two candidates already equal on Score. <0 means a sorts first, >0
// means b sorts first, 0 means this key cannot separate them and the caller must fall
// through to the next key. int rather than bool so B-038's saturation measurement
// (cmd/forge) can count "cannot separate" pairs, not just observe an ordering.
type TieBreak func(a, b Candidate) int

// PathTieBreak is today's sole shipped tie-break, extracted verbatim from the old
// sortByScore so RankPool's byte-identical guarantee has an exact reference to pin
// against (see TestRankPoolWithBodyPassMatchesRankPoolWithTieBreakAtPathTieBreak).
func PathTieBreak(a, b Candidate) int {
	switch {
	case a.Path < b.Path:
		return -1
	case a.Path > b.Path:
		return 1
	default:
		return 0
	}
}

// RecencyTieBreak favors the more recently Updated note — one of B-038 question (b)'s
// three named candidates (docs/TODO.md).
func RecencyTieBreak(a, b Candidate) int { return compareDates(a.Updated, b.Updated) }

// VerifiedTieBreak favors the more recently Verified note — the second of the three.
func VerifiedTieBreak(a, b Candidate) int { return compareDates(a.Verified, b.Verified) }

// compareDates parses both sides with parseDate (freshness.go) and prefers the newer
// one. A parseable date always sorts before an unparseable one — an undatable note is
// not evidence of recency either way, so it should not win a tie by default — and two
// unparseable dates tie (0), same as two notes with no signal at all.
func compareDates(a, b string) int {
	ta, oka := parseDate(a)
	tb, okb := parseDate(b)
	switch {
	case oka && okb && ta.After(tb):
		return -1
	case oka && okb && ta.Before(tb):
		return 1
	case oka && okb:
		return 0
	case oka:
		return -1
	case okb:
		return 1
	default:
		return 0
	}
}

// DocFreqTieBreak favors the more topic-specific note among score-tied candidates: the
// mean idf() (score.go, same idfCap) over the note's own Tags ∪ Stack, computed once
// from docFreq(docs) — the same corpus-wide count newScope already takes. Mean, not
// min: BACKLOG B-036's closure note found min-idf did not separate two "universal" notes
// from labelled-wanted neighbours, because several wanted notes shared the exact same
// min value; mean over the whole tag/stack set is a considered response to that finding,
// not a re-run of it — but it is not assumed to have fixed the saturation problem, which
// B-038's measurement harness (cmd/forge) checks directly.
//
// A note with no tags or stack (31 of 91 real notes, including this item's own two
// motivating rows — transactional-outbox-pattern.md, cqrs-and-event-driven-messaging.md)
// has no specificity signal at all. The comparator returns 0 (falls through to path) for
// any pair where either side lacks one, rather than sorting untagged notes last: the
// "untagged sorts last" alternative was checked by hand against those exact two rows and
// found to bury them below where today's path tie-break already puts them — worse than
// the status quo on the row that opened B-038. Do not simplify this back to that shape.
func DocFreqTieBreak(docs []Doc) TieBreak {
	spec := specificityOf(docs)
	return func(a, b Candidate) int {
		sa, oka := spec[a.Path]
		sb, okb := spec[b.Path]
		if !oka || !okb {
			return 0
		}
		switch {
		case sa > sb:
			return -1
		case sa < sb:
			return 1
		default:
			return 0
		}
	}
}

// specificityOf maps each doc's Rel path to the mean idf() over its own Tags ∪ Stack.
// A doc with neither is absent from the map, not present at 0 — 0 would read as "the
// least specific note there is" and win no ties it should lose, when the honest answer
// is "no signal."
func specificityOf(docs []Doc) map[string]float64 {
	tagDF, stackDF := docFreq(docs)
	n := len(docs)
	out := make(map[string]float64, len(docs))
	for _, d := range docs {
		sum, count := 0.0, 0
		for t := range setOf(d.Tags) {
			sum, count = sum+idf(tagDF[t], n), count+1
		}
		for t := range setOf(d.Stack) {
			sum, count = sum+idf(stackDF[t], n), count+1
		}
		if count > 0 {
			out[d.Rel] = sum / float64(count)
		}
	}
	return out
}

// sortByScoreWith orders by score descending, then tb, then always PathTieBreak as the
// final fallback — the same determinism guarantee sortByScore has always made (two runs
// over the same tree return byte-identical JSON), whatever tb is.
func sortByScoreWith(c []Candidate, tb TieBreak) {
	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Score != c[j].Score {
			return c[i].Score > c[j].Score
		}
		if d := tb(c[i], c[j]); d != 0 {
			return d < 0
		}
		return PathTieBreak(c[i], c[j]) < 0
	})
}

// sortByScore is the shipped rule: PathTieBreak alone. Production code calls only this;
// RankPoolWithTieBreak (rank.go) is B-038's measurement seam for the alternatives above.
func sortByScore(c []Candidate) { sortByScoreWith(c, PathTieBreak) }
