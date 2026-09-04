package recall

import (
	"math"
	"sort"
	"time"
)

// BodyPassSize is the spec's "top 20 candidates".
const BodyPassSize = 20

// TopN is the size of the emitted array (recall-spec.md §4: `candidates` is "at most 10
// entries").
const TopN = 10

// NeighbourWindow bounds how many top-ranked, already body-scored candidates
// Thresholds.Neighbours may draw from.
const NeighbourWindow = 20

// RankPool scores every doc and returns every nonzero-scoring candidate, sorted,
// highest first — the same computation Rank has always done.
func RankPool(q Query, docs []Doc, now time.Time) []Candidate {
	return RankPoolWithBodyPass(q, docs, now, BodyPassSize)
}

// RankPoolWithBodyPass is RankPool with an explicit body-pass window instead of the
// package constant BodyPassSize.
func RankPoolWithBodyPass(q Query, docs []Doc, now time.Time, bodyPassSize int) []Candidate {
	s := newScope(q, docs)
	cands := make([]Candidate, 0, len(docs))
	for _, d := range docs {
		cands = append(cands, s.frontmatterScore(d, now))
	}
	sortByScore(cands)
	s.bodyPass(cands, docs, bodyPassSize)
	sortByScore(cands)
	return round(nonZero(cands))
}

// Rank returns the top TopN candidates, highest first — recall-spec.md §4's
// `candidates` contract.
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

// NeighbourPool is the slice Thresholds.ResultFrom band-filters for Neighbours.
func NeighbourPool(pool []Candidate) []Candidate {
	return truncate(pool, NeighbourWindow)
}

// nonZero drops candidates no channel matched at all.
func nonZero(cands []Candidate) []Candidate {
	for i, c := range cands {
		if c.Score == 0 {
			return cands[:i]
		}
	}
	return cands
}

// round trims scores to three decimals.
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

// bodyPass opens the leading candidates and rescores them with the 0.1 channel.
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

// sortByScore orders by score descending, breaking ties on path so two runs over the
// same tree return byte-identical JSON. Phase 2b re-measures against these numbers.
func sortByScore(c []Candidate) {
	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Score != c[j].Score {
			return c[i].Score > c[j].Score
		}
		return c[i].Path < c[j].Path
	})
}

// Neighbours are the candidates a new note links to on a CREATE verdict (the original
// spec: "then link to the 0.3–0.55 neighbours").
func (t Thresholds) Neighbours(cands []Candidate) []Candidate {
	var out []Candidate
	for _, c := range cands {
		if c.Score >= t.Neighbour && c.Score < t.Update {
			out = append(out, c)
		}
	}
	return out
}
