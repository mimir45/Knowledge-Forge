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
// The outcome is partial, not absent (BACKLOG B-035, closed 2026-08-25). Each pair now
// carries RunID, minted by telemetry.NewRunID in runRecall and emitted in forge recall's
// JSON envelope. A `forge gate --run-id <id>` call that threads it back appends a
// separate D1Outcome record (d1_outcome.go) keyed by the same RunID — a second file, not
// a rewritten field, because this line is already on disk and immutable by the time the
// gate call happens, sometimes minutes later in a different process. Export joins the two
// by RunID (export_records.go's loadD1); a gate call that never received --run-id simply
// never joins. That is the documented, honest degradation — see the D1 datasheet's join
// rate, not a silent guess either way.
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
	RunID          string    `json:"run_id,omitempty"`
	QHash          string    `json:"q_hash"`
	Topic          string    `json:"topic"`
	Decision       string    `json:"decision"`
	Stack          []string  `json:"stack,omitempty"`
	RecallTopScore float64   `json:"recall_top_score"`
	Candidates     int       `json:"candidates"`
	CapturedAt     time.Time `json:"captured_at"`
	// Outcome is never written by AppendD1 — it stays nil on the capture file forever.
	// export_records.go's loadD1 fills it in memory, after reading, by joining against
	// D1Outcome on RunID; a pointer so "never joined" (nil) and "joined, gate published
	// nothing" (non-nil, false) stay distinguishable in the rendered export.
	Outcome *bool `json:"outcome,omitempty"`
}

// AppendD1 writes one pair as a JSONL line.
func AppendD1(vaultRoot string, p D1Pair) error { return D1.Append(vaultRoot, p) }
