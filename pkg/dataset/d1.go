package dataset

import "time"

// D1 is the routing dataset: (question features → recall decision) pairs.
const (
	D1Kind = "d1-routing"
	D1Path = ".forge/datasets/d1.jsonl"
	D1Tag  = "d1"
)

// D1Pair is one JSONL record.
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
	// export_records.go's loadD1 fills it in memory, after reading.
	Outcome *bool `json:"outcome,omitempty"`
}

// AppendD1 writes one pair as a JSONL line.
func AppendD1(vaultRoot string, p D1Pair) error { return D1.Append(vaultRoot, p) }
