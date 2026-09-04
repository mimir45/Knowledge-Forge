package dataset

import "time"

// D4 is the gate-repair dataset: (failing draft, gate error, fixed draft) triples.
const (
	D4Kind = "d4-gate-repair"
	D4Path = ".forge/datasets/d4.jsonl"
	D4Tag  = "d4" // the cfg.Dataset.Capture entry that gates this
)

// D4Pair is one JSONL record. Stage is always "gate" today — forge gate is D4's only
// producer — recorded explicitly rather than left blank.
type D4Pair struct {
	Kind         string    `json:"kind"`
	Stage        string    `json:"stage"`
	FailingDraft string    `json:"failing_draft"`
	GateError    string    `json:"gate_error"`
	FixedDraft   string    `json:"fixed_draft"`
	CapturedAt   time.Time `json:"captured_at"`
}

// AppendD4 writes one triple as a JSONL line.
func AppendD4(vaultRoot string, p D4Pair) error { return D4.Append(vaultRoot, p) }
