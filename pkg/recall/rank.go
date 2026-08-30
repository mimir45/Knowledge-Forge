package recall

import (
	"math"
	"time"
)

// BodyPassSize is DESIGN §8 step 3's "top 20 files". The three frontmatter channels
// rank first and only the leaders are opened, which is what keeps the body pass cheap
// on a vault that no longer fits in a few hundred kilobytes.
const BodyPassSize = 20

// TopN is the size of the emitted array (recall-spec.md §4: `candidates` is "at most 10
// entries"). This bounds only the `candidates` output — `neighbours` is a separate field
// in the §4 contract with no stated cap there (see NeighbourWindow).
const TopN = 10

// NeighbourWindow bounds how many top-ranked, already body-scored candidates
// Thresholds.Neighbours may draw from — wider than TopN=10 so a genuinely-scoring 11th+
// candidate is not discarded by TopN truncation before the neighbour band ever sees it
// (BACKLOG B-036: on a broad query, all 10 truncated candidates can already clear the
// neighbour floor, so no floor value can surface an 11th that Rank never computed).
//
// It equals BodyPassSize today because both cite DESIGN §8 step 3's "top 20 files" — a
// neighbour can only be as informed as a candidate that was actually body-scored. But the
// two are kept as separate named constants on purpose: BodyPassSize is a scoring-cost
// boundary (which candidates get opened and body-rescored) and NeighbourWindow is an
// output-truncation boundary (how much of the ranked list a downstream caller may see).
// B-038 may move the former without silently moving the latter, and vice versa — see
// rank_test.go's TestNeighbourWindowMatchesBodyPassSizeToday, which fails on purpose if
// the two drift apart without a deliberate decision either way.
const NeighbourWindow = 20

// RankPool scores every doc and returns every nonzero-scoring candidate, sorted, highest
// first — the same computation Rank has always done, minus its final TopN truncation.
// Rank and Thresholds.ResultFrom each take their own truncated view of this; RankPool
// itself belongs to neither. `now` is an argument rather than time.Now() so staleness is
// testable and so a run is a pure function of its inputs.
func RankPool(q Query, docs []Doc, now time.Time) []Candidate {
	return RankPoolWithBodyPass(q, docs, now, BodyPassSize)
}

// RankPoolWithBodyPass is RankPool with an explicit body-pass window instead of the
// package constant BodyPassSize. It exists for B-038's measurement harness in cmd/forge,
// which needs to compare today's window against a widened or removed one without moving
// the shipped constant — that would be exactly the unmeasured performance change B-038's
// own text forbids. Production code calls RankPool, never this, at any size other than
// BodyPassSize; treat a second call site as a review flag, not a convenience.
func RankPoolWithBodyPass(q Query, docs []Doc, now time.Time, bodyPassSize int) []Candidate {
	return RankPoolWithTieBreak(q, docs, now, bodyPassSize, PathTieBreak)
}

// RankPoolWithTieBreak is RankPoolWithBodyPass with an explicit tie-break instead of the
// shipped PathTieBreak. It exists for B-038 question (b)'s measurement harness in
// cmd/forge, which needs to compare today's path tie-break against an alternative
// (recency, verified date, document-frequency — tiebreak.go) without moving shipped
// behavior. Production code calls RankPoolWithBodyPass, never this, at any tie-break
// other than PathTieBreak; treat a second call site as a review flag, not a convenience.
func RankPoolWithTieBreak(q Query, docs []Doc, now time.Time, bodyPassSize int, tb TieBreak) []Candidate {
	s := newScope(q, docs)
	cands := make([]Candidate, 0, len(docs))
	for _, d := range docs {
		cands = append(cands, s.frontmatterScore(d, now))
	}
	sortByScoreWith(cands, tb)
	s.bodyPass(cands, docs, bodyPassSize)
	sortByScoreWith(cands, tb)
	return round(nonZero(cands))
}

