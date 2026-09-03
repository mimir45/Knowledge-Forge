package recall

// Result is the recall output envelope (recall-spec.md §4).
type Result struct {
	Question   string      `json:"question"`
	Verdict    Decision    `json:"verdict"`
	TopScore   float64     `json:"top_score"`
	Candidates []Candidate `json:"candidates"`
	Neighbours []Candidate `json:"neighbours"`
}

// ResultFrom decides the verdict for a full, untruncated candidate pool (RankPool's
// output) and packages both output views: Candidates is truncated to TopN.
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
