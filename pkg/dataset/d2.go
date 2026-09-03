package dataset

import "time"

// D2 is the advisor-distillation dataset: (draft, critique) pairs. Where D3
// dedupes on a git commit a hook can legitimately re-fire on, D2's trigger is one CLI call
// to `forge engine run` that made a real advisor call — there is no re-fire to guard
// against, so every capture is appended as its own line, no Key() or idempotency check.
const (
	D2Kind = "d2-advisor-critique"
	D2Path = ".forge/datasets/d2.jsonl"
	D2Tag  = "d2" // the cfg.Dataset.Capture entry that gates this
)

// D2Pair is one JSONL record: the draft sent to the advisor and its critique, verbatim.
// Both the original specs stated the same requirement: "log the critique verbatim — it's
// dataset D2." Nothing here judges or reshapes what the advisor said.
type D2Pair struct {
	Kind       string    `json:"kind"`
	Stage      string    `json:"stage"`
	Draft      string    `json:"draft"`
	Critique   string    `json:"critique"`
	CapturedAt time.Time `json:"captured_at"`
}

// AppendD2 writes one pair as a JSONL line, creating .forge/datasets/ on first use.
// The gate that used to live here as Enabled() is now D2.Enabled (tier.go): it read as
// general but hardcoded D2Tag, which is exactly the trap a third tier would have fallen
// into.
func AppendD2(vaultRoot string, p D2Pair) error { return D2.Append(vaultRoot, p) }
