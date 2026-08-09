package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// D2 is ADDENDUM §D.1's advisor-distillation dataset: (draft, critique) pairs. Where D3
// dedupes on a git commit a hook can legitimately re-fire on, D2's trigger is one CLI call
// to `forge engine run` that made a real advisor call — there is no re-fire to guard
// against, so every capture is appended as its own line, no Key() or idempotency check.
const (
	D2Kind = "d2-advisor-critique"
	D2Path = ".forge/datasets/d2.jsonl"
	D2Tag  = "d2_advisor" // the cfg.Dataset.Capture entry that gates this
)

// D2Pair is one JSONL record: the draft sent to the advisor and its critique, verbatim.
// ADDENDUM §B.4 step 4 and §14 both say the same thing: "log the critique verbatim — it's
// dataset D2." Nothing here judges or reshapes what the advisor said.
type D2Pair struct {
	Kind       string    `json:"kind"`
	Stage      string    `json:"stage"`
	Draft      string    `json:"draft"`
	Critique   string    `json:"critique"`
	CapturedAt time.Time `json:"captured_at"`
}

// Enabled reports whether the config chain turned D2 capture on.
func Enabled(capture []string) bool {
	for _, c := range capture {
		if c == D2Tag {
			return true
		}
	}
	return false
}

// AppendD2 writes one pair as a JSONL line, creating .forge/datasets/ on first use.
func AppendD2(vaultRoot string, p D2Pair) error {
	path := filepath.Join(vaultRoot, D2Path)
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
