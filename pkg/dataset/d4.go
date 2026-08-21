package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

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

// D4Enabled reports whether the config chain turned D4 capture on. A second function
// rather than a parameter on Enabled: Enabled already hardcodes D2Tag (see d2.go) and is
// exercised by an existing call site, so it stays as-is.
func D4Enabled(capture []string) bool {
	for _, c := range capture {
		if c == D4Tag {
			return true
		}
	}
	return false
}

// AppendD4 writes one triple as a JSONL line, mirroring AppendD2 exactly.
func AppendD4(vaultRoot string, p D4Pair) error {
	path := filepath.Join(vaultRoot, D4Path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
