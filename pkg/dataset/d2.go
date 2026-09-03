package dataset

import "time"

// D2 is the advisor-distillation dataset: (draft, critique) pairs.
const (
	D2Kind = "d2-advisor-critique"
	D2Path = ".forge/datasets/d2.jsonl"
	D2Tag  = "d2" // the cfg.Dataset.Capture entry that gates this
)

// D2Pair is one JSONL record: the draft sent to the advisor and its critique, verbatim.
type D2Pair struct {
	Kind       string    `json:"kind"`
	Stage      string    `json:"stage"`
	Draft      string    `json:"draft"`
	Critique   string    `json:"critique"`
	CapturedAt time.Time `json:"captured_at"`
}

// AppendD2 writes one pair as a JSONL line, creating .forge/datasets/ on first use.
func AppendD2(vaultRoot string, p D2Pair) error { return D2.Append(vaultRoot, p) }
