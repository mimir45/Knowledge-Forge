package dataset

import (
	"path/filepath"
	"time"
)

// D1OutcomePath is deliberately its own file.
const (
	D1OutcomeKind = "d1-outcome"
	D1OutcomePath = ".forge/datasets/d1-outcomes.jsonl"
)

// D1Outcome is one JSONL record: the note write that followed a recall call, keyed back
// to it by RunID.
type D1Outcome struct {
	Kind       string    `json:"kind"`
	RunID      string    `json:"run_id"`
	Published  bool      `json:"published"`
	CapturedAt time.Time `json:"captured_at"`
}

// AppendD1Outcome writes one outcome record.
func AppendD1Outcome(vaultRoot string, o D1Outcome) error {
	return appendJSONL(filepath.Join(vaultRoot, D1OutcomePath), o)
}
