package dataset

import "time"

// D5 is the style dataset: (question, profile, sources) → the note that was accepted.
const (
	D5Kind = "d5-style"
	D5Path = ".forge/datasets/d5.jsonl"
	D5Tag  = "d5"
)

// D5Pair is one JSONL record.
type D5Pair struct {
	Kind       string            `json:"kind"`
	Topic      string            `json:"topic"`
	Rel        string            `json:"rel"`  // vault-relative path; hashed on export
	Note       string            `json:"note"` // the accepted body — the training target
	Profile    map[string]string `json:"profile,omitempty"`
	Stack      []string          `json:"stack,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	CapturedAt time.Time         `json:"captured_at"`
}

// AppendD5 writes one pair as a JSONL line.
func AppendD5(vaultRoot string, p D5Pair) error { return D5.Append(vaultRoot, p) }
