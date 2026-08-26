package recall

// Result is the recall output envelope (recall-spec.md §4). The verdict travels with
// the candidates rather than being left for the caller to derive: AUDIT §8.4 D-7 moves
// the thresholds into Phase 3's config union, and a skill that restated DESIGN §5.3's
// tree in prose would silently diverge from the config the next phase introduces. One
// implementation, in Go, is the only copy.
type Result struct {
	Question   string      `json:"question"`
	Verdict    Decision    `json:"verdict"`
	TopScore   float64     `json:"top_score"`
	Candidates []Candidate `json:"candidates"`
	Neighbours []Candidate `json:"neighbours"`
}

// ResultFrom decides the verdict for a full, untruncated candidate pool (RankPool's
// output) and packages both output views: Candidates is truncated to TopN (recall-spec.md
// §4's contract, unchanged), while Neighbours — on a CREATE verdict only — band-filters
// NeighbourPool's wider view of the same pool (BACKLOG B-036), so a real 11th+ candidate
// that TopN truncation would otherwise discard can still be admitted as a neighbour.
//
// Neighbours are populated on CREATE only. The 0.150–0.55 band exists to answer "what
// should this new note link to" (DESIGN §5.3); on an ANSWER or UPDATE verdict the same
// band is a list of notes the caller was told to ignore, and emitting it invites a
// consumer to link them anyway.
func (t Thresholds) ResultFrom(q Query, pool []Candidate) Result {
	r := Result{Question: q.Question, Candidates: truncate(pool, TopN), Neighbours: []Candidate{}}
	if r.Candidates == nil {
		r.Candidates = []Candidate{}
	}
	var top *Candidate
	if len(r.Candidates) > 0 {
		top = &r.Candidates[0]
		r.TopScore = top.Score
	}
	if r.Verdict = t.Decide(top); r.Verdict == Create {
		r.Neighbours = append(r.Neighbours, t.Neighbours(NeighbourPool(pool))...)
	}
	return r
}