// BodyPassWindow returns the Path of every candidate the body pass would open — the
// leading bodyPassSize candidates after the frontmatter-only sort, before body scores
// blend in. It exists because RankPoolWithTieBreak's own output can't show this: nonZero
// drops a body-passed note that gets no body-term hits, which is exactly the population
// B-038's tie-break comparison needs to diff. Never called by production code.
func BodyPassWindow(q Query, docs []Doc, now time.Time, bodyPassSize int, tb TieBreak) []string {
	s := newScope(q, docs)
	cands := make([]Candidate, 0, len(docs))
	for _, d := range docs {
		cands = append(cands, s.frontmatterScore(d, now))
	}
	sortByScoreWith(cands, tb)
	n := min(len(cands), bodyPassSize)
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = cands[i].Path
	}
	return out
}

// Rank returns the top TopN candidates, highest first — recall-spec.md §4's `candidates`
// contract. Byte-identical in shape and behaviour to before B-036: forge intent's top-1
// pick and every caller that only wants candidates keeps using this unchanged.
func Rank(q Query, docs []Doc, now time.Time) []Candidate {
	return truncate(RankPool(q, docs, now), TopN)
}

// truncate returns the leading n candidates, or all of them if there are fewer than n.
func truncate(cands []Candidate, n int) []Candidate {
	if len(cands) > n {
		return cands[:n]
	}
	return cands
}

// NeighbourPool is the slice Thresholds.ResultFrom band-filters for Neighbours. Exported
// so cmd/forge's sweep and frequency harnesses can mirror production truncation exactly
// instead of reimplementing it.
func NeighbourPool(pool []Candidate) []Candidate {
	return truncate(pool, NeighbourWindow)
}

// nonZero drops candidates no channel matched at all. A vault rarely has ten notes with
// any overlap with a genuinely new question, and padding the array to TopN with 0.000
// rows put `index.md` and `log.md` at the top of a CREATE verdict — noise a caller has
// to know to ignore. An empty array is the honest answer to "what covers this".
func nonZero(cands []Candidate) []Candidate {
	for i, c := range cands {
		if c.Score == 0 {
			return cands[:i]
		}
	}
	return cands
}

// round trims scores to three decimals. Float noise from the renormalizing division
// ("0.6954545454545457") is not signal, and a threshold comparison should not turn on
// the seventeenth digit of a division the spec writes as 0.695.
func round(cands []Candidate) []Candidate {
	for i := range cands {
		cands[i].Score = math.Round(cands[i].Score*1000) / 1000
	}
	return cands
}

// frontmatterScore is the ranking pass over the 0.4/0.3/0.2 channels. The body channel
// is appended later, for the leaders only.
func (s scope) frontmatterScore(d Doc, now time.Time) Candidate {
	chs := []Channel{s.titleChannel(d), s.tagsChannel(d), s.stackChannel(d)}
	score, matched := blend(chs)
	return Candidate{
		Slug: d.Slug, Path: d.Rel, Title: d.Title, Score: score,
		Updated: d.Updated, Verified: d.Verified,
		Stale: IsStale(d, now), MatchedOn: matched, Channels: chs,
	}
}

// bodyPass opens the leading candidates and rescores them with the 0.1 channel. Notes
// outside the window keep their frontmatter-only score, which is correct: the body
// channel can move a score by at most 0.1 and never lifts a non-match into the band.
// size is BodyPassSize in production (see RankPool); RankPoolWithBodyPass lets B-038's
// measurement harness vary it without touching the shipped constant.
func (s scope) bodyPass(cands []Candidate, docs []Doc, size int) {
	byRel := map[string]Doc{}
	for _, d := range docs {
		byRel[d.Rel] = d
	}
	for i := range cands[:min(len(cands), size)] {
		d, ok := byRel[cands[i].Path]
		if !ok || d.LoadBody == nil {
			continue
		}
		cands[i].Channels = append(cands[i].Channels, s.bodyChannel(d.LoadBody()))
		cands[i].Score, cands[i].MatchedOn = blend(cands[i].Channels)
	}
}

// Neighbours are the candidates a new note links to on a CREATE verdict (DESIGN §5.3:
// "then link to the 0.3–0.55 neighbours").
func (t Thresholds) Neighbours(cands []Candidate) []Candidate {
	var out []Candidate
	for _, c := range cands {
		if c.Score >= t.Neighbour && c.Score < t.Update {
			out = append(out, c)
		}
	}
	return out
}
