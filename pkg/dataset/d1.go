package dataset

import "time"

// D1 is ADDENDUM §D.1's routing dataset: (question features → recall decision) pairs,
// captured by `forge recall`. Two scope limits are deliberate and both are restated in
// every export's datasheet rather than left for a reader to infer.
//
// It captures recall calls, not the hook path's ranking. ADDENDUM says "every run", but
// `forge intent` also ranks the vault on every UserPromptSubmit and it is the wrong
// producer: it carries a 50ms budget and a contract never to disturb the session, and a
// passive prompt hint is not a question anyone asked. intent.go builds its own
// recall.Query and never reaches this path, so the limit is structural, not a convention.
//
// The outcome is not captured. A pair here is (question features → the routing decision),
// with nothing recording whether that decision turned out right, because there is no
// run_id to join a later note-write back to the recall call that preceded it. That makes
// D1 supervised on the router's own output — good enough to distil the routing rule into
// a small model, not evidence the rule is correct. The correlation key is BACKLOG B-035.
const (
	D1Kind = "d1-routing"
	D1Path = ".forge/datasets/d1.jsonl"
	D1Tag  = "d1"
)

// D1Pair is one JSONL record. The raw question never appears: QHash is the caller's
// telemetry.QHash and Topic is the slug, the same two-field shape DESIGN §14's ask event
// uses — "never store raw question text, hash + extracted topic only" (ADDENDUM §D).
type D1Pair struct {
	Kind           string    `json:"kind"`
	QHash          string    `json:"q_hash"`
	Topic          string    `json:"topic"`
	Decision       string    `json:"decision"`
	Stack          []string  `json:"stack,omitempty"`
	RecallTopScore float64   `json:"recall_top_score"`
	Candidates     int       `json:"candidates"`
	CapturedAt     time.Time `json:"captured_at"`
}

// AppendD1 writes one pair as a JSONL line.
func AppendD1(vaultRoot string, p D1Pair) error { return D1.Append(vaultRoot, p) }
