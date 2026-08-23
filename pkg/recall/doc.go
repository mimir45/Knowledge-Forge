package recall

// Doc is one note as recall sees it: frontmatter, plus a lazy handle on the body. The
// body is a func because DESIGN §8 step 3 reads only the top 20 candidates — the other
// channels rank first, and most notes are never opened.
type Doc struct {
	Rel, Slug, Title  string
	Tags, Stack       []string
	Updated, Verified string
	FreshnessDays     int
	LoadBody          func() []byte
}

// Query is the question plus the caller's stack hints (`--stack java,spring-boot`).
type Query struct {
	Question string
	Stack    []string
}

// Thresholds is DESIGN §5.3's decision tree. Phase 2 hardcodes the defaults; Phase 3
// wires them to the config chain (AUDIT §8.4 D-7). They live here so no literal is
// scattered — a threshold in two places is a threshold that drifts.
type Thresholds struct {
	Answer    float64 // ≥ this and fresh → ANSWER_FROM_VAULT; ≥ this and stale → UPDATE(refresh)
	Update    float64 // ≥ this → UPDATE(extend)
	Neighbour float64 // ≥ this on a CREATE → link as a neighbour
}

// DefaultThresholds are DESIGN §10's recall.answer_threshold / update_threshold.
//
// Answer and Update are DESIGN §5.3's and do not move — B-008 forbids it by name.
// Neighbour was re-derived from 0.30 to 0.125 closing BACKLOG B-033: 0.30 was calibrated
// against the pre-B-008 scale, and after that change it left 6 of 15 adjacent-topic
// queries with zero neighbours, i.e. writing a new note orphaned in a vault whose graph
// report already tracks 21 orphans of 94. Swept over labelled ground truth
// (cmd/forge/testdata/neighbour-labels.txt, TestNeighbourFloorSweep), 0.125 is F1's
// maximum and the only value that recovers the case B-033 was opened for. The full
// argument is in docs/BACKLOG.md's B-033 closure note; changing this number without
// re-running that sweep is tuning.
//
// This is the Go default and every un-configured install sees it. It has a twin in
// config/forge.config.example.md's neighbour_min_score, which must move with it.
var DefaultThresholds = Thresholds{Answer: 0.85, Update: 0.55, Neighbour: 0.125}

// Decision is what stage 2 does with the top candidate.
type Decision string

const (
	AnswerFromVault Decision = "ANSWER_FROM_VAULT"
	UpdateRefresh   Decision = "UPDATE(refresh)"
	UpdateExtend    Decision = "UPDATE(extend)"
	Create          Decision = "CREATE"
)

// Decide applies §5.3's tree to the top candidate. An empty result is CREATE: no note
// scored, so there is nothing to answer from or extend.
func (t Thresholds) Decide(top *Candidate) Decision {
	switch {
	case top == nil:
		return Create
	case top.Score >= t.Answer && !top.Stale:
		return AnswerFromVault
	case top.Score >= t.Answer:
		return UpdateRefresh
	case top.Score >= t.Update:
		return UpdateExtend
	}
	return Create
}

// Channel is one weighted component of the score. Active records whether the query
// supplied any input for it — an inactive channel is undefined, not zero, and drops out
// of the denominator. See recall-spec.md §2.5 for why that distinction decides the
// verdict rather than merely the scale.
// Terms carries the per-term IDF weights behind Value on the channels that use them.
// Without it the number is unauditable: after B-008 "tags 0.136, hits: [spring]" cannot
// be re-derived from the hit list, because the terms no longer count equally. The map is
// query-scope and shared by every candidate, so it is read-only and costs no allocation.
//
// DF carries the raw document frequency behind each of those weights. B-008's second pass
// had to count it by hand to explain why the weighting had changed nothing, because a
// printed weight of 0.00 reads identically whether a term is on every note or on none —
// and after this fix those two cases have opposite consequences. Both maps stay out of
// the §4 output contract.
type Channel struct {
	Name   string             `json:"name"`
	Weight float64            `json:"weight"`
	Value  float64            `json:"value"`
	Active bool               `json:"active"`
	Hits   []string           `json:"hits,omitempty"`
	Terms  map[string]float64 `json:"-"`
	DF     map[string]int     `json:"-"`
}

// Candidate is one scored note. The JSON tags are the output contract in
// recall-spec.md §4; Channels is carried for --explain and omitted from the array.
type Candidate struct {
	Slug      string    `json:"slug"`
	Path      string    `json:"path"`
	Title     string    `json:"title"`
	Score     float64   `json:"score"`
	Updated   string    `json:"updated"`
	Verified  string    `json:"verified"`
	Stale     bool      `json:"stale"`
	MatchedOn []string  `json:"matched_on"`
	Channels  []Channel `json:"-"`
}
