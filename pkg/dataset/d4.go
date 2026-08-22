package dataset

import "time"

// D4 is ADDENDUM §D.1's gate-repair dataset: (failing draft, gate error, fixed draft)
// triples, captured when `forge gate --previous-draft` sees a retry that now passes.
// Like D2, its trigger is one CLI call with no legitimate re-fire to guard against, so
// every capture is appended as its own line — no Key(), no dedup. D4Tag matches the
// packaged dataset.capture list verbatim ("d4"), as D2Tag now does too — B-024 closed the
// mismatch that once made D2 inert under the shipped config. TestPackagedCaptureListGates
// pins both against the packaged layer so neither can drift again.
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

// AppendD4 writes one triple as a JSONL line. D4Enabled is gone: it only ever existed
// because Enabled() had already claimed the general name for D2, and tier.go's
// D4.Enabled removes the reason for a per-tier gate function at all.
func AppendD4(vaultRoot string, p D4Pair) error { return D4.Append(vaultRoot, p) }
