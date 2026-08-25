package dataset

import (
	"path/filepath"
	"time"
)

// D1OutcomePath is deliberately its own file, not a field rewritten onto an existing
// D1Pair line — see d1.go's doc comment for why. It has no entry in Tiers(): it is not a
// sixth capture surface with its own `dataset.capture` tag, it is D1's own outcome half,
// gated the same way D1 itself is (see cmd/forge/gate.go's captureD1Outcome) and read only
// by D1's export join (export_records.go's loadD1).
const (
	D1OutcomeKind = "d1-outcome"
	D1OutcomePath = ".forge/datasets/d1-outcomes.jsonl"
)

// D1Outcome is one JSONL record: the note write that followed a recall call, keyed back
// to it by RunID. Published spells out qualitygate.Report.Quarantine's negation rather
// than carrying a field named after the gate package's own vocabulary — a D1 consumer
// should not have to know what "Quarantine" means to read this file.
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
